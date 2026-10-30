package patient

import (
	"errors"
	"testing"
	"time"
)

func TestValidateRequiresGuardianForMinorAndMedicalHistory(t *testing.T) {
	now := time.Date(2026, 10, 30, 12, 0, 0, 0, time.FixedZone("America/Guayaquil", -5*60*60))
	minor := validPatient(now)
	minor.DateOfBirth = time.Date(2015, 3, 4, 0, 0, 0, 0, time.UTC)
	if err := Validate(minor, now); !errors.Is(err, ErrInvalidPatient) {
		t.Fatalf("Validate(minor without guardian) = %v, want %v", err, ErrInvalidPatient)
	}
	minor.LegalGuardian = &Guardian{Name: "Legal Guardian", Identification: "1234567890", Phone: "+593991234567"}
	if err := Validate(minor, now); err != nil {
		t.Fatalf("Validate(minor with guardian) error = %v", err)
	}
	missingHistory := validPatient(now)
	missingHistory.MedicalHistory = MedicalHistory{}
	if err := Validate(missingHistory, now); err != nil {
		t.Fatalf("empty but provided medical history should be valid: %v", err)
	}
}

func TestMedicalHistoryCriticalAlerts(t *testing.T) {
	history := MedicalHistory{
		Allergies:      []Allergy{{Name: "Penicillin", IsCritical: true}, {Name: "Latex"}},
		Coagulopathies: []string{"anticoagulant treatment"},
	}
	alerts := history.CriticalAlerts()
	if len(alerts) != 2 || alerts[0] != "Alergia: Penicillin" || alerts[1] != "Coagulopatía: anticoagulant treatment" {
		t.Fatalf("CriticalAlerts() = %#v", alerts)
	}
}

func validPatient(now time.Time) Patient {
	return Patient{
		FirstName: "Ada", LastName: "Dentist", DateOfBirth: time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
		Email: "ada@example.com", Phone: "+593991234567", EmergencyContactName: "Grace Dentist",
		EmergencyContactPhone: "+593991234568", MedicalHistory: MedicalHistory{},
	}
}
