// Package config loads and validates service configuration from the environment.
//
// Configuration is read once at startup and passed down explicitly. No other
// package calls os.Getenv, so a missing variable fails the boot instead of a
// request in production.
package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
)

// Environment values accepted by SIGO_ENV.
const (
	EnvDevelopment = "development"
	EnvTest        = "test"
	EnvStaging     = "staging"
	EnvProduction  = "production"
)

// Config is the full configuration of one service.
type Config struct {
	Env      string
	LogLevel slog.Level
	Auth     Auth
	Postgres Postgres
	Redis    Redis
	CORS     CORS
}

// Load reads the environment and returns a validated Config. Every problem is
// reported at once, so a misconfigured boot does not take one restart per typo.
func Load() (Config, error) {
	l := &loader{}

	cfg := Config{
		Env:      l.oneOf("SIGO_ENV", EnvDevelopment, EnvDevelopment, EnvTest, EnvStaging, EnvProduction),
		LogLevel: l.level("SIGO_LOG_LEVEL", slog.LevelInfo),
		Auth:     loadAuth(l),
		Postgres: loadPostgres(l),
		Redis:    loadRedis(l),
		CORS:     loadCORS(l, l.optional("SIGO_ENV")),
	}

	if cfg.Postgres.MinConns > cfg.Postgres.MaxConns {
		l.add("SIGO_PG_MIN_CONNS (%d) cannot exceed SIGO_PG_MAX_CONNS (%d)",
			cfg.Postgres.MinConns, cfg.Postgres.MaxConns)
	}
	if cfg.Env == EnvProduction && cfg.Postgres.SSLMode == "disable" {
		l.add("SIGO_PG_SSLMODE cannot be disable when SIGO_ENV=%s", EnvProduction)
	}

	if len(l.problems) > 0 {
		return Config{}, fmt.Errorf("invalid configuration:\n  - %s", strings.Join(l.problems, "\n  - "))
	}
	return cfg, nil
}

// IsProduction reports whether the service runs in the production environment.
func (c Config) IsProduction() bool { return c.Env == EnvProduction }

type loader struct {
	problems []string
}

func (l *loader) add(format string, args ...any) {
	l.problems = append(l.problems, fmt.Sprintf(format, args...))
}

func (l *loader) required(key string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		l.add("%s is required", key)
	}
	return v
}

func (l *loader) optional(key string) string {
	return strings.TrimSpace(os.Getenv(key))
}

func (l *loader) value(key, fallback string) string {
	if value := l.optional(key); value != "" {
		return value
	}
	return fallback
}

func (l *loader) integer(key string, fallback, min, max int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		l.add("%s must be an integer, got %q", key, raw)
		return fallback
	}
	if v < min || v > max {
		l.add("%s must be between %d and %d, got %d", key, min, max, v)
		return fallback
	}
	return v
}

func (l *loader) oneOf(key, fallback string, allowed ...string) string {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	for _, a := range allowed {
		if raw == a {
			return raw
		}
	}
	l.add("%s must be one of %s, got %q", key, strings.Join(allowed, ", "), raw)
	return fallback
}

func (l *loader) level(key string, fallback slog.Level) slog.Level {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	var lvl slog.Level
	if err := lvl.UnmarshalText([]byte(raw)); err != nil {
		l.add("%s must be one of debug, info, warn, error, got %q", key, raw)
		return fallback
	}
	return lvl
}
