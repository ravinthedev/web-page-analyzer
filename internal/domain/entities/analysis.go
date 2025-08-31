package entities

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Analysis struct {
	ID        uuid.UUID       `json:"id" db:"id"`
	URL       string          `json:"url" db:"url"`
	Status    AnalysisStatus  `json:"status" db:"status"`
	Result    *AnalysisResult `json:"result,omitempty" db:"result"`
	CreatedAt time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt time.Time       `json:"updated_at" db:"updated_at"`
}

type AnalysisStatus string

const (
	StatusPending   AnalysisStatus = "pending"
	StatusCompleted AnalysisStatus = "completed"
	StatusFailed    AnalysisStatus = "failed"
)

type AnalysisResult struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Keywords    []string          `json:"keywords"`
	Links       []LinkAnalysis    `json:"links"`
	Images      []ImageAnalysis   `json:"images"`
	MetaTags    map[string]string `json:"meta_tags"`
	WordCount   int               `json:"word_count"`
	LoadTime    time.Duration     `json:"load_time"`
	StatusCode  int               `json:"status_code"`
}

type LinkAnalysis struct {
	URL        string `json:"url"`
	Text       string `json:"text"`
	IsExternal bool   `json:"is_external"`
	IsValid    bool   `json:"is_valid"`
}

type ImageAnalysis struct {
	URL     string `json:"url"`
	Alt     string `json:"alt"`
	Width   int    `json:"width"`
	Height  int    `json:"height"`
	IsValid bool   `json:"is_valid"`
}

func NewAnalysis(url string) *Analysis {
	now := time.Now()
	return &Analysis{
		ID:        uuid.New(),
		URL:       url,
		Status:    StatusPending,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (a *Analysis) SetResult(result *AnalysisResult) {
	a.Result = result
	a.Status = StatusCompleted
	a.UpdatedAt = time.Now()
}

func (a *Analysis) SetFailed() {
	a.Status = StatusFailed
	a.UpdatedAt = time.Now()
}

func (a *Analysis) ToJSON() ([]byte, error) {
	return json.Marshal(a)
}
