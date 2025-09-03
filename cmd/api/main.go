package main

import (
	"context"
	"database/sql"
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
	"webpage-analyzer/pkg/migrate"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
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

	db, err := sql.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.Password, cfg.Database.Name, cfg.Database.SSLMode))
	if err != nil {
		appLogger.Fatal("Failed to connect to database", zap.Error(err))
	}

	migrator := migrate.NewMigrator(db)
	migrationsFS := os.DirFS("migrations")
	if err := migrator.Up(migrationsFS); err != nil {
		appLogger.Fatal("Failed to run migrations", zap.Error(err))
	}

	analysisRepo, err := postgres.NewAnalysisRepository(&cfg.Database)
	if err != nil {
		appLogger.Fatal("Failed to initialize database", zap.Error(err))
	}

	httpClient := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
			DisableKeepAlives:   false,
		},
	}
	wrappedClient := services.NewHTTPClient(httpClient)
	parser := services.NewHTMLParser(wrappedClient)

	analyzerConfig := &services.AnalyzerConfig{
		LinkCheckTimeout:        cfg.Analysis.LinkCheckTimeout,
		MaxLinksToCheck:         cfg.Analysis.MaxLinksToCheck,
		MaxConcurrentLinkChecks: cfg.Analysis.MaxConcurrentLinkChecks,
		MaxHTMLDepth:            cfg.Analysis.MaxHTMLDepth,
		MaxURLLength:            cfg.Analysis.MaxURLLength,
	}
	analyzer := services.NewAnalyzerService(wrappedClient, parser, analyzerConfig)

	analysisUC := usecases.NewAnalysisUseCase(
		analysisRepo,
		cacheRepo,
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
