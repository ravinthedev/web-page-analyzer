package dto

import (
	"time"

	"web-page-analyzer/internal/domain/entities"
)

type AnalysisRequest struct {
	URL string `json:"url" binding:"required,url"`
}

type AnalysisResponse struct {
	ID            string                  `json:"id"`
	URL           string                  `json:"url"`
	Status        string                  `json:"status"`
	Result        *AnalysisResultResponse `json:"result,omitempty"`
	Error         string                  `json:"error,omitempty"`
	CreatedAt     time.Time               `json:"created_at"`
	UpdatedAt     time.Time               `json:"updated_at"`
	CompletedAt   *time.Time              `json:"completed_at,omitempty"`
	RetryCount    int                     `json:"retry_count"`
	Priority      int                     `json:"priority"`
	UserID        string                  `json:"user_id,omitempty"`
	CorrelationID string                  `json:"correlation_id"`
}

type AnalysisResultResponse struct {
	HTMLVersion   string            `json:"html_version"`
	Title         string            `json:"title"`
	Headings      map[string]int    `json:"headings"`
	Links         LinkResponse      `json:"links"`
	HasLoginForm  bool              `json:"has_login_form"`
	LoadTime      string            `json:"load_time"`
	ContentLength int64             `json:"content_length"`
	StatusCode    int               `json:"status_code"`
	Metadata      map[string]string `json:"metadata,omitempty"`
}

type LinkResponse struct {
	Internal      int      `json:"internal"`
	External      int      `json:"external"`
	Inaccessible  int      `json:"inaccessible"`
	BrokenLinks   []string `json:"broken_links,omitempty"`
	ExternalHosts []string `json:"external_hosts,omitempty"`
}

func ToResponse(analysis *entities.Analysis) *AnalysisResponse {
	response := &AnalysisResponse{
		ID:            analysis.ID.String(),
		URL:           analysis.URL,
		Status:        string(analysis.Status),
		Error:         analysis.Error,
		CreatedAt:     analysis.CreatedAt,
		UpdatedAt:     analysis.UpdatedAt,
		CompletedAt:   analysis.CompletedAt,
		RetryCount:    analysis.RetryCount,
		Priority:      analysis.Priority,
		UserID:        analysis.UserID,
		CorrelationID: analysis.CorrelationID,
	}

	if analysis.Result != nil {
		response.Result = &AnalysisResultResponse{
			HTMLVersion: analysis.Result.HTMLVersion,
			Title:       analysis.Result.Title,
			Headings:    analysis.Result.Headings,
			Links: LinkResponse{
				Internal:      analysis.Result.Links.Internal,
				External:      analysis.Result.Links.External,
				Inaccessible:  analysis.Result.Links.Inaccessible,
				BrokenLinks:   analysis.Result.Links.BrokenLinks,
				ExternalHosts: analysis.Result.Links.ExternalHosts,
			},
			HasLoginForm:  analysis.Result.HasLoginForm,
			LoadTime:      analysis.Result.LoadTime.String(),
			ContentLength: analysis.Result.ContentLength,
			StatusCode:    analysis.Result.StatusCode,
			Metadata:      analysis.Result.Metadata,
		}
	}

	return response
}
