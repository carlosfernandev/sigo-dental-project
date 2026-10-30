package billing

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/carlosfpatino/sigo-dental/platform/authn"
)

type cashCloseRepository struct {
	Repository
	tx *cashCloseTransaction
}

func (r cashCloseRepository) BeginCashCloseTx(context.Context) (CashCloseTx, error) {
	return r.tx, nil
}

type cashCloseTransaction struct {
	totals CashTotals
	saved  CashClose
}

func (tx *cashCloseTransaction) SystemTotals(context.Context, time.Time, time.Time) (CashTotals, error) {
	return tx.totals, nil
}

func (tx *cashCloseTransaction) SaveCashClose(_ context.Context, close CashClose) (CashClose, error) {
	tx.saved = close
	close.ID = "closure-id"
	close.CreatedAt = time.Date(2026, 10, 30, 15, 0, 0, 0, time.UTC)
	return close, nil
}

func (tx *cashCloseTransaction) Commit(context.Context) error   { return nil }
func (tx *cashCloseTransaction) Rollback(context.Context) error { return nil }

func TestCloseCashRegister(t *testing.T) {
	now := time.Date(2026, 10, 30, 12, 0, 0, 0, time.FixedZone("America/Guayaquil", -5*60*60))
	tests := []struct {
		name       string
		reported   cashCloseRequest
		totals     CashTotals
		wantStatus string
	}{
		{"matched", cashCloseRequest{ReportedCash: 12500, ReportedTransfers: 8000}, CashTotals{Cash: 12500, Transfers: 8000}, "matched"},
		{"discrepant", cashCloseRequest{ReportedCash: 12000, ReportedTransfers: 8000}, CashTotals{Cash: 12500, Transfers: 8000}, "discrepant"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &cashCloseTransaction{totals: tt.totals}
			service := NewService(cashCloseRepository{tx: tx})
			service.now = func() time.Time { return now }
			close, err := service.CloseCashRegister(context.Background(), authn.Principal{
				UserID: "123e4567-e89b-12d3-a456-426614174000", Role: authn.RoleReceptionistCashier,
			}, tt.reported)
			if err != nil {
				t.Fatalf("CloseCashRegister() error = %v", err)
			}
			if close.Status != tt.wantStatus || close.ClosureDate != "2026-10-30" ||
				close.SystemCash != tt.totals.Cash || close.SystemTransfers != tt.totals.Transfers {
				t.Fatalf("CloseCashRegister() = %+v", close)
			}
			if tx.saved.UserID != "123e4567-e89b-12d3-a456-426614174000" {
				t.Fatalf("saved user ID = %q", tx.saved.UserID)
			}
		})
	}
}

func TestCloseCashRegisterRejectsInvalidAndUnauthorizedRequests(t *testing.T) {
	service := NewService(cashCloseRepository{tx: &cashCloseTransaction{}})
	principal := authn.Principal{UserID: "123e4567-e89b-12d3-a456-426614174000", Role: authn.RoleReceptionistCashier}
	if _, err := service.CloseCashRegister(context.Background(), principal, cashCloseRequest{ReportedCash: -1}); !errors.Is(err, ErrInvalidCashClose) {
		t.Fatalf("negative count error = %v, want %v", err, ErrInvalidCashClose)
	}
	principal.Role = authn.RoleDentist
	if _, err := service.CloseCashRegister(context.Background(), principal, cashCloseRequest{}); !errors.Is(err, authn.ErrForbidden) {
		t.Fatalf("unauthorized role error = %v, want %v", err, authn.ErrForbidden)
	}
}
