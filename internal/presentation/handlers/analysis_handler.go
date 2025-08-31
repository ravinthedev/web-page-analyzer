package handlers

import (
	"net/http"

	"web-page-analyzer/internal/application/dto"
	"web-page-analyzer/internal/application/usecases"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AnalysisHandler struct {
	analysisUseCase *usecases.AnalysisUseCase
}

func NewAnalysisHandler(analysisUseCase *usecases.AnalysisUseCase) *AnalysisHandler {
	return &AnalysisHandler{
		analysisUseCase: analysisUseCase,
	}
}

func (h *AnalysisHandler) AnalyzeURL(c *gin.Context) {
	var request dto.AnalysisRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"message": err.Error(),
		})
		return
	}

	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		userID = "anonymous"
	}

	correlationID := c.GetHeader("X-Correlation-ID")
	if correlationID == "" {
		correlationID = uuid.New().String()
	}

	analysis, err := h.analysisUseCase.AnalyzeURL(c.Request.Context(), request.URL, userID, correlationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Analysis failed",
			"message": err.Error(),
		})
		return
	}

	response := dto.ToResponse(analysis)

	c.JSON(http.StatusOK, response)
}

func (h *AnalysisHandler) GetAnalysis(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"message": "Analysis ID is required",
		})
		return
	}

	analysis, err := h.analysisUseCase.GetAnalysis(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Analysis not found",
			"message": err.Error(),
		})
		return
	}

	response := dto.ToResponse(analysis)

	c.JSON(http.StatusOK, response)
}

func (h *AnalysisHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"service": "web-page-analyzer",
	})
}
