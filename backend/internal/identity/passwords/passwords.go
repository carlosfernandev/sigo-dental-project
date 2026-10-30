// Package passwords provides identity password validation and bcrypt hashing.
package passwords

import (
	"errors"
	"strings"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"
)

const (
	MinimumLength = 12
	BcryptCost    = 12
)

var (
	ErrTooShort       = errors.New("password must contain at least 12 characters")
	ErrTooLong        = errors.New("password exceeds bcrypt's 72-byte limit")
	ErrCommon         = errors.New("password is too common")
	ErrContainsEmail  = errors.New("password must not contain the account email")
	ErrInvalidHash    = errors.New("password hash is invalid")
	ErrHashGeneration = errors.New("password hash could not be generated")
)

var commonPasswords = map[string]struct{}{
	"password123456": {},
	"123456789012":   {},
	"qwertyuiop123":  {},
	"letmeinplease":  {},
	"adminadmin123":  {},
	"odontologia123": {},
}

// Validate enforces the password policy, including the account-specific email check.
func Validate(password, email string) error {
	if utf8.RuneCountInString(password) < MinimumLength {
		return ErrTooShort
	}
	if len([]byte(password)) > 72 {
		return ErrTooLong
	}
	if _, ok := commonPasswords[strings.ToLower(strings.TrimSpace(password))]; ok {
		return ErrCommon
	}
	if normalizedEmail := strings.ToLower(strings.TrimSpace(email)); normalizedEmail != "" &&
		strings.Contains(strings.ToLower(password), normalizedEmail) {
		return ErrContainsEmail
	}
	return nil
}

// Hash validates a password and returns its bcrypt hash at BcryptCost.
func Hash(password string) (string, error) {
	if err := Validate(password, ""); err != nil {
		return "", err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
	if err != nil {
		return "", ErrHashGeneration
	}
	return string(hash), nil
}

// Verify reports whether password matches the encoded bcrypt hash.
func Verify(password, encodedHash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(encodedHash), []byte(password)) == nil
}

// NeedsRehash reports whether a hash is malformed or below the current cost.
func NeedsRehash(encodedHash string) bool {
	cost, err := bcrypt.Cost([]byte(encodedHash))
	return err != nil || cost < BcryptCost
}

// VerifyAndRehash verifies a password and returns a replacement hash when its
// stored bcrypt cost is obsolete. A blank replacement means no rehash is needed.
func VerifyAndRehash(password, encodedHash string) (valid bool, replacementHash string, err error) {
	if !Verify(password, encodedHash) {
		return false, "", nil
	}
	if !NeedsRehash(encodedHash) {
		return true, "", nil
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
	if err != nil {
		return true, "", ErrHashGeneration
	}
	return true, string(hash), nil
}
