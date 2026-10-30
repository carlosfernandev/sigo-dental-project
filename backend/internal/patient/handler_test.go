package patient

import (
	"io"
	"log/slog"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/carlosfpatino/sigo-dental/internal/patient/routes"
	"github.com/carlosfpatino/sigo-dental/middleware"
	"github.com/carlosfpatino/sigo-dental/platform/authn"
	"github.com/carlosfpatino/sigo-dental/platform/httpx"
	"github.com/gofiber/fiber/v3"
)

func TestPatientHTTPFlowRequiresRoleAndReturnsCriticalAlerts(t *testing.T) {
	repository := &repositoryFake{}
	service := NewService(repository)
	service.now = func() time.Time { return time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC) }
	tokens, err := authn.NewTokenManager([]byte("patient-test-secret-with-more-than-32-bytes"), authn.AccessTokenTTL)
	if err != nil {
		t.Fatal(err)
	}
	app := fiber.New(fiber.Config{ErrorHandler: httpx.FiberErrorHandler(slog.New(slog.NewTextHandler(io.Discard, nil)))})
	routes.Register(app, NewHandler(service), middleware.RequireAuth(tokens, authn.NewActivityTracker()))

	const receptionistID = "d53a806e-e017-4e31-a57c-9778cd7c61f1"
	accessToken, _, err := tokens.Issue(receptionistID, authn.RoleReceptionistCashier)
	if err != nil {
		t.Fatal(err)
	}
	body := `{"first_name":"Ada","last_name":"Patient","date_of_birth":"1990-01-01","email":"ada@example.com","phone":"+593991234567","address":"Quito","emergency_contact_name":"Grace","emergency_contact_phone":"+593991234568","insurance_plan":"","medical_history":{"systemic_diseases":[],"allergies":[{"name":"Penicillin","is_critical":true}],"surgical_history":[],"coagulopathies":[]}}`
	request := httptest.NewRequest("POST", "/patients", strings.NewReader(body))
	request.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	request.Header.Set(fiber.HeaderAuthorization, "Bearer "+accessToken)
	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("POST /patients error = %v", err)
	}
	createdBody, _ := io.ReadAll(response.Body)
	response.Body.Close()
	if response.StatusCode != fiber.StatusCreated || !strings.Contains(string(createdBody), `"Alergia: Penicillin"`) {
		t.Fatalf("POST /patients = %d %s", response.StatusCode, createdBody)
	}

	listRequest := httptest.NewRequest("GET", "/patients", nil)
	listRequest.Header.Set(fiber.HeaderAuthorization, "Bearer "+accessToken)
	listResponse, err := app.Test(listRequest)
	if err != nil {
		t.Fatalf("GET /patients error = %v", err)
	}
	listBody, _ := io.ReadAll(listResponse.Body)
	listResponse.Body.Close()
	if listResponse.StatusCode != fiber.StatusOK || strings.Contains(string(listBody), "medical_history") {
		t.Fatalf("GET /patients = %d %s; list must contain summaries only", listResponse.StatusCode, listBody)
	}

	auditorToken, _, err := tokens.Issue("d53a806e-e017-4e31-a57c-9778cd7c61f2", authn.RoleAuditorAccountant)
	if err != nil {
		t.Fatal(err)
	}
	forbiddenRequest := httptest.NewRequest("GET", "/patients", nil)
	forbiddenRequest.Header.Set(fiber.HeaderAuthorization, "Bearer "+auditorToken)
	forbiddenResponse, err := app.Test(forbiddenRequest)
	if err != nil {
		t.Fatalf("auditor GET /patients error = %v", err)
	}
	forbiddenResponse.Body.Close()
	if forbiddenResponse.StatusCode != fiber.StatusForbidden {
		t.Fatalf("auditor GET /patients status = %d, want 403", forbiddenResponse.StatusCode)
	}
}

var _ Repository = (*repositoryFake)(nil)
