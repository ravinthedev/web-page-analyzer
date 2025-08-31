package handlers

import (
	"net/http"
	"strconv"
	"webpage-analyzer/internal/application/usecases"
	"webpage-analyzer/internal/domain/entities"
	"webpage-analyzer/internal/domain/repositories"
	"webpage-analyzer/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type AnalysisHandler struct {
	analysisUC usecases.AnalysisUseCase
	logger     logger.Logger
}

type AnalyzeRequest struct {
	URL      string `json:"url" binding:"required"`
	Priority int    `json:"priority,omitempty"`
	Async    bool   `json:"async,omitempty"`
}

type AnalyzeResponse struct {
	ID            string      `json:"id"`
	AnalysisID    string      `json:"analysis_id,omitempty"`
	URL           string      `json:"url"`
	Status        string      `json:"status"`
	Result        interface{} `json:"result,omitempty"`
	Error         string      `json:"error,omitempty"`
	CorrelationID string      `json:"correlation_id"`
}

func NewAnalysisHandler(analysisUC usecases.AnalysisUseCase, logger logger.Logger) *AnalysisHandler {
	return &AnalysisHandler{
		analysisUC: analysisUC,
		logger:     logger,
	}
}

func (h *AnalysisHandler) AnalyzeURL(c *gin.Context) {
	var req AnalyzeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	userID, ok := c.Request.Context().Value(logger.UserIDKey).(string)
	if !ok {
		userID = "anonymous"
	}
	correlationID, ok := c.Request.Context().Value(logger.CorrelationIDKey).(string)
	if !ok {
		correlationID = "unknown"
	}

	log := h.logger.WithContext(c.Request.Context()).With(
		zap.String(logger.URLKey, req.URL),
		zap.Bool("async", req.Async),
		zap.Int("priority", req.Priority),
	)

	if req.Async {
		job, analysis, err := h.analysisUC.SubmitAnalysisJob(c.Request.Context(), req.URL, userID, req.Priority)
		if err != nil {
			log.Error("Failed to submit analysis job", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":          "Failed to submit analysis job",
				"details":        err.Error(),
				"correlation_id": correlationID,
			})
			return
		}

		c.JSON(http.StatusAccepted, AnalyzeResponse{
			ID:            job.ID.String(),
			AnalysisID:    analysis.ID.String(),
			URL:           req.URL,
			Status:        "pending",
			CorrelationID: correlationID,
		})
	} else {
		analysis, err := h.analysisUC.AnalyzeURL(c.Request.Context(), req.URL, userID)
		if err != nil {
			log.Error("Analysis failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":          "Analysis failed",
				"details":        err.Error(),
				"correlation_id": correlationID,
			})
			return
		}

		response := AnalyzeResponse{
			ID:            analysis.ID.String(),
			URL:           analysis.URL,
			Status:        string(analysis.Status),
			CorrelationID: correlationID,
		}

		if analysis.Result != nil {
			response.Result = analysis.Result
		}

		if analysis.Error != "" {
			response.Error = analysis.Error
		}

		statusCode := http.StatusOK
		if analysis.Status == "failed" {
			statusCode = http.StatusUnprocessableEntity
		}

		c.JSON(statusCode, response)
	}
}

func (h *AnalysisHandler) GetAnalysis(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid analysis ID format",
		})
		return
	}

	log := h.logger.WithContext(c.Request.Context()).With(
		zap.String("analysis_id", id.String()),
	)

	analysis, err := h.analysisUC.GetAnalysis(c.Request.Context(), id)
	if err != nil {
		log.Error("Failed to get analysis", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Analysis not found",
		})
		return
	}

	response := AnalyzeResponse{
		ID:            analysis.ID.String(),
		URL:           analysis.URL,
		Status:        string(analysis.Status),
		CorrelationID: analysis.CorrelationID,
	}

	if analysis.Result != nil {
		response.Result = analysis.Result
	}

	if analysis.Error != "" {
		response.Error = analysis.Error
	}

	c.JSON(http.StatusOK, response)
}

func (h *AnalysisHandler) ListAnalyses(c *gin.Context) {
	filters := repositories.AnalysisFilters{
		Status: entities.AnalysisStatus(c.Query("status")),
		UserID: c.Query("user_id"),
		URL:    c.Query("url"),
		Limit:  10,
		Offset: 0,
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit <= 100 {
			filters.Limit = limit
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			filters.Offset = offset
		}
	}

	log := h.logger.WithContext(c.Request.Context()).With(
		zap.String("status", string(filters.Status)),
		zap.String("user_id", filters.UserID),
		zap.Int("limit", filters.Limit),
		zap.Int("offset", filters.Offset),
	)

	analyses, err := h.analysisUC.ListAnalyses(c.Request.Context(), filters)
	if err != nil {
		log.Error("Failed to list analyses", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve analyses",
		})
		return
	}

	responses := make([]AnalyzeResponse, len(analyses))
	for i, analysis := range analyses {
		responses[i] = AnalyzeResponse{
			ID:            analysis.ID.String(),
			URL:           analysis.URL,
			Status:        string(analysis.Status),
			CorrelationID: analysis.CorrelationID,
		}

		if analysis.Result != nil {
			responses[i].Result = analysis.Result
		}

		if analysis.Error != "" {
			responses[i].Error = analysis.Error
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"analyses": responses,
		"total":    len(responses),
		"limit":    filters.Limit,
		"offset":   filters.Offset,
	})
}

func (h *AnalysisHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"service":   "webpage-analyzer",
		"timestamp": c.Request.Context().Value("timestamp"),
	})
}
