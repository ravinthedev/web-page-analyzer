package usecases

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAnalysisUseCase(t *testing.T) {
	uc := &AnalysisUseCase{}

	assert.NotNil(t, uc)
}

func TestAnalysisUseCaseFields(t *testing.T) {
	uc := &AnalysisUseCase{
		timeout: 300 * time.Second,
	}

	assert.Equal(t, 300*time.Second, uc.timeout)
}
