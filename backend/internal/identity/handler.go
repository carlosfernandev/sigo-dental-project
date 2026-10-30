package identity

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/mail"
	"strings"

	"github.com/carlosfpatino/sigo-dental/platform/httpx"
	"github.com/gofiber/fiber/v3"
)

const maxLoginBodyBytes = 4096

type LoginHandler struct {
	service *AuthService
}

func NewLoginHandler(service *AuthService) *LoginHandler {
	return &LoginHandler{service: service}
}

func (h *LoginHandler) RegisterRoutes(router fiber.Router) {
	router.Post("/auth/login", h.Login)
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *LoginHandler) Login(c fiber.Ctx) error {
	contentType, _, contentTypeErr := mime.ParseMediaType(c.Get(fiber.HeaderContentType))
	if contentTypeErr != nil || !strings.EqualFold(contentType, fiber.MIMEApplicationJSON) {
		return invalidLoginRequest()
	}
	body := c.Body()
	if len(body) == 0 || len(body) > maxLoginBodyBytes {
		return invalidLoginRequest()
	}

	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	var request loginRequest
	if err := decoder.Decode(&request); err != nil {
		return invalidLoginRequest()
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return invalidLoginRequest()
	}

	request.Email = strings.TrimSpace(request.Email)
	if len(request.Email) > 254 || request.Password == "" || len([]byte(request.Password)) > 72 {
		return invalidLoginRequest()
	}
	parsed, err := mail.ParseAddress(request.Email)
	if err != nil || parsed.Address != request.Email {
		return invalidLoginRequest()
	}
	email := strings.ToLower(request.Email)

	user, err := h.service.Authenticate(c.Context(), email, request.Password)
	if errors.Is(err, ErrInvalidCredentials) {
		return httpx.NewError(fiber.StatusUnauthorized, "invalid_credentials", "Credenciales inválidas", err)
	}
	if err != nil {
		return httpx.NewError(fiber.StatusInternalServerError, "internal_error", "Error interno del servidor", err)
	}

	return c.JSON(fiber.Map{"data": fiber.Map{
		"status":               "authenticated",
		"must_change_password": user.MustChangePassword,
	}})
}

func invalidLoginRequest() error {
	return httpx.NewError(fiber.StatusBadRequest, "invalid_request", "Solicitud inválida", nil)
}
