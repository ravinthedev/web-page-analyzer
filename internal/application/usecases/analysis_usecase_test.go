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
