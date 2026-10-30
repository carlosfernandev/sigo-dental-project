package config_test

import (
	"log/slog"
	"net/url"
	"strings"
	"testing"

	"github.com/carlosfpatino/sigo-dental/platform/config"
)

// testJWTSecret is long enough to pass token manager validation (32 bytes).
const testJWTSecret = "test-only-secret-with-more-than-32-bytes"

// baseline is a complete, valid environment. Each case overrides one key so a
// failure points at that key and not at a missing neighbour.
func baseline() map[string]string {
	return map[string]string{
		"SIGO_ENV":                  "",
		"SIGO_LOG_LEVEL":            "",
		"SIGO_JWT_SECRET":           testJWTSecret,
		"SIGO_PG_HOST":              "localhost",
		"SIGO_PG_PORT":              "",
		"SIGO_PG_USER":              "sigo",
		"SIGO_PG_PASSWORD":          "sigo_dev_password",
		"SIGO_PG_DATABASE":          "sigo",
		"SIGO_PG_SSLMODE":           "",
		"SIGO_PG_MAX_CONNS":         "",
		"SIGO_PG_MIN_CONNS":         "",
		"SIGO_REDIS_HOST":           "",
		"SIGO_REDIS_PORT":           "",
		"SIGO_REDIS_PASSWORD":       "",
		"SIGO_REDIS_DATABASE":       "",
		"SIGO_CORS_ALLOWED_ORIGINS": "",
	}
}

func apply(t *testing.T, overrides map[string]string) {
	t.Helper()
	env := baseline()
	for k, v := range overrides {
		env[k] = v
	}
	for k, v := range env {
		t.Setenv(k, v)
	}
}

func TestLoad(t *testing.T) {
	tests := []struct {
		name    string
		env     map[string]string
		wantErr string // substring; empty means success
	}{
		{
			name: "valid with defaults",
			env:  map[string]string{},
		},
		{
			name:    "missing password",
			env:     map[string]string{"SIGO_PG_PASSWORD": ""},
			wantErr: "SIGO_PG_PASSWORD is required",
		},
		{
			name:    "missing database",
			env:     map[string]string{"SIGO_PG_DATABASE": ""},
			wantErr: "SIGO_PG_DATABASE is required",
		},
		{
			name:    "port not a number",
			env:     map[string]string{"SIGO_PG_PORT": "postgres"},
			wantErr: "SIGO_PG_PORT must be an integer",
		},
		{
			name:    "port out of range",
			env:     map[string]string{"SIGO_PG_PORT": "70000"},
			wantErr: "SIGO_PG_PORT must be between",
		},
		{
			name: "min connections above max",
			env: map[string]string{
				"SIGO_PG_MIN_CONNS": "20",
				"SIGO_PG_MAX_CONNS": "10",
			},
			wantErr: "cannot exceed SIGO_PG_MAX_CONNS",
		},
		{
			name:    "unknown environment",
			env:     map[string]string{"SIGO_ENV": "prod"},
			wantErr: "SIGO_ENV must be one of",
		},
		{
			name: "production without TLS",
			env: map[string]string{
				"SIGO_ENV":        config.EnvProduction,
				"SIGO_PG_SSLMODE": "disable",
			},
			wantErr: "SIGO_PG_SSLMODE cannot be disable",
		},
		{
			name:    "missing JWT secret",
			env:     map[string]string{"SIGO_JWT_SECRET": ""},
			wantErr: "SIGO_JWT_SECRET is required",
		},
		{
			name:    "unknown log level",
			env:     map[string]string{"SIGO_LOG_LEVEL": "verbose"},
			wantErr: "SIGO_LOG_LEVEL must be one of",
		},
		{
			name:    "invalid CORS origin",
			env:     map[string]string{"SIGO_CORS_ALLOWED_ORIGINS": "ftp://127.0.0.1:5173"},
			wantErr: "SIGO_CORS_ALLOWED_ORIGINS contains an invalid origin",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apply(t, tt.env)

			cfg, err := config.Load()

			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("Load() succeeded, want error containing %q", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("Load() error = %q, want it to contain %q", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("Load() unexpected error: %v", err)
			}
			if cfg.Postgres.Port != 5432 {
				t.Errorf("default port = %d, want 5432", cfg.Postgres.Port)
			}
			if cfg.Postgres.MaxConns != 10 {
				t.Errorf("default MaxConns = %d, want 10", cfg.Postgres.MaxConns)
			}
			if cfg.LogLevel != slog.LevelInfo {
				t.Errorf("default LogLevel = %v, want info", cfg.LogLevel)
			}
			if got := string(cfg.Auth.JWTSecret); got != testJWTSecret {
				t.Errorf("JWTSecret = %q, want %q", got, testJWTSecret)
			}
			if cfg.Redis.Host != "127.0.0.1" || cfg.Redis.Port != 6379 || cfg.Redis.Database != 0 {
				t.Errorf("Redis defaults = %+v, want host 127.0.0.1, port 6379, database 0", cfg.Redis)
			}
		})
	}
}

func TestLoadCORSOrigins(t *testing.T) {
	apply(t, map[string]string{"SIGO_CORS_ALLOWED_ORIGINS": "https://app.example.com, https://admin.example.com/"})

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	want := []string{"https://app.example.com", "https://admin.example.com"}
	if len(cfg.CORS.AllowedOrigins) != len(want) {
		t.Fatalf("allowed origins = %v, want %v", cfg.CORS.AllowedOrigins, want)
	}
	for i := range want {
		if cfg.CORS.AllowedOrigins[i] != want[i] {
			t.Errorf("allowed origins = %v, want %v", cfg.CORS.AllowedOrigins, want)
			break
		}
	}
}

func TestLoadReportsEveryProblemAtOnce(t *testing.T) {
	apply(t, map[string]string{
		"SIGO_PG_HOST":     "",
		"SIGO_PG_USER":     "",
		"SIGO_PG_PASSWORD": "",
	})

	_, err := config.Load()
	if err == nil {
		t.Fatal("Load() succeeded, want error")
	}

	for _, key := range []string{"SIGO_PG_HOST", "SIGO_PG_USER", "SIGO_PG_PASSWORD"} {
		if !strings.Contains(err.Error(), key) {
			t.Errorf("error does not mention %s: %v", key, err)
		}
	}
}

func TestPostgresDSNEscapesPassword(t *testing.T) {
	const password = "p@ss/w:rd?#1"

	p := config.Postgres{
		Host:     "localhost",
		Port:     5432,
		User:     "sigo",
		Password: password,
		Database: "sigo",
		SSLMode:  "disable",
	}

	u, err := url.Parse(p.DSN())
	if err != nil {
		t.Fatalf("DSN() is not a valid URL: %v", err)
	}

	got, ok := u.User.Password()
	if !ok {
		t.Fatal("DSN() dropped the password")
	}
	if got != password {
		t.Errorf("password round-trip = %q, want %q", got, password)
	}
	if u.Path != "/sigo" {
		t.Errorf("path = %q, want /sigo", u.Path)
	}
	if u.Query().Get("sslmode") != "disable" {
		t.Errorf("sslmode = %q, want disable", u.Query().Get("sslmode"))
	}
}

func TestPostgresRedactedHidesPassword(t *testing.T) {
	const password = "sigo_dev_password"

	p := config.Postgres{
		Host:     "localhost",
		Port:     5432,
		User:     "sigo",
		Password: password,
		Database: "sigo",
		SSLMode:  "disable",
	}

	if got := p.Redacted(); strings.Contains(got, password) {
		t.Fatalf("Redacted() leaked the password: %s", got)
	}
}
