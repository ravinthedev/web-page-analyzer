package migrate

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMigrator(t *testing.T) {
	migrator := NewMigrator(nil)

	assert.NotNil(t, migrator)
}

func TestExtractVersion(t *testing.T) {
	migrator := &Migrator{}

	version, err := migrator.extractVersion("001_create_table.sql")

	assert.NoError(t, err)
	assert.Equal(t, int64(1), version)
}

func TestExtractVersionWithInvalidFilename(t *testing.T) {
	migrator := &Migrator{}

	version, err := migrator.extractVersion("invalid.sql")

	assert.Error(t, err)
	assert.Equal(t, int64(0), version)
}

func TestExtractVersionWithValidFilename(t *testing.T) {
	migrator := &Migrator{}

	version, err := migrator.extractVersion("002_add_table.sql")

	assert.NoError(t, err)
	assert.Equal(t, int64(2), version)
}

func TestExtractVersionWithLongFilename(t *testing.T) {
	migrator := &Migrator{}

	version, err := migrator.extractVersion("999_create_complex_table.sql")

	assert.NoError(t, err)
	assert.Equal(t, int64(999), version)
}

func TestExtractVersionWithZeroVersion(t *testing.T) {
	migrator := &Migrator{}

	version, err := migrator.extractVersion("000_initial.sql")

	assert.NoError(t, err)
	assert.Equal(t, int64(0), version)
}
