// Package cache implements the Redis Cache-Aside adapter used by domain services.
package cache

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// CacheRepository exposes failure-tolerant cache operations to application code.
type CacheRepository interface {
	Get(ctx context.Context, key string, destination any) bool
	Set(ctx context.Context, key string, value any, ttl time.Duration)
	Delete(ctx context.Context, keys ...string)
	DeleteByPattern(ctx context.Context, pattern string)
}

type redisCommands interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Delete(ctx context.Context, keys ...string) error
	Scan(ctx context.Context, cursor uint64, pattern string, count int64) ([]string, uint64, error)
}

type redisAdapter struct {
	client *redis.Client
}

func (r redisAdapter) Get(ctx context.Context, key string) ([]byte, error) {
	return r.client.Get(ctx, key).Bytes()
}

func (r redisAdapter) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return r.client.Set(ctx, key, value, ttl).Err()
}

func (r redisAdapter) Delete(ctx context.Context, keys ...string) error {
	return r.client.Del(ctx, keys...).Err()
}

func (r redisAdapter) Scan(ctx context.Context, cursor uint64, pattern string, count int64) ([]string, uint64, error) {
	return r.client.Scan(ctx, cursor, pattern, count).Result()
}

// RedisCache implements CacheRepository. A nil Redis client disables caching.
type RedisCache struct {
	client redisCommands
	logger *slog.Logger
}

func NewRedisCache(client *redis.Client, logger *slog.Logger) *RedisCache {
	if logger == nil {
		logger = slog.Default()
	}
	cache := &RedisCache{logger: logger}
	if client != nil {
		cache.client = redisAdapter{client: client}
	}
	return cache
}

func (r *RedisCache) Get(ctx context.Context, key string, destination any) bool {
	if r.client == nil {
		return false
	}

	data, err := r.client.Get(ctx, key)
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			r.warn("redis cache read failed", "error", err)
		}
		return false
	}
	if err := json.Unmarshal(data, destination); err != nil {
		r.warn("redis cache entry is invalid", "error", err)
		return false
	}
	return true
}

func (r *RedisCache) Set(ctx context.Context, key string, value any, ttl time.Duration) {
	if r.client == nil {
		return
	}
	if ttl <= 0 {
		r.warn("redis cache write skipped because TTL is not positive")
		return
	}

	data, err := json.Marshal(value)
	if err != nil {
		r.warn("redis cache value could not be encoded", "error", err)
		return
	}
	if err := r.client.Set(ctx, key, data, ttl); err != nil {
		r.warn("redis cache write failed", "error", err)
	}
}

func (r *RedisCache) Delete(ctx context.Context, keys ...string) {
	if r.client == nil || len(keys) == 0 {
		return
	}
	if err := r.client.Delete(ctx, keys...); err != nil {
		r.warn("redis cache invalidation failed", "error", err)
	}
}

// DeleteByPattern uses SCAN in batches rather than Redis KEYS, which can block
// Redis while traversing a large keyspace.
func (r *RedisCache) DeleteByPattern(ctx context.Context, pattern string) {
	if r.client == nil || strings.TrimSpace(pattern) == "" {
		return
	}

	var cursor uint64
	for {
		keys, nextCursor, err := r.client.Scan(ctx, cursor, pattern, 100)
		if err != nil {
			r.warn("redis cache pattern scan failed", "error", err)
			return
		}
		if len(keys) > 0 {
			if err := r.client.Delete(ctx, keys...); err != nil {
				r.warn("redis cache pattern invalidation failed", "error", err)
				return
			}
		}
		if nextCursor == 0 {
			return
		}
		cursor = nextCursor
	}
}

func (r *RedisCache) warn(message string, args ...any) {
	if r.logger != nil {
		r.logger.Warn(message, args...)
	}
}
