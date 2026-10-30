// Package billing owns medical catalog, quotes, invoices, and payments.
package billing

import (
	"errors"
	"time"
)

var (
	ErrInvalidProcedure  = errors.New("invalid procedure")
	ErrProcedureNotFound = errors.New("procedure not found")
	ErrInvalidQuote      = errors.New("invalid quote")
	ErrQuoteNotFound     = errors.New("quote not found")
	ErrInvalidInvoice    = errors.New("invalid invoice")
	ErrInvoiceNotFound   = errors.New("invoice not found")
)

type Procedure struct {
	ID          string
	Code        string
	Name        string
	Description string
	Category    string
	BasePrice   int64 // cents
	Active      bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type QuoteItem struct {
	ProcedureID   string
	ProcedureName string
	Quantity      int
	UnitPrice     int64 // cents
	Discount      int64 // cents
	Total         int64 // cents
}

type Quote struct {
	ID         string
	PatientID  string
	Items      []QuoteItem
	Subtotal   int64  // cents
	Discount   int64  // cents
	Total      int64  // cents
	Status     string // draft, approved, rejected
	ApprovedAt *time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type InvoiceItem struct {
	ProcedureID   string
	ProcedureName string
	Quantity      int
	UnitPrice     int64 // cents
	Total         int64 // cents
}

type Payment struct {
	Method string // cash, card, transfer, insurance
	Amount int64  // cents
}

type Invoice struct {
	ID        string
	PatientID string
	QuoteID   *string
	Items     []InvoiceItem
	Subtotal  int64 // cents
	Tax       int64 // cents
	Total     int64 // cents
	Payments  []Payment
	Status    string // draft, issued, paid, cancelled
	IssuedAt  *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}
