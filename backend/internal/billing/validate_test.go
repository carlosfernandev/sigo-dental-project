// Package billing owns medical catalog, quotes, invoices, and payments.
package billing

import (
	"errors"
	"math"
	"testing"
	"time"
)

func TestValidateProcedure(t *testing.T) {
	tests := []struct {
		name    string
		p       Procedure
		wantErr error
	}{
		{"valid", Procedure{Code: "PROC001", Name: "Limpieza Dental", BasePrice: 5000}, nil},
		{"empty code", Procedure{Code: "", Name: "Test", BasePrice: 1000}, ErrInvalidProcedure},
		{"empty name", Procedure{Code: "PROC001", Name: "", BasePrice: 1000}, ErrInvalidProcedure},
		{"negative price", Procedure{Code: "PROC001", Name: "Test", BasePrice: -100}, ErrInvalidProcedure},
		{"price too high", Procedure{Code: "PROC001", Name: "Test", BasePrice: 100000001}, ErrInvalidProcedure},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateProcedure(tt.p)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("ValidateProcedure() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateInvoice(t *testing.T) {
	now := time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC)
	patientID := "123e4567-e89b-12d3-a456-426614174000"
	tests := []struct {
		name    string
		invoice Invoice
		wantErr error
	}{
		{"valid cents", Invoice{PatientID: patientID, Items: []InvoiceItem{{Description: "Cleaning", Quantity: 2, UnitPrice: 1250}}, Tax: 300}, nil},
		{"invalid patient id", Invoice{PatientID: "patient", Items: []InvoiceItem{{Description: "Cleaning", Quantity: 1, UnitPrice: 100}}}, ErrInvalidInvoice},
		{"empty description", Invoice{PatientID: patientID, Items: []InvoiceItem{{Quantity: 1, UnitPrice: 100}}}, ErrInvalidInvoice},
		{"nonpositive amount", Invoice{PatientID: patientID, Items: []InvoiceItem{{Description: "Cleaning", Quantity: 1, UnitPrice: 0}}}, ErrInvalidInvoice},
		{"multiplication overflow", Invoice{PatientID: patientID, Items: []InvoiceItem{{Description: "Cleaning", Quantity: 2, UnitPrice: math.MaxInt64}}}, ErrInvalidInvoice},
		{"total overflow", Invoice{PatientID: patientID, Items: []InvoiceItem{{Description: "Cleaning", Quantity: 1, UnitPrice: math.MaxInt64}}, Tax: 1}, ErrInvalidInvoice},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateInvoice(tt.invoice, now); !errors.Is(err, tt.wantErr) {
				t.Fatalf("ValidateInvoice() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidatePayment(t *testing.T) {
	reference := "  voucher-123  "
	tests := []struct {
		name    string
		payment Payment
		wantErr error
	}{
		{"cash without reference", Payment{Method: "cash", Amount: 100}, nil},
		{"card with reference", Payment{Method: "card", Amount: 100, Reference: &reference}, nil},
		{"zero amount", Payment{Method: "cash"}, ErrInvalidInvoice},
		{"transfer requires reference", Payment{Method: "transfer", Amount: 100}, ErrInvalidInvoice},
		{"unsupported method", Payment{Method: "crypto", Amount: 100}, ErrInvalidInvoice},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidatePayment(tt.payment); !errors.Is(err, tt.wantErr) {
				t.Fatalf("ValidatePayment() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateQuote(t *testing.T) {
	now := time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name    string
		q       Quote
		wantErr error
	}{
		{"valid", Quote{
			PatientID: "patient-1",
			Items:     []QuoteItem{{ProcedureID: "proc-1", ProcedureName: "Test", Quantity: 1, UnitPrice: 5000, Total: 5000}},
			Subtotal:  5000, Total: 5000, Status: "draft",
		}, nil},
		{"empty patient", Quote{Items: []QuoteItem{{ProcedureID: "proc-1", ProcedureName: "Test", Quantity: 1, UnitPrice: 5000}}}, ErrInvalidQuote},
		{"no items", Quote{PatientID: "patient-1", Items: []QuoteItem{}}, ErrInvalidQuote},
		{"invalid status", Quote{PatientID: "patient-1", Items: []QuoteItem{{ProcedureID: "proc-1", ProcedureName: "Test", Quantity: 1, UnitPrice: 5000}}, Status: "invalid"}, ErrInvalidQuote},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateQuote(tt.q, now)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("ValidateQuote() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
