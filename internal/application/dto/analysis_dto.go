package dto

import (
	"time"

	"web-page-analyzer/internal/domain/entities"
)

type AnalysisRequest struct {
	URL string `json:"url" binding:"required,url"`
}

type AnalysisResponse struct {
	ID        string                  `json:"id"`
	URL       string                  `json:"url"`
	Status    string                  `json:"status"`
	Result    *AnalysisResultResponse `json:"result,omitempty"`
	CreatedAt time.Time               `json:"created_at"`
	UpdatedAt time.Time               `json:"updated_at"`
}

type AnalysisResultResponse struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Keywords    []string          `json:"keywords"`
	Links       []LinkResponse    `json:"links"`
	Images      []ImageResponse   `json:"images"`
	MetaTags    map[string]string `json:"meta_tags"`
	WordCount   int               `json:"word_count"`
	LoadTime    string            `json:"load_time"`
	StatusCode  int               `json:"status_code"`
}

type LinkResponse struct {
	URL        string `json:"url"`
	Text       string `json:"text"`
	IsExternal bool   `json:"is_external"`
	IsValid    bool   `json:"is_valid"`
}

type ImageResponse struct {
	URL     string `json:"url"`
	Alt     string `json:"alt"`
	Width   int    `json:"width"`
	Height  int    `json:"height"`
	IsValid bool   `json:"is_valid"`
}

func ToResponse(analysis *entities.Analysis) *AnalysisResponse {
	response := &AnalysisResponse{
		ID:        analysis.ID.String(),
		URL:       analysis.URL,
		Status:    string(analysis.Status),
		CreatedAt: analysis.CreatedAt,
		UpdatedAt: analysis.UpdatedAt,
	}

	if analysis.Result != nil {
		response.Result = &AnalysisResultResponse{
			Title:       analysis.Result.Title,
			Description: analysis.Result.Description,
			Keywords:    analysis.Result.Keywords,
			MetaTags:    analysis.Result.MetaTags,
			WordCount:   analysis.Result.WordCount,
			LoadTime:    analysis.Result.LoadTime.String(),
			StatusCode:  analysis.Result.StatusCode,
		}

		response.Result.Links = make([]LinkResponse, len(analysis.Result.Links))
		for i, link := range analysis.Result.Links {
			response.Result.Links[i] = LinkResponse{
				URL:        link.URL,
				Text:       link.Text,
				IsExternal: link.IsExternal,
				IsValid:    link.IsValid,
			}
		}

		response.Result.Images = make([]ImageResponse, len(analysis.Result.Images))
		for i, img := range analysis.Result.Images {
			response.Result.Images[i] = ImageResponse{
				URL:     img.URL,
				Alt:     img.Alt,
				Width:   img.Width,
				Height:  img.Height,
				IsValid: img.IsValid,
			}
		}
	}

	return response
}
