package identity

import (
	"context"
	"fmt"
	"net/mail"
	"strings"

	"github.com/carlosfpatino/sigo-dental/internal/identity/passwords"
)

// AdminSeedRepository stores the initial system administrator without
// replacing an existing account.
type AdminSeedRepository interface {
	InsertAdminIfAbsent(ctx context.Context, email, passwordHash string) error
}

// SeedInitialAdmin normalizes the configured email and stores only a bcrypt
// hash of the initial password. Existing accounts are left untouched.
func SeedInitialAdmin(ctx context.Context, repository AdminSeedRepository, email, password string) error {
	email, err := normalizeEmail(email)
	if err != nil {
		return err
	}
	if err := passwords.Validate(password, email); err != nil {
		return fmt.Errorf("invalid initial administrator password: %w", err)
	}

	hash, err := passwords.Hash(password)
	if err != nil {
		return fmt.Errorf("hash initial administrator password: %w", err)
	}
	if err := repository.InsertAdminIfAbsent(ctx, email, hash); err != nil {
		return fmt.Errorf("seed initial administrator: %w", err)
	}
	return nil
}

func normalizeEmail(value string) (string, error) {
	email := strings.ToLower(strings.TrimSpace(value))
	if email == "" || len(email) > 254 {
		return "", fmt.Errorf("a valid administrator email is required")
	}
	parsed, err := mail.ParseAddress(email)
	if err != nil || parsed.Address != email {
		return "", fmt.Errorf("a valid administrator email is required")
	}
	return email, nil
}
