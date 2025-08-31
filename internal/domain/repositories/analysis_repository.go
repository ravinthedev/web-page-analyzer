package repositories

import (
	"context"

	"web-page-analyzer/internal/domain/entities"
)

type AnalysisRepository interface {
	Create(ctx context.Context, analysis *entities.Analysis) error
	
	GetByID(ctx context.Context, id string) (*entities.Analysis, error)
	
	GetByURL(ctx context.Context, url string) (*entities.Analysis, error)
	
	Update(ctx context.Context, analysis *entities.Analysis) error
	
	Delete(ctx context.Context, id string) error
}
