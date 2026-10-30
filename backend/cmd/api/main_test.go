package main

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/carlosfpatino/sigo-dental/internal/identity"
	"github.com/gofiber/fiber/v3"
	"golang.org/x/crypto/bcrypt"
)

func TestHealth(t *testing.T) {
	app := newApp(slog.New(slog.NewTextHandler(io.Discard, nil)))
	req := httptest.NewRequest("GET", "/health", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusOK)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}
	if got, want := string(body), `{"status":"ok"}`; got != want {
		t.Errorf("body = %s, want %s", got, want)
	}
}

type fakeAuthRepository struct {
	users       map[string]identity.AuthUser
	updatedHash string
}

func (r *fakeAuthRepository) FindByEmail(_ context.Context, email string) (identity.AuthUser, error) {
	user, ok := r.users[email]
	if !ok {
		return identity.AuthUser{}, identity.ErrUserNotFound
	}
	return user, nil
}

func (r *fakeAuthRepository) UpdatePasswordHash(_ context.Context, _ string, passwordHash string) error {
	r.updatedHash = passwordHash
	return nil
}

func lowCostTestHash(t *testing.T, password string) string {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("create test password hash: %v", err)
	}
	return string(hash)
}

func TestLoginErrorsAreIndistinguishable(t *testing.T) {
	const password = "clinic-user-password-2026"
	passwordHash := lowCostTestHash(t, password)
	dummyHash := lowCostTestHash(t, "dummy-test-password-value")
	repository := &fakeAuthRepository{users: map[string]identity.AuthUser{
		"wrong@example.com":    {Email: "wrong@example.com", PasswordHash: passwordHash, IsActive: true},
		"inactive@example.com": {Email: "inactive@example.com", PasswordHash: passwordHash, IsActive: false},
	}}
	app := newApp(slog.New(slog.NewTextHandler(io.Discard, nil)))
	identity.NewLoginHandler(identity.NewAuthService(repository, dummyHash)).RegisterRoutes(app)

	requests := []struct {
		name     string
		email    string
		password string
	}{
		{name: "wrong password", email: "wrong@example.com", password: "incorrect-password-2026"},
		{name: "unknown user", email: "missing@example.com", password: "incorrect-password-2026"},
		{name: "inactive user", email: "inactive@example.com", password: password},
	}

	var firstBody []byte
	for i, tc := range requests {
		requestIndex := i
		t.Run(tc.name, func(t *testing.T) {
			body := `{"email":"` + tc.email + `","password":"` + tc.password + `"}`
			req := httptest.NewRequest("POST", "/auth/login", strings.NewReader(body))
			req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
			resp, err := app.Test(req, fiber.TestConfig{Timeout: 15 * time.Second})
			if err != nil {
				t.Fatalf("app.Test() error = %v", err)
			}
			defer resp.Body.Close()
			responseBody, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("read response body: %v", err)
			}
			if resp.StatusCode != fiber.StatusUnauthorized {
				t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusUnauthorized)
			}
			if requestIndex == 0 {
				firstBody = responseBody
			} else if !bytes.Equal(responseBody, firstBody) {
				t.Errorf("error response = %s, want identical response %s", responseBody, firstBody)
			}
		})
	}
	if got, want := string(firstBody), `{"error":{"code":"invalid_credentials","details":null,"message":"Credenciales inválidas"}}`; got != want {
		t.Errorf("generic login error body = %s, want %s", got, want)
	}
}

func TestLoginRejectsMalformedRequests(t *testing.T) {
	app := newApp(slog.New(slog.NewTextHandler(io.Discard, nil)))
	identity.NewLoginHandler(identity.NewAuthService(&fakeAuthRepository{}, "unused")).RegisterRoutes(app)

	tests := []struct {
		name        string
		contentType string
		body        string
	}{
		{name: "malformed JSON", contentType: fiber.MIMEApplicationJSON, body: `{"email":`},
		{name: "unknown field", contentType: fiber.MIMEApplicationJSON, body: `{"email":"admin@example.com","password":"password-123456","extra":true}`},
		{name: "trailing JSON", contentType: fiber.MIMEApplicationJSON, body: `{"email":"admin@example.com","password":"password-123456"}{}`},
		{name: "missing password", contentType: fiber.MIMEApplicationJSON, body: `{"email":"admin@example.com"}`},
		{name: "invalid email", contentType: fiber.MIMEApplicationJSON, body: `{"email":"nope","password":"password-123456"}`},
		{name: "email too long", contentType: fiber.MIMEApplicationJSON, body: `{"email":"` + strings.Repeat("a", 250) + `@example.com","password":"password-123456"}`},
		{name: "password too long", contentType: fiber.MIMEApplicationJSON, body: `{"email":"admin@example.com","password":"` + strings.Repeat("a", 73) + `"}`},
		{name: "body too large", contentType: fiber.MIMEApplicationJSON, body: strings.Repeat(" ", 4097)},
		{name: "wrong content type", contentType: "text/plain", body: `{"email":"admin@example.com","password":"password-123456"}`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/auth/login", strings.NewReader(tc.body))
			req.Header.Set(fiber.HeaderContentType, tc.contentType)
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("app.Test() error = %v", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != fiber.StatusBadRequest {
				t.Errorf("status = %d, want %d", resp.StatusCode, fiber.StatusBadRequest)
			}
		})
	}
}

func TestLoginReturnsInternalAuthenticatedResult(t *testing.T) {
	const password = "clinic-user-password-2026"
	passwordHash := lowCostTestHash(t, password)
	dummyHash := lowCostTestHash(t, "dummy-test-password-value")
	repository := &fakeAuthRepository{users: map[string]identity.AuthUser{
		"admin@example.com": {
			ID: "d53a806e-e017-4e31-a57c-9778cd7c61f1", Email: "admin@example.com",
			PasswordHash: passwordHash, Role: "admin_ti", IsActive: true, MustChangePassword: true,
		},
	}}
	app := newApp(slog.New(slog.NewTextHandler(io.Discard, nil)))
	identity.NewLoginHandler(identity.NewAuthService(repository, dummyHash)).RegisterRoutes(app)
	req := httptest.NewRequest("POST", "/auth/login", strings.NewReader(`{"email":"ADMIN@example.com","password":"`+password+`"}`))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 15 * time.Second})
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status = %d, want %d, body %s", resp.StatusCode, fiber.StatusOK, body)
	}
	if got, want := string(body), `{"data":{"must_change_password":true,"status":"authenticated"}}`; got != want {
		t.Errorf("body = %s, want %s", got, want)
	}
}
