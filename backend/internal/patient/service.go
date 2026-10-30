package patient

import (
	"context"
	"fmt"
	"net"
	"regexp"
	"strings"
	"time"

	"github.com/carlosfpatino/sigo-dental/platform/authn"
)

var patientIDPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

type Service struct {
	repository Repository
	now        func() time.Time
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository, now: func() time.Time {
		return time.Now().In(time.FixedZone("America/Guayaquil", -5*60*60))
	}}
}

func (s *Service) Create(ctx context.Context, principal authn.Principal, sourceIP string, patient Patient) (Patient, error) {
	if err := requirePatientAccess(principal); err != nil {
		return Patient{}, err
	}
	if !patientIDPattern.MatchString(principal.UserID) || net.ParseIP(sourceIP) == nil {
		return Patient{}, ErrInvalidPatient
	}
	patient = normalize(patient)
	if err := Validate(patient, s.now()); err != nil {
		return Patient{}, err
	}
	created, err := s.repository.Create(ctx, principal.UserID, sourceIP, patient)
	if err != nil {
		return Patient{}, fmt.Errorf("create patient: %w", err)
	}
	return created, nil
}

func (s *Service) List(ctx context.Context, principal authn.Principal, sourceIP string) ([]PatientSummary, error) {
	if err := requirePatientAccess(principal); err != nil {
		return nil, err
	}
	if !patientIDPattern.MatchString(principal.UserID) || net.ParseIP(sourceIP) == nil {
		return nil, ErrInvalidPatient
	}
	patients, err := s.repository.List(ctx, principal.UserID, sourceIP)
	if err != nil {
		return nil, fmt.Errorf("list patients: %w", err)
	}
	return patients, nil
}

func (s *Service) Get(ctx context.Context, principal authn.Principal, sourceIP, patientID string) (Patient, error) {
	if err := requirePatientAccess(principal); err != nil {
		return Patient{}, err
	}
	if !patientIDPattern.MatchString(principal.UserID) || net.ParseIP(sourceIP) == nil || !patientIDPattern.MatchString(patientID) {
		return Patient{}, ErrInvalidPatient
	}
	patient, err := s.repository.Get(ctx, principal.UserID, sourceIP, patientID)
	if err != nil {
		return Patient{}, fmt.Errorf("get patient: %w", err)
	}
	return patient, nil
}

func (s *Service) Update(ctx context.Context, principal authn.Principal, sourceIP string, patient Patient) (Patient, error) {
	if err := requirePatientAccess(principal); err != nil {
		return Patient{}, err
	}
	if !patientIDPattern.MatchString(principal.UserID) || net.ParseIP(sourceIP) == nil || !patientIDPattern.MatchString(patient.ID) {
		return Patient{}, ErrInvalidPatient
	}
	patient = normalize(patient)
	if err := Validate(patient, s.now()); err != nil {
		return Patient{}, err
	}
	updated, err := s.repository.Update(ctx, principal.UserID, sourceIP, patient)
	if err != nil {
		return Patient{}, fmt.Errorf("update patient: %w", err)
	}
	return updated, nil
}

func requirePatientAccess(principal authn.Principal) error {
	return authn.RequireAnyRole(principal,
		authn.RoleAdminTI,
		authn.RoleReceptionistCashier,
		authn.RoleDentist,
		authn.RoleDentalAssistant,
	)
}

func normalize(patient Patient) Patient {
	patient.FirstName = strings.TrimSpace(patient.FirstName)
	patient.LastName = strings.TrimSpace(patient.LastName)
	patient.Identification = strings.TrimSpace(patient.Identification)
	patient.Email = strings.ToLower(strings.TrimSpace(patient.Email))
	patient.Phone = strings.TrimSpace(patient.Phone)
	patient.Address = strings.TrimSpace(patient.Address)
	patient.EmergencyContactName = strings.TrimSpace(patient.EmergencyContactName)
	patient.EmergencyContactPhone = strings.TrimSpace(patient.EmergencyContactPhone)
	patient.InsurancePlan = strings.TrimSpace(patient.InsurancePlan)
	if patient.LegalGuardian != nil {
		patient.LegalGuardian.Name = strings.TrimSpace(patient.LegalGuardian.Name)
		patient.LegalGuardian.Identification = strings.TrimSpace(patient.LegalGuardian.Identification)
		patient.LegalGuardian.Phone = strings.TrimSpace(patient.LegalGuardian.Phone)
	}
	patient.MedicalHistory.SystemicDiseases = normalizeStrings(patient.MedicalHistory.SystemicDiseases)
	patient.MedicalHistory.SurgicalHistory = normalizeStrings(patient.MedicalHistory.SurgicalHistory)
	patient.MedicalHistory.Coagulopathies = normalizeStrings(patient.MedicalHistory.Coagulopathies)
	for i := range patient.MedicalHistory.Allergies {
		patient.MedicalHistory.Allergies[i].Name = strings.TrimSpace(patient.MedicalHistory.Allergies[i].Name)
	}
	if patient.MedicalHistory.SystemicDiseases == nil {
		patient.MedicalHistory.SystemicDiseases = []string{}
	}
	if patient.MedicalHistory.Allergies == nil {
		patient.MedicalHistory.Allergies = []Allergy{}
	}
	if patient.MedicalHistory.SurgicalHistory == nil {
		patient.MedicalHistory.SurgicalHistory = []string{}
	}
	if patient.MedicalHistory.Coagulopathies == nil {
		patient.MedicalHistory.Coagulopathies = []string{}
	}
	return patient
}

func normalizeStrings(values []string) []string {
	for i := range values {
		values[i] = strings.TrimSpace(values[i])
	}
	return values
}
