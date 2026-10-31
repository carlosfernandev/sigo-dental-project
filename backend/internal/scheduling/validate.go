// Package scheduling owns operatory, chair, dentist, and appointment management.
package scheduling

import (
	"regexp"
	"strings"
	"time"
)

var schedulingUUIDPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

func ValidateOperatory(o Operatory) error {
	o.Name = strings.TrimSpace(o.Name)
	o.Description = strings.TrimSpace(o.Description)
	if o.Name == "" || len(o.Name) > 100 || len(o.Description) > 300 {
		return ErrInvalidOperatory
	}
	return nil
}

func ValidateChair(c Chair) error {
	c.Name = strings.TrimSpace(c.Name)
	if c.Name == "" || len(c.Name) > 100 || c.OperatoryID == "" {
		return ErrInvalidChair
	}
	return nil
}

func ValidateDentist(d Dentist) error {
	d.FirstName = strings.TrimSpace(d.FirstName)
	d.LastName = strings.TrimSpace(d.LastName)
	d.Specialty = strings.TrimSpace(d.Specialty)
	if d.FirstName == "" || len(d.FirstName) > 100 || d.LastName == "" || len(d.LastName) > 100 {
		return ErrInvalidDentist
	}
	return nil
}

func ValidateAppointment(a Appointment, now time.Time) error {
	if a.ChairID == "" || a.DentistID == "" || a.PatientID == "" {
		return ErrInvalidAppointment
	}
	if a.StartTime.IsZero() || a.EndTime.IsZero() {
		return ErrInvalidAppointment
	}
	if a.StartTime.After(a.EndTime) {
		return ErrInvalidAppointment
	}
	if a.StartTime.Before(now.Add(-24 * time.Hour)) {
		return ErrInvalidAppointment
	}
	duration := a.EndTime.Sub(a.StartTime)
	if duration < 15*time.Minute || duration > 8*time.Hour {
		return ErrInvalidAppointment
	}
	validStates := map[AppointmentState]bool{
		AppointmentStateScheduled:  true,
		AppointmentStateConfirmed:  true,
		AppointmentStateCheckedIn:  true,
		AppointmentStateInProgress: true,
		AppointmentStateCompleted:  true,
		AppointmentStateCancelled:  true,
		AppointmentStateNoShow:     true,
	}
	if !validStates[a.State] {
		return ErrInvalidAppointment
	}
	return nil
}

func ValidateStateTransition(from, to AppointmentState) bool {
	transitions := map[AppointmentState][]AppointmentState{
		AppointmentStateScheduled: {
			AppointmentStateConfirmed,
			AppointmentStateCancelled,
			AppointmentStateNoShow,
		},
		AppointmentStateConfirmed: {
			AppointmentStateCheckedIn,
			AppointmentStateCancelled,
			AppointmentStateNoShow,
		},
		AppointmentStateCheckedIn: {
			AppointmentStateInProgress,
			AppointmentStateCancelled,
		},
		AppointmentStateInProgress: {
			AppointmentStateCompleted,
		},
		AppointmentStateCompleted: {},
		AppointmentStateCancelled: {},
		AppointmentStateNoShow:    {},
	}
	allowed, exists := transitions[from]
	if !exists {
		return false
	}
	for _, state := range allowed {
		if state == to {
			return true
		}
	}
	return false
}
