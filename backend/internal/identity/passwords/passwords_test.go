package passwords

import (
	"errors"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestHashAndVerify(t *testing.T) {
	const password = "secure-clinic-password-2026"
	hash, err := Hash(password)
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}
	if !Verify(password, hash) {
		t.Fatal("Verify() rejected the correct password")
	}
	if Verify("another-clinic-password", hash) {
		t.Fatal("Verify() accepted an incorrect password")
	}
	cost, err := bcrypt.Cost([]byte(hash))
	if err != nil {
		t.Fatalf("bcrypt.Cost() error = %v", err)
	}
	if cost != BcryptCost {
		t.Errorf("hash cost = %d, want %d", cost, BcryptCost)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name     string
		password string
		email    string
		wantErr  error
	}{
		{name: "valid", password: "secure-clinic-password", email: "admin@example.com"},
		{name: "too short", password: "short-123", wantErr: ErrTooShort},
		{name: "common password", password: " Password123456 ", wantErr: ErrCommon},
		{name: "contains account email", password: "admin@example.comSecure", email: "ADMIN@example.com", wantErr: ErrContainsEmail},
		{name: "exceeds bcrypt byte limit", password: "é" + string(make([]byte, 0)) + "123456789012345678901234567890123456789012345678901234567890123456789012", wantErr: ErrTooLong},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.password, tt.email)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Validate() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestVerifyAndRehash(t *testing.T) {
	const password = "legacy-clinic-password"
	oldHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("create legacy hash: %v", err)
	}

	valid, replacement, err := VerifyAndRehash(password, string(oldHash))
	if err != nil {
		t.Fatalf("VerifyAndRehash() error = %v", err)
	}
	if !valid || replacement == "" {
		t.Fatalf("VerifyAndRehash() = (%v, %q), want valid with replacement", valid, replacement)
	}
	if !Verify(password, replacement) {
		t.Error("replacement hash does not verify")
	}
	if NeedsRehash(replacement) {
		t.Error("replacement hash still needs rehashing")
	}

	valid, replacement, err = VerifyAndRehash("incorrect-password", string(oldHash))
	if err != nil {
		t.Fatalf("VerifyAndRehash() for incorrect password error = %v", err)
	}
	if valid || replacement != "" {
		t.Errorf("incorrect password result = (%v, %q), want false and no replacement", valid, replacement)
	}
}

func TestNeedsRehashRejectsMalformedHash(t *testing.T) {
	if !NeedsRehash("not-a-bcrypt-hash") {
		t.Fatal("NeedsRehash() = false for malformed hash")
	}
}
