package entities

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewAnalysis(t *testing.T) {
	url := "https://example.com"
	userID := "test-user"
	correlationID := "test-correlation-id"

	analysis := NewAnalysis(url, userID, correlationID)

	assert.NotNil(t, analysis)
	assert.Equal(t, url, analysis.URL)
	assert.Equal(t, StatusPending, analysis.Status)
	assert.Equal(t, userID, analysis.UserID)
	assert.Equal(t, correlationID, analysis.CorrelationID)
}

func TestMarkAsProcessing(t *testing.T) {
	analysis := NewAnalysis("https://example.com", "test-user", "test-correlation-id")

	analysis.MarkAsProcessing()
	assert.Equal(t, StatusProcessing, analysis.Status)
}

func TestMarkAsCompleted(t *testing.T) {
	analysis := NewAnalysis("https://example.com", "test-user", "test-correlation-id")
	result := &AnalysisResult{
		HTMLVersion:  "HTML5",
		Title:        "Test Page",
		Headings:     map[string]int{"h1": 1},
		Links:        LinkAnalysis{Internal: 2, External: 1, Inaccessible: 0},
		HasLoginForm: false,
	}

	analysis.MarkAsCompleted(result)
	assert.Equal(t, result, analysis.Result)
	assert.Equal(t, StatusCompleted, analysis.Status)
}

func TestMarkAsFailed(t *testing.T) {
	analysis := NewAnalysis("https://example.com", "test-user", "test-correlation-id")

	analysis.MarkAsFailed("test error")
	assert.Equal(t, StatusFailed, analysis.Status)
	assert.Equal(t, "test error", analysis.Error)
}

func TestCanRetry(t *testing.T) {
	analysis := NewAnalysis("https://example.com", "test-user", "test-correlation-id")

	assert.True(t, analysis.CanRetry(3))

	analysis.RetryCount = 3
	assert.False(t, analysis.CanRetry(3))
}

func TestAnalysisResult(t *testing.T) {
	result := &AnalysisResult{
		HTMLVersion:  "HTML5",
		Title:        "Test Page",
		Headings:     map[string]int{"h1": 1},
		Links:        LinkAnalysis{Internal: 2, External: 1, Inaccessible: 0},
		HasLoginForm: false,
	}

	assert.Equal(t, "HTML5", result.HTMLVersion)
	assert.Equal(t, "Test Page", result.Title)
	assert.Equal(t, 1, result.Headings["h1"])
}

func TestLinkAnalysis(t *testing.T) {
	links := LinkAnalysis{
		Internal:     5,
		External:     3,
		Inaccessible: 1,
	}

	assert.Equal(t, 5, links.Internal)
	assert.Equal(t, 3, links.External)
	assert.Equal(t, 1, links.Inaccessible)
}

func TestNewAnalysisJob(t *testing.T) {
	url := "https://example.com"
	userID := "test-user"
	correlationID := "test-correlation-id"
	priority := 1

	job := NewAnalysisJob(url, userID, correlationID, priority)

	assert.NotNil(t, job)
	assert.Equal(t, url, job.URL)
	assert.Equal(t, priority, job.Priority)
	assert.Equal(t, userID, job.UserID)
}

func TestAnalysisStatus(t *testing.T) {
	assert.Equal(t, "pending", string(StatusPending))
	assert.Equal(t, "processing", string(StatusProcessing))
	assert.Equal(t, "completed", string(StatusCompleted))
	assert.Equal(t, "failed", string(StatusFailed))
}

func TestAnalysisFields(t *testing.T) {
	analysis := &Analysis{
		URL:           "https://example.com",
		Status:        StatusCompleted,
		UserID:        "user1",
		CorrelationID: "corr1",
		Priority:      1,
		RetryCount:    0,
	}

	assert.Equal(t, "https://example.com", analysis.URL)
	assert.Equal(t, StatusCompleted, analysis.Status)
	assert.Equal(t, "user1", analysis.UserID)
	assert.Equal(t, "corr1", analysis.CorrelationID)
	assert.Equal(t, 1, analysis.Priority)
	assert.Equal(t, 0, analysis.RetryCount)
}

func TestAnalysisResultFields(t *testing.T) {
	result := &AnalysisResult{
		StatusCode:   200,
		LoadTime:     time.Second,
		Title:        "Test Page",
		HTMLVersion:  "HTML5",
		Headings:     map[string]int{"h1": 1, "h2": 2},
		HasLoginForm: false,
	}

	assert.Equal(t, 200, result.StatusCode)
	assert.Equal(t, time.Second, result.LoadTime)
	assert.Equal(t, "Test Page", result.Title)
	assert.Equal(t, "HTML5", result.HTMLVersion)
	assert.Len(t, result.Headings, 2)
	assert.False(t, result.HasLoginForm)
}

func TestAnalysisJobFields(t *testing.T) {
	job := &AnalysisJob{
		URL:           "https://example.com",
		UserID:        "user1",
		CorrelationID: "corr1",
		Priority:      1,
	}

	assert.Equal(t, "https://example.com", job.URL)
	assert.Equal(t, "user1", job.UserID)
	assert.Equal(t, "corr1", job.CorrelationID)
	assert.Equal(t, 1, job.Priority)
}
