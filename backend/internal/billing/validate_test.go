// Package billing owns medical catalog, quotes, invoices, and payments.
package billing

import (
	"errors"
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
