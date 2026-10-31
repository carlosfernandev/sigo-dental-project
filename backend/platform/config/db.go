// Package config provides database initialization using the loaded configuration.
package config

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strconv"

	"github.com/carlosfpatino/sigo-dental/platform/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Postgres holds the connection settings for the monolith's database.
type Postgres struct {
	Host        string
	Port        int
	User        string
	Password    string
	Database    string
	SSLMode     string
	SSLRootCert string
	MaxConns    int32
	MinConns    int32
}

func loadPostgres(l *loader) Postgres {
	return Postgres{
		Host:        l.required("SIGO_PG_HOST"),
		Port:        l.integer("SIGO_PG_PORT", 5432, 1, 65535),
		User:        l.required("SIGO_PG_USER"),
		Password:    l.required("SIGO_PG_PASSWORD"),
		Database:    l.required("SIGO_PG_DATABASE"),
		SSLMode:     l.oneOf("SIGO_PG_SSLMODE", "disable", "disable", "require", "verify-ca", "verify-full"),
		SSLRootCert: l.optional("SIGO_PG_SSLROOTCERT"),
		MaxConns:    int32(l.integer("SIGO_PG_MAX_CONNS", 10, 1, 500)),
		MinConns:    int32(l.integer("SIGO_PG_MIN_CONNS", 0, 0, 500)),
	}
}

// DSN renders a libpq connection string with URL-escaped credentials.
func (p Postgres) DSN() string {
	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(p.User, p.Password),
		Host:     net.JoinHostPort(p.Host, strconv.Itoa(p.Port)),
		Path:     "/" + p.Database,
		RawQuery: p.query().Encode(),
	}
	return u.String()
}

func (p Postgres) query() url.Values {
	values := url.Values{"sslmode": {p.SSLMode}}
	if p.SSLRootCert != "" {
		values.Set("sslrootcert", p.SSLRootCert)
	}
	return values
}

// Redacted renders a safe-to-log DSN without the password.
func (p Postgres) Redacted() string {
	return fmt.Sprintf("postgres://%s@%s/%s?sslmode=%s",
		p.User, net.JoinHostPort(p.Host, strconv.Itoa(p.Port)), p.Database, p.SSLMode)
}

// InitDB creates the PostgreSQL connection pool using the loaded configuration.
// The pool is returned for dependency injection; there are no global variables.
func InitDB(ctx context.Context, cfg Config) (*pgxpool.Pool, error) {
	pool, err := postgres.New(ctx, postgres.Config{
		DSN:      cfg.Postgres.DSN(),
		MaxConns: cfg.Postgres.MaxConns,
		MinConns: cfg.Postgres.MinConns,
	})
	if err != nil {
		return nil, fmt.Errorf("init postgres pool: %w", err)
	}
	return pool, nil
}
