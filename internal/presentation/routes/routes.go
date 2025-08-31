package routes

import (
	"web-page-analyzer/internal/presentation/handlers"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine, analysisHandler *handlers.AnalysisHandler) {
	v1 := router.Group("/api/v1")
	{
		v1.POST("/analyze", analysisHandler.AnalyzeURL)
		v1.GET("/analyze/:id", analysisHandler.GetAnalysis)
	}

	router.GET("/health", analysisHandler.Health)

	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Web Page Analyzer API",
			"version": "1.0.0",
		})
	})
}
