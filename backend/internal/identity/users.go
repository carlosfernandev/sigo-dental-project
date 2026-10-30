package identity

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"regexp"
	"sort"
	"strings"

	"github.com/carlosfpatino/sigo-dental/internal/identity/passwords"
	"github.com/carlosfpatino/sigo-dental/platform/authn"
)

var ErrInvalidUserInput = errors.New("invalid user input")
var ErrRegistrationClosed = errors.New("initial administrator registration is closed")
var ErrCannotDeactivateSelf = errors.New("cannot deactivate the current user")
var userIDPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

type UserAdminRepository interface {
	ListUsers(ctx context.Context) ([]ManagedUser, error)
	CreateUser(ctx context.Context, email, passwordHash, role string) (ManagedUser, error)
	UpdateUser(ctx context.Context, userID, role string, active bool) (ManagedUser, error)
}

type paginatedUserAdminRepository interface {
	ListUsersPage(ctx context.Context, filter UserListFilter) (UserListResult, error)
	CreateUserWithAudit(ctx context.Context, actorID, email, passwordHash, role string) (ManagedUser, error)
	UpdateUserWithAudit(ctx context.Context, actorID, userID, role string, active *bool) (ManagedUser, error)
	DeactivateUserWithAudit(ctx context.Context, actorID, userID, reason string) (ManagedUser, error)
}

type initialAdminRegistrationRepository interface {
	RegistrationEnabled(ctx context.Context) (bool, error)
	RegisterInitialAdmin(ctx context.Context, email, passwordHash string) (ManagedUser, error)
}

type UserListFilter struct {
	Page      int
	PageSize  int
	Search    string
	Role      string
	IsActive  *bool
	SortBy    string
	SortOrder string
}

type UserListResult struct {
	Items    []ManagedUser `json:"items"`
	Total    int64         `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"page_size"`
}

type UserAdminService struct {
	repository any
}

func NewUserAdminService(repository any) *UserAdminService {
	return &UserAdminService{repository: repository}
}

// ListUsers limits identity administration to the SystemAdmin role.
func (s *UserAdminService) ListUsers(ctx context.Context, principal authn.Principal, filter UserListFilter) (UserListResult, error) {
	if err := authn.RequireAnyRole(principal, authn.RoleAdminTI); err != nil {
		return UserListResult{}, err
	}
	filter, err := validateUserListFilter(filter)
	if err != nil {
		return UserListResult{}, err
	}
	if repository, ok := s.repository.(paginatedUserAdminRepository); ok {
		users, err := repository.ListUsersPage(ctx, filter)
		if err != nil {
			return UserListResult{}, fmt.Errorf("list users: %w", err)
		}
		return users, nil
	}
	repository, ok := s.repository.(UserAdminRepository)
	if !ok {
		return UserListResult{}, fmt.Errorf("list users: unsupported repository")
	}
	users, err := repository.ListUsers(ctx)
	if err != nil {
		return UserListResult{}, fmt.Errorf("list users: %w", err)
	}
	return filterUsersInMemory(users, filter), nil
}

func (s *UserAdminService) CreateUser(ctx context.Context, principal authn.Principal, email, password, role string) (ManagedUser, error) {
	if err := authn.RequireAnyRole(principal, authn.RoleAdminTI); err != nil {
		return ManagedUser{}, err
	}
	return s.createUser(ctx, principal.UserID, email, password, role)
}

func (s *UserAdminService) RegistrationEnabled(ctx context.Context) (bool, error) {
	repository, ok := s.repository.(initialAdminRegistrationRepository)
	if !ok {
		return false, nil
	}
	enabled, err := repository.RegistrationEnabled(ctx)
	if err != nil {
		return false, fmt.Errorf("check initial administrator registration: %w", err)
	}
	return enabled, nil
}

// Register creates the initial administrator only while the user table is empty.
func (s *UserAdminService) Register(ctx context.Context, email, password string) (ManagedUser, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	parsed, err := mail.ParseAddress(email)
	if err != nil || parsed.Address != email || len(email) > 254 {
		return ManagedUser{}, ErrInvalidUserInput
	}
	if err := passwords.Validate(password, email); err != nil {
		return ManagedUser{}, fmt.Errorf("%w: %v", ErrInvalidUserInput, err)
	}
	passwordHash, err := passwords.Hash(password)
	if err != nil {
		return ManagedUser{}, fmt.Errorf("hash initial administrator password: %w", err)
	}
	repository, ok := s.repository.(initialAdminRegistrationRepository)
	if !ok {
		return ManagedUser{}, ErrRegistrationClosed
	}
	user, err := repository.RegisterInitialAdmin(ctx, email, passwordHash)
	if err != nil {
		return ManagedUser{}, fmt.Errorf("register initial administrator: %w", err)
	}
	return user, nil
}

func (s *UserAdminService) createUser(ctx context.Context, actorID, email, password, role string) (ManagedUser, error) {
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
	var user ManagedUser
	if repository, ok := s.repository.(paginatedUserAdminRepository); ok {
		user, err = repository.CreateUserWithAudit(ctx, actorID, email, passwordHash, role)
	} else if repository, ok := s.repository.(UserAdminRepository); ok {
		user, err = repository.CreateUser(ctx, email, passwordHash, role)
	} else {
		return ManagedUser{}, fmt.Errorf("create user: unsupported repository")
	}
	if err != nil {
		return ManagedUser{}, fmt.Errorf("create user: %w", err)
	}
	return user, nil
}

func (s *UserAdminService) UpdateUser(ctx context.Context, principal authn.Principal, userID, role string, active *bool) (ManagedUser, error) {
	if err := authn.RequireAnyRole(principal, authn.RoleAdminTI); err != nil {
		return ManagedUser{}, err
	}
	if !userIDPattern.MatchString(userID) || !authn.IsKnownRole(role) {
		return ManagedUser{}, ErrInvalidUserInput
	}
	if active != nil && !*active {
		return ManagedUser{}, ErrInvalidUserInput
	}
	var user ManagedUser
	var err error
	if repository, ok := s.repository.(paginatedUserAdminRepository); ok {
		user, err = repository.UpdateUserWithAudit(ctx, principal.UserID, userID, role, active)
	} else if repository, ok := s.repository.(UserAdminRepository); ok {
		activeValue := true
		if active != nil {
			activeValue = *active
		} else {
			users, listErr := repository.ListUsers(ctx)
			if listErr != nil {
				return ManagedUser{}, fmt.Errorf("read user before update: %w", listErr)
			}
			found := false
			for _, current := range users {
				if current.ID == userID {
					activeValue = current.IsActive
					found = true
					break
				}
			}
			if !found {
				return ManagedUser{}, ErrUserNotFound
			}
		}
		user, err = repository.UpdateUser(ctx, userID, role, activeValue)
	} else {
		return ManagedUser{}, fmt.Errorf("update user: unsupported repository")
	}
	if err != nil {
		return ManagedUser{}, fmt.Errorf("update user: %w", err)
	}
	return user, nil
}

func (s *UserAdminService) DeactivateUser(ctx context.Context, principal authn.Principal, userID, reason string) (ManagedUser, error) {
	if err := authn.RequireAnyRole(principal, authn.RoleAdminTI); err != nil {
		return ManagedUser{}, err
	}
	if !userIDPattern.MatchString(userID) {
		return ManagedUser{}, ErrInvalidUserInput
	}
	if userID == principal.UserID {
		return ManagedUser{}, ErrCannotDeactivateSelf
	}
	reason = strings.TrimSpace(reason)
	if len([]rune(reason)) < 3 || len([]rune(reason)) > 500 {
		return ManagedUser{}, ErrInvalidUserInput
	}
	var user ManagedUser
	var err error
	if repository, ok := s.repository.(paginatedUserAdminRepository); ok {
		user, err = repository.DeactivateUserWithAudit(ctx, principal.UserID, userID, reason)
	} else if repository, ok := s.repository.(UserAdminRepository); ok {
		users, listErr := repository.ListUsers(ctx)
		if listErr != nil {
			return ManagedUser{}, fmt.Errorf("read user before deactivation: %w", listErr)
		}
		var target *ManagedUser
		for index := range users {
			if users[index].ID == userID {
				target = &users[index]
				break
			}
		}
		if target == nil {
			return ManagedUser{}, ErrUserNotFound
		}
		user, err = repository.UpdateUser(ctx, userID, target.Role, false)
	} else {
		return ManagedUser{}, fmt.Errorf("deactivate user: unsupported repository")
	}
	if err != nil {
		return ManagedUser{}, fmt.Errorf("deactivate user: %w", err)
	}
	return user, nil
}

func filterUsersInMemory(users []ManagedUser, filter UserListFilter) UserListResult {
	filtered := make([]ManagedUser, 0, len(users))
	for _, user := range users {
		if filter.Search != "" && !strings.Contains(strings.ToLower(user.Email+" "+user.Role), strings.ToLower(filter.Search)) {
			continue
		}
		if filter.Role != "" && user.Role != filter.Role {
			continue
		}
		if filter.IsActive != nil && user.IsActive != *filter.IsActive {
			continue
		}
		filtered = append(filtered, user)
	}
	sort.Slice(filtered, func(i, j int) bool {
		left, right := filtered[i], filtered[j]
		less := left.Email < right.Email
		switch filter.SortBy {
		case "role":
			less = left.Role < right.Role
		case "is_active":
			less = !left.IsActive && right.IsActive
		}
		if filter.SortOrder == "desc" {
			return !less
		}
		return less
	})
	total := len(filtered)
	start := (filter.Page - 1) * filter.PageSize
	if start > total {
		start = total
	}
	end := start + filter.PageSize
	if end > total {
		end = total
	}
	return UserListResult{Items: filtered[start:end], Total: int64(total), Page: filter.Page, PageSize: filter.PageSize}
}

func validateUserListFilter(filter UserListFilter) (UserListFilter, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Page > 1_000_000 {
		return UserListFilter{}, ErrInvalidUserInput
	}
	if filter.PageSize == 0 {
		filter.PageSize = 25
	}
	if filter.PageSize != 25 {
		return UserListFilter{}, ErrInvalidUserInput
	}
	if filter.SortBy == "" {
		filter.SortBy = "email"
	}
	if filter.SortOrder == "" {
		filter.SortOrder = "asc"
	}
	if filter.SortBy != "email" && filter.SortBy != "role" && filter.SortBy != "is_active" {
		return UserListFilter{}, ErrInvalidUserInput
	}
	if filter.SortOrder != "asc" && filter.SortOrder != "desc" {
		return UserListFilter{}, ErrInvalidUserInput
	}
	if filter.Role != "" && !authn.IsKnownRole(filter.Role) {
		return UserListFilter{}, ErrInvalidUserInput
	}
	filter.Search = strings.TrimSpace(filter.Search)
	if len([]rune(filter.Search)) > 100 {
		return UserListFilter{}, ErrInvalidUserInput
	}
	return filter, nil
}
