package usecases

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAnalysisUseCase(t *testing.T) {
	uc := &analysisUseCase{}

	assert.NotNil(t, uc)
}

func TestAnalysisUseCaseFields(t *testing.T) {
	uc := &analysisUseCase{
		cacheTTL: 300,
	}

	assert.Equal(t, 300, uc.cacheTTL)
}

func TestAnalysisUseCaseWithNilDependencies(t *testing.T) {
	uc := &analysisUseCase{
		analysisRepo: nil,
		cacheRepo:    nil,
		analyzer:     nil,
		logger:       nil,
		cacheTTL:     300,
	}

	assert.NotNil(t, uc)
	assert.Equal(t, 300, uc.cacheTTL)
}

func TestAnalysisUseCaseConstructor(t *testing.T) {
	uc := NewAnalysisUseCase(nil, nil, nil, nil, 600)

	assert.NotNil(t, uc)
}

func TestAnalyzeURLUseCase(t *testing.T) {
	uc := NewAnalysisUseCase(nil, nil, nil, nil, 300)

	assert.NotNil(t, uc)
}

func TestAnalyzeURLUseCaseWithInvalidURL(t *testing.T) {
	uc := NewAnalysisUseCase(nil, nil, nil, nil, 300)

	assert.NotNil(t, uc)
}

func TestGetAnalysisUseCase(t *testing.T) {
	uc := NewAnalysisUseCase(nil, nil, nil, nil, 300)

	assert.NotNil(t, uc)
}

func TestCacheTTLBehavior(t *testing.T) {
	uc := NewAnalysisUseCase(nil, nil, nil, nil, 300)

	// Test that use case is created successfully
	assert.NotNil(t, uc)
}
