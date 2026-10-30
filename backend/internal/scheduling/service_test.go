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
	generateReminderCalls int
	generatedAt           time.Time
	reminders             []Reminder
	updatedReminder       Reminder
	schedulableTasks      []SchedulableTask
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
func (r *repositoryFake) GeneratePendingReminders(_ context.Context, now time.Time) error {
	r.generateReminderCalls++
	r.generatedAt = now
	return nil
}
func (r *repositoryFake) ListPendingReminders(context.Context) ([]Reminder, error) {
	return r.reminders, nil
}
func (r *repositoryFake) GetReminder(_ context.Context, id string) (Reminder, error) {
	if r.updatedReminder.ID != id {
		return Reminder{}, ErrReminderNotFound
	}
	return r.updatedReminder, nil
}
func (r *repositoryFake) UpdateReminder(_ context.Context, id, status, actorID string, notes *string) (Reminder, error) {
	r.updatedReminder.ID = id
	r.updatedReminder.Status = status
	r.updatedReminder.HandledBy = &actorID
	r.updatedReminder.Notes = ""
	if notes != nil {
		r.updatedReminder.Notes = *notes
	}
	return r.updatedReminder, nil
}
func (r *repositoryFake) CreateSchedulableTasks(_ context.Context, _ Executor, tasks []SchedulableTask) error {
	r.schedulableTasks = append(r.schedulableTasks, tasks...)
	return nil
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

func TestListPendingRemindersGeneratesTheQueue(t *testing.T) {
	repo := &repositoryFake{reminders: []Reminder{{ID: "reminder-1", Status: "pending"}}}
	service := NewService(repo)
	now := time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC)
	service.now = func() time.Time { return now }
	principal := authn.Principal{UserID: "user-1", Role: authn.RoleReceptionistCashier}

	reminders, err := service.ListPendingReminders(context.Background(), principal)
	if err != nil {
		t.Fatalf("ListPendingReminders() error = %v", err)
	}
	if len(reminders) != 1 || repo.generateReminderCalls != 1 || !repo.generatedAt.Equal(now) {
		t.Fatalf("reminders=%+v, generation calls=%d, generated at=%v", reminders, repo.generateReminderCalls, repo.generatedAt)
	}
	if _, err := service.ListPendingReminders(context.Background(), authn.Principal{Role: authn.RoleDentist}); !errors.Is(err, authn.ErrForbidden) {
		t.Fatalf("ListPendingReminders(dentist) error = %v, want forbidden", err)
	}
}

func TestUpdateReminderValidatesStatusAndRecordsReceptionist(t *testing.T) {
	repo := &repositoryFake{}
	service := NewService(repo)
	principal := authn.Principal{UserID: "123e4567-e89b-12d3-a456-426614174000", Role: authn.RoleReceptionistCashier}
	id := "223e4567-e89b-12d3-a456-426614174000"
	notes := "  Llamada realizada  "

	updated, err := service.UpdateReminder(context.Background(), principal, id, updateReminderRequest{Status: "contacted", Notes: &notes})
	if err != nil {
		t.Fatalf("UpdateReminder() error = %v", err)
	}
	if updated.Status != "contacted" || updated.Notes != "Llamada realizada" || updated.HandledBy == nil || *updated.HandledBy != principal.UserID {
		t.Fatalf("UpdateReminder() = %+v", updated)
	}
	if _, err := service.UpdateReminder(context.Background(), principal, id, updateReminderRequest{Status: "pending"}); !errors.Is(err, ErrInvalidReminder) {
		t.Fatalf("UpdateReminder(pending) error = %v, want invalid reminder", err)
	}
}
