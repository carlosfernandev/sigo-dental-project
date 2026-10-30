package identity

import (
	"context"
	"encoding/json"
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
	var userID string
	err = tx.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, role_id, is_active, must_change_password)
		VALUES ($1, $2, $3, true, true)
		ON CONFLICT (email) DO NOTHING
		RETURNING id::text
	`, email, passwordHash, roleID).Scan(&userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return tx.Commit(ctx)
	}
	if err != nil {
		return fmt.Errorf("insert initial administrator: %w", err)
	}
	newValues, err := json.Marshal(map[string]any{"email": email, "role": authn.RoleAdminTI, "is_active": true})
	if err != nil {
		return fmt.Errorf("encode seeded administrator audit values: %w", err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO identity_user_audit_events (actor_user_id, target_user_id, action, reason, previous_values, new_values)
		VALUES ($1::uuid, $1::uuid, 'created', 'initial administrator seed', '{}'::jsonb, $2::jsonb)
	`, userID, newValues); err != nil {
		return fmt.Errorf("audit seeded administrator creation: %w", err)
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
func (r *PostgresRepository) ListUsersPage(ctx context.Context, filter UserListFilter) (UserListResult, error) {
	pattern := "%" + filter.Search + "%"
	var activeFilter any
	if filter.IsActive != nil {
		activeFilter = *filter.IsActive
	}
	var result UserListResult
	result.Page = filter.Page
	result.PageSize = filter.PageSize
	if err := r.db.QueryRow(ctx, `
		SELECT count(*)
		FROM users
		JOIN roles ON roles.id = users.role_id
		WHERE ($1 = '%%' OR users.email ILIKE $1 OR roles.name ILIKE $1)
		  AND ($2 = '' OR roles.name = $2)
		  AND ($3::boolean IS NULL OR users.is_active = $3)
	`, pattern, filter.Role, activeFilter).Scan(&result.Total); err != nil {
		return UserListResult{}, fmt.Errorf("count identity users: %w", err)
	}
	orderColumns := map[string]string{"email": "users.email", "role": "roles.name", "is_active": "users.is_active"}
	orderColumn := orderColumns[filter.SortBy]
	orderDirection := "ASC"
	if filter.SortOrder == "desc" {
		orderDirection = "DESC"
	}
	offset := (filter.Page - 1) * filter.PageSize
	query := fmt.Sprintf(`
		SELECT users.id::text, users.email, roles.name, users.is_active
		FROM users
		JOIN roles ON roles.id = users.role_id
		WHERE ($1 = '%%' OR users.email ILIKE $1 OR roles.name ILIKE $1)
		  AND ($2 = '' OR roles.name = $2)
		  AND ($3::boolean IS NULL OR users.is_active = $3)
		ORDER BY %s %s, users.email ASC
		LIMIT $4 OFFSET $5
	`, orderColumn, orderDirection)
	rows, err := r.db.Query(ctx, query, pattern, filter.Role, activeFilter, filter.PageSize, offset)
	if err != nil {
		return UserListResult{}, fmt.Errorf("list identity users: %w", err)
	}
	defer rows.Close()

	result.Items = make([]ManagedUser, 0, filter.PageSize)
	for rows.Next() {
		var user ManagedUser
		if err := rows.Scan(&user.ID, &user.Email, &user.Role, &user.IsActive); err != nil {
			return UserListResult{}, fmt.Errorf("scan identity user: %w", err)
		}
		result.Items = append(result.Items, user)
	}
	if err := rows.Err(); err != nil {
		return UserListResult{}, fmt.Errorf("read identity users: %w", err)
	}
	return result, nil
}

func (r *PostgresRepository) CreateUserWithAudit(ctx context.Context, actorID, email, passwordHash, role string) (ManagedUser, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return ManagedUser{}, fmt.Errorf("begin identity user creation: %w", err)
	}
	defer tx.Rollback(ctx)
	var user ManagedUser
	user.Role = role
	err = tx.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, role_id, is_active, must_change_password)
		SELECT $1, $2, roles.id, true, false
		FROM roles
		WHERE roles.name = $3
		RETURNING id::text, email, is_active
	`, email, passwordHash, role).Scan(&user.ID, &user.Email, &user.IsActive)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ManagedUser{}, ErrInvalidUserInput
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ManagedUser{}, ErrEmailAlreadyExists
		}
		return ManagedUser{}, fmt.Errorf("insert identity user: %w", err)
	}
	newValues, err := json.Marshal(map[string]any{"email": user.Email, "role": user.Role, "is_active": user.IsActive})
	if err != nil {
		return ManagedUser{}, fmt.Errorf("encode identity user audit values: %w", err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO identity_user_audit_events (actor_user_id, target_user_id, action, previous_values, new_values)
		VALUES ($1::uuid, $2::uuid, 'created', '{}'::jsonb, $3::jsonb)
	`, actorID, user.ID, newValues); err != nil {
		return ManagedUser{}, fmt.Errorf("audit identity user creation: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return ManagedUser{}, fmt.Errorf("commit identity user creation: %w", err)
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
	newValues, err := json.Marshal(map[string]any{"email": user.Email, "role": user.Role, "is_active": user.IsActive})
	if err != nil {
		return ManagedUser{}, fmt.Errorf("encode initial administrator audit values: %w", err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO identity_user_audit_events (actor_user_id, target_user_id, action, reason, previous_values, new_values)
		VALUES ($1::uuid, $1::uuid, 'created', 'initial administrator self-registration', '{}'::jsonb, $2::jsonb)
	`, user.ID, newValues); err != nil {
		return ManagedUser{}, fmt.Errorf("audit initial administrator creation: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return ManagedUser{}, fmt.Errorf("commit initial administrator registration: %w", err)
	}
	return user, nil
}

func (r *PostgresRepository) UpdateUserWithAudit(ctx context.Context, actorID, userID, role string, active *bool) (ManagedUser, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return ManagedUser{}, fmt.Errorf("begin identity user update: %w", err)
	}
	defer tx.Rollback(ctx)
	var previousUser ManagedUser
	if err := tx.QueryRow(ctx, `
		SELECT users.id::text, users.email, roles.name, users.is_active
		FROM users JOIN roles ON roles.id = users.role_id
		WHERE users.id = $1::uuid FOR UPDATE
	`, userID).Scan(&previousUser.ID, &previousUser.Email, &previousUser.Role, &previousUser.IsActive); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ManagedUser{}, ErrUserNotFound
		}
		return ManagedUser{}, fmt.Errorf("read identity user before update: %w", err)
	}
	var user ManagedUser
	user.Role = role
	err = tx.QueryRow(ctx, `
		UPDATE users
		SET role_id = (SELECT id FROM roles WHERE name = $2),
		    is_active = COALESCE($3::boolean, users.is_active),
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
	previousValues, err := json.Marshal(map[string]any{"email": previousUser.Email, "role": previousUser.Role, "is_active": previousUser.IsActive})
	if err != nil {
		return ManagedUser{}, fmt.Errorf("encode previous identity user values: %w", err)
	}
	newValues, err := json.Marshal(map[string]any{"email": user.Email, "role": user.Role, "is_active": user.IsActive})
	if err != nil {
		return ManagedUser{}, fmt.Errorf("encode updated identity user values: %w", err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO identity_user_audit_events (actor_user_id, target_user_id, action, previous_values, new_values)
		VALUES ($1::uuid, $2::uuid, 'updated', $3::jsonb, $4::jsonb)
	`, actorID, user.ID, previousValues, newValues); err != nil {
		return ManagedUser{}, fmt.Errorf("audit identity user update: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return ManagedUser{}, fmt.Errorf("commit identity user update: %w", err)
	}
	return user, nil
}

func (r *PostgresRepository) DeactivateUserWithAudit(ctx context.Context, actorID, userID, reason string) (ManagedUser, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return ManagedUser{}, fmt.Errorf("begin identity user deactivation: %w", err)
	}
	defer tx.Rollback(ctx)
	var previousUser ManagedUser
	if err := tx.QueryRow(ctx, `
		SELECT users.id::text, users.email, roles.name, users.is_active
		FROM users JOIN roles ON roles.id = users.role_id
		WHERE users.id = $1::uuid FOR UPDATE
	`, userID).Scan(&previousUser.ID, &previousUser.Email, &previousUser.Role, &previousUser.IsActive); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ManagedUser{}, ErrUserNotFound
		}
		return ManagedUser{}, fmt.Errorf("read identity user before deactivation: %w", err)
	}
	if !previousUser.IsActive {
		return previousUser, nil
	}
	var user ManagedUser
	err = tx.QueryRow(ctx, `
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
	previousValues, err := json.Marshal(map[string]any{"email": previousUser.Email, "role": previousUser.Role, "is_active": previousUser.IsActive})
	if err != nil {
		return ManagedUser{}, fmt.Errorf("encode previous identity user values: %w", err)
	}
	newValues, err := json.Marshal(map[string]any{"email": user.Email, "role": user.Role, "is_active": user.IsActive})
	if err != nil {
		return ManagedUser{}, fmt.Errorf("encode deactivated identity user values: %w", err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO identity_user_audit_events (actor_user_id, target_user_id, action, reason, previous_values, new_values)
		VALUES ($1::uuid, $2::uuid, 'deactivated', $3, $4::jsonb, $5::jsonb)
	`, actorID, user.ID, reason, previousValues, newValues); err != nil {
		return ManagedUser{}, fmt.Errorf("audit identity user deactivation: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return ManagedUser{}, fmt.Errorf("commit identity user deactivation: %w", err)
	}
	return user, nil
}
