package usecases

import (
	"context"
	"fmt"
	"time"
	"webpage-analyzer/internal/domain/entities"
	"webpage-analyzer/internal/domain/repositories"
	"webpage-analyzer/internal/domain/services"
	"webpage-analyzer/pkg/logger"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	DefaultCorrelationID = "unknown"
)

type AnalysisUseCase interface {
	AnalyzeURL(ctx context.Context, url, userID string) (*entities.Analysis, error)
	GetAnalysis(ctx context.Context, id uuid.UUID) (*entities.Analysis, error)
	GetAnalysisByURL(ctx context.Context, url string) (*entities.Analysis, error)
	SubmitAnalysisJob(ctx context.Context, url, userID string, priority int) (*entities.AnalysisJob, *entities.Analysis, error)
	ProcessAnalysisAsync(ctx context.Context, analysis *entities.Analysis)
	ListAnalyses(ctx context.Context, filters repositories.AnalysisFilters) ([]*entities.Analysis, error)
}

type analysisUseCase struct {
	analysisRepo repositories.AnalysisRepository
	cacheRepo    repositories.CacheRepository
	analyzer     services.AnalyzerService
	logger       logger.Logger
	cacheTTL     int
}

func NewAnalysisUseCase(
	analysisRepo repositories.AnalysisRepository,
	cacheRepo repositories.CacheRepository,
	analyzer services.AnalyzerService,
	logger logger.Logger,
	cacheTTL int,
) AnalysisUseCase {
	return &analysisUseCase{
		analysisRepo: analysisRepo,
		cacheRepo:    cacheRepo,
		analyzer:     analyzer,
		logger:       logger,
		cacheTTL:     cacheTTL,
	}
}

func (uc *analysisUseCase) AnalyzeURL(ctx context.Context, url, userID string) (*entities.Analysis, error) {
	correlationID, ok := ctx.Value(logger.CorrelationIDKey).(string)
	if !ok {
		correlationID = DefaultCorrelationID
	}
	log := uc.logger.WithContext(ctx).With(
		zap.String(string(logger.URLKey), url),
		zap.String(string(logger.UserIDKey), userID),
	)

	log.Info("Starting URL analysis")

	cacheKey := fmt.Sprintf("analysis:%s", url)
	var cachedResult entities.AnalysisResult
	if err := uc.cacheRepo.Get(ctx, cacheKey, &cachedResult); err == nil {
		log.Info("Analysis result found in cache")
		analysis := entities.NewAnalysis(url, userID, correlationID)
		analysis.MarkAsCompleted(&cachedResult)
		return analysis, nil
	}

	if existing, err := uc.analysisRepo.GetByURL(ctx, url); err == nil {
		if existing.Status == entities.StatusCompleted && existing.Result != nil {
			// check if the analysis is still fresh (within cache TTL)
			if time.Since(existing.CreatedAt) < time.Duration(uc.cacheTTL)*time.Second {
				log.Info("Analysis already completed and still fresh",
					zap.String("analysis_id", existing.ID.String()),
					zap.Duration("age", time.Since(existing.CreatedAt)))
				return existing, nil
			} else {
				log.Info("Analysis exists but expired, will re-analyze",
					zap.String("analysis_id", existing.ID.String()),
					zap.Duration("age", time.Since(existing.CreatedAt)))
			}
		}
	}

	analysis := entities.NewAnalysis(url, userID, correlationID)
	if err := uc.analysisRepo.Create(ctx, analysis); err != nil {
		log.Error("Failed to create analysis record", zap.Error(err))
		return nil, fmt.Errorf("failed to create analysis: %w", err)
	}

	analysis.MarkAsProcessing()
	if err := uc.analysisRepo.Update(ctx, analysis); err != nil {
		log.Error("Failed to update analysis status", zap.Error(err))
	}

	result, err := uc.analyzer.AnalyzeURL(ctx, url)
	if err != nil {
		log.Error("Analysis failed", zap.Error(err))
		analysis.MarkAsFailed(err.Error())
		_ = uc.analysisRepo.Update(ctx, analysis)
		return analysis, fmt.Errorf("analysis failed: %w", err)
	}

	analysis.MarkAsCompleted(result)
	if err := uc.analysisRepo.Update(ctx, analysis); err != nil {
		log.Error("Failed to update analysis result", zap.Error(err))
	}

	if err := uc.cacheRepo.Set(ctx, cacheKey, result, uc.cacheTTL); err != nil {
		log.Warn("Failed to cache analysis result", zap.Error(err))
	}

	log.Info("Analysis completed successfully",
		zap.Duration(string(logger.DurationKey), result.LoadTime),
		zap.Int(string(logger.StatusCodeKey), result.StatusCode),
	)

	return analysis, nil
}

func (uc *analysisUseCase) SubmitAnalysisJob(ctx context.Context, url, userID string, priority int) (*entities.AnalysisJob, *entities.Analysis, error) {
	correlationID, ok := ctx.Value(logger.CorrelationIDKey).(string)
	if !ok {
		correlationID = DefaultCorrelationID
	}
	log := uc.logger.WithContext(ctx).With(
		zap.String(string(logger.URLKey), url),
		zap.String(string(logger.UserIDKey), userID),
		zap.Int("priority", priority),
	)

	log.Info("Submitting analysis job")

	if err := uc.analyzer.ValidateURL(url); err != nil {
		log.Error("Invalid URL", zap.Error(err))
		return nil, nil, fmt.Errorf("invalid URL: %w", err)
	}

	analysis := entities.NewAnalysis(url, userID, correlationID)
	if err := uc.analysisRepo.Create(ctx, analysis); err != nil {
		log.Error("Failed to create analysis record", zap.Error(err))
		return nil, nil, fmt.Errorf("failed to create analysis: %w", err)
	}

	uc.ProcessAnalysisAsync(ctx, analysis)

	job := entities.NewAnalysisJob(url, userID, correlationID, priority)

	log.Info("Analysis job submitted successfully",
		zap.String("job_id", job.ID.String()),
		zap.String("analysis_id", analysis.ID.String()),
	)
	return job, analysis, nil
}

func (uc *analysisUseCase) ProcessAnalysisAsync(ctx context.Context, analysis *entities.Analysis) {
	go func() {
		asyncCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		asyncCtx = context.WithValue(asyncCtx, logger.CorrelationIDKey, analysis.CorrelationID)
		asyncCtx = context.WithValue(asyncCtx, logger.UserIDKey, analysis.UserID)

		log := uc.logger.WithContext(asyncCtx).With(
			zap.String(string(logger.URLKey), analysis.URL),
			zap.String(string(logger.UserIDKey), analysis.UserID),
			zap.String("analysis_id", analysis.ID.String()),
		)

		log.Info("Starting async analysis processing")

		cacheKey := fmt.Sprintf("analysis:%s", analysis.URL)
		var cachedResult entities.AnalysisResult
		if err := uc.cacheRepo.Get(asyncCtx, cacheKey, &cachedResult); err == nil {
			log.Info("Analysis result found in cache")
			analysis.MarkAsCompleted(&cachedResult)
		} else {
			result, err := uc.analyzer.AnalyzeURL(asyncCtx, analysis.URL)
			if err != nil {
				log.Error("Analysis failed", zap.Error(err))
				analysis.MarkAsFailed(err.Error())
			} else {
				log.Info("Analysis completed successfully")
				analysis.MarkAsCompleted(result)

				if err := uc.cacheRepo.Set(asyncCtx, cacheKey, result, uc.cacheTTL); err != nil {
					log.Warn("Failed to cache analysis result", zap.Error(err))
				}
			}
		}

		if err := uc.analysisRepo.Update(asyncCtx, analysis); err != nil {
			log.Error("Failed to update analysis in database", zap.Error(err))
		}

		log.Info("Async analysis processing completed",
			zap.String("status", string(analysis.Status)),
		)
	}()
}

func (uc *analysisUseCase) GetAnalysis(ctx context.Context, id uuid.UUID) (*entities.Analysis, error) {
	log := uc.logger.WithContext(ctx).With(zap.String("analysis_id", id.String()))
	log.Debug("Retrieving analysis")

	analysis, err := uc.analysisRepo.GetByID(ctx, id)
	if err != nil {
		log.Error("Failed to retrieve analysis", zap.Error(err))
		return nil, fmt.Errorf("failed to get analysis: %w", err)
	}

	return analysis, nil
}

func (uc *analysisUseCase) GetAnalysisByURL(ctx context.Context, url string) (*entities.Analysis, error) {
	log := uc.logger.WithContext(ctx).With(zap.String(string(logger.URLKey), url))
	log.Debug("Retrieving analysis by URL")

	analysis, err := uc.analysisRepo.GetByURL(ctx, url)
	if err != nil {
		log.Error("Failed to retrieve analysis by URL", zap.Error(err))
		return nil, fmt.Errorf("failed to get analysis: %w", err)
	}

	return analysis, nil
}

func (uc *analysisUseCase) ListAnalyses(ctx context.Context, filters repositories.AnalysisFilters) ([]*entities.Analysis, error) {
	log := uc.logger.WithContext(ctx).With(
		zap.String("status", string(filters.Status)),
		zap.String("user_id", filters.UserID),
		zap.Int("limit", filters.Limit),
		zap.Int("offset", filters.Offset),
	)
	log.Debug("Listing analyses")

	analyses, err := uc.analysisRepo.List(ctx, filters)
	if err != nil {
		log.Error("Failed to list analyses", zap.Error(err))
		return nil, fmt.Errorf("failed to list analyses: %w", err)
	}

	log.Debug("Retrieved analyses", zap.Int("count", len(analyses)))
	return analyses, nil
}
