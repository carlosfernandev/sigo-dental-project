// Package scheduling owns operatory, chair, dentist, and appointment management.
package scheduling

import "time"

// Operatory DTOs
type operatoryRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type operatoryResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Chair DTOs
type chairRequest struct {
	OperatoryID string `json:"operatory_id"`
	Name        string `json:"name"`
}

type chairResponse struct {
	ID          string    `json:"id"`
	OperatoryID string    `json:"operatory_id"`
	Name        string    `json:"name"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Dentist DTOs
type dentistRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Specialty string `json:"specialty"`
	Active    bool   `json:"active"`
}

type dentistResponse struct {
	ID        string    `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Specialty string    `json:"specialty"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Appointment DTOs
type appointmentRequest struct {
	ChairID   string           `json:"chair_id"`
	DentistID string           `json:"dentist_id"`
	PatientID string           `json:"patient_id"`
	StartTime string           `json:"start_time"`
	EndTime   string           `json:"end_time"`
	State     AppointmentState `json:"state"`
	Notes     string           `json:"notes"`
}

type appointmentResponse struct {
	ID        string           `json:"id"`
	ChairID   string           `json:"chair_id"`
	DentistID string           `json:"dentist_id"`
	PatientID string           `json:"patient_id"`
	StartTime string           `json:"start_time"`
	EndTime   string           `json:"end_time"`
	State     AppointmentState `json:"state"`
	Notes     string           `json:"notes"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
}

type appointmentSummaryResponse struct {
	ID          string           `json:"id"`
	ChairID     string           `json:"chair_id"`
	ChairName   string           `json:"chair_name"`
	DentistID   string           `json:"dentist_id"`
	DentistName string           `json:"dentist_name"`
	PatientID   string           `json:"patient_id"`
	PatientName string           `json:"patient_name"`
	StartTime   string           `json:"start_time"`
	EndTime     string           `json:"end_time"`
	State       AppointmentState `json:"state"`
}

// Request conversion methods
func (r operatoryRequest) entity() Operatory {
	return Operatory{
		Name:        r.Name,
		Description: r.Description,
	}
}

func (r chairRequest) entity() Chair {
	return Chair{
		OperatoryID: r.OperatoryID,
		Name:        r.Name,
	}
}

func (r dentistRequest) entity() Dentist {
	return Dentist{
		FirstName: r.FirstName,
		LastName:  r.LastName,
		Specialty: r.Specialty,
		Active:    r.Active,
	}
}

func (r appointmentRequest) entity() (Appointment, error) {
	start, err := time.Parse(time.RFC3339, r.StartTime)
	if err != nil {
		return Appointment{}, ErrInvalidAppointment
	}
	end, err := time.Parse(time.RFC3339, r.EndTime)
	if err != nil {
		return Appointment{}, ErrInvalidAppointment
	}
	return Appointment{
		ChairID:   r.ChairID,
		DentistID: r.DentistID,
		PatientID: r.PatientID,
		StartTime: start,
		EndTime:   end,
		State:     r.State,
		Notes:     r.Notes,
	}, nil
}

// Response conversion methods
func operatoryResponseFrom(o Operatory) operatoryResponse {
	return operatoryResponse{
		ID:          o.ID,
		Name:        o.Name,
		Description: o.Description,
		CreatedAt:   o.CreatedAt,
		UpdatedAt:   o.UpdatedAt,
	}
}

func chairResponseFrom(c Chair) chairResponse {
	return chairResponse{
		ID:          c.ID,
		OperatoryID: c.OperatoryID,
		Name:        c.Name,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}
}

func dentistResponseFrom(d Dentist) dentistResponse {
	return dentistResponse{
		ID:        d.ID,
		FirstName: d.FirstName,
		LastName:  d.LastName,
		Specialty: d.Specialty,
		Active:    d.Active,
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
	}
}

func appointmentResponseFrom(a Appointment) appointmentResponse {
	return appointmentResponse{
		ID:        a.ID,
		ChairID:   a.ChairID,
		DentistID: a.DentistID,
		PatientID: a.PatientID,
		StartTime: a.StartTime.Format(time.RFC3339),
		EndTime:   a.EndTime.Format(time.RFC3339),
		State:     a.State,
		Notes:     a.Notes,
		CreatedAt: a.CreatedAt,
		UpdatedAt: a.UpdatedAt,
	}
}

func appointmentSummaryResponseFrom(a AppointmentSummary) appointmentSummaryResponse {
	return appointmentSummaryResponse{
		ID:          a.ID,
		ChairID:     a.ChairID,
		ChairName:   a.ChairName,
		DentistID:   a.DentistID,
		DentistName: a.DentistName,
		PatientID:   a.PatientID,
		PatientName: a.PatientName,
		StartTime:   a.StartTime.Format(time.RFC3339),
		EndTime:     a.EndTime.Format(time.RFC3339),
		State:       a.State,
	}
}
