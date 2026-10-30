package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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
	app := newTestApp()
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
	if got, want := string(body), `{"data":{"dependencies":{"postgres":"ok","redis":"ok"},"status":"ok"}}`; got != want {
		t.Errorf("body = %s, want %s", got, want)
	}
}

func TestCORSPreflightAllowsViteOriginsAndAuthHeaders(t *testing.T) {
	for _, origin := range []string{"http://localhost:5173", "http://127.0.0.1:5173"} {
		t.Run(origin, func(t *testing.T) {
			req := httptest.NewRequest(fiber.MethodOptions, "/auth/login", nil)
			req.Header.Set(fiber.HeaderOrigin, origin)
			req.Header.Set(fiber.HeaderAccessControlRequestMethod, fiber.MethodPost)
			req.Header.Set(fiber.HeaderAccessControlRequestHeaders, "authorization,content-type")
			resp, err := newTestApp().Test(req)
			if err != nil {
				t.Fatalf("preflight app.Test() error = %v", err)
			}
			defer resp.Body.Close()
			if resp.Header.Get(fiber.HeaderAccessControlAllowOrigin) != origin {
				t.Errorf("Access-Control-Allow-Origin = %q, want %q", resp.Header.Get(fiber.HeaderAccessControlAllowOrigin), origin)
			}
			if !strings.Contains(strings.ToLower(resp.Header.Get(fiber.HeaderAccessControlAllowHeaders)), "authorization") ||
				!strings.Contains(strings.ToLower(resp.Header.Get(fiber.HeaderAccessControlAllowHeaders)), "content-type") {
				t.Errorf("Access-Control-Allow-Headers = %q, want authorization and content-type", resp.Header.Get(fiber.HeaderAccessControlAllowHeaders))
			}
			if resp.StatusCode < fiber.StatusOK || resp.StatusCode >= fiber.StatusMultipleChoices {
				t.Errorf("preflight status = %d, want 2xx", resp.StatusCode)
			}
		})
	}
}

func TestCORSDoesNotAllowUnlistedOrigin(t *testing.T) {
	req := httptest.NewRequest(fiber.MethodOptions, "/auth/login", nil)
	req.Header.Set(fiber.HeaderOrigin, "http://untrusted.example")
	req.Header.Set(fiber.HeaderAccessControlRequestMethod, fiber.MethodPost)
	resp, err := newTestApp().Test(req)
	if err != nil {
		t.Fatalf("preflight app.Test() error = %v", err)
	}
	defer resp.Body.Close()
	if got := resp.Header.Get(fiber.HeaderAccessControlAllowOrigin); got != "" {
		t.Fatalf("Access-Control-Allow-Origin = %q, want no header", got)
	}
}

func TestHealthDegradesForRedisAndFailsForPostgres(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	for _, tc := range []struct {
		name       string
		pgErr      error
		redisErr   error
		wantStatus int
		wantBody   string
	}{
		{name: "redis degraded", redisErr: errors.New("redis down"), wantStatus: fiber.StatusOK, wantBody: `"status":"degraded"`},
		{name: "postgres unavailable", pgErr: errors.New("postgres down"), wantStatus: fiber.StatusServiceUnavailable, wantBody: `"code":"postgres_unavailable"`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			app := newApp(logger, fakeHealthPinger{err: tc.pgErr}, fakeHealthPinger{err: tc.redisErr})
			resp, err := app.Test(httptest.NewRequest("GET", "/health", nil))
			if err != nil {
				t.Fatalf("app.Test() error = %v", err)
			}
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)
			if resp.StatusCode != tc.wantStatus || !bytes.Contains(body, []byte(tc.wantBody)) {
				t.Fatalf("health = %d %s, want status %d containing %s", resp.StatusCode, body, tc.wantStatus, tc.wantBody)
			}
		})
	}
}

type fakeHealthPinger struct{ err error }

func (p fakeHealthPinger) Ping(context.Context) error { return p.err }

func newTestApp() *fiber.App {
	return newApp(slog.New(slog.NewTextHandler(io.Discard, nil)), fakeHealthPinger{}, fakeHealthPinger{})
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

func (r *fakeAuthRepository) ListUsers(_ context.Context) ([]identity.ManagedUser, error) {
	users := make([]identity.ManagedUser, 0, len(r.users))
	for _, user := range r.users {
		users = append(users, identity.ManagedUser{ID: user.ID, Email: user.Email, Role: user.Role, IsActive: user.IsActive})
	}
	return users, nil
}

func (r *fakeAuthRepository) CreateUser(_ context.Context, email, _ string, role string) (identity.ManagedUser, error) {
	user := identity.ManagedUser{ID: "d53a806e-e017-4e31-a57c-9778cd7c61f2", Email: email, Role: role, IsActive: true}
	if r.users == nil {
		r.users = make(map[string]identity.AuthUser)
	}
	r.users[email] = identity.AuthUser{ID: user.ID, Email: email, Role: role, IsActive: true}
	return user, nil
}

func (r *fakeAuthRepository) UpdateUser(_ context.Context, userID, role string, active bool) (identity.ManagedUser, error) {
	for email, user := range r.users {
		if user.ID == userID {
			user.Role, user.IsActive = role, active
			r.users[email] = user
			return identity.ManagedUser{ID: user.ID, Email: email, Role: role, IsActive: active}, nil
		}
	}
	return identity.ManagedUser{}, identity.ErrUserNotFound
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
		identity.NewUserAdminHandler(identity.NewUserAdminService(repository)),
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
	app := newTestApp()
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
	app := newTestApp()
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
	app := newTestApp()
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

	usersReq := httptest.NewRequest("GET", "/users", nil)
	usersReq.Header.Set(fiber.HeaderAuthorization, "Bearer "+body.Data.AccessToken)
	usersResp, err := app.Test(usersReq)
	if err != nil {
		t.Fatalf("admin /users app.Test() error = %v", err)
	}
	usersBody, _ := io.ReadAll(usersResp.Body)
	usersResp.Body.Close()
	if usersResp.StatusCode != fiber.StatusOK || !bytes.Contains(usersBody, []byte(`"email":"admin@example.com"`)) || bytes.Contains(usersBody, []byte("password_hash")) {
		t.Fatalf("admin /users = %d %s", usersResp.StatusCode, usersBody)
	}

	createReq := httptest.NewRequest("POST", "/users", strings.NewReader(`{"email":"dentist@example.com","password":"Dentist-Start-2026!","role":"odontologo"}`))
	createReq.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	createReq.Header.Set(fiber.HeaderAuthorization, "Bearer "+body.Data.AccessToken)
	createResp, err := app.Test(createReq, fiber.TestConfig{Timeout: 15 * time.Second})
	if err != nil {
		t.Fatalf("admin POST /users app.Test() error = %v", err)
	}
	createBody, _ := io.ReadAll(createResp.Body)
	createResp.Body.Close()
	if createResp.StatusCode != fiber.StatusCreated || !bytes.Contains(createBody, []byte(`"role":"odontologo"`)) || bytes.Contains(createBody, []byte("password")) {
		t.Fatalf("admin POST /users = %d %s", createResp.StatusCode, createBody)
	}

	updateReq := httptest.NewRequest("PUT", "/users/d53a806e-e017-4e31-a57c-9778cd7c61f2", strings.NewReader(`{"role":"odontologo","is_active":false}`))
	updateReq.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	updateReq.Header.Set(fiber.HeaderAuthorization, "Bearer "+body.Data.AccessToken)
	updateResp, err := app.Test(updateReq)
	if err != nil {
		t.Fatalf("admin PUT /users app.Test() error = %v", err)
	}
	updateBody, _ := io.ReadAll(updateResp.Body)
	updateResp.Body.Close()
	if updateResp.StatusCode != fiber.StatusOK || !bytes.Contains(updateBody, []byte(`"is_active":false`)) {
		t.Fatalf("admin PUT /users = %d %s", updateResp.StatusCode, updateBody)
	}

	receptionistToken, _, err := tokens.Issue("user-2", authn.RoleReceptionistCashier)
	if err != nil {
		t.Fatalf("issue receptionist token: %v", err)
	}
	forbiddenReq := httptest.NewRequest("GET", "/users", nil)
	forbiddenReq.Header.Set(fiber.HeaderAuthorization, "Bearer "+receptionistToken)
	forbiddenResp, err := app.Test(forbiddenReq)
	if err != nil {
		t.Fatalf("non-admin /users app.Test() error = %v", err)
	}
	forbiddenBody, _ := io.ReadAll(forbiddenResp.Body)
	forbiddenResp.Body.Close()
	if forbiddenResp.StatusCode != fiber.StatusForbidden || !bytes.Contains(forbiddenBody, []byte(`"code":"forbidden"`)) {
		t.Fatalf("non-admin /users = %d %s, want forbidden", forbiddenResp.StatusCode, forbiddenBody)
	}
	createForbiddenReq := httptest.NewRequest("POST", "/users", strings.NewReader(`{}`))
	createForbiddenReq.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	createForbiddenReq.Header.Set(fiber.HeaderAuthorization, "Bearer "+receptionistToken)
	createForbiddenResp, err := app.Test(createForbiddenReq)
	if err != nil {
		t.Fatalf("non-admin POST /users app.Test() error = %v", err)
	}
	createForbiddenBody, _ := io.ReadAll(createForbiddenResp.Body)
	createForbiddenResp.Body.Close()
	if createForbiddenResp.StatusCode != fiber.StatusForbidden || !bytes.Contains(createForbiddenBody, []byte(`"code":"forbidden"`)) {
		t.Fatalf("non-admin POST /users = %d %s, want forbidden", createForbiddenResp.StatusCode, createForbiddenBody)
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
