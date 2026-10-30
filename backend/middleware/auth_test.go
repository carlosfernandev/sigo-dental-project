package middleware

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/carlosfpatino/sigo-dental/platform/httpx"
	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

func TestRequireAuthAddsPrincipalToContext(t *testing.T) {
	secret := []byte(strings.Repeat("s", 32))
	guard, err := RequireAuth(secret)
	if err != nil {
		t.Fatalf("RequireAuth() error = %v", err)
	}
	app := authTestApp(guard)
	token := issueTestToken(t, jwt.SigningMethodHS256, secret, time.Now().Add(time.Hour), true)
	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set(fiber.HeaderAuthorization, "bEaReR "+token)
	response, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != fiber.StatusOK {
		t.Fatalf("status = %d, want %d", response.StatusCode, fiber.StatusOK)
	}
	var body struct {
		Data struct {
			Subject string `json:"sub"`
			Role    string `json:"role"`
		} `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Data.Subject != "d53a806e-e017-4e31-a57c-9778cd7c61f1" || body.Data.Role != "admin_ti" {
		t.Errorf("principal = %+v, want authenticated user and role", body.Data)
	}
}

func TestRequireAuthReturnsSameErrorForInvalidTokens(t *testing.T) {
	secret := []byte(strings.Repeat("s", 32))
	guard, err := RequireAuth(secret)
	if err != nil {
		t.Fatalf("RequireAuth() error = %v", err)
	}
	expired := issueTestToken(t, jwt.SigningMethodHS256, secret, time.Now().Add(-time.Minute), true)
	wrongMethod := issueTestToken(t, jwt.SigningMethodHS384, secret, time.Now().Add(time.Hour), true)
	noIssuedAt := issueTestToken(t, jwt.SigningMethodHS256, secret, time.Now().Add(time.Hour), false)
	app := authTestApp(guard)

	requests := []struct {
		name   string
		header string
	}{
		{name: "missing header"},
		{name: "malformed token", header: "Bearer not-a-jwt"},
		{name: "expired token", header: "Bearer " + expired},
		{name: "wrong signing method", header: "Bearer " + wrongMethod},
		{name: "missing issued-at", header: "Bearer " + noIssuedAt},
	}
	var expectedBody string
	for i, test := range requests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/protected", nil)
			if test.header != "" {
				req.Header.Set(fiber.HeaderAuthorization, test.header)
			}
			response, err := app.Test(req)
			if err != nil {
				t.Fatalf("app.Test() error = %v", err)
			}
			defer response.Body.Close()
			if response.StatusCode != fiber.StatusUnauthorized {
				t.Fatalf("status = %d, want %d", response.StatusCode, fiber.StatusUnauthorized)
			}
			var body strings.Builder
			if _, err := io.Copy(&body, response.Body); err != nil {
				t.Fatalf("read response body: %v", err)
			}
			if i == 0 {
				expectedBody = body.String()
			} else if body.String() != expectedBody {
				t.Errorf("body = %s, want identical error %s", body.String(), expectedBody)
			}
		})
	}
	if expectedBody != `{"error":{"code":"unauthorized","message":"Autenticación requerida","details":null}}` {
		t.Errorf("unauthorized body = %s", expectedBody)
	}
}

func TestRequireAuthRejectsShortSecrets(t *testing.T) {
	if handler, err := RequireAuth([]byte("too-short")); err == nil || handler != nil {
		t.Fatalf("RequireAuth() = (%v, %v), want nil handler and an error", handler, err)
	}
}

func authTestApp(guard fiber.Handler) *fiber.App {
	app := fiber.New(fiber.Config{ErrorHandler: httpx.FiberErrorHandler(slog.New(slog.NewTextHandler(io.Discard, nil)))})
	app.Get("/protected", guard, func(c fiber.Ctx) error {
		principal, ok := PrincipalFromContext(c)
		if !ok {
			return fiber.NewError(fiber.StatusInternalServerError)
		}
		return c.JSON(fiber.Map{"data": fiber.Map{"sub": principal.UserID, "role": principal.Role}})
	})
	return app
}

func issueTestToken(t *testing.T, method jwt.SigningMethod, secret []byte, expiresAt time.Time, includeIssuedAt bool) string {
	t.Helper()
	now := time.Now()
	claims := Claims{
		Role: "admin_ti",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "d53a806e-e017-4e31-a57c-9778cd7c61f1",
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}
	if includeIssuedAt {
		claims.IssuedAt = jwt.NewNumericDate(now.Add(-time.Minute))
	}
	token, err := jwt.NewWithClaims(method, claims).SignedString(secret)
	if err != nil {
		t.Fatalf("sign test token: %v", err)
	}
	return token
}
