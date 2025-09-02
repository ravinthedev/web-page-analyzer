package middleware

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"
	"webpage-analyzer/internal/infrastructure/monitoring"
	"webpage-analyzer/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type RateLimiter struct {
	visitors map[string]*Visitor
	mu       sync.RWMutex
	rate     int
	window   time.Duration
}

type Visitor struct {
	requests []time.Time
	mu       sync.Mutex
}

func NewRateLimiter(rate int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*Visitor),
		rate:     rate,
		window:   window,
	}

	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		for ip, visitor := range rl.visitors {
			visitor.mu.Lock()
			cutoff := time.Now().Add(-rl.window)
			validRequests := make([]time.Time, 0)

			for _, reqTime := range visitor.requests {
				if reqTime.After(cutoff) {
					validRequests = append(validRequests, reqTime)
				}
			}

			if len(validRequests) == 0 {
				delete(rl.visitors, ip)
			} else {
				visitor.requests = validRequests
			}
			visitor.mu.Unlock()
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	visitor, exists := rl.visitors[ip]
	if !exists {
		visitor = &Visitor{requests: make([]time.Time, 0)}
		rl.visitors[ip] = visitor
	}
	rl.mu.Unlock()

	visitor.mu.Lock()
	defer visitor.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	validRequests := make([]time.Time, 0)
	for _, reqTime := range visitor.requests {
		if reqTime.After(cutoff) {
			validRequests = append(validRequests, reqTime)
		}
	}

	if len(validRequests) >= rl.rate {
		return false
	}

	visitor.requests = append(validRequests, now)
	return true
}

func CorrelationIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		correlationID := c.GetHeader("X-Correlation-ID")
		if correlationID == "" {
			correlationID = uuid.New().String()
		}

		ctx := context.WithValue(c.Request.Context(), logger.CorrelationIDKey, correlationID)
		c.Request = c.Request.WithContext(ctx)
		c.Header("X-Correlation-ID", correlationID)

		c.Next()
	}
}

func LoggingMiddleware(log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		duration := time.Since(start)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method
		userAgent := c.Request.UserAgent()

		if raw != "" {
			path = path + "?" + raw
		}

		correlationID := c.Request.Context().Value(string(logger.CorrelationIDKey))

		fields := []zap.Field{
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status_code", statusCode),
			zap.String("client_ip", clientIP),
			zap.String("user_agent", userAgent),
			zap.Duration("duration", duration),
		}

		if correlationID != nil {
			fields = append(fields, zap.String("correlation_id", correlationID.(string)))
		}

		monitoring.RecordHTTPRequest(method, path, statusCode, duration)

		if statusCode >= 500 {
			log.Error("HTTP request completed with server error", fields...)
		} else if statusCode >= 400 {
			log.Warn("HTTP request completed with client error", fields...)
		} else {
			log.Info("HTTP request completed", fields...)
		}
	}
}

func RateLimitMiddleware(rateLimiter *RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		if !rateLimiter.Allow(clientIP) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
				"message": fmt.Sprintf("Maximum %d requests per %v allowed",
					rateLimiter.rate, rateLimiter.window),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func ErrorHandlingMiddleware(log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				correlationID := c.Request.Context().Value(string(logger.CorrelationIDKey))

				fields := []zap.Field{
					zap.Any("panic", err),
					zap.String("method", c.Request.Method),
					zap.String("path", c.Request.URL.Path),
					zap.String("client_ip", c.ClientIP()),
				}

				if correlationID != nil {
					fields = append(fields, zap.String("correlation_id", correlationID.(string)))
				}

				log.Error("Panic recovered", fields...)

				c.JSON(http.StatusInternalServerError, gin.H{
					"error":          "Internal server error",
					"correlation_id": correlationID,
				})
				c.Abort()
			}
		}()

		c.Next()
	}
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Correlation-ID")
		c.Header("Access-Control-Expose-Headers", "X-Correlation-ID")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func RequestSizeLimitMiddleware(maxSize int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > maxSize {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"error":    "Request entity too large",
				"max_size": maxSize,
			})
			c.Abort()
			return
		}

		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSize)
		c.Next()
	}
}

func TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)

		done := make(chan struct{})
		go func() {
			c.Next()
			close(done)
		}()

		select {
		case <-done:
			return
		case <-ctx.Done():
			c.JSON(http.StatusRequestTimeout, gin.H{
				"error":   "Request timeout",
				"timeout": timeout.String(),
			})
			c.Abort()
		}
	}
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetHeader("X-User-ID")
		if userID == "" {
			userID = "anonymous"
		}

		ctx := context.WithValue(c.Request.Context(), logger.UserIDKey, userID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
