// Package routes registers HTTP endpoints owned by the patient domain.
package routes

import "github.com/gofiber/fiber/v3"

type handler interface {
	Create(c fiber.Ctx) error
	List(c fiber.Ctx) error
	Get(c fiber.Ctx) error
	Update(c fiber.Ctx) error
}

func Register(router fiber.Router, patientHandler handler, authMiddleware fiber.Handler) {
	patients := router.Group("/patients", authMiddleware)
	patients.Post("/", patientHandler.Create)
	patients.Get("/", patientHandler.List)
	patients.Get("/:id", patientHandler.Get)
	patients.Put("/:id", patientHandler.Update)
}
