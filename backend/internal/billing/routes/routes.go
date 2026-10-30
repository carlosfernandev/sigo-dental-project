// Package routes registers HTTP endpoints owned by the billing domain.
package routes

import "github.com/gofiber/fiber/v3"

type handler interface {
	// Procedure
	CreateProcedure(c fiber.Ctx) error
	ListProcedures(c fiber.Ctx) error
	GetProcedure(c fiber.Ctx) error
	UpdateProcedure(c fiber.Ctx) error

	// Quote
	CreateQuote(c fiber.Ctx) error
	ListQuotes(c fiber.Ctx) error
	GetQuote(c fiber.Ctx) error
	UpdateQuote(c fiber.Ctx) error

	CreateInvoice(c fiber.Ctx) error
	RegisterPayment(c fiber.Ctx) error
	ListInvoicesByPatient(c fiber.Ctx) error
	CloseCashRegister(c fiber.Ctx) error
}

func Register(router fiber.Router, h handler, authMiddleware fiber.Handler) {
	billing := router.Group("/billing", authMiddleware)

	// Procedures
	procedures := billing.Group("/procedures")
	procedures.Post("/", h.CreateProcedure)
	procedures.Get("/", h.ListProcedures)
	procedures.Get("/:id", h.GetProcedure)
	procedures.Put("/:id", h.UpdateProcedure)

	// Quotes
	quotes := billing.Group("/quotes")
	quotes.Post("/", h.CreateQuote)
	quotes.Get("/", h.ListQuotes)
	quotes.Get("/:id", h.GetQuote)
	quotes.Put("/:id", h.UpdateQuote)

	// Invoices and mixed payments (RF-12)
	invoices := router.Group("/api/v1/invoices", authMiddleware)
	invoices.Post("/", h.CreateInvoice)
	invoices.Patch("/:id/pay", h.RegisterPayment)
	invoices.Get("/patient/:patient_id", h.ListInvoicesByPatient)

	cashRegister := router.Group("/cash-register", authMiddleware)
	cashRegister.Post("/close", h.CloseCashRegister)
}
