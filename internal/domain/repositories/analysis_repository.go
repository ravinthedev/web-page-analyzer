package repositories

import (
	"context"
	"webpage-analyzer/internal/domain/entities"

	"github.com/google/uuid"
)

type AnalysisRepository interface {
	Create(ctx context.Context, analysis *entities.Analysis) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Analysis, error)
	GetByURL(ctx context.Context, url string) (*entities.Analysis, error)
	Update(ctx context.Context, analysis *entities.Analysis) error
	List(ctx context.Context, filters AnalysisFilters) ([]*entities.Analysis, error)
}

type CacheRepository interface {
	Set(ctx context.Context, key string, value interface{}, ttl int) error
	Get(ctx context.Context, key string, dest interface{}) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
}

type AnalysisFilters struct {
	Status    entities.AnalysisStatus
	UserID    string
	URL       string
	Limit     int
	Offset    int
	SortBy    string
	SortOrder string
}
