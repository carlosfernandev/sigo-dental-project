package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/carlosfpatino/sigo-dental/internal/identity"
	identityroutes "github.com/carlosfpatino/sigo-dental/internal/identity/routes"
	"github.com/carlosfpatino/sigo-dental/middleware"
	"github.com/carlosfpatino/sigo-dental/platform/authn"
	"github.com/gofiber/fiber/v3"
	"golang.org/x/crypto/bcrypt"
)

const testJWTSecret = "test-only-secret-with-more-than-32-bytes"

func TestHealth(t *testing.T) {
	app := newApp(slog.New(slog.NewTextHandler(io.Discard, nil)))
	resp, err := app.Test(httptest.NewRequest("GET", "/health", nil))
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

func registerLogin(app *fiber.App, repository *fakeAuthRepository, dummyHash string) *authn.TokenManager {
	tokens, err := authn.NewTokenManager([]byte(testJWTSecret), authn.AccessTokenTTL)
	if err != nil {
		panic(err)
	}
	identityroutes.Register(app,
		identity.NewLoginHandler(identity.NewAuthService(repository, dummyHash), tokens),
		middleware.RequireAuth(tokens, authn.NewActivityTracker()))
	return tokens
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
	registerLogin(app, repository, dummyHash)
	requests := []struct {
		name, email, password string
	}{
		{name: "wrong password", email: "wrong@example.com", password: "incorrect-password-2026"},
		{name: "unknown user", email: "missing@example.com", password: "incorrect-password-2026"},
		{name: "inactive user", email: "inactive@example.com", password: password},
	}
	var firstBody []byte
	for i, tc := range requests {
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
			if i == 0 {
				firstBody = responseBody
			} else if !bytes.Equal(responseBody, firstBody) {
				t.Errorf("error response = %s, want identical response %s", responseBody, firstBody)
			}
		})
	}
}

func TestLoginRejectsMalformedRequest(t *testing.T) {
	app := newApp(slog.New(slog.NewTextHandler(io.Discard, nil)))
	registerLogin(app, &fakeAuthRepository{}, "unused")
	req := httptest.NewRequest("POST", "/auth/login", strings.NewReader(`{"email":`))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusBadRequest)
	}
}

func TestLoginIssuesBearerTokenAndMeValidatesIt(t *testing.T) {
	const password = "clinic-user-password-2026"
	passwordHash := lowCostTestHash(t, password)
	dummyHash := lowCostTestHash(t, "dummy-test-password-value")
	user := identity.AuthUser{
		ID: "d53a806e-e017-4e31-a57c-9778cd7c61f1", Email: "admin@example.com",
		PasswordHash: passwordHash, Role: "admin_ti", IsActive: true, MustChangePassword: true,
	}
	app := newApp(slog.New(slog.NewTextHandler(io.Discard, nil)))
	tokens := registerLogin(app, &fakeAuthRepository{users: map[string]identity.AuthUser{user.Email: user}}, dummyHash)
	req := httptest.NewRequest("POST", "/auth/login", strings.NewReader(`{"email":"admin@example.com","password":"`+password+`"}`))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 15 * time.Second})
	if err != nil {
		t.Fatalf("login app.Test() error = %v", err)
	}
	var body struct {
		Data struct {
			Status             string    `json:"status"`
			MustChangePassword bool      `json:"must_change_password"`
			AccessToken        string    `json:"access_token"`
			ExpiresAt          time.Time `json:"expires_at"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		resp.Body.Close()
		t.Fatalf("decode login response: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != fiber.StatusOK || body.Data.Status != "authenticated" || !body.Data.MustChangePassword || body.Data.AccessToken == "" || body.Data.ExpiresAt.IsZero() {
		t.Fatalf("login = %d %+v, want authenticated response with access token", resp.StatusCode, body.Data)
	}
	if _, err := tokens.Validate(body.Data.AccessToken); err != nil {
		t.Fatalf("validate issued token: %v", err)
	}

	meReq := httptest.NewRequest("GET", "/auth/me", nil)
	meReq.Header.Set(fiber.HeaderAuthorization, "Bearer "+body.Data.AccessToken)
	meResp, err := app.Test(meReq)
	if err != nil {
		t.Fatalf("/auth/me app.Test() error = %v", err)
	}
	meBody, _ := io.ReadAll(meResp.Body)
	meResp.Body.Close()
	if meResp.StatusCode != fiber.StatusOK || !bytes.Contains(meBody, []byte(`"role":"admin_ti"`)) {
		t.Fatalf("/auth/me = %d %s", meResp.StatusCode, meBody)
	}

	invalidReq := httptest.NewRequest("GET", "/auth/me", nil)
	invalidReq.Header.Set(fiber.HeaderAuthorization, "Bearer invalid")
	invalidResp, err := app.Test(invalidReq)
	if err != nil {
		t.Fatalf("invalid /auth/me app.Test() error = %v", err)
	}
	invalidResp.Body.Close()
	if invalidResp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("invalid token status = %d, want %d", invalidResp.StatusCode, fiber.StatusUnauthorized)
	}
}
