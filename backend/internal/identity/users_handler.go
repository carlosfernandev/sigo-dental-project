package identity

import (
	"errors"

	"github.com/carlosfpatino/sigo-dental/middleware"
	"github.com/carlosfpatino/sigo-dental/platform/authn"
	"github.com/carlosfpatino/sigo-dental/platform/httpx"
	"github.com/gofiber/fiber/v3"
)

type UserAdminHandler struct {
	service *UserAdminService
}

func NewUserAdminHandler(service *UserAdminService) *UserAdminHandler {
	return &UserAdminHandler{service: service}
}

func (h *UserAdminHandler) ListUsers(c fiber.Ctx) error {
	principal, ok := middleware.PrincipalFromContext(c)
	if !ok {
		return httpx.NewError(fiber.StatusUnauthorized, "unauthorized", "Autenticación requerida", nil)
	}
	users, err := h.service.ListUsers(c.Context(), principal)
	if errors.Is(err, authn.ErrForbidden) {
		return httpx.NewError(fiber.StatusForbidden, "forbidden", "Permiso insuficiente", err)
	}
	if err != nil {
		return httpx.NewError(fiber.StatusInternalServerError, "internal_error", "Error interno del servidor", err)
	}
	return c.JSON(fiber.Map{"data": users})
}
