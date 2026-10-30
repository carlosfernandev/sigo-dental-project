// Package authn contains request identity types shared across HTTP boundaries.
package authn

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const AccessTokenTTL = 2 * time.Hour

var ErrInvalidToken = errors.New("invalid token")
var ErrForbidden = errors.New("forbidden")

type Principal struct {
	UserID   string
	Role     string
	TokenID  string
	IssuedAt time.Time
}

type claims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

type TokenManager struct {
	secret []byte
	ttl    time.Duration
}

func NewTokenManager(secret []byte, ttl time.Duration) (*TokenManager, error) {
	if len(secret) < 32 {
		return nil, errors.New("JWT secret must contain at least 32 bytes")
	}
	if ttl <= 0 {
		return nil, errors.New("JWT lifetime must be positive")
	}
	return &TokenManager{secret: append([]byte(nil), secret...), ttl: ttl}, nil
}

func (m *TokenManager) Issue(userID, role string) (string, time.Time, error) {
	var randomID [16]byte
	if _, err := rand.Read(randomID[:]); err != nil {
		return "", time.Time{}, fmt.Errorf("generate token id: %w", err)
	}
	now := time.Now().UTC()
	expiresAt := now.Add(m.ttl)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims{
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			ID:        base64.RawURLEncoding.EncodeToString(randomID[:]),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	})
	signed, err := token.SignedString(m.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign access token: %w", err)
	}
	return signed, expiresAt, nil
}

func (m *TokenManager) Validate(raw string) (Principal, error) {
	if len(raw) == 0 || len(raw) > 8192 {
		return Principal{}, ErrInvalidToken
	}
	parsedClaims := &claims{}
	token, err := jwt.ParseWithClaims(raw, parsedClaims, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, ErrInvalidToken
		}
		return m.secret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), jwt.WithExpirationRequired(), jwt.WithIssuedAt())
	if err != nil || token == nil || !token.Valid || parsedClaims.Subject == "" || parsedClaims.Role == "" || parsedClaims.ID == "" || parsedClaims.IssuedAt == nil {
		return Principal{}, ErrInvalidToken
	}
	return Principal{
		UserID: parsedClaims.Subject, Role: parsedClaims.Role,
		TokenID: parsedClaims.ID, IssuedAt: parsedClaims.IssuedAt.Time,
	}, nil
}

// RequireAnyRole is called by domain service use cases before they act on a request.
func RequireAnyRole(principal Principal, roles ...string) error {
	for _, role := range roles {
		if principal.Role == role {
			return nil
		}
	}
	return ErrForbidden
}
