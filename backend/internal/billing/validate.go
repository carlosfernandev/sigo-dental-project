// Package billing owns medical catalog, quotes, invoices, and payments.
package billing

import (
	"math"
	"regexp"
	"strings"
	"time"
)

var billingUUIDPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

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
	validStatuses := map[string]bool{"draft": true, "approved": true, "partially_approved": true, "rejected": true}
	if !validStatuses[q.Status] {
		return ErrInvalidQuote
	}
	return nil
}

func ValidateInvoice(i Invoice, now time.Time) error {
	if !billingUUIDPattern.MatchString(i.PatientID) {
		return ErrInvalidInvoice
	}
	if i.QuoteID != nil && !billingUUIDPattern.MatchString(*i.QuoteID) {
		return ErrInvalidInvoice
	}
	if len(i.Items) == 0 {
		return ErrInvalidInvoice
	}
	var subtotal int64
	for index := range i.Items {
		item := &i.Items[index]
		item.Description = strings.TrimSpace(item.Description)
		if item.Description == "" || len(item.Description) > 500 || item.Quantity <= 0 || item.Quantity > 100 || item.UnitPrice <= 0 {
			return ErrInvalidInvoice
		}
		if int64(item.Quantity) > math.MaxInt64/item.UnitPrice {
			return ErrInvalidInvoice
		}
		item.Subtotal = item.UnitPrice * int64(item.Quantity)
		if subtotal > math.MaxInt64-item.Subtotal {
			return ErrInvalidInvoice
		}
		subtotal += item.Subtotal
	}
	if i.Tax < 0 || subtotal > math.MaxInt64-i.Tax {
		return ErrInvalidInvoice
	}
	i.Subtotal = subtotal
	i.Total = subtotal + i.Tax
	if i.Total <= 0 {
		return ErrInvalidInvoice
	}
	return nil
}

func ValidatePayment(payment Payment) error {
	if payment.Amount <= 0 {
		return ErrInvalidInvoice
	}
	switch payment.Method {
	case "cash":
	case "card", "transfer", "insurance":
		if payment.Reference == nil || strings.TrimSpace(*payment.Reference) == "" || len(strings.TrimSpace(*payment.Reference)) > 150 {
			return ErrInvalidInvoice
		}
	default:
		return ErrInvalidInvoice
	}
	if payment.Reference != nil {
		reference := strings.TrimSpace(*payment.Reference)
		*payment.Reference = reference
	}
	return nil
}
