// Package billing owns medical catalog, quotes, invoices, and payments.
package billing

import (
	"context"
	"fmt"
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

// CreateInvoice calculates all line totals on the server and issues the invoice atomically.
func (s *Service) CreateInvoice(ctx context.Context, principal authn.Principal, invoice Invoice) (Invoice, error) {
	if err := requirePaymentAccess(principal); err != nil {
		return Invoice{}, err
	}
	if err := ValidateInvoice(invoice, s.now()); err != nil {
		return Invoice{}, err
	}
	invoice.Status = "issued"
	now := s.now().UTC()
	invoice.IssuedAt = &now
	tx, err := s.repo.BeginInvoiceTx(ctx)
	if err != nil {
		return Invoice{}, err
	}
	defer tx.Rollback(ctx)
	created, err := tx.CreateInvoice(ctx, invoice)
	if err != nil {
		return Invoice{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Invoice{}, err
	}
	return created, nil
}

func (s *Service) RegisterPayment(ctx context.Context, principal authn.Principal, invoiceID string, payment Payment) (Invoice, error) {
	if err := requirePaymentAccess(principal); err != nil {
		return Invoice{}, err
	}
	if !billingUUIDPattern.MatchString(invoiceID) {
		return Invoice{}, ErrInvalidInvoice
	}
	if err := ValidatePayment(payment); err != nil {
		return Invoice{}, err
	}
	tx, err := s.repo.BeginInvoiceTx(ctx)
	if err != nil {
		return Invoice{}, err
	}
	defer tx.Rollback(ctx)
	balance, err := tx.LockInvoice(ctx, invoiceID)
	if err != nil {
		return Invoice{}, err
	}
	if balance.Status != "issued" {
		return Invoice{}, ErrInvoiceNotPayable
	}
	if balance.Paid > balance.Total || payment.Amount > balance.Total-balance.Paid {
		return Invoice{}, ErrPaymentExceedsBalance
	}
	if _, err := tx.CreatePayment(ctx, invoiceID, payment); err != nil {
		return Invoice{}, err
	}
	status := "issued"
	if balance.Paid+payment.Amount == balance.Total {
		status = "paid"
	}
	if err := tx.SetInvoiceStatus(ctx, invoiceID, status); err != nil {
		return Invoice{}, err
	}
	updated, err := tx.LoadInvoice(ctx, invoiceID)
	if err != nil {
		return Invoice{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Invoice{}, err
	}
	return updated, nil
}

func (s *Service) ListInvoicesByPatient(ctx context.Context, principal authn.Principal, patientID string) ([]Invoice, error) {
	if err := authn.RequireAnyRole(principal,
		authn.RoleAdminTI, authn.RoleReceptionistCashier, authn.RoleDentist, authn.RoleAuditorAccountant,
	); err != nil {
		return nil, err
	}
	if !billingUUIDPattern.MatchString(patientID) {
		return nil, ErrInvalidInvoice
	}
	return s.repo.ListInvoicesByPatient(ctx, patientID)
}

func (s *Service) CloseCashRegister(ctx context.Context, principal authn.Principal, request cashCloseRequest) (CashClose, error) {
	if err := authn.RequireAnyRole(principal, authn.RoleReceptionistCashier); err != nil {
		return CashClose{}, err
	}
	if !billingUUIDPattern.MatchString(principal.UserID) || request.ReportedCash < 0 || request.ReportedTransfers < 0 {
		return CashClose{}, ErrInvalidCashClose
	}
	now := s.now()
	localStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	nextLocalStart := localStart.AddDate(0, 0, 1)
	close := CashClose{
		UserID: principal.UserID, ClosureDate: localStart.Format("2006-01-02"),
		ReportedCash: request.ReportedCash, ReportedTransfers: request.ReportedTransfers,
	}
	tx, err := s.repo.BeginCashCloseTx(ctx)
	if err != nil {
		return CashClose{}, err
	}
	defer tx.Rollback(ctx)
	totals, err := tx.SystemTotals(ctx, localStart.UTC(), nextLocalStart.UTC())
	if err != nil {
		return CashClose{}, err
	}
	close.SystemCash = totals.Cash
	close.SystemTransfers = totals.Transfers
	close.Status = "matched"
	if close.ReportedCash != close.SystemCash || close.ReportedTransfers != close.SystemTransfers {
		close.Status = "discrepant"
	}
	saved, err := tx.SaveCashClose(ctx, close)
	if err != nil {
		return CashClose{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return CashClose{}, fmt.Errorf("commit cash close: %w", err)
	}
	return saved, nil
}

func requireAdminAccess(principal authn.Principal) error {
	return authn.RequireAnyRole(principal, authn.RoleAdminTI)
}

func requireBillingAccess(principal authn.Principal) error {
	return authn.RequireAnyRole(principal,
		authn.RoleAdminTI,
		authn.RoleReceptionistCashier,
		authn.RoleDentist,
		authn.RoleAuditorAccountant,
	)
}

func requirePaymentAccess(principal authn.Principal) error {
	return authn.RequireAnyRole(principal, authn.RoleAdminTI, authn.RoleReceptionistCashier)
}
