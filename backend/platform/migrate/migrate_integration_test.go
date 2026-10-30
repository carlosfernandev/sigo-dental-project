package migrate

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"os"
	"testing"
	"testing/fstest"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestUpPostgresIntegration(t *testing.T) {
	dsn := os.Getenv("SIGO_TEST_PG_DSN")
	if dsn == "" {
		t.Skip("SIGO_TEST_PG_DSN is not set")
	}

	ctx := context.Background()
	adminPool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("connect to integration database: %v", err)
	}
	t.Cleanup(adminPool.Close)

	var random [12]byte
	if _, err := rand.Read(random[:]); err != nil {
		t.Fatalf("generate test schema name: %v", err)
	}
	schema := "migration_test_" + hex.EncodeToString(random[:])
	if _, err := adminPool.Exec(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatalf("create isolated test schema: %v", err)
	}
	t.Cleanup(func() {
		if _, err := adminPool.Exec(context.Background(), "DROP SCHEMA "+schema+" CASCADE"); err != nil {
			t.Errorf("drop isolated test schema: %v", err)
		}
	})

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatalf("parse integration database DSN: %v", err)
	}
	poolConfig.ConnConfig.RuntimeParams["search_path"] = schema
	testPool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		t.Fatalf("connect to isolated test schema: %v", err)
	}
	t.Cleanup(testPool.Close)

	migrations := fstest.MapFS{
		"001_create_example.sql": {Data: []byte("CREATE TABLE example (id integer PRIMARY KEY)")},
	}
	if err := Up(ctx, testPool, migrations); err != nil {
		t.Fatalf("first Up() error = %v", err)
	}
	if err := Up(ctx, testPool, migrations); err != nil {
		t.Fatalf("second Up() error = %v", err)
	}

	var count int
	if err := testPool.QueryRow(ctx, "SELECT count(*) FROM schema_migrations").Scan(&count); err != nil {
		t.Fatalf("count applied migrations: %v", err)
	}
	if count != 1 {
		t.Errorf("applied migration count = %d, want 1", count)
	}
}
