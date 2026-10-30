// Package patient owns patient identity and general medical history.
package patient

import (
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"
)

var ErrInvalidPatient = errors.New("invalid patient")
var ErrPatientNotFound = errors.New("patient not found")

type Guardian struct {
	Name           string `json:"name"`
	Identification string `json:"identification"`
	Phone          string `json:"phone"`
}

type Allergy struct {
	Name       string `json:"name"`
	IsCritical bool   `json:"is_critical"`
}

type MedicalHistory struct {
	SystemicDiseases []string  `json:"systemic_diseases"`
	Allergies        []Allergy `json:"allergies"`
	SurgicalHistory  []string  `json:"surgical_history"`
	Coagulopathies   []string  `json:"coagulopathies"`
}

func (h MedicalHistory) CriticalAlerts() []string {
	alerts := make([]string, 0)
	for _, allergy := range h.Allergies {
		if allergy.IsCritical {
			alerts = append(alerts, "Alergia: "+allergy.Name)
		}
	}
	for _, condition := range h.Coagulopathies {
		alerts = append(alerts, "Coagulopatía: "+condition)
	}
	return alerts
}

type Patient struct {
	ID                    string
	FirstName             string
	LastName              string
	DateOfBirth           time.Time
	Email                 string
	Phone                 string
	Address               string
	EmergencyContactName  string
	EmergencyContactPhone string
	LegalGuardian         *Guardian
	InsurancePlan         string
	MedicalHistory        MedicalHistory
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

type PatientSummary struct {
	ID          string
	FirstName   string
	LastName    string
	DateOfBirth time.Time
	Phone       string
}

func Validate(p Patient, now time.Time) error {
	p.FirstName = strings.TrimSpace(p.FirstName)
	p.LastName = strings.TrimSpace(p.LastName)
	p.Email = strings.TrimSpace(p.Email)
	p.Phone = strings.TrimSpace(p.Phone)
	p.EmergencyContactName = strings.TrimSpace(p.EmergencyContactName)
	p.EmergencyContactPhone = strings.TrimSpace(p.EmergencyContactPhone)
	if p.FirstName == "" || len(p.FirstName) > 100 || p.LastName == "" || len(p.LastName) > 100 ||
		p.DateOfBirth.IsZero() || p.DateOfBirth.After(now) || p.Phone == "" || len(p.Phone) > 32 ||
		p.EmergencyContactName == "" || len(p.EmergencyContactName) > 120 ||
		p.EmergencyContactPhone == "" || len(p.EmergencyContactPhone) > 32 || len(p.Address) > 300 || len(p.InsurancePlan) > 200 {
		return ErrInvalidPatient
	}
	if p.Email != "" {
		parsed, err := mail.ParseAddress(p.Email)
		if err != nil || parsed.Address != p.Email || len(p.Email) > 254 {
			return ErrInvalidPatient
		}
	}
	if p.LegalGuardian != nil {
		p.LegalGuardian.Name = strings.TrimSpace(p.LegalGuardian.Name)
		p.LegalGuardian.Identification = strings.TrimSpace(p.LegalGuardian.Identification)
		p.LegalGuardian.Phone = strings.TrimSpace(p.LegalGuardian.Phone)
		if p.LegalGuardian.Name == "" || len(p.LegalGuardian.Name) > 120 || p.LegalGuardian.Identification == "" ||
			len(p.LegalGuardian.Identification) > 32 || p.LegalGuardian.Phone == "" || len(p.LegalGuardian.Phone) > 32 {
			return ErrInvalidPatient
		}
	}
	birthDate := dateOnly(p.DateOfBirth)
	adultOn := dateOnly(now).AddDate(-18, 0, 0)
	if birthDate.After(adultOn) && p.LegalGuardian == nil {
		return fmt.Errorf("%w: legal guardian is required for a minor", ErrInvalidPatient)
	}
	if err := validateHistory(p.MedicalHistory); err != nil {
		return err
	}
	return nil
}

func validateHistory(h MedicalHistory) error {
	if len(h.SystemicDiseases) > 50 || len(h.Allergies) > 50 || len(h.SurgicalHistory) > 50 || len(h.Coagulopathies) > 50 {
		return ErrInvalidPatient
	}
	for _, item := range [][]string{h.SystemicDiseases, h.SurgicalHistory, h.Coagulopathies} {
		for _, value := range item {
			if strings.TrimSpace(value) == "" || len(value) > 200 {
				return ErrInvalidPatient
			}
		}
	}
	for _, allergy := range h.Allergies {
		if strings.TrimSpace(allergy.Name) == "" || len(allergy.Name) > 200 {
			return ErrInvalidPatient
		}
	}
	return nil
}

func dateOnly(value time.Time) time.Time {
	year, month, day := value.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}
