package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"webpage-analyzer/internal/domain/repositories"
	"webpage-analyzer/pkg/config"

	"github.com/go-redis/redis/v8"
)

type cacheRepository struct {
	client *redis.Client
}

func NewCacheRepository(cfg *config.RedisConfig) repositories.CacheRepository {
	rdb := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	})

	return &cacheRepository{client: rdb}
}

func (r *cacheRepository) Set(ctx context.Context, key string, value interface{}, ttl int) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	duration := time.Duration(ttl) * time.Second
	return r.client.Set(ctx, key, data, duration).Err()
}

func (r *cacheRepository) Get(ctx context.Context, key string, dest interface{}) error {
	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("key not found")
		}
		return fmt.Errorf("failed to get value: %w", err)
	}

	return json.Unmarshal([]byte(data), dest)
}

func (r *cacheRepository) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

func (r *cacheRepository) Exists(ctx context.Context, key string) (bool, error) {
	count, err := r.client.Exists(ctx, key).Result()
	return count > 0, err
}
