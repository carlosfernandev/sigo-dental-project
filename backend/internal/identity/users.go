package identity

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"regexp"
	"strings"

	"github.com/carlosfpatino/sigo-dental/internal/identity/passwords"
	"github.com/carlosfpatino/sigo-dental/platform/authn"
)

var ErrInvalidUserInput = errors.New("invalid user input")
var userIDPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

type UserAdminRepository interface {
	ListUsers(ctx context.Context) ([]ManagedUser, error)
	CreateUser(ctx context.Context, email, passwordHash, role string) (ManagedUser, error)
	UpdateUser(ctx context.Context, userID, role string, active bool) (ManagedUser, error)
}

type UserAdminService struct {
	repository UserAdminRepository
}

func NewUserAdminService(repository UserAdminRepository) *UserAdminService {
	return &UserAdminService{repository: repository}
}

// ListUsers limits identity administration to the SystemAdmin role.
func (s *UserAdminService) ListUsers(ctx context.Context, principal authn.Principal) ([]ManagedUser, error) {
	if err := authn.RequireAnyRole(principal, authn.RoleAdminTI); err != nil {
		return nil, err
	}
	users, err := s.repository.ListUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	return users, nil
}

func (s *UserAdminService) CreateUser(ctx context.Context, principal authn.Principal, email, password, role string) (ManagedUser, error) {
	if err := authn.RequireAnyRole(principal, authn.RoleAdminTI); err != nil {
		return ManagedUser{}, err
	}
	email = strings.ToLower(strings.TrimSpace(email))
	parsed, err := mail.ParseAddress(email)
	if err != nil || parsed.Address != email || len(email) > 254 || !authn.IsKnownRole(role) {
		return ManagedUser{}, ErrInvalidUserInput
	}
	if err := passwords.Validate(password, email); err != nil {
		return ManagedUser{}, fmt.Errorf("%w: %v", ErrInvalidUserInput, err)
	}
	passwordHash, err := passwords.Hash(password)
	if err != nil {
		return ManagedUser{}, fmt.Errorf("hash initial user password: %w", err)
	}
	user, err := s.repository.CreateUser(ctx, email, passwordHash, role)
	if err != nil {
		return ManagedUser{}, fmt.Errorf("create user: %w", err)
	}
	return user, nil
}

func (s *UserAdminService) UpdateUser(ctx context.Context, principal authn.Principal, userID, role string, active bool) (ManagedUser, error) {
	if err := authn.RequireAnyRole(principal, authn.RoleAdminTI); err != nil {
		return ManagedUser{}, err
	}
	if !userIDPattern.MatchString(userID) || !authn.IsKnownRole(role) {
		return ManagedUser{}, ErrInvalidUserInput
	}
	user, err := s.repository.UpdateUser(ctx, userID, role, active)
	if err != nil {
		return ManagedUser{}, fmt.Errorf("update user: %w", err)
	}
	return user, nil
}
