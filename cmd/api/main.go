package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"webpage-analyzer/internal/application/usecases"
	"webpage-analyzer/internal/domain/services"
	"webpage-analyzer/internal/infrastructure/persistence/postgres"
	"webpage-analyzer/internal/infrastructure/persistence/redis"
	"webpage-analyzer/internal/presentation/middleware"
	"webpage-analyzer/internal/presentation/routes"
	"webpage-analyzer/pkg/config"
	"webpage-analyzer/pkg/logger"

	"github.com/gin-gonic/gin"
	redisclient "github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	appLogger, err := logger.New(cfg.Logger.Level, cfg.Logger.Development)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	appLogger.Info("Starting Web Page Analyzer API",
		zap.String("version", "1.0.0"),
		zap.String("port", cfg.Server.Port),
	)

	cacheRepo := redis.NewCacheRepository(&cfg.Redis)

	redisClient := redisclient.NewClient(&redisclient.Options{
		Addr:         fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		PoolSize:     cfg.Redis.PoolSize,
		MinIdleConns: cfg.Redis.MinIdleConns,
		DialTimeout:  cfg.Redis.DialTimeout,
		ReadTimeout:  cfg.Redis.ReadTimeout,
		WriteTimeout: cfg.Redis.WriteTimeout,
	})

	queueRepo := redis.NewJobQueueRepository(redisClient)

	analysisRepo, err := postgres.NewAnalysisRepository(&cfg.Database)
	if err != nil {
		appLogger.Fatal("Failed to initialize database", zap.Error(err))
	}

	httpClient := &http.Client{
		Timeout: cfg.Analysis.RequestTimeout,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		},
	}
	parser := services.NewHTMLParser(httpClient)
	analyzer := services.NewAnalyzerService(httpClient, parser)

	analysisUC := usecases.NewAnalysisUseCase(
		analysisRepo,
		cacheRepo,
		queueRepo,
		analyzer,
		appLogger,
		int(cfg.Analysis.CacheTTL.Seconds()),
	)

	if !cfg.Logger.Development {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	rateLimiter := middleware.NewRateLimiter(cfg.Analysis.RateLimitPerIP, cfg.Analysis.RateLimitWindow)

	routes.SetupRoutes(router, analysisUC, appLogger, rateLimiter, cfg.Analysis.MaxContentLength, int(cfg.Analysis.RequestTimeout.Seconds()))

	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	go func() {
		appLogger.Info("HTTP server starting", zap.String("address", server.Addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	appLogger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		appLogger.Error("Server forced to shutdown", zap.Error(err))
	}

	appLogger.Info("Server shutdown complete")
}
