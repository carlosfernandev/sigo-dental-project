package identity

import (
	"context"
	"testing"

	"github.com/carlosfpatino/sigo-dental/internal/identity/passwords"
	"golang.org/x/crypto/bcrypt"
)

type authRepositoryFake struct {
	user        AuthUser
	findErr     error
	updatedHash string
}

func (r *authRepositoryFake) FindByEmail(_ context.Context, _ string) (AuthUser, error) {
	return r.user, r.findErr
}

func (r *authRepositoryFake) UpdatePasswordHash(_ context.Context, _ string, hash string) error {
	r.updatedHash = hash
	return nil
}

func TestAuthenticateRehashesObsoletePasswordHash(t *testing.T) {
	const password = "legacy-user-password-2026"
	legacyHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("create legacy password hash: %v", err)
	}
	repository := &authRepositoryFake{
		user: AuthUser{
			ID: "d53a806e-e017-4e31-a57c-9778cd7c61f1", Email: "admin@example.com",
			PasswordHash: string(legacyHash), Role: "admin_ti", IsActive: true,
		},
	}
	service := NewAuthService(repository, "unused")
	if _, err := service.Authenticate(context.Background(), "ADMIN@example.com", password); err != nil {
		t.Fatalf("Authenticate() error = %v", err)
	}
	if repository.updatedHash == "" {
		t.Fatal("obsolete hash was not replaced")
	}
	newCost, err := bcrypt.Cost([]byte(repository.updatedHash))
	if err != nil || newCost != passwords.BcryptCost {
		t.Error("replacement hash is invalid or still obsolete")
	}
}
