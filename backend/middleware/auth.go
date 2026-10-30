// Package middleware contains HTTP middleware for authentication.
package middleware

import (
	"strings"

	"github.com/carlosfpatino/sigo-dental/platform/authn"
	"github.com/carlosfpatino/sigo-dental/platform/httpx"
	"github.com/gofiber/fiber/v3"
)

const principalContextKey = "sigo.authenticated_principal"

type Principal = authn.Principal

// RequireAuth validates a Bearer JWT and stores its subject and role in Fiber's
// request context. Authorization decisions remain in owning domain services.
func RequireAuth(tokens *authn.TokenManager) fiber.Handler {
	return func(c fiber.Ctx) error {
		scheme, raw, ok := getReqBearerAuthToken(c)
		if !ok || !strings.EqualFold(scheme, "Bearer") {
			return unauthorized()
		}
		principal, err := tokens.Validate(raw)
		if err != nil {
			return unauthorized()
		}
		c.Locals(principalContextKey, principal)
		return c.Next()
	}
}

func PrincipalFromContext(c fiber.Ctx) (Principal, bool) {
	principal, ok := c.Locals(principalContextKey).(Principal)
	return principal, ok
}

func getReqBearerAuthToken(c fiber.Ctx) (scheme, token string, ok bool) {
	parts := strings.Fields(c.Get(fiber.HeaderAuthorization))
	if len(parts) != 2 {
		return "", "", false
	}
	return parts[0], parts[1], true
}

func unauthorized() error {
	return httpx.NewError(fiber.StatusUnauthorized, "unauthorized", "Autenticación requerida", nil)
}
