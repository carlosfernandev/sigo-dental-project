// Package httpx contains HTTP error values mapped by the central Fiber handler.
package httpx

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gofiber/fiber/v3"
)

type ErrorEnvelope struct {
	Error ErrorBody `json:"error"`
}

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details"`
}

// FiberErrorHandler maps typed and framework errors to the shared response envelope.
func FiberErrorHandler(logger *slog.Logger) fiber.ErrorHandler {
	return func(c fiber.Ctx, err error) error {
		status := fiber.StatusInternalServerError
		code := "internal_error"
		message := "Internal Server Error"
		var apiErr *Error
		if errors.As(err, &apiErr) {
			status = apiErr.Status
			code = apiErr.Code
			message = apiErr.Message
		} else if fiberErr, ok := err.(*fiber.Error); ok {
			status = fiberErr.Code
			code = "http_error"
			if statusText := http.StatusText(status); statusText != "" {
				message = statusText
			}
		}
		if logger != nil {
			logger.Error("request failed", "code", code, "path", c.Path(), "status", status)
		}
		return c.Status(status).JSON(ErrorEnvelope{Error: ErrorBody{
			Code: code, Message: message, Details: nil,
		}})
	}
}

type Error struct {
	Status  int
	Code    string
	Message string
	Cause   error
}

func NewError(status int, code, message string, cause error) *Error {
	return &Error{Status: status, Code: code, Message: message, Cause: cause}
}

func (e *Error) Error() string { return e.Code }

func (e *Error) Unwrap() error { return e.Cause }
