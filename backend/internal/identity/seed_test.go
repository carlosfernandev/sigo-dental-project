package identity

import (
	"context"
	"testing"

	"golang.org/x/crypto/bcrypt"
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
	const password = "temporary-admin-password"
	if err := SeedInitialAdmin(context.Background(), repository, " Admin@Example.com ", password); err != nil {
		t.Fatalf("SeedInitialAdmin() error = %v", err)
	}
	if repository.email != "admin@example.com" {
		t.Errorf("email = %q, want normalized email", repository.email)
	}
	if repository.passwordHash == password {
		t.Fatal("repository received a plaintext password")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(repository.passwordHash), []byte(password)); err != nil {
		t.Errorf("stored hash does not match the initial password: %v", err)
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
