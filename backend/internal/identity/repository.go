package identity

import (
	"context"
	"errors"
	"fmt"

	"github.com/carlosfpatino/sigo-dental/platform/authn"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresRepository implements identity persistence for the monolith.
type PostgresRepository struct {
	db *pgxpool.Pool
}

type ManagedUser struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	IsActive bool   `json:"is_active"`
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

// ListUsers returns account fields suitable for the administrator user list.
func (r *PostgresRepository) ListUsers(ctx context.Context) ([]ManagedUser, error) {
	rows, err := r.db.Query(ctx, `
		SELECT users.id::text, users.email, roles.name, users.is_active
		FROM users
	JOIN roles ON roles.id = users.role_id
		ORDER BY users.email
	`)
	if err != nil {
		return nil, fmt.Errorf("list identity users: %w", err)
	}
	defer rows.Close()

	users := make([]ManagedUser, 0)
	for rows.Next() {
		var user ManagedUser
		if err := rows.Scan(&user.ID, &user.Email, &user.Role, &user.IsActive); err != nil {
			return nil, fmt.Errorf("scan identity user: %w", err)
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read identity users: %w", err)
	}
	return users, nil
}

func (r *PostgresRepository) CreateUser(ctx context.Context, email, passwordHash, role string) (ManagedUser, error) {
	var user ManagedUser
	user.Role = role
	err := r.db.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, role_id, is_active, must_change_password)
		SELECT $1, $2, roles.id, true, false
		FROM roles
		WHERE roles.name = $3
		RETURNING id::text, email, is_active
	`, email, passwordHash, role).Scan(&user.ID, &user.Email, &user.IsActive)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ManagedUser{}, ErrEmailAlreadyExists
		}
		return ManagedUser{}, fmt.Errorf("insert identity user: %w", err)
	}
	return user, nil
}

func (r *PostgresRepository) RegistrationEnabled(ctx context.Context) (bool, error) {
	var enabled bool
	if err := r.db.QueryRow(ctx, `SELECT NOT EXISTS (SELECT 1 FROM users)`).Scan(&enabled); err != nil {
		return false, fmt.Errorf("check initial administrator registration: %w", err)
	}
	return enabled, nil
}

func (r *PostgresRepository) RegisterInitialAdmin(ctx context.Context, email, passwordHash string) (ManagedUser, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return ManagedUser{}, fmt.Errorf("begin initial administrator registration: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `LOCK TABLE users IN EXCLUSIVE MODE`); err != nil {
		return ManagedUser{}, fmt.Errorf("lock users for initial administrator registration: %w", err)
	}
	var hasUsers bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM users)`).Scan(&hasUsers); err != nil {
		return ManagedUser{}, fmt.Errorf("check users before initial administrator registration: %w", err)
	}
	if hasUsers {
		return ManagedUser{}, ErrRegistrationClosed
	}

	var user ManagedUser
	user.Role = authn.RoleAdminTI
	err = tx.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, role_id, is_active, must_change_password)
		SELECT $1, $2, roles.id, true, false
		FROM roles
		WHERE roles.name = $3
		RETURNING id::text, email, is_active
	`, email, passwordHash, user.Role).Scan(&user.ID, &user.Email, &user.IsActive)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ManagedUser{}, ErrEmailAlreadyExists
		}
		return ManagedUser{}, fmt.Errorf("insert initial administrator: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return ManagedUser{}, fmt.Errorf("commit initial administrator registration: %w", err)
	}
	return user, nil
}

func (r *PostgresRepository) UpdateUser(ctx context.Context, userID, role string, active bool) (ManagedUser, error) {
	var user ManagedUser
	user.Role = role
	err := r.db.QueryRow(ctx, `
		UPDATE users
		SET role_id = (SELECT id FROM roles WHERE name = $2),
		    is_active = $3,
		    updated_at = now()
		WHERE id = $1::uuid
		RETURNING id::text, email, is_active
	`, userID, role, active).Scan(&user.ID, &user.Email, &user.IsActive)
	if errors.Is(err, pgx.ErrNoRows) {
		return ManagedUser{}, ErrUserNotFound
	}
	if err != nil {
		return ManagedUser{}, fmt.Errorf("update identity user: %w", err)
	}
	return user, nil
}

func (r *PostgresRepository) DeactivateUser(ctx context.Context, userID string) (ManagedUser, error) {
	var user ManagedUser
	err := r.db.QueryRow(ctx, `
		UPDATE users AS user_account
		SET is_active = false, updated_at = now()
		FROM roles
		WHERE user_account.id = $1::uuid AND roles.id = user_account.role_id
		RETURNING user_account.id::text, user_account.email, roles.name, user_account.is_active
	`, userID).Scan(&user.ID, &user.Email, &user.Role, &user.IsActive)
	if errors.Is(err, pgx.ErrNoRows) {
		return ManagedUser{}, ErrUserNotFound
	}
	if err != nil {
		return ManagedUser{}, fmt.Errorf("deactivate identity user: %w", err)
	}
	return user, nil
}
