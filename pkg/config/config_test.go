package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfigWithInvalidPath(t *testing.T) {
	cfg, err := Load("invalid/path.yaml")

	assert.Error(t, err)
	assert.Nil(t, cfg)
}

func TestConfigStruct(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{
			Port: "8080",
		},
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     "5432",
			User:     "user",
			Password: "password",
			Name:     "dbname",
		},
		Redis: RedisConfig{
			Host:     "localhost",
			Port:     "6379",
			Password: "",
			DB:       0,
		},
	}

	assert.Equal(t, "8080", cfg.Server.Port)
	assert.Equal(t, "localhost", cfg.Database.Host)
	assert.Equal(t, "5432", cfg.Database.Port)
	assert.Equal(t, "user", cfg.Database.User)
	assert.Equal(t, "password", cfg.Database.Password)
	assert.Equal(t, "dbname", cfg.Database.Name)
	assert.Equal(t, "localhost", cfg.Redis.Host)
	assert.Equal(t, "6379", cfg.Redis.Port)
	assert.Equal(t, "", cfg.Redis.Password)
	assert.Equal(t, 0, cfg.Redis.DB)
}
