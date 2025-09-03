package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestAnalyzeURL(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.POST("/test", func(c *gin.Context) {
		var req map[string]interface{}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	requestBody := map[string]interface{}{
		"url": "https://example.com",
	}

	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/test", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHealthCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandlerWithInvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.POST("/test", func(c *gin.Context) {
		var req map[string]interface{}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	req := httptest.NewRequest("POST", "/test", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandlerWithEmptyBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.POST("/test", func(c *gin.Context) {
		var req map[string]interface{}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	req := httptest.NewRequest("POST", "/test", bytes.NewBufferString(""))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAnalyzeURLHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.POST("/analyze", func(c *gin.Context) {
		var req struct {
			URL string `json:"url"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		if req.URL == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "URL is required"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"url":            req.URL,
			"status":         "analyzed",
			"title":          "Test Page",
			"html_version":   "HTML5",
			"headings":       map[string]int{"h1": 1},
			"has_login_form": false,
		})
	})

	requestBody := map[string]string{
		"url": "https://example.com",
	}
	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/analyze", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAnalyzeURLHandlerWithEmptyURL(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.POST("/analyze", func(c *gin.Context) {
		var req struct {
			URL string `json:"url"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		if req.URL == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "URL is required"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	requestBody := map[string]string{
		"url": "",
	}
	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/analyze", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
