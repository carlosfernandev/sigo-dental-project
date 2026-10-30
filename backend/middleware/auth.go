// Package middleware contains HTTP middleware for authentication.
package middleware

import (
	"errors"
	"strings"

	"github.com/carlosfpatino/sigo-dental/platform/httpx"
	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

const (
	principalContextKey = "sigo.authenticated_principal"
	maxBearerTokenBytes = 8192
	minimumHMACKeyBytes = 32
)

type Claims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

type Principal struct {
	UserID string
	Role   string
}

// RequireAuth validates an HS256 bearer token and attaches its identity to the
// Fiber context. Authorization decisions remain in the owning domain services.
func RequireAuth(secret []byte) (fiber.Handler, error) {
	if len(secret) < minimumHMACKeyBytes {
		return nil, errors.New("JWT secret must be at least 32 bytes")
	}
	key := append([]byte(nil), secret...)

	return func(c fiber.Ctx) error {
		parts := strings.Fields(c.Get(fiber.HeaderAuthorization))
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || len(parts[1]) > maxBearerTokenBytes {
			return unauthorized()
		}

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(parts[1], claims, func(token *jwt.Token) (any, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, errors.New("unexpected signing method")
			}
			return key, nil
		}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), jwt.WithExpirationRequired(), jwt.WithIssuedAt())
		if err != nil || token == nil || !token.Valid || claims.Subject == "" || claims.Role == "" || claims.IssuedAt == nil {
			return unauthorized()
		}

		c.Locals(principalContextKey, Principal{UserID: claims.Subject, Role: claims.Role})
		return c.Next()
	}, nil
}

func PrincipalFromContext(c fiber.Ctx) (Principal, bool) {
	principal, ok := c.Locals(principalContextKey).(Principal)
	return principal, ok
}

func unauthorized() error {
	return httpx.NewError(fiber.StatusUnauthorized, "unauthorized", "Autenticación requerida", nil)
}
