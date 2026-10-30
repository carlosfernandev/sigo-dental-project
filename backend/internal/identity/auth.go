package identity

import (
	"context"
	"errors"
	"fmt"

	"github.com/carlosfpatino/sigo-dental/internal/identity/passwords"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailAlreadyExists = errors.New("email already exists")
)

type AuthUser struct {
	ID                 string
	Email              string
	PasswordHash       string
	Role               string
	IsActive           bool
	MustChangePassword bool
}

type AuthRepository interface {
	FindByEmail(ctx context.Context, email string) (AuthUser, error)
	UpdatePasswordHash(ctx context.Context, userID, passwordHash string) error
}

type AuthenticatedUser struct {
	ID                 string
	Email              string
	Role               string
	MustChangePassword bool
}

type AuthService struct {
	repository AuthRepository
	dummyHash  string
}

func NewAuthService(repository AuthRepository, dummyHash string) *AuthService {
	return &AuthService{repository: repository, dummyHash: dummyHash}
}

// Authenticate returns an internal authenticated result without creating a
// session. Session issuance is handled by the later session task.
func (s *AuthService) Authenticate(ctx context.Context, email, password string) (AuthenticatedUser, error) {
	email, err := normalizeEmail(email)
	if err != nil {
		return AuthenticatedUser{}, ErrInvalidCredentials
	}

	user, err := s.repository.FindByEmail(ctx, email)
	if errors.Is(err, ErrUserNotFound) {
		passwords.Verify(password, s.dummyHash)
		return AuthenticatedUser{}, ErrInvalidCredentials
	}
	if err != nil {
		return AuthenticatedUser{}, fmt.Errorf("load credentials: %w", err)
	}

	if !user.IsActive {
		passwords.Verify(password, user.PasswordHash)
		return AuthenticatedUser{}, ErrInvalidCredentials
	}
	valid, replacementHash, err := passwords.VerifyAndRehash(password, user.PasswordHash)
	if err != nil {
		return AuthenticatedUser{}, fmt.Errorf("rehash identity password: %w", err)
	}
	if !valid {
		return AuthenticatedUser{}, ErrInvalidCredentials
	}
	if replacementHash != "" {
		if err := s.repository.UpdatePasswordHash(ctx, user.ID, replacementHash); err != nil {
			return AuthenticatedUser{}, fmt.Errorf("persist rehashed identity password: %w", err)
		}
	}

	return AuthenticatedUser{
		ID:                 user.ID,
		Email:              user.Email,
		Role:               user.Role,
		MustChangePassword: user.MustChangePassword,
	}, nil
}
