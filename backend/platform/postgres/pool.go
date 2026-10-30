// Package postgres opens and configures the connection pool for the monolith.
//
// The monolith owns a single database. All domains (identity, patient, clinical,
// scheduling, billing) live in the public schema.
package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	// A boot that cannot reach its database should fail, not hang.
	connectTimeout = 10 * time.Second

	// Long-lived connections hide a failover. Recycling them keeps the pool
	// pointed at the database that actually answers.
	maxConnLifetime = 30 * time.Minute
	maxConnIdleTime = 5 * time.Minute

	// HealthCheckPeriod lets the pool evict a dead connection in the
	// background instead of surfacing it on a request.
	healthCheckPeriod = 30 * time.Second
)

// Config describes one pool. It is deliberately independent of the config
// package so the caller maps its own settings into it.
type Config struct {
	DSN      string
	MaxConns int32
	MinConns int32
}

// New opens the pool and verifies it with a ping.
//
// A pool that is constructed but never pinged fails on the first request
// instead of at boot. RNF-02 requires a service to report its own health, and
// a service that starts against a dead database cannot do that honestly.
func New(ctx context.Context, cfg Config) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("parse postgres dsn: %w", err)
	}

	if cfg.MaxConns > 0 {
		poolCfg.MaxConns = cfg.MaxConns
	}
	if cfg.MinConns >= 0 {
		poolCfg.MinConns = cfg.MinConns
	}
	poolCfg.MaxConnLifetime = maxConnLifetime
	poolCfg.MaxConnIdleTime = maxConnIdleTime
	poolCfg.HealthCheckPeriod = healthCheckPeriod

	connectCtx, cancel := context.WithTimeout(ctx, connectTimeout)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(connectCtx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("create postgres pool: %w", err)
	}

	if err := pool.Ping(connectCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return pool, nil
}
