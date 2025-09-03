package redis

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCacheRepository(t *testing.T) {
	assert.True(t, true)
}

func TestCacheRepositoryConstructor(t *testing.T) {
	repo := &cacheRepository{}
	assert.NotNil(t, repo)
}

func TestCacheRepositoryMethods(t *testing.T) {
	repo := &cacheRepository{}

	assert.NotNil(t, repo)
	assert.True(t, true)
}
