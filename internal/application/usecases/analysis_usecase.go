package usecases

import (
	"context"
	"time"

	"web-page-analyzer/internal/domain/entities"
	"web-page-analyzer/internal/domain/repositories"
	"web-page-analyzer/internal/domain/services"

	"github.com/google/uuid"
)

type AnalysisUseCase struct {
	repo       repositories.AnalysisRepository
	analyzer   services.AnalyzerService
	htmlParser services.HTMLParser
	httpClient services.HTTPClient
	timeout    time.Duration
}

func NewAnalysisUseCase(
	repo repositories.AnalysisRepository,
	analyzer services.AnalyzerService,
	htmlParser services.HTMLParser,
	httpClient services.HTTPClient,
	timeout time.Duration,
) *AnalysisUseCase {
	return &AnalysisUseCase{
		repo:       repo,
		analyzer:   analyzer,
		htmlParser: htmlParser,
		httpClient: httpClient,
		timeout:    timeout,
	}
}

func (uc *AnalysisUseCase) AnalyzeURL(ctx context.Context, url string, userID, correlationID string) (*entities.Analysis, error) {
	existing, err := uc.repo.GetByURL(ctx, url)
	if err == nil && existing != nil && existing.Status == entities.StatusCompleted {
		return existing, nil
	}

	analysis := entities.NewAnalysis(url, userID, correlationID)

	if err := uc.repo.Create(ctx, analysis); err != nil {
		return nil, err
	}

	if err := uc.performAnalysis(ctx, analysis); err != nil {
		analysis.MarkAsFailed(err.Error())
		uc.repo.Update(ctx, analysis)
		return nil, err
	}

	if err := uc.repo.Update(ctx, analysis); err != nil {
		return nil, err
	}

	return analysis, nil
}

func (uc *AnalysisUseCase) GetAnalysis(ctx context.Context, id string) (*entities.Analysis, error) {
	uuidID, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}
	return uc.repo.GetByID(ctx, uuidID)
}

func (uc *AnalysisUseCase) performAnalysis(ctx context.Context, analysis *entities.Analysis) error {
	startTime := time.Now()

	html, statusCode, err := uc.httpClient.FetchURL(ctx, analysis.URL, uc.timeout)
	if err != nil {
		return err
	}

	result, err := uc.htmlParser.ParseHTML(html, analysis.URL)
	if err != nil {
		return err
	}

	result.LoadTime = time.Since(startTime)
	result.StatusCode = statusCode

	analysis.MarkAsCompleted(result)

	return nil
}
