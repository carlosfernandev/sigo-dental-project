// Package billing owns medical catalog, quotes, invoices, and payments.
package billing

import (
	"context"
	"time"

	"github.com/carlosfpatino/sigo-dental/platform/authn"
)

type Service struct {
	repo Repository
	now  func() time.Time
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
		now: func() time.Time {
			return time.Now().In(time.FixedZone("America/Guayaquil", -5*60*60))
		},
	}
}

// Procedure operations
func (s *Service) CreateProcedure(ctx context.Context, principal authn.Principal, p Procedure) (Procedure, error) {
	if err := requireAdminAccess(principal); err != nil {
		return Procedure{}, err
	}
	if err := ValidateProcedure(p); err != nil {
		return Procedure{}, err
	}
	return s.repo.CreateProcedure(ctx, p)
}

func (s *Service) ListProcedures(ctx context.Context, principal authn.Principal) ([]Procedure, error) {
	if err := requireBillingAccess(principal); err != nil {
		return nil, err
	}
	return s.repo.ListProcedures(ctx)
}

func (s *Service) GetProcedure(ctx context.Context, principal authn.Principal, id string) (Procedure, error) {
	if err := requireBillingAccess(principal); err != nil {
		return Procedure{}, err
	}
	return s.repo.GetProcedure(ctx, id)
}

func (s *Service) UpdateProcedure(ctx context.Context, principal authn.Principal, p Procedure) (Procedure, error) {
	if err := requireAdminAccess(principal); err != nil {
		return Procedure{}, err
	}
	if err := ValidateProcedure(p); err != nil {
		return Procedure{}, err
	}
	return s.repo.UpdateProcedure(ctx, p)
}

// Quote operations
func (s *Service) CreateQuote(ctx context.Context, principal authn.Principal, q Quote) (Quote, error) {
	if err := requireBillingAccess(principal); err != nil {
		return Quote{}, err
	}
	if q.Status == "" {
		q.Status = "draft"
	}
	if err := ValidateQuote(q, s.now()); err != nil {
		return Quote{}, err
	}
	return s.repo.CreateQuote(ctx, q)
}

func (s *Service) ListQuotes(ctx context.Context, principal authn.Principal, patientID *string) ([]Quote, error) {
	if err := requireBillingAccess(principal); err != nil {
		return nil, err
	}
	return s.repo.ListQuotes(ctx, patientID)
}

func (s *Service) GetQuote(ctx context.Context, principal authn.Principal, id string) (Quote, error) {
	if err := requireBillingAccess(principal); err != nil {
		return Quote{}, err
	}
	return s.repo.GetQuote(ctx, id)
}

func (s *Service) UpdateQuote(ctx context.Context, principal authn.Principal, q Quote) (Quote, error) {
	if err := requireBillingAccess(principal); err != nil {
		return Quote{}, err
	}
	if err := ValidateQuote(q, s.now()); err != nil {
		return Quote{}, err
	}
	if q.Status == "approved" && q.ApprovedAt == nil {
		now := s.now()
		q.ApprovedAt = &now
	}
	return s.repo.UpdateQuote(ctx, q)
}

func requireAdminAccess(principal authn.Principal) error {
	return authn.RequireAnyRole(principal, authn.RoleAdminTI)
}

func requireBillingAccess(principal authn.Principal) error {
	return authn.RequireAnyRole(principal,
		authn.RoleReceptionistCashier,
		authn.RoleDentist,
	)
}
