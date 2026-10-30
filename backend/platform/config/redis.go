// Package config owns Redis cache settings and client initialization.
package config

import (
	"context"
	"log/slog"
	"net"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// Redis holds the connection settings for the cache.
type Redis struct {
	Host     string
	Port     int
	Password string
	Database int
}

func loadRedis(l *loader) Redis {
	return Redis{
		Host:     l.value("SIGO_REDIS_HOST", "127.0.0.1"),
		Port:     l.integer("SIGO_REDIS_PORT", 6379, 1, 65535),
		Password: l.optional("SIGO_REDIS_PASSWORD"),
		Database: l.integer("SIGO_REDIS_DATABASE", 0, 0, 15),
	}
}

// InitRedis creates a Redis client and performs a bounded startup health check.
// An unavailable cache is logged and returned as a client: Redis is optional,
// and callers must let cache operations fall back to PostgreSQL.
func InitRedis(ctx context.Context, cfg Redis, logger *slog.Logger) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:         net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port)),
		Password:     cfg.Password,
		DB:           cfg.Database,
		DialTimeout:  3 * time.Second,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
		MaxRetries:   3,
	})

	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	if err := client.Ping(pingCtx).Err(); err != nil {
		logger.Warn("redis is unavailable; continuing without cache", "address", net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port)), "error", err)
	}
	return client
}
