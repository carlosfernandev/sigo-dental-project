package billing

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"os"
	"testing"

	"github.com/carlosfpatino/sigo-dental/internal/scheduling"
	"github.com/carlosfpatino/sigo-dental/platform/authn"
	"github.com/carlosfpatino/sigo-dental/platform/migrate"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestInvoiceRepositoryPostgresIntegration(t *testing.T) {
	dsn := os.Getenv("SIGO_TEST_PG_DSN")
	if dsn == "" {
		t.Skip("SIGO_TEST_PG_DSN is not set")
	}
	ctx := context.Background()
	adminPool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("connect to integration database: %v", err)
	}
	t.Cleanup(adminPool.Close)

	var random [12]byte
	if _, err := rand.Read(random[:]); err != nil {
		t.Fatalf("generate isolated schema name: %v", err)
	}
	schema := "billing_test_" + hex.EncodeToString(random[:])
	if _, err := adminPool.Exec(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatalf("create isolated schema: %v", err)
	}
	t.Cleanup(func() {
		if _, err := adminPool.Exec(context.Background(), "DROP SCHEMA "+schema+" CASCADE"); err != nil {
			t.Errorf("drop isolated schema: %v", err)
		}
	})
	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatalf("parse integration DSN: %v", err)
	}
	poolConfig.ConnConfig.RuntimeParams["search_path"] = schema
	db, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		t.Fatalf("connect to isolated schema: %v", err)
	}
	t.Cleanup(db.Close)
	if err := migrate.Up(ctx, db, os.DirFS("../../migrations")); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}
	var patientID string
	err = db.QueryRow(ctx, `
		INSERT INTO patients (first_name, last_name, date_of_birth, phone, emergency_contact_name, emergency_contact_phone)
		VALUES ('Ada', 'Lovelace', '1990-01-01', '+593991234567', 'Grace Lovelace', '+593991234568')
		RETURNING id::text
	`).Scan(&patientID)
	if err != nil {
		t.Fatalf("insert patient: %v", err)
	}
	var cashierID string
	err = db.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, role_id)
		SELECT 'cashier-rn02@example.com', 'test-hash', id FROM roles WHERE name = $1
		RETURNING id::text
	`, authn.RoleReceptionistCashier).Scan(&cashierID)
	if err != nil {
		t.Fatalf("insert cashier: %v", err)
	}

	repository := NewPostgresRepository(db)
	schedulingService := scheduling.NewService(scheduling.NewPostgresRepository(db))
	service := NewService(repository, schedulingService)
	principal := authn.Principal{UserID: cashierID, Role: authn.RoleReceptionistCashier}
	quote, err := service.CreateQuote(ctx, principal, Quote{
		PatientID: patientID,
		Items: []QuoteItem{
			{ProcedureID: "procedure-1", ProcedureName: "Cleaning", Quantity: 1, UnitPrice: 2500, Total: 2500},
			{ProcedureID: "procedure-2", ProcedureName: "Filling", Quantity: 1, UnitPrice: 5000, Total: 5000},
		},
		Subtotal: 7500, Total: 7500,
	})
	if err != nil {
		t.Fatalf("CreateQuote() error = %v", err)
	}
	partialQuote, err := service.ApproveQuote(ctx, principal, quote.ID, ApproveQuoteRequest{ApprovedItemIDs: []string{quote.Items[0].ID}})
	if err != nil || partialQuote.Status != "partially_approved" || partialQuote.Items[0].Status != "approved" || partialQuote.Items[1].Status != "pending" {
		t.Fatalf("partial ApproveQuote() = %+v, error %v", partialQuote, err)
	}
	var taskCount int
	if err := db.QueryRow(ctx, `SELECT count(*) FROM schedulable_tasks WHERE quote_id=$1::uuid`, quote.ID).Scan(&taskCount); err != nil || taskCount != 1 {
		t.Fatalf("partial task count = %d, error %v", taskCount, err)
	}
	approved, err := service.ApproveQuote(ctx, principal, quote.ID, ApproveQuoteRequest{ApprovedItemIDs: []string{quote.Items[1].ID}})
	if err != nil || approved.Status != "approved" {
		t.Fatalf("final ApproveQuote() = status %q, error %v", approved.Status, err)
	}
	if err := db.QueryRow(ctx, `SELECT count(*) FROM schedulable_tasks WHERE quote_id=$1::uuid`, quote.ID).Scan(&taskCount); err != nil || taskCount != 2 {
		t.Fatalf("final task count = %d, error %v", taskCount, err)
	}
	invoice, err := service.CreateInvoice(ctx, principal, Invoice{
		PatientID: patientID,
		Items:     []InvoiceItem{{Description: "Dental cleaning", Quantity: 2, UnitPrice: 2500}},
		Tax:       500,
	})
	if err != nil {
		t.Fatalf("CreateInvoice() error = %v", err)
	}
	reference := "transfer-123"
	partial, err := service.RegisterPayment(ctx, principal, invoice.ID, Payment{Method: "cash", Amount: 1500})
	if err != nil || partial.Status != "issued" || len(partial.Payments) != 1 {
		t.Fatalf("partial RegisterPayment() = status %q, payments %d, error %v", partial.Status, len(partial.Payments), err)
	}
	if _, err := service.RegisterPayment(ctx, principal, invoice.ID, Payment{Method: "transfer", Amount: 5000, Reference: &reference}); !errors.Is(err, ErrPaymentExceedsBalance) {
		t.Fatalf("overpayment error = %v, want %v", err, ErrPaymentExceedsBalance)
	}
	paid, err := service.RegisterPayment(ctx, principal, invoice.ID, Payment{Method: "transfer", Amount: 4000, Reference: &reference})
	if err != nil || paid.Status != "paid" || len(paid.Payments) != 2 {
		t.Fatalf("final RegisterPayment() = status %q, payments %d, error %v", paid.Status, len(paid.Payments), err)
	}
	listed, err := service.ListInvoicesByPatient(ctx, principal, patientID)
	foundReference := false
	if len(listed) == 1 {
		for _, payment := range listed[0].Payments {
			if payment.Reference != nil && *payment.Reference == reference {
				foundReference = true
			}
		}
	}
	if err != nil || len(listed) != 1 || listed[0].Items[0].Subtotal != 5000 || !foundReference {
		t.Fatalf("ListInvoicesByPatient() = %+v, error %v", listed, err)
	}
	close, err := service.CloseCashRegister(ctx, principal, cashCloseRequest{ReportedCash: 1500, ReportedTransfers: 4000})
	if err != nil || close.Status != "matched" || close.SystemCash != 1500 || close.SystemTransfers != 4000 {
		t.Fatalf("CloseCashRegister() = %+v, error %v", close, err)
	}
	if _, err := service.CloseCashRegister(ctx, principal, cashCloseRequest{}); !errors.Is(err, ErrCashCloseExists) {
		t.Fatalf("duplicate cash close error = %v, want %v", err, ErrCashCloseExists)
	}
}
