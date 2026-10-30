// Package scheduling owns operatory, chair, dentist, and appointment management.
package scheduling

import (
	"errors"
	"testing"
	"time"
)

func TestValidateOperatory(t *testing.T) {
	tests := []struct {
		name    string
		o       Operatory
		wantErr error
	}{
		{"valid", Operatory{Name: "Consultorio A", Description: "Primer piso"}, nil},
		{"empty name", Operatory{Name: "", Description: "Test"}, ErrInvalidOperatory},
		{"name too long", Operatory{Name: string(make([]byte, 101)), Description: "Test"}, ErrInvalidOperatory},
		{"description too long", Operatory{Name: "Test", Description: string(make([]byte, 301))}, ErrInvalidOperatory},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateOperatory(tt.o)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("ValidateOperatory() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateChair(t *testing.T) {
	tests := []struct {
		name    string
		c       Chair
		wantErr error
	}{
		{"valid", Chair{OperatoryID: "uuid-123", Name: "Sillón 1"}, nil},
		{"empty name", Chair{OperatoryID: "uuid-123", Name: ""}, ErrInvalidChair},
		{"empty operatory", Chair{OperatoryID: "", Name: "Sillón 1"}, ErrInvalidChair},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateChair(tt.c)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("ValidateChair() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateDentist(t *testing.T) {
	tests := []struct {
		name    string
		d       Dentist
		wantErr error
	}{
		{"valid", Dentist{FirstName: "Juan", LastName: "Pérez", Specialty: "General"}, nil},
		{"empty first name", Dentist{FirstName: "", LastName: "Pérez"}, ErrInvalidDentist},
		{"empty last name", Dentist{FirstName: "Juan", LastName: ""}, ErrInvalidDentist},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDentist(tt.d)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("ValidateDentist() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateAppointment(t *testing.T) {
	now := time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name    string
		a       Appointment
		wantErr error
	}{
		{"valid", Appointment{
			ChairID: "chair-1", DentistID: "dentist-1", PatientID: "patient-1",
			StartTime: now.Add(2 * time.Hour), EndTime: now.Add(3 * time.Hour), State: AppointmentStateScheduled,
		}, nil},
		{"empty chair", Appointment{DentistID: "d", PatientID: "p", StartTime: now.Add(2 * time.Hour), EndTime: now.Add(3 * time.Hour)}, ErrInvalidAppointment},
		{"start after end", Appointment{
			ChairID: "c", DentistID: "d", PatientID: "p",
			StartTime: now.Add(3 * time.Hour), EndTime: now.Add(2 * time.Hour),
		}, ErrInvalidAppointment},
		{"too short", Appointment{
			ChairID: "c", DentistID: "d", PatientID: "p",
			StartTime: now.Add(2 * time.Hour), EndTime: now.Add(2*time.Hour + 10*time.Minute),
		}, ErrInvalidAppointment},
		{"too long", Appointment{
			ChairID: "c", DentistID: "d", PatientID: "p",
			StartTime: now.Add(2 * time.Hour), EndTime: now.Add(24 * time.Hour),
		}, ErrInvalidAppointment},
		{"invalid state", Appointment{
			ChairID: "c", DentistID: "d", PatientID: "p",
			StartTime: now.Add(2 * time.Hour), EndTime: now.Add(3 * time.Hour), State: "invalid",
		}, ErrInvalidAppointment},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAppointment(tt.a, now)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("ValidateAppointment() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateStateTransition(t *testing.T) {
	tests := []struct {
		name string
		from AppointmentState
		to   AppointmentState
		want bool
	}{
		{"scheduled to confirmed", AppointmentStateScheduled, AppointmentStateConfirmed, true},
		{"scheduled to cancelled", AppointmentStateScheduled, AppointmentStateCancelled, true},
		{"scheduled to completed", AppointmentStateScheduled, AppointmentStateCompleted, false},
		{"confirmed to checked_in", AppointmentStateConfirmed, AppointmentStateCheckedIn, true},
		{"checked_in to in_progress", AppointmentStateCheckedIn, AppointmentStateInProgress, true},
		{"in_progress to completed", AppointmentStateInProgress, AppointmentStateCompleted, true},
		{"completed to cancelled", AppointmentStateCompleted, AppointmentStateCancelled, false},
		{"cancelled to confirmed", AppointmentStateCancelled, AppointmentStateConfirmed, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidateStateTransition(tt.from, tt.to)
			if got != tt.want {
				t.Fatalf("ValidateStateTransition(%s, %s) = %v, want %v", tt.from, tt.to, got, tt.want)
			}
		})
	}
}
