package identity

import (
	"context"
	"fmt"

	"github.com/carlosfpatino/sigo-dental/platform/authn"
)

type UserAdminRepository interface {
	ListUsers(ctx context.Context) ([]ManagedUser, error)
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
