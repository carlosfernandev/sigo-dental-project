package identity

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
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

// FindByEmail returns only fields required to authenticate a user.
func (r *PostgresRepository) FindByEmail(ctx context.Context, email string) (AuthUser, error) {
	var user AuthUser
	err := r.db.QueryRow(ctx, `
		SELECT users.id::text, users.email, users.password_hash, roles.name,
		       users.is_active, users.must_change_password
		FROM users
	JOIN roles ON roles.id = users.role_id
		WHERE users.email = $1
	`, email).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Role, &user.IsActive, &user.MustChangePassword)
	if errors.Is(err, pgx.ErrNoRows) {
		return AuthUser{}, ErrUserNotFound
	}
	if err != nil {
		return AuthUser{}, fmt.Errorf("find identity user: %w", err)
	}
	return user, nil
}

func (r *PostgresRepository) UpdatePasswordHash(ctx context.Context, userID, passwordHash string) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE users
		SET password_hash = $1, updated_at = now()
		WHERE id = $2::uuid
	`, passwordHash, userID)
	if err != nil {
		return fmt.Errorf("update identity password hash: %w", err)
	}
	if tag.RowsAffected() != 1 {
		return fmt.Errorf("update identity password hash: user no longer exists")
	}
	return nil
}
