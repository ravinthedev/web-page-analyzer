package entities

import (
	"time"

	"github.com/google/uuid"
)

type AnalysisStatus string

const (
	StatusPending    AnalysisStatus = "pending"
	StatusProcessing AnalysisStatus = "processing"
	StatusCompleted  AnalysisStatus = "completed"
	StatusFailed     AnalysisStatus = "failed"
	StatusRetrying   AnalysisStatus = "retrying"
)

type Analysis struct {
	ID            uuid.UUID       `json:"id" db:"id"`
	URL           string          `json:"url" db:"url"`
	Status        AnalysisStatus  `json:"status" db:"status"`
	Result        *AnalysisResult `json:"result,omitempty" db:"result"`
	Error         string          `json:"error,omitempty" db:"error"`
	CreatedAt     time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at" db:"updated_at"`
	CompletedAt   *time.Time      `json:"completed_at,omitempty" db:"completed_at"`
	RetryCount    int             `json:"retry_count" db:"retry_count"`
	Priority      int             `json:"priority" db:"priority"`
	UserID        string          `json:"user_id,omitempty" db:"user_id"`
	CorrelationID string          `json:"correlation_id" db:"correlation_id"`
}

type AnalysisResult struct {
	HTMLVersion   string            `json:"html_version"`
	Title         string            `json:"title"`
	Headings      map[string]int    `json:"headings"`
	Links         LinkAnalysis      `json:"links"`
	HasLoginForm  bool              `json:"has_login_form"`
	LoadTime      time.Duration     `json:"load_time"`
	ContentLength int64             `json:"content_length"`
	StatusCode    int               `json:"status_code"`
	Metadata      map[string]string `json:"metadata,omitempty"`
}

type LinkAnalysis struct {
	Internal      int      `json:"internal"`
	External      int      `json:"external"`
	Inaccessible  int      `json:"inaccessible"`
	BrokenLinks   []string `json:"broken_links,omitempty"`
	ExternalHosts []string `json:"external_hosts,omitempty"`
}

type AnalysisJob struct {
	ID            uuid.UUID `json:"id"`
	URL           string    `json:"url"`
	Priority      int       `json:"priority"`
	RetryCount    int       `json:"retry_count"`
	MaxRetries    int       `json:"max_retries"`
	UserID        string    `json:"user_id,omitempty"`
	CorrelationID string    `json:"correlation_id"`
	CreatedAt     time.Time `json:"created_at"`
}

func NewAnalysis(url, userID, correlationID string) *Analysis {
	return &Analysis{
		ID:            uuid.New(),
		URL:           url,
		Status:        StatusPending,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		UserID:        userID,
		CorrelationID: correlationID,
		Priority:      1,
	}
}

func NewAnalysisJob(url, userID, correlationID string, priority int) *AnalysisJob {
	return &AnalysisJob{
		ID:            uuid.New(),
		URL:           url,
		Priority:      priority,
		UserID:        userID,
		CorrelationID: correlationID,
		CreatedAt:     time.Now(),
		MaxRetries:    3,
	}
}

func (a *Analysis) MarkAsProcessing() {
	a.Status = StatusProcessing
	a.UpdatedAt = time.Now()
}

func (a *Analysis) MarkAsCompleted(result *AnalysisResult) {
	a.Status = StatusCompleted
	a.Result = result
	a.UpdatedAt = time.Now()
	now := time.Now()
	a.CompletedAt = &now
}

func (a *Analysis) MarkAsFailed(err string) {
	a.Status = StatusFailed
	a.Error = err
	a.UpdatedAt = time.Now()
}

func (a *Analysis) MarkAsRetrying() {
	a.Status = StatusRetrying
	a.RetryCount++
	a.UpdatedAt = time.Now()
}

func (a *Analysis) CanRetry(maxRetries int) bool {
	return a.RetryCount < maxRetries
}
