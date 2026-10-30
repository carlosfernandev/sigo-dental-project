// Package billing owns medical catalog, quotes, invoices, and payments.
package billing

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"strings"

	"github.com/carlosfpatino/sigo-dental/middleware"
	"github.com/carlosfpatino/sigo-dental/platform/authn"
	"github.com/carlosfpatino/sigo-dental/platform/httpx"
	"github.com/gofiber/fiber/v3"
)

const maxRequestBytes = 32 * 1024

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Procedure handlers
func (h *Handler) CreateProcedure(c fiber.Ctx) error {
	principal, ok := middleware.PrincipalFromContext(c)
	if !ok {
		return unauthorized()
	}
	var request procedureRequest
	if err := decodeRequest(c, &request); err != nil {
		return invalidRequest()
	}
	procedure := request.entity()
	created, err := h.service.CreateProcedure(c.Context(), principal, procedure)
	if err != nil {
		return billingError(err)
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": procedureResponseFrom(created)})
}

func (h *Handler) ListProcedures(c fiber.Ctx) error {
	principal, ok := middleware.PrincipalFromContext(c)
	if !ok {
		return unauthorized()
	}
	procedures, err := h.service.ListProcedures(c.Context(), principal)
	if err != nil {
		return billingError(err)
	}
	items := make([]procedureResponse, len(procedures))
	for i, p := range procedures {
		items[i] = procedureResponseFrom(p)
	}
	return c.JSON(fiber.Map{"data": items})
}

func (h *Handler) GetProcedure(c fiber.Ctx) error {
	principal, ok := middleware.PrincipalFromContext(c)
	if !ok {
		return unauthorized()
	}
	procedure, err := h.service.GetProcedure(c.Context(), principal, c.Params("id"))
	if err != nil {
		return billingError(err)
	}
	return c.JSON(fiber.Map{"data": procedureResponseFrom(procedure)})
}

func (h *Handler) UpdateProcedure(c fiber.Ctx) error {
	principal, ok := middleware.PrincipalFromContext(c)
	if !ok {
		return unauthorized()
	}
	var request procedureRequest
	if err := decodeRequest(c, &request); err != nil {
		return invalidRequest()
	}
	procedure := request.entity()
	procedure.ID = c.Params("id")
	updated, err := h.service.UpdateProcedure(c.Context(), principal, procedure)
	if err != nil {
		return billingError(err)
	}
	return c.JSON(fiber.Map{"data": procedureResponseFrom(updated)})
}

// Quote handlers
func (h *Handler) CreateQuote(c fiber.Ctx) error {
	principal, ok := middleware.PrincipalFromContext(c)
	if !ok {
		return unauthorized()
	}
	var request quoteRequest
	if err := decodeRequest(c, &request); err != nil {
		return invalidRequest()
	}
	quote := request.entity()
	created, err := h.service.CreateQuote(c.Context(), principal, quote)
	if err != nil {
		return billingError(err)
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": quoteResponseFrom(created)})
}

func (h *Handler) ListQuotes(c fiber.Ctx) error {
	principal, ok := middleware.PrincipalFromContext(c)
	if !ok {
		return unauthorized()
	}
	var patientID *string
	if v := c.Query("patient_id"); v != "" {
		patientID = &v
	}
	quotes, err := h.service.ListQuotes(c.Context(), principal, patientID)
	if err != nil {
		return billingError(err)
	}
	items := make([]quoteResponse, len(quotes))
	for i, q := range quotes {
		items[i] = quoteResponseFrom(q)
	}
	return c.JSON(fiber.Map{"data": items})
}

func (h *Handler) GetQuote(c fiber.Ctx) error {
	principal, ok := middleware.PrincipalFromContext(c)
	if !ok {
		return unauthorized()
	}
	quote, err := h.service.GetQuote(c.Context(), principal, c.Params("id"))
	if err != nil {
		return billingError(err)
	}
	return c.JSON(fiber.Map{"data": quoteResponseFrom(quote)})
}

func (h *Handler) UpdateQuote(c fiber.Ctx) error {
	principal, ok := middleware.PrincipalFromContext(c)
	if !ok {
		return unauthorized()
	}
	var request quoteRequest
	if err := decodeRequest(c, &request); err != nil {
		return invalidRequest()
	}
	quote := request.entity()
	quote.ID = c.Params("id")
	updated, err := h.service.UpdateQuote(c.Context(), principal, quote)
	if err != nil {
		return billingError(err)
	}
	return c.JSON(fiber.Map{"data": quoteResponseFrom(updated)})
}

// CreateInvoice records and issues an internal invoice. Monetary inputs and outputs are cents.
func (h *Handler) CreateInvoice(c fiber.Ctx) error {
	principal, ok := middleware.PrincipalFromContext(c)
	if !ok {
		return unauthorized()
	}
	var request invoiceRequest
	if err := decodeRequest(c, &request); err != nil {
		return invalidRequest()
	}
	created, err := h.service.CreateInvoice(c.Context(), principal, request.entity())
	if err != nil {
		return billingError(err)
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": invoiceResponseFrom(created)})
}

// RegisterPayment records one component of a payment; repeated calls support mixed payments.
func (h *Handler) RegisterPayment(c fiber.Ctx) error {
	principal, ok := middleware.PrincipalFromContext(c)
	if !ok {
		return unauthorized()
	}
	var request paymentRequest
	if err := decodeRequest(c, &request); err != nil {
		return invalidRequest()
	}
	invoice, err := h.service.RegisterPayment(c.Context(), principal, c.Params("id"), Payment{
		Method: request.Method, Amount: request.Amount, Reference: request.Reference,
	})
	if err != nil {
		return billingError(err)
	}
	return c.JSON(fiber.Map{"data": invoiceResponseFrom(invoice)})
}

func (h *Handler) ListInvoicesByPatient(c fiber.Ctx) error {
	principal, ok := middleware.PrincipalFromContext(c)
	if !ok {
		return unauthorized()
	}
	invoices, err := h.service.ListInvoicesByPatient(c.Context(), principal, c.Params("patient_id"))
	if err != nil {
		return billingError(err)
	}
	responses := make([]invoiceResponse, len(invoices))
	for i, invoice := range invoices {
		responses[i] = invoiceResponseFrom(invoice)
	}
	return c.JSON(fiber.Map{"data": responses})
}

func decodeRequest(c fiber.Ctx, target any) error {
	contentType, _, err := mime.ParseMediaType(c.Get(fiber.HeaderContentType))
	if err != nil || !strings.EqualFold(contentType, fiber.MIMEApplicationJSON) || len(c.Body()) == 0 || len(c.Body()) > maxRequestBytes {
		return ErrInvalidQuote
	}
	decoder := json.NewDecoder(bytes.NewReader(c.Body()))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return ErrInvalidQuote
	}
	return nil
}

func billingError(err error) error {
	switch {
	case errors.Is(err, authn.ErrForbidden):
		return httpx.NewError(fiber.StatusForbidden, "forbidden", "Permiso insuficiente", err)
	case errors.Is(err, ErrInvalidProcedure), errors.Is(err, ErrInvalidQuote), errors.Is(err, ErrInvalidInvoice):
		return invalidRequest()
	case errors.Is(err, ErrProcedureNotFound):
		return httpx.NewError(fiber.StatusNotFound, "procedure_not_found", "Procedimiento no encontrado", err)
	case errors.Is(err, ErrQuoteNotFound):
		return httpx.NewError(fiber.StatusNotFound, "quote_not_found", "Presupuesto no encontrado", err)
	case errors.Is(err, ErrInvoiceNotFound):
		return httpx.NewError(fiber.StatusNotFound, "invoice_not_found", "Factura no encontrada", err)
	case errors.Is(err, ErrPaymentExceedsBalance):
		return httpx.NewError(fiber.StatusConflict, "payment_exceeds_balance", "El pago supera el saldo pendiente", err)
	case errors.Is(err, ErrInvoiceNotPayable):
		return httpx.NewError(fiber.StatusConflict, "invoice_not_payable", "La factura no acepta pagos", err)
	default:
		return httpx.NewError(fiber.StatusInternalServerError, "internal_error", "Error interno del servidor", err)
	}
}

func unauthorized() error {
	return httpx.NewError(fiber.StatusUnauthorized, "unauthorized", "Autenticación requerida", nil)
}

func invalidRequest() error {
	return httpx.NewError(fiber.StatusBadRequest, "invalid_request", "Solicitud inválida", nil)
}
