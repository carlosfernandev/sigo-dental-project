package identity

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresRepository implements identity persistence for the monolith.
type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) InsertAdminIfAbsent(ctx context.Context, email, passwordHash string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin admin seed transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var roleID string
	if err := tx.QueryRow(ctx, "SELECT id FROM roles WHERE name = $1", "admin_ti").Scan(&roleID); err != nil {
		return fmt.Errorf("find admin_ti role: %w", err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO users (email, password_hash, role_id, is_active, must_change_password)
		VALUES ($1, $2, $3, true, true)
		ON CONFLICT (email) DO NOTHING
	`, email, passwordHash, roleID); err != nil {
		return fmt.Errorf("insert initial administrator: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit admin seed transaction: %w", err)
	}
	return nil
}
