package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLogger(t *testing.T) {
	logger, err := New("info", false)

	assert.NoError(t, err)
	assert.NotNil(t, logger)
}

func TestNewLoggerWithDebug(t *testing.T) {
	logger, err := New("debug", true)

	assert.NoError(t, err)
	assert.NotNil(t, logger)
}

func TestNewLoggerWithWarn(t *testing.T) {
	logger, err := New("warn", false)

	assert.NoError(t, err)
	assert.NotNil(t, logger)
}

func TestNewLoggerWithError(t *testing.T) {
	logger, err := New("error", false)

	assert.NoError(t, err)
	assert.NotNil(t, logger)
}

func TestNewLoggerWithFatal(t *testing.T) {
	logger, err := New("fatal", false)

	assert.NoError(t, err)
	assert.NotNil(t, logger)
}
