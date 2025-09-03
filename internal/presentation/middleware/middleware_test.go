package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRateLimiter(t *testing.T) {
	rateLimiter := NewRateLimiter(2, time.Second)

	ip := "192.168.1.1"

	assert.True(t, rateLimiter.Allow(ip))
	assert.True(t, rateLimiter.Allow(ip))
	assert.False(t, rateLimiter.Allow(ip))
}

func TestRateLimiterDifferentIPs(t *testing.T) {
	rateLimiter := NewRateLimiter(1, time.Second)

	ip1 := "192.168.1.1"
	ip2 := "192.168.1.2"

	assert.True(t, rateLimiter.Allow(ip1))
	assert.True(t, rateLimiter.Allow(ip2))
	assert.False(t, rateLimiter.Allow(ip1))
}

func TestRateLimitMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	rateLimiter := NewRateLimiter(1, time.Second)

	router.Use(RateLimitMiddleware(rateLimiter))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	w1 := httptest.NewRecorder()
	req1 := httptest.NewRequest("GET", "/test", nil)
	req1.RemoteAddr = "192.168.1.1:12345"
	router.ServeHTTP(w1, req1)

	assert.Equal(t, http.StatusOK, w1.Code)

	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.RemoteAddr = "192.168.1.1:12345"
	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusTooManyRequests, w2.Code)
}

func TestCORSMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(CORSMiddleware())

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
}

func TestRequestSizeLimitMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestSizeLimitMiddleware(1024))

	router.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/test", nil)
	req.ContentLength = 512
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestNewRateLimiter(t *testing.T) {
	rateLimiter := NewRateLimiter(10, time.Minute)

	assert.NotNil(t, rateLimiter)
}

func TestCorrelationIDMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.Use(CorrelationIDMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"correlation_id": c.GetString("correlation_id")})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLoggingMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestErrorHandlingMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.Use(ErrorHandlingMiddleware(nil))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "test error"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.Use(AuthMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRateLimiterWithDifferentLimits(t *testing.T) {
	rateLimiter := NewRateLimiter(5, time.Second)
	assert.NotNil(t, rateLimiter)

	rateLimiter2 := NewRateLimiter(100, time.Minute)
	assert.NotNil(t, rateLimiter2)
}

func TestRateLimiterWithZeroLimit(t *testing.T) {
	rateLimiter := NewRateLimiter(0, time.Second)
	assert.NotNil(t, rateLimiter)
}

func TestMiddlewareIntegration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.Use(CorrelationIDMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
