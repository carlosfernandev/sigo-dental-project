package identity

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"strconv"
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
	filter, err := userListFilterFromRequest(c)
	if err != nil {
		return invalidUserAdminRequest()
	}
	users, err := h.service.ListUsers(c.Context(), principal, filter)
	if err != nil {
		return userAdminError(err)
	}
	return c.JSON(fiber.Map{"data": users})
}

type createUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
	RoleID   string `json:"role_id"`
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
	role := request.RoleID
	if role == "" {
		role = request.Role
	}
	user, err := h.service.CreateUser(c.Context(), principal, request.Email, request.Password, role)
	if err != nil {
		return userAdminError(err)
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": user})
}

func (h *UserAdminHandler) Register(c fiber.Ctx) error {
	var request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := decodeUserAdminRequest(c, &request); err != nil {
		return invalidUserAdminRequest()
	}
	user, err := h.service.Register(c.Context(), request.Email, request.Password)
	if err != nil {
		return userAdminError(err)
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": user})
}

func (h *UserAdminHandler) RegistrationStatus(c fiber.Ctx) error {
	enabled, err := h.service.RegistrationEnabled(c.Context())
	if err != nil {
		return httpx.NewError(fiber.StatusInternalServerError, "internal_error", "No se pudo consultar el estado del registro", err)
	}
	return c.JSON(fiber.Map{"data": fiber.Map{"enabled": enabled}})
}

func (h *UserAdminHandler) UpdateUser(c fiber.Ctx) error {
	principal, ok := middleware.PrincipalFromContext(c)
	if !ok {
		return unauthorizedUserAdmin()
	}
	var request updateUserRequest
	if err := decodeUserAdminRequest(c, &request); err != nil || request.Role == "" {
		return invalidUserAdminRequest()
	}
	user, err := h.service.UpdateUser(c.Context(), principal, c.Params("id"), request.Role, request.IsActive)
	if err != nil {
		return userAdminError(err)
	}
	return c.JSON(fiber.Map{"data": user})
}

func (h *UserAdminHandler) DeleteUser(c fiber.Ctx) error {
	principal, ok := middleware.PrincipalFromContext(c)
	if !ok {
		return unauthorizedUserAdmin()
	}
	var request struct {
		Reason string `json:"reason"`
	}
	if err := decodeUserAdminRequest(c, &request); err != nil {
		return invalidUserAdminRequest()
	}
	if _, err := h.service.DeactivateUser(c.Context(), principal, c.Params("id"), request.Reason); err != nil {
		return userAdminError(err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func userListFilterFromRequest(c fiber.Ctx) (UserListFilter, error) {
	page, err := queryInt(c.Query("page"), 1)
	if err != nil {
		return UserListFilter{}, err
	}
	pageSize, err := queryInt(c.Query("page_size"), 25)
	if err != nil {
		return UserListFilter{}, err
	}
	filter := UserListFilter{
		Page: page, PageSize: pageSize, Search: c.Query("search"), Role: c.Query("role"),
		SortBy: c.Query("sort_by"), SortOrder: c.Query("sort_order"),
	}
	if active := c.Query("is_active"); active != "" {
		parsed, err := strconv.ParseBool(active)
		if err != nil {
			return UserListFilter{}, err
		}
		filter.IsActive = &parsed
	}
	return filter, nil
}

func queryInt(value string, fallback int) (int, error) {
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, err
	}
	return parsed, nil
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
	case errors.Is(err, ErrRegistrationClosed):
		return httpx.NewError(fiber.StatusConflict, "registration_closed", "El registro inicial ya se completó. Solicita una cuenta a un administrador", err)
	case errors.Is(err, ErrUserNotFound):
		return httpx.NewError(fiber.StatusNotFound, "user_not_found", "Usuario no encontrado", err)
	case errors.Is(err, ErrCannotDeactivateSelf):
		return httpx.NewError(fiber.StatusConflict, "cannot_deactivate_self", "No puedes desactivar tu propia cuenta", err)
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
