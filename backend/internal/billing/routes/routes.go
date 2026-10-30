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
}
