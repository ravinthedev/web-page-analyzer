package services

import (
	"context"
	"time"

	"web-page-analyzer/internal/domain/entities"
)

type AnalyzerService interface {
	AnalyzeURL(ctx context.Context, url string) (*entities.Analysis, error)
	
	GetAnalysis(ctx context.Context, id string) (*entities.Analysis, error)
}

type HTMLParser interface {
	ParseHTML(html string, baseURL string) (*entities.AnalysisResult, error)
}

type HTTPClient interface {
	FetchURL(ctx context.Context, url string, timeout time.Duration) (string, int, error)
}
