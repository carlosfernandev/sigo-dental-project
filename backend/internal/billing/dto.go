// Package billing owns medical catalog, quotes, invoices, and payments.
package billing

import "time"

// Procedure DTOs
type procedureRequest struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
	BasePrice   int64  `json:"base_price"`
	Active      bool   `json:"active"`
}

type procedureResponse struct {
	ID          string    `json:"id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	BasePrice   int64     `json:"base_price"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Quote DTOs
type quoteItemRequest struct {
	ProcedureID   string `json:"procedure_id"`
	ProcedureName string `json:"procedure_name"`
	Quantity      int    `json:"quantity"`
	UnitPrice     int64  `json:"unit_price"`
	Discount      int64  `json:"discount"`
	Total         int64  `json:"total"`
}

type quoteRequest struct {
	PatientID string             `json:"patient_id"`
	Items     []quoteItemRequest `json:"items"`
	Subtotal  int64              `json:"subtotal"`
	Discount  int64              `json:"discount"`
	Total     int64              `json:"total"`
	Status    string             `json:"status"`
}

type quoteResponse struct {
	ID         string      `json:"id"`
	PatientID  string      `json:"patient_id"`
	Items      []QuoteItem `json:"items"`
	Subtotal   int64       `json:"subtotal"`
	Discount   int64       `json:"discount"`
	Total      int64       `json:"total"`
	Status     string      `json:"status"`
	ApprovedAt *time.Time  `json:"approved_at,omitempty"`
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
}

// Invoice DTOs
type paymentRequest struct {
	Method string `json:"method"`
	Amount int64  `json:"amount"`
}

type invoiceItemRequest struct {
	ProcedureID   string `json:"procedure_id"`
	ProcedureName string `json:"procedure_name"`
	Quantity      int    `json:"quantity"`
	UnitPrice     int64  `json:"unit_price"`
	Total         int64  `json:"total"`
}

type invoiceRequest struct {
	PatientID string               `json:"patient_id"`
	QuoteID   *string              `json:"quote_id"`
	Items     []invoiceItemRequest `json:"items"`
	Subtotal  int64                `json:"subtotal"`
	Tax       int64                `json:"tax"`
	Total     int64                `json:"total"`
	Payments  []paymentRequest     `json:"payments"`
	Status    string               `json:"status"`
}

type invoiceResponse struct {
	ID        string        `json:"id"`
	PatientID string        `json:"patient_id"`
	QuoteID   *string       `json:"quote_id,omitempty"`
	Items     []InvoiceItem `json:"items"`
	Subtotal  int64         `json:"subtotal"`
	Tax       int64         `json:"tax"`
	Total     int64         `json:"total"`
	Payments  []Payment     `json:"payments"`
	Status    string        `json:"status"`
	IssuedAt  *time.Time    `json:"issued_at,omitempty"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}

// Request conversion methods
func (r procedureRequest) entity() Procedure {
	return Procedure{
		Code:        r.Code,
		Name:        r.Name,
		Description: r.Description,
		Category:    r.Category,
		BasePrice:   r.BasePrice,
		Active:      r.Active,
	}
}

func (r quoteRequest) entity() Quote {
	items := make([]QuoteItem, len(r.Items))
	for i, item := range r.Items {
		items[i] = QuoteItem{
			ProcedureID:   item.ProcedureID,
			ProcedureName: item.ProcedureName,
			Quantity:      item.Quantity,
			UnitPrice:     item.UnitPrice,
			Discount:      item.Discount,
			Total:         item.Total,
		}
	}
	return Quote{
		PatientID: r.PatientID,
		Items:     items,
		Subtotal:  r.Subtotal,
		Discount:  r.Discount,
		Total:     r.Total,
		Status:    r.Status,
	}
}

func (r invoiceRequest) entity() Invoice {
	items := make([]InvoiceItem, len(r.Items))
	for i, item := range r.Items {
		items[i] = InvoiceItem{
			ProcedureID:   item.ProcedureID,
			ProcedureName: item.ProcedureName,
			Quantity:      item.Quantity,
			UnitPrice:     item.UnitPrice,
			Total:         item.Total,
		}
	}
	payments := make([]Payment, len(r.Payments))
	for i, payment := range r.Payments {
		payments[i] = Payment{
			Method: payment.Method,
			Amount: payment.Amount,
		}
	}
	return Invoice{
		PatientID: r.PatientID,
		QuoteID:   r.QuoteID,
		Items:     items,
		Subtotal:  r.Subtotal,
		Tax:       r.Tax,
		Total:     r.Total,
		Payments:  payments,
		Status:    r.Status,
	}
}

// Response conversion methods
func procedureResponseFrom(p Procedure) procedureResponse {
	return procedureResponse{
		ID:          p.ID,
		Code:        p.Code,
		Name:        p.Name,
		Description: p.Description,
		Category:    p.Category,
		BasePrice:   p.BasePrice,
		Active:      p.Active,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

func quoteResponseFrom(q Quote) quoteResponse {
	return quoteResponse{
		ID:         q.ID,
		PatientID:  q.PatientID,
		Items:      q.Items,
		Subtotal:   q.Subtotal,
		Discount:   q.Discount,
		Total:      q.Total,
		Status:     q.Status,
		ApprovedAt: q.ApprovedAt,
		CreatedAt:  q.CreatedAt,
		UpdatedAt:  q.UpdatedAt,
	}
}

func invoiceResponseFrom(i Invoice) invoiceResponse {
	return invoiceResponse{
		ID:        i.ID,
		PatientID: i.PatientID,
		QuoteID:   i.QuoteID,
		Items:     i.Items,
		Subtotal:  i.Subtotal,
		Tax:       i.Tax,
		Total:     i.Total,
		Payments:  i.Payments,
		Status:    i.Status,
		IssuedAt:  i.IssuedAt,
		CreatedAt: i.CreatedAt,
		UpdatedAt: i.UpdatedAt,
	}
}
