package identity

import (
	"context"
	"testing"

	"github.com/carlosfpatino/sigo-dental/internal/identity/passwords"
)

type seedRepository struct {
	email        string
	passwordHash string
}

func (r *seedRepository) InsertAdminIfAbsent(_ context.Context, email, passwordHash string) error {
	r.email = email
	r.passwordHash = passwordHash
	return nil
}

func TestSeedInitialAdminNormalizesEmailAndHashesPassword(t *testing.T) {
	repository := &seedRepository{}
	const password = "temporary-admin-password-2026"
	if err := SeedInitialAdmin(context.Background(), repository, " Admin@Example.com ", password); err != nil {
		t.Fatalf("SeedInitialAdmin() error = %v", err)
	}
	if repository.email != "admin@example.com" {
		t.Errorf("email = %q, want normalized email", repository.email)
	}
	if repository.passwordHash == password {
		t.Fatal("repository received a plaintext password")
	}
	if !passwords.Verify(password, repository.passwordHash) {
		t.Error("stored hash does not match the initial password")
	}
}

func TestSeedInitialAdminRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		password string
	}{
		{name: "missing email", email: "", password: "secret"},
		{name: "invalid email", email: "not-an-email", password: "secret"},
		{name: "missing password", email: "admin@example.com", password: ""},
		{name: "email in password", email: "admin@example.com", password: "admin@example.comSecure"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repository := &seedRepository{}
			if err := SeedInitialAdmin(context.Background(), repository, tt.email, tt.password); err == nil {
				t.Fatal("SeedInitialAdmin() succeeded, want validation error")
			}
			if repository.email != "" || repository.passwordHash != "" {
				t.Fatal("repository was called with invalid input")
			}
		})
	}
}
