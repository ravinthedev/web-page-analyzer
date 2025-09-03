package routes

import (
	"testing"
	"time"

	"webpage-analyzer/internal/application/usecases"
	"webpage-analyzer/internal/presentation/middleware"
	"webpage-analyzer/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSetupRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	var uc usecases.AnalysisUseCase
	var log logger.Logger
	rateLimiter := middleware.NewRateLimiter(100, time.Minute)

	SetupRoutes(router, uc, log, rateLimiter, 1024*1024, 30)

	assert.NotNil(t, router)
}

func TestSetupRoutesWithDifferentConfigs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	var uc usecases.AnalysisUseCase
	var log logger.Logger
	rateLimiter := middleware.NewRateLimiter(50, time.Second)

	SetupRoutes(router, uc, log, rateLimiter, 512*1024, 60)

	assert.NotNil(t, router)
}

func TestSetupRoutesWithZeroLimits(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	var uc usecases.AnalysisUseCase
	var log logger.Logger
	rateLimiter := middleware.NewRateLimiter(0, time.Second)

	SetupRoutes(router, uc, log, rateLimiter, 0, 0)

	assert.NotNil(t, router)
}
