package identity

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"os"
	"testing"

	"github.com/carlosfpatino/sigo-dental/internal/identity/passwords"
	"github.com/carlosfpatino/sigo-dental/platform/migrate"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestSeedInitialAdminPostgresIntegration(t *testing.T) {
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
	schema := "identity_test_" + hex.EncodeToString(random[:])
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

	if err := migrate.Up(ctx, testPool, os.DirFS("../../migrations")); err != nil {
		t.Fatalf("apply identity migrations: %v", err)
	}
	repository := NewPostgresRepository(testPool)
	if err := SeedInitialAdmin(ctx, repository, " ADMIN@Example.com ", "first-temporary-password-2026"); err != nil {
		t.Fatalf("seed administrator: %v", err)
	}
	if err := SeedInitialAdmin(ctx, repository, "admin@example.com", "replacement-password"); err != nil {
		t.Fatalf("repeat administrator seed: %v", err)
	}

	var email, passwordHash, role string
	var active, mustChange bool
	err = testPool.QueryRow(ctx, `
		SELECT users.email, users.password_hash, roles.name, users.is_active, users.must_change_password
		FROM users JOIN roles ON roles.id = users.role_id
		WHERE users.email = $1
	`, "admin@example.com").Scan(&email, &passwordHash, &role, &active, &mustChange)
	if err != nil {
		t.Fatalf("query seeded administrator: %v", err)
	}
	if email != "admin@example.com" || role != "admin_ti" || !active || !mustChange {
		t.Errorf("seeded administrator = (%q, %q, %v, %v), want normalized active admin_ti requiring password change", email, role, active, mustChange)
	}
	if !passwords.Verify("first-temporary-password-2026", passwordHash) {
		t.Error("stored password hash was replaced or invalid")
	}
	if passwords.Verify("replacement-password", passwordHash) {
		t.Error("repeated seed replaced the existing administrator password")
	}

	var count int
	if err := testPool.QueryRow(ctx, "SELECT count(*) FROM users WHERE email = $1", "admin@example.com").Scan(&count); err != nil {
		t.Fatalf("count seeded administrators: %v", err)
	}
	if count != 1 {
		t.Errorf("seeded administrator rows = %d, want 1", count)
	}
}
