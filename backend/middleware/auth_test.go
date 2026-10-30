package middleware

import (
	"io"
	"log/slog"
	"net/http/httptest"
	"testing"

	"github.com/carlosfpatino/sigo-dental/platform/authn"
	"github.com/carlosfpatino/sigo-dental/platform/httpx"
	"github.com/gofiber/fiber/v3"
)

func TestRequireAuthAddsPrincipal(t *testing.T) {
	tokens, err := authn.NewTokenManager([]byte("test-only-secret-with-more-than-32-bytes"), authn.AccessTokenTTL)
	if err != nil {
		t.Fatal(err)
	}
	accessToken, _, err := tokens.Issue("user-1", "odontologo")
	if err != nil {
		t.Fatal(err)
	}
	app := testApp()
	app.Get("/protected", RequireAuth(tokens, authn.NewActivityTracker()), func(c fiber.Ctx) error {
		principal, ok := PrincipalFromContext(c)
		if !ok || principal.UserID != "user-1" || principal.Role != "odontologo" {
			t.Fatalf("principal = %+v, %v", principal, ok)
		}
		return c.SendStatus(fiber.StatusNoContent)
	})
	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set(fiber.HeaderAuthorization, "Bearer "+accessToken)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusNoContent {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusNoContent)
	}
}

func TestRequireAuthRejectsMissingAndMalformedBearerTokens(t *testing.T) {
	tokens, err := authn.NewTokenManager([]byte("test-only-secret-with-more-than-32-bytes"), authn.AccessTokenTTL)
	if err != nil {
		t.Fatal(err)
	}
	app := testApp()
	app.Get("/protected", RequireAuth(tokens, authn.NewActivityTracker()), func(c fiber.Ctx) error { return c.SendStatus(fiber.StatusNoContent) })
	for _, authorization := range []string{"", "Basic abc", "Bearer", "Bearer not-a-jwt"} {
		req := httptest.NewRequest("GET", "/protected", nil)
		if authorization != "" {
			req.Header.Set(fiber.HeaderAuthorization, authorization)
		}
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("app.Test() error = %v", err)
		}
		resp.Body.Close()
		if resp.StatusCode != fiber.StatusUnauthorized {
			t.Errorf("authorization %q status = %d, want %d", authorization, resp.StatusCode, fiber.StatusUnauthorized)
		}
	}
}

func TestNewTokenManagerRequiresStrongSecret(t *testing.T) {
	if _, err := authn.NewTokenManager([]byte("short"), authn.AccessTokenTTL); err == nil {
		t.Fatal("NewTokenManager() accepted a short secret")
	}
}

func testApp() *fiber.App {
	return fiber.New(fiber.Config{ErrorHandler: httpx.FiberErrorHandler(slog.New(slog.NewTextHandler(io.Discard, nil)))})
}
