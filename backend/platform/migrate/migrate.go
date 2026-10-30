// Package migrate applies versioned, forward-only SQL migrations.
package migrate

import (
	"context"
	"fmt"
	"io/fs"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const advisoryLockKey int64 = 0x5349474f

var migrationName = regexp.MustCompile(`^([0-9]{3,})_([a-z0-9_]+)\.sql$`)

type migration struct {
	version int64
	name    string
	sql     string
}

// Up applies each unapplied migration in filename order. Migrations, their
// tracking records, and the advisory lock share a transaction.
func Up(ctx context.Context, db *pgxpool.Pool, source fs.FS) error {
	migrations, err := readMigrations(source)
	if err != nil {
		return err
	}

	conn, err := db.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("acquire migration connection: %w", err)
	}
	defer conn.Release()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin migration transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock($1)", advisoryLockKey); err != nil {
		return fmt.Errorf("lock migrations: %w", err)
	}
	if _, err := tx.Exec(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
		version bigint PRIMARY KEY,
		name text NOT NULL,
		applied_at timestamptz NOT NULL DEFAULT now()
	)`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	rows, err := tx.Query(ctx, "SELECT version, name FROM schema_migrations ORDER BY version")
	if err != nil {
		return fmt.Errorf("read schema_migrations: %w", err)
	}
	applied := make(map[int64]string)
	for rows.Next() {
		var version int64
		var name string
		if err := rows.Scan(&version, &name); err != nil {
			rows.Close()
			return fmt.Errorf("scan schema_migrations: %w", err)
		}
		applied[version] = name
	}
	rowsErr := rows.Err()
	rows.Close()
	if rowsErr != nil {
		return fmt.Errorf("iterate schema_migrations: %w", rowsErr)
	}

	byVersion := make(map[int64]string, len(migrations))
	for _, item := range migrations {
		byVersion[item.version] = item.name
	}
	for version, name := range applied {
		if current, ok := byVersion[version]; !ok || current != name {
			return fmt.Errorf("applied migration %d (%s) is missing or renamed", version, name)
		}
	}

	for _, item := range migrations {
		if _, ok := applied[item.version]; ok {
			continue
		}
		if strings.TrimSpace(item.sql) == "" {
			return fmt.Errorf("migration %s is empty", item.name)
		}
		if _, err := tx.Exec(ctx, item.sql, pgx.QueryExecModeSimpleProtocol); err != nil {
			return fmt.Errorf("apply migration %s: %w", item.name, err)
		}
		if _, err := tx.Exec(ctx,
			"INSERT INTO schema_migrations (version, name) VALUES ($1, $2)",
			item.version, item.name,
		); err != nil {
			return fmt.Errorf("record migration %s: %w", item.name, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit migrations: %w", err)
	}
	return nil
}

func readMigrations(source fs.FS) ([]migration, error) {
	entries, err := fs.ReadDir(source, ".")
	if err != nil {
		return nil, fmt.Errorf("read migration directory: %w", err)
	}

	var migrations []migration
	versions := make(map[int64]string)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		matches := migrationName.FindStringSubmatch(entry.Name())
		if matches == nil {
			continue
		}
		version, err := strconv.ParseInt(matches[1], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid migration version in %q: %w", entry.Name(), err)
		}
		if previous, exists := versions[version]; exists {
			return nil, fmt.Errorf("duplicate migration version %d in %q and %q", version, previous, entry.Name())
		}
		contents, err := fs.ReadFile(source, entry.Name())
		if err != nil {
			return nil, fmt.Errorf("read migration %q: %w", entry.Name(), err)
		}
		versions[version] = entry.Name()
		migrations = append(migrations, migration{version: version, name: entry.Name(), sql: string(contents)})
	}
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].version < migrations[j].version
	})
	return migrations, nil
}
