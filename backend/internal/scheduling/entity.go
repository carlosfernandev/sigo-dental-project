// Package scheduling owns operatory, chair, dentist, and appointment management.
package scheduling

import (
	"context"
	"errors"
	"time"
)

var (
	ErrInvalidOperatory       = errors.New("invalid operatory")
	ErrInvalidChair           = errors.New("invalid chair")
	ErrInvalidDentist         = errors.New("invalid dentist")
	ErrInvalidAppointment     = errors.New("invalid appointment")
	ErrAppointmentNotFound    = errors.New("appointment not found")
	ErrChairOverlap           = errors.New("chair_overlap")
	ErrInvalidReminder        = errors.New("invalid reminder")
	ErrReminderNotFound       = errors.New("reminder not found")
	ErrReminderHandled        = errors.New("reminder already handled")
	ErrInvalidSchedulableTask = errors.New("invalid schedulable task")
)

type Operatory struct {
	ID          string
	Name        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Chair struct {
	ID          string
	OperatoryID string
	Name        string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Dentist struct {
	ID        string
	FirstName string
	LastName  string
	Specialty string
	Active    bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type AppointmentState string

const (
	AppointmentStateScheduled  AppointmentState = "scheduled"
	AppointmentStateConfirmed  AppointmentState = "confirmed"
	AppointmentStateCheckedIn  AppointmentState = "checked_in"
	AppointmentStateInProgress AppointmentState = "in_progress"
	AppointmentStateCompleted  AppointmentState = "completed"
	AppointmentStateCancelled  AppointmentState = "cancelled"
	AppointmentStateNoShow     AppointmentState = "no_show"
)

type Appointment struct {
	ID        string
	ChairID   string
	DentistID string
	PatientID string
	StartTime time.Time
	EndTime   time.Time
	State     AppointmentState
	Notes     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type AppointmentSummary struct {
	ID          string
	ChairID     string
	ChairName   string
	DentistID   string
	DentistName string
	PatientID   string
	PatientName string
	StartTime   time.Time
	EndTime     time.Time
	State       AppointmentState
}

type Reminder struct {
	ID                   string     `json:"id"`
	AppointmentID        string     `json:"appointment_id"`
	ReminderType         string     `json:"reminder_type"`
	Status               string     `json:"status"`
	Notes                string     `json:"notes"`
	AppointmentStartTime time.Time  `json:"appointment_start_time"`
	AppointmentEndTime   time.Time  `json:"appointment_end_time"`
	PatientID            string     `json:"patient_id"`
	PatientName          string     `json:"patient_name"`
	PatientPhone         string     `json:"patient_phone"`
	DentistName          string     `json:"dentist_name"`
	ChairName            string     `json:"chair_name"`
	HandledBy            *string    `json:"handled_by,omitempty"`
	HandledAt            *time.Time `json:"handled_at,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

type SchedulableTask struct {
	ID            string    `json:"id"`
	PatientID     string    `json:"patient_id"`
	QuoteID       string    `json:"quote_id"`
	ItemID        string    `json:"item_id"`
	ProcedureName string    `json:"procedure_name"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// Executor allows a same-process cross-domain write to share the caller's transaction.
type Executor interface {
	Exec(ctx context.Context, query string, args ...any) error
}
