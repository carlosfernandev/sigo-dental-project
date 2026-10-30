package patient

import (
	"time"
)

type patientRequest struct {
	FirstName             string          `json:"first_name"`
	LastName              string          `json:"last_name"`
	Identification        string          `json:"identification"`
	DateOfBirth           string          `json:"date_of_birth"`
	Email                 string          `json:"email"`
	Phone                 string          `json:"phone"`
	Address               string          `json:"address"`
	EmergencyContactName  string          `json:"emergency_contact_name"`
	EmergencyContactPhone string          `json:"emergency_contact_phone"`
	LegalGuardian         *Guardian       `json:"legal_guardian"`
	InsurancePlan         string          `json:"insurance_plan"`
	MedicalHistory        *MedicalHistory `json:"medical_history"`
}

type patientResponse struct {
	ID                    string         `json:"id"`
	FirstName             string         `json:"first_name"`
	LastName              string         `json:"last_name"`
	Identification        string         `json:"identification"`
	DateOfBirth           string         `json:"date_of_birth"`
	Email                 string         `json:"email"`
	Phone                 string         `json:"phone"`
	Address               string         `json:"address"`
	EmergencyContactName  string         `json:"emergency_contact_name"`
	EmergencyContactPhone string         `json:"emergency_contact_phone"`
	LegalGuardian         *Guardian      `json:"legal_guardian,omitempty"`
	InsurancePlan         string         `json:"insurance_plan"`
	MedicalHistory        MedicalHistory `json:"medical_history"`
	CriticalAlerts        []string       `json:"critical_alerts"`
	CreatedAt             time.Time      `json:"created_at"`
	UpdatedAt             time.Time      `json:"updated_at"`
}

type patientSummaryResponse struct {
	ID             string `json:"id"`
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	Identification string `json:"identification"`
	DateOfBirth    string `json:"date_of_birth"`
	Phone          string `json:"phone"`
	InsurancePlan  string `json:"insurance_plan"`
}

func (r patientRequest) entity() (Patient, error) {
	dob, err := time.Parse("2006-01-02", r.DateOfBirth)
	if err != nil || r.MedicalHistory == nil {
		return Patient{}, ErrInvalidPatient
	}
	return Patient{
		FirstName: r.FirstName, LastName: r.LastName, Identification: r.Identification, DateOfBirth: dob,
		Email: r.Email, Phone: r.Phone, Address: r.Address,
		EmergencyContactName: r.EmergencyContactName, EmergencyContactPhone: r.EmergencyContactPhone,
		LegalGuardian: r.LegalGuardian, InsurancePlan: r.InsurancePlan,
		MedicalHistory: *r.MedicalHistory,
	}, nil
}

func response(patient Patient) patientResponse {
	return patientResponse{
		ID: patient.ID, FirstName: patient.FirstName, LastName: patient.LastName, Identification: patient.Identification,
		DateOfBirth: patient.DateOfBirth.Format("2006-01-02"), Email: patient.Email,
		Phone: patient.Phone, Address: patient.Address,
		EmergencyContactName: patient.EmergencyContactName, EmergencyContactPhone: patient.EmergencyContactPhone,
		LegalGuardian: patient.LegalGuardian, InsurancePlan: patient.InsurancePlan,
		MedicalHistory: patient.MedicalHistory, CriticalAlerts: patient.MedicalHistory.CriticalAlerts(),
		CreatedAt: patient.CreatedAt, UpdatedAt: patient.UpdatedAt,
	}
}
