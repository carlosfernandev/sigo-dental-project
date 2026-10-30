// Package scheduling owns operatory, chair, dentist, and appointment management.
package scheduling

import (
	"errors"
	"time"
)

var (
	ErrInvalidOperatory    = errors.New("invalid operatory")
	ErrInvalidChair        = errors.New("invalid chair")
	ErrInvalidDentist      = errors.New("invalid dentist")
	ErrInvalidAppointment  = errors.New("invalid appointment")
	ErrAppointmentNotFound = errors.New("appointment not found")
	ErrChairOverlap        = errors.New("chair_overlap")
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
