package patient

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/carlosfpatino/sigo-dental/platform/authn"
)

type repositoryFake struct {
	createCalls int
	patient     Patient
}

func (r *repositoryFake) Create(_ context.Context, _, _ string, patient Patient) (Patient, error) {
	r.createCalls++
	patient.ID = "d53a806e-e017-4e31-a57c-9778cd7c61f1"
	r.patient = patient
	return patient, nil
}
func (r *repositoryFake) List(_ context.Context, _ string, _ string) ([]PatientSummary, error) {
	return []PatientSummary{{ID: r.patient.ID}}, nil
}
func (r *repositoryFake) Get(_ context.Context, _ string, _ string, _ string) (Patient, error) {
	return r.patient, nil
}
func (r *repositoryFake) Update(_ context.Context, _, _ string, patient Patient) (Patient, error) {
	r.patient = patient
	return patient, nil
}

func TestCreateRequiresPatientCareRoleAndValidatesMinor(t *testing.T) {
	repo := &repositoryFake{}
	service := NewService(repo)
	service.now = func() time.Time { return time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC) }
	patient := Patient{
		FirstName: "Ada", LastName: "Patient", DateOfBirth: time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
		Phone: "+593991234567", EmergencyContactName: "Grace", EmergencyContactPhone: "+593991234568",
		MedicalHistory: MedicalHistory{},
	}
	principal := authn.Principal{UserID: "d53a806e-e017-4e31-a57c-9778cd7c61f1", Role: authn.RoleReceptionistCashier}
	if _, err := service.Create(context.Background(), authn.Principal{UserID: principal.UserID, Role: authn.RoleAuditorAccountant}, "203.0.113.8", patient); !errors.Is(err, authn.ErrForbidden) {
		t.Fatalf("Create(auditor) error = %v, want forbidden", err)
	}
	if repo.createCalls != 0 {
		t.Fatalf("repository Create called %d times for forbidden role", repo.createCalls)
	}
	created, err := service.Create(context.Background(), principal, "203.0.113.8", patient)
	if err != nil {
		t.Fatalf("Create(receptionist) error = %v", err)
	}
	if created.ID == "" || repo.createCalls != 1 {
		t.Fatalf("created = %+v, create calls = %d", created, repo.createCalls)
	}
	patient.DateOfBirth = time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC)
	if _, err := service.Create(context.Background(), authn.Principal{UserID: principal.UserID, Role: authn.RoleDentist}, "203.0.113.8", patient); !errors.Is(err, ErrInvalidPatient) {
		t.Fatalf("Create(minor without guardian) error = %v, want invalid patient", err)
	}
}
