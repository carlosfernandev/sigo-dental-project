package postgres_test

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/carlosfpatino/sigo-dental/platform/postgres"
)

// testDSN returns the connection string for the integration test, or skips.
// The suite must stay green on a machine with no database: a test that fails
// for lack of infrastructure gets disabled, and then it protects nothing.
func testDSN(t *testing.T) string {
	t.Helper()
	dsn := os.Getenv("SIGO_TEST_PG_DSN")
	if dsn == "" {
		t.Skip("SIGO_TEST_PG_DSN is not set; skipping database integration test")
	}
	return dsn
}

func TestNewConnects(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	pool, err := postgres.New(ctx, postgres.Config{
		DSN:      testDSN(t),
		MaxConns: 4,
		MinConns: 0,
	})
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	defer pool.Close()

	var version string
	if err := pool.QueryRow(ctx, "select version()").Scan(&version); err != nil {
		t.Fatalf("query version: %v", err)
	}
	if !strings.Contains(version, "PostgreSQL 16") {
		t.Errorf("server = %q, want PostgreSQL 16", version)
	}

	if got := pool.Stat().MaxConns(); got != 4 {
		t.Errorf("MaxConns = %d, want 4", got)
	}
}

func TestNewRejectsInvalidDSN(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := postgres.New(ctx, postgres.Config{DSN: "this is not a dsn"}); err == nil {
		t.Fatal("New() accepted an invalid DSN")
	}
}

func TestNewFailsWhenUnreachable(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Port 1 on loopback is reserved and never listens. The pool must fail
	// at construction, not on the first request.
	dsn := "postgres://sigo:secret@127.0.0.1:1/sigo?sslmode=disable&connect_timeout=2"

	if _, err := postgres.New(ctx, postgres.Config{DSN: dsn}); err == nil {
		t.Fatal("New() succeeded against an unreachable database")
	}
}
