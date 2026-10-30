// Package routes registers the HTTP endpoints owned by the identity domain.
package routes

import "github.com/gofiber/fiber/v3"

type loginHandler interface {
	Login(c fiber.Ctx) error
	Me(c fiber.Ctx) error
}

type userAdminHandler interface {
	ListUsers(c fiber.Ctx) error
}

// Register mounts identity endpoints on the shared monolith router.
func Register(router fiber.Router, handler loginHandler, users userAdminHandler, authMiddleware fiber.Handler) {
	router.Post("/auth/login", handler.Login)
	protected := router.Group("/auth", authMiddleware)
	protected.Get("/me", handler.Me)
	router.Get("/users", authMiddleware, users.ListUsers)
}
