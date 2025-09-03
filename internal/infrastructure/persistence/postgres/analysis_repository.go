package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"webpage-analyzer/internal/domain/entities"
	"webpage-analyzer/internal/domain/repositories"
	"webpage-analyzer/pkg/config"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type analysisRepository struct {
	db *sql.DB
}

func NewAnalysisRepository(cfg *config.DatabaseConfig) (repositories.AnalysisRepository, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxConnections)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &analysisRepository{db: db}, nil
}

func (r *analysisRepository) Create(ctx context.Context, analysis *entities.Analysis) error {
	query := `
		INSERT INTO analyses (id, url, status, result, error, created_at, updated_at, 
			completed_at, retry_count, priority, user_id, correlation_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	var resultJSON interface{}
	if analysis.Result != nil {
		resultBytes, err := json.Marshal(analysis.Result)
		if err != nil {
			return fmt.Errorf("failed to marshal result to JSON: %w", err)
		}
		resultJSON = resultBytes
	} else {
		resultJSON = nil
	}

	_, err := r.db.ExecContext(ctx, query,
		analysis.ID,
		analysis.URL,
		analysis.Status,
		resultJSON,
		analysis.Error,
		analysis.CreatedAt,
		analysis.UpdatedAt,
		analysis.CompletedAt,
		analysis.RetryCount,
		analysis.Priority,
		analysis.UserID,
		analysis.CorrelationID,
	)

	if err != nil {
		return fmt.Errorf("failed to create analysis: %w", err)
	}

	return nil
}

func (r *analysisRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Analysis, error) {
	query := `
		SELECT id, url, status, result, error, created_at, updated_at, 
			completed_at, retry_count, priority, user_id, correlation_id
		FROM analyses WHERE id = $1`

	row := r.db.QueryRowContext(ctx, query, id)
	return r.scanAnalysis(row)
}

func (r *analysisRepository) GetByURL(ctx context.Context, url string) (*entities.Analysis, error) {
	query := `
		SELECT id, url, status, result, error, created_at, updated_at, 
			completed_at, retry_count, priority, user_id, correlation_id
		FROM analyses WHERE url = $1 ORDER BY created_at DESC LIMIT 1`

	row := r.db.QueryRowContext(ctx, query, url)
	return r.scanAnalysis(row)
}

func (r *analysisRepository) Update(ctx context.Context, analysis *entities.Analysis) error {
	query := `
		UPDATE analyses SET 
			status = $2, result = $3, error = $4, updated_at = $5, 
			completed_at = $6, retry_count = $7
		WHERE id = $1`

	var resultJSON interface{}
	if analysis.Result != nil {
		resultBytes, err := json.Marshal(analysis.Result)
		if err != nil {
			return fmt.Errorf("failed to marshal result to JSON: %w", err)
		}
		resultJSON = resultBytes
	} else {
		resultJSON = nil
	}

	_, err := r.db.ExecContext(ctx, query,
		analysis.ID,
		analysis.Status,
		resultJSON,
		analysis.Error,
		analysis.UpdatedAt,
		analysis.CompletedAt,
		analysis.RetryCount,
	)

	if err != nil {
		return fmt.Errorf("failed to update analysis: %w", err)
	}

	return nil
}

func (r *analysisRepository) List(ctx context.Context, filters repositories.AnalysisFilters) ([]*entities.Analysis, error) {
	query := `
		SELECT id, url, status, result, error, created_at, updated_at, 
			completed_at, retry_count, priority, user_id, correlation_id
		FROM analyses WHERE 1=1`

	args := make([]interface{}, 0)
	argCount := 0

	if filters.Status != "" {
		argCount++
		query += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, filters.Status)
	}

	if filters.UserID != "" {
		argCount++
		query += fmt.Sprintf(" AND user_id = $%d", argCount)
		args = append(args, filters.UserID)
	}

	if filters.URL != "" {
		argCount++
		query += fmt.Sprintf(" AND url ILIKE $%d", argCount)
		args = append(args, "%"+filters.URL+"%")
	}

	validSortFields := map[string]bool{
		"created_at": true,
		"updated_at": true,
		"status":     true,
		"priority":   true,
	}

	if filters.SortBy != "" && validSortFields[filters.SortBy] {
		query += fmt.Sprintf(" ORDER BY %s", filters.SortBy)
		if filters.SortOrder == "desc" {
			query += " DESC"
		} else {
			query += " ASC"
		}
	} else {
		query += " ORDER BY created_at DESC"
	}

	if filters.Limit > 0 {
		argCount++
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, filters.Limit)
	}

	if filters.Offset > 0 {
		argCount++
		query += fmt.Sprintf(" OFFSET $%d", argCount)
		args = append(args, filters.Offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list analyses: %w", err)
	}
	defer rows.Close()

	analyses := make([]*entities.Analysis, 0)
	for rows.Next() {
		analysis, err := r.scanAnalysisFromRows(rows)
		if err != nil {
			return nil, err
		}
		analyses = append(analyses, analysis)
	}

	return analyses, nil
}

func (r *analysisRepository) scanAnalysis(row *sql.Row) (*entities.Analysis, error) {
	var analysis entities.Analysis
	var resultJSON []byte

	err := row.Scan(
		&analysis.ID,
		&analysis.URL,
		&analysis.Status,
		&resultJSON,
		&analysis.Error,
		&analysis.CreatedAt,
		&analysis.UpdatedAt,
		&analysis.CompletedAt,
		&analysis.RetryCount,
		&analysis.Priority,
		&analysis.UserID,
		&analysis.CorrelationID,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("analysis not found")
		}
		return nil, fmt.Errorf("failed to scan analysis: %w", err)
	}

	if resultJSON != nil {
		var result entities.AnalysisResult
		if err := json.Unmarshal(resultJSON, &result); err != nil {
			return nil, fmt.Errorf("failed to unmarshal result: %w", err)
		}
		analysis.Result = &result
	}

	return &analysis, nil
}

func (r *analysisRepository) scanAnalysisFromRows(rows *sql.Rows) (*entities.Analysis, error) {
	var analysis entities.Analysis
	var resultJSON []byte

	err := rows.Scan(
		&analysis.ID,
		&analysis.URL,
		&analysis.Status,
		&resultJSON,
		&analysis.Error,
		&analysis.CreatedAt,
		&analysis.UpdatedAt,
		&analysis.CompletedAt,
		&analysis.RetryCount,
		&analysis.Priority,
		&analysis.UserID,
		&analysis.CorrelationID,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to scan analysis: %w", err)
	}

	if resultJSON != nil {
		var result entities.AnalysisResult
		if err := json.Unmarshal(resultJSON, &result); err != nil {
			return nil, fmt.Errorf("failed to unmarshal result: %w", err)
		}
		analysis.Result = &result
	}

	return &analysis, nil
}
