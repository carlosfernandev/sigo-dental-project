package identity

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

type createUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type updateUserRequest struct {
	Role     string `json:"role"`
	IsActive *bool  `json:"is_active"`
}

func (h *UserAdminHandler) CreateUser(c fiber.Ctx) error {
	principal, ok := middleware.PrincipalFromContext(c)
	if !ok {
		return unauthorizedUserAdmin()
	}
	var request createUserRequest
	if err := decodeUserAdminRequest(c, &request); err != nil {
		return invalidUserAdminRequest()
	}
	user, err := h.service.CreateUser(c.Context(), principal, request.Email, request.Password, request.Role)
	if err != nil {
		return userAdminError(err)
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": user})
}

func (h *UserAdminHandler) UpdateUser(c fiber.Ctx) error {
	principal, ok := middleware.PrincipalFromContext(c)
	if !ok {
		return unauthorizedUserAdmin()
	}
	var request updateUserRequest
	if err := decodeUserAdminRequest(c, &request); err != nil || request.IsActive == nil {
		return invalidUserAdminRequest()
	}
	user, err := h.service.UpdateUser(c.Context(), principal, c.Params("id"), request.Role, *request.IsActive)
	if err != nil {
		return userAdminError(err)
	}
	return c.JSON(fiber.Map{"data": user})
}

func decodeUserAdminRequest(c fiber.Ctx, target any) error {
	contentType, _, err := mime.ParseMediaType(c.Get(fiber.HeaderContentType))
	if err != nil || !strings.EqualFold(contentType, fiber.MIMEApplicationJSON) || len(c.Body()) == 0 || len(c.Body()) > 4096 {
		return ErrInvalidUserInput
	}
	decoder := json.NewDecoder(bytes.NewReader(c.Body()))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return ErrInvalidUserInput
	}
	return nil
}

func userAdminError(err error) error {
	switch {
	case errors.Is(err, authn.ErrForbidden):
		return httpx.NewError(fiber.StatusForbidden, "forbidden", "Permiso insuficiente", err)
	case errors.Is(err, ErrInvalidUserInput):
		return invalidUserAdminRequest()
	case errors.Is(err, ErrEmailAlreadyExists):
		return httpx.NewError(fiber.StatusConflict, "email_already_exists", "El correo ya está registrado", err)
	case errors.Is(err, ErrUserNotFound):
		return httpx.NewError(fiber.StatusNotFound, "user_not_found", "Usuario no encontrado", err)
	default:
		return httpx.NewError(fiber.StatusInternalServerError, "internal_error", "Error interno del servidor", err)
	}
}

func unauthorizedUserAdmin() error {
	return httpx.NewError(fiber.StatusUnauthorized, "unauthorized", "Autenticación requerida", nil)
}

func invalidUserAdminRequest() error {
	return httpx.NewError(fiber.StatusBadRequest, "invalid_request", "Solicitud inválida", nil)
}
