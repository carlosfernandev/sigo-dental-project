// Package billing owns medical catalog, quotes, invoices, and payments.
package billing

import (
	"strings"
	"time"
)

func ValidateProcedure(p Procedure) error {
	p.Code = strings.TrimSpace(p.Code)
	p.Name = strings.TrimSpace(p.Name)
	p.Description = strings.TrimSpace(p.Description)
	p.Category = strings.TrimSpace(p.Category)
	if p.Code == "" || len(p.Code) > 50 || p.Name == "" || len(p.Name) > 200 || len(p.Description) > 500 || len(p.Category) > 100 {
		return ErrInvalidProcedure
	}
	if p.BasePrice < 0 || p.BasePrice > 100000000 {
		return ErrInvalidProcedure
	}
	return nil
}

func ValidateQuote(q Quote, now time.Time) error {
	if q.PatientID == "" {
		return ErrInvalidQuote
	}
	if len(q.Items) == 0 {
		return ErrInvalidQuote
	}
	for _, item := range q.Items {
		if item.ProcedureID == "" || item.ProcedureName == "" || item.Quantity <= 0 || item.UnitPrice < 0 || item.Discount < 0 {
			return ErrInvalidQuote
		}
		if item.Quantity > 100 {
			return ErrInvalidQuote
		}
	}
	if q.Subtotal < 0 || q.Discount < 0 || q.Total < 0 {
		return ErrInvalidQuote
	}
	validStatuses := map[string]bool{"draft": true, "approved": true, "rejected": true}
	if !validStatuses[q.Status] {
		return ErrInvalidQuote
	}
	return nil
}

func ValidateInvoice(i Invoice, now time.Time) error {
	if i.PatientID == "" {
		return ErrInvalidInvoice
	}
	if len(i.Items) == 0 {
		return ErrInvalidInvoice
	}
	for _, item := range i.Items {
		if item.ProcedureID == "" || item.ProcedureName == "" || item.Quantity <= 0 || item.UnitPrice < 0 {
			return ErrInvalidInvoice
		}
	}
	if i.Subtotal < 0 || i.Tax < 0 || i.Total < 0 {
		return ErrInvalidInvoice
	}
	for _, payment := range i.Payments {
		if payment.Amount < 0 {
			return ErrInvalidInvoice
		}
		validMethods := map[string]bool{"cash": true, "card": true, "transfer": true, "insurance": true}
		if !validMethods[payment.Method] {
			return ErrInvalidInvoice
		}
	}
	validStatuses := map[string]bool{"draft": true, "issued": true, "paid": true, "cancelled": true}
	if !validStatuses[i.Status] {
		return ErrInvalidInvoice
	}
	return nil
}
