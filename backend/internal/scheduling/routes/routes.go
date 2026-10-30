// Package routes registers HTTP endpoints owned by the scheduling domain.
package routes

import "github.com/gofiber/fiber/v3"

type handler interface {
	// Operatory
	CreateOperatory(c fiber.Ctx) error
	ListOperatories(c fiber.Ctx) error
	GetOperatory(c fiber.Ctx) error
	UpdateOperatory(c fiber.Ctx) error

	// Chair
	CreateChair(c fiber.Ctx) error
	ListChairs(c fiber.Ctx) error
	GetChair(c fiber.Ctx) error
	UpdateChair(c fiber.Ctx) error

	// Dentist
	CreateDentist(c fiber.Ctx) error
	ListDentists(c fiber.Ctx) error
	GetDentist(c fiber.Ctx) error
	UpdateDentist(c fiber.Ctx) error

	// Appointment
	CreateAppointment(c fiber.Ctx) error
	ListAppointments(c fiber.Ctx) error
	GetAppointment(c fiber.Ctx) error
	UpdateAppointment(c fiber.Ctx) error
	ListPendingReminders(c fiber.Ctx) error
	UpdateReminder(c fiber.Ctx) error
}

func Register(router fiber.Router, h handler, authMiddleware fiber.Handler) {
	scheduling := router.Group("/scheduling", authMiddleware)

	// Operatories
	operatories := scheduling.Group("/operatories")
	operatories.Post("/", h.CreateOperatory)
	operatories.Get("/", h.ListOperatories)
	operatories.Get("/:id", h.GetOperatory)
	operatories.Put("/:id", h.UpdateOperatory)

	// Chairs
	chairs := scheduling.Group("/chairs")
	chairs.Post("/", h.CreateChair)
	chairs.Get("/", h.ListChairs)
	chairs.Get("/:id", h.GetChair)
	chairs.Put("/:id", h.UpdateChair)

	// Dentists
	dentists := scheduling.Group("/dentists")
	dentists.Post("/", h.CreateDentist)
	dentists.Get("/", h.ListDentists)
	dentists.Get("/:id", h.GetDentist)
	dentists.Put("/:id", h.UpdateDentist)

	// Appointments
	appointments := scheduling.Group("/appointments")
	appointments.Post("/", h.CreateAppointment)
	appointments.Get("/", h.ListAppointments)
	appointments.Get("/:id", h.GetAppointment)
	appointments.Put("/:id", h.UpdateAppointment)

	reminders := router.Group("/reminders", authMiddleware)
	reminders.Get("/pending", h.ListPendingReminders)
	reminders.Patch("/:id", h.UpdateReminder)
}
