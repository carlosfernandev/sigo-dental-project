// Package scheduling owns operatory, chair, dentist, and appointment management.
package scheduling

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/carlosfpatino/sigo-dental/platform/authn"
)

type repositoryFake struct {
	createAppointmentCall int
	appointment           Appointment
}

func (r *repositoryFake) CreateOperatory(_ context.Context, o Operatory) (Operatory, error) {
	o.ID = "operatory-1"
	return o, nil
}
func (r *repositoryFake) ListOperatories(_ context.Context) ([]Operatory, error) {
	return []Operatory{{ID: "operatory-1"}}, nil
}
func (r *repositoryFake) GetOperatory(_ context.Context, _ string) (Operatory, error) {
	return Operatory{ID: "operatory-1"}, nil
}
func (r *repositoryFake) UpdateOperatory(_ context.Context, o Operatory) (Operatory, error) {
	return o, nil
}
func (r *repositoryFake) CreateChair(_ context.Context, c Chair) (Chair, error) {
	c.ID = "chair-1"
	return c, nil
}
func (r *repositoryFake) ListChairs(_ context.Context) ([]Chair, error) {
	return []Chair{{ID: "chair-1"}}, nil
}
func (r *repositoryFake) GetChair(_ context.Context, _ string) (Chair, error) {
	return Chair{ID: "chair-1"}, nil
}
func (r *repositoryFake) UpdateChair(_ context.Context, c Chair) (Chair, error) {
	return c, nil
}
func (r *repositoryFake) CreateDentist(_ context.Context, d Dentist) (Dentist, error) {
	d.ID = "dentist-1"
	return d, nil
}
func (r *repositoryFake) ListDentists(_ context.Context) ([]Dentist, error) {
	return []Dentist{{ID: "dentist-1"}}, nil
}
func (r *repositoryFake) GetDentist(_ context.Context, _ string) (Dentist, error) {
	return Dentist{ID: "dentist-1"}, nil
}
func (r *repositoryFake) UpdateDentist(_ context.Context, d Dentist) (Dentist, error) {
	return d, nil
}
func (r *repositoryFake) CreateAppointment(_ context.Context, a Appointment) (Appointment, error) {
	r.createAppointmentCall++
	a.ID = "appointment-1"
	r.appointment = a
	return a, nil
}
func (r *repositoryFake) ListAppointments(_ context.Context, _, _, _, _, _ *string) ([]AppointmentSummary, error) {
	return []AppointmentSummary{{ID: "appointment-1"}}, nil
}
func (r *repositoryFake) GetAppointment(_ context.Context, id string) (Appointment, error) {
	if id == "appointment-1" {
		return r.appointment, nil
	}
	return Appointment{}, ErrAppointmentNotFound
}
func (r *repositoryFake) UpdateAppointment(_ context.Context, a Appointment) (Appointment, error) {
	r.appointment = a
	return a, nil
}

func TestCreateAppointmentRequiresSchedulingRoleAndValidates(t *testing.T) {
	repo := &repositoryFake{}
	service := NewService(repo)
	service.now = func() time.Time { return time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC) }

	appointment := Appointment{
		ChairID:   "chair-1",
		DentistID: "dentist-1",
		PatientID: "patient-1",
		StartTime: service.now().Add(2 * time.Hour),
		EndTime:   service.now().Add(3 * time.Hour),
	}

	principal := authn.Principal{UserID: "user-1", Role: authn.RoleReceptionistCashier}
	created, err := service.CreateAppointment(context.Background(), principal, appointment)
	if err != nil {
		t.Fatalf("CreateAppointment(receptionist) error = %v", err)
	}
	if created.ID == "" || repo.createAppointmentCall != 1 {
		t.Fatalf("created = %+v, create calls = %d", created, repo.createAppointmentCall)
	}

	if _, err := service.CreateAppointment(context.Background(), authn.Principal{UserID: "user-2", Role: authn.RoleAuditorAccountant}, appointment); !errors.Is(err, authn.ErrForbidden) {
		t.Fatalf("CreateAppointment(auditor) error = %v, want forbidden", err)
	}
}

func TestUpdateAppointmentValidatesStateTransition(t *testing.T) {
	repo := &repositoryFake{}
	service := NewService(repo)
	service.now = func() time.Time { return time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC) }

	principal := authn.Principal{UserID: "user-1", Role: authn.RoleReceptionistCashier}
	existing := Appointment{
		ID:        "appointment-1",
		ChairID:   "chair-1",
		DentistID: "dentist-1",
		PatientID: "patient-1",
		StartTime: service.now().Add(2 * time.Hour),
		EndTime:   service.now().Add(3 * time.Hour),
		State:     AppointmentStateScheduled,
	}
	repo.appointment = existing

	update := existing
	update.State = AppointmentStateConfirmed
	updated, err := service.UpdateAppointment(context.Background(), principal, update)
	if err != nil {
		t.Fatalf("UpdateAppointment(scheduled->confirmed) error = %v", err)
	}
	if updated.State != AppointmentStateConfirmed {
		t.Fatalf("updated state = %v, want confirmed", updated.State)
	}

	update.State = AppointmentStateCompleted
	if _, err := service.UpdateAppointment(context.Background(), principal, update); err == nil {
		t.Fatalf("UpdateAppointment(scheduled->completed) should fail, got nil error")
	}
}
