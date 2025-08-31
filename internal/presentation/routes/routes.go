package routes

import (
	"webpage-analyzer/internal/application/usecases"
	"webpage-analyzer/internal/presentation/handlers"
	"webpage-analyzer/internal/presentation/middleware"
	"webpage-analyzer/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func SetupRoutes(
	router *gin.Engine,
	analysisUC usecases.AnalysisUseCase,
	logger logger.Logger,
	rateLimiter *middleware.RateLimiter,
	maxContentLength int64,
	requestTimeout int,
) {
	analysisHandler := handlers.NewAnalysisHandler(analysisUC, logger)

	router.Use(middleware.ErrorHandlingMiddleware(logger))
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.CorrelationIDMiddleware())
	router.Use(middleware.AuthMiddleware())
	router.Use(middleware.LoggingMiddleware(logger))
	router.Use(middleware.RateLimitMiddleware(rateLimiter))
	router.Use(middleware.RequestSizeLimitMiddleware(maxContentLength))

	router.GET("/health", analysisHandler.HealthCheck)
	router.GET("/health/live", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "alive"})
	})
	router.GET("/health/ready", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ready"})
	})

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	v1 := router.Group("/api/v1")
	{
		v1.POST("/analyze", analysisHandler.AnalyzeURL)
		v1.GET("/analysis/:id", analysisHandler.GetAnalysis)
		v1.GET("/analyses", analysisHandler.ListAnalyses)
	}

	router.POST("/api/analyze", analysisHandler.AnalyzeURL)
}
