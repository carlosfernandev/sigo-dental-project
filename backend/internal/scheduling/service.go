// Package scheduling owns operatory, chair, dentist, and appointment management.
package scheduling

import (
	"context"
	"fmt"
	"time"

	"github.com/carlosfpatino/sigo-dental/platform/authn"
)

type Service struct {
	repo Repository
	now  func() time.Time
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
		now: func() time.Time {
			return time.Now().In(time.FixedZone("America/Guayaquil", -5*60*60))
		},
	}
}

// Operatory operations
func (s *Service) CreateOperatory(ctx context.Context, principal authn.Principal, o Operatory) (Operatory, error) {
	if err := requireAdminAccess(principal); err != nil {
		return Operatory{}, err
	}
	if err := ValidateOperatory(o); err != nil {
		return Operatory{}, err
	}
	return s.repo.CreateOperatory(ctx, o)
}

func (s *Service) ListOperatories(ctx context.Context, principal authn.Principal) ([]Operatory, error) {
	if err := requireSchedulingAccess(principal); err != nil {
		return nil, err
	}
	return s.repo.ListOperatories(ctx)
}

func (s *Service) GetOperatory(ctx context.Context, principal authn.Principal, id string) (Operatory, error) {
	if err := requireSchedulingAccess(principal); err != nil {
		return Operatory{}, err
	}
	return s.repo.GetOperatory(ctx, id)
}

func (s *Service) UpdateOperatory(ctx context.Context, principal authn.Principal, o Operatory) (Operatory, error) {
	if err := requireAdminAccess(principal); err != nil {
		return Operatory{}, err
	}
	if err := ValidateOperatory(o); err != nil {
		return Operatory{}, err
	}
	return s.repo.UpdateOperatory(ctx, o)
}

// Chair operations
func (s *Service) CreateChair(ctx context.Context, principal authn.Principal, c Chair) (Chair, error) {
	if err := requireAdminAccess(principal); err != nil {
		return Chair{}, err
	}
	if err := ValidateChair(c); err != nil {
		return Chair{}, err
	}
	return s.repo.CreateChair(ctx, c)
}

func (s *Service) ListChairs(ctx context.Context, principal authn.Principal) ([]Chair, error) {
	if err := requireSchedulingAccess(principal); err != nil {
		return nil, err
	}
	return s.repo.ListChairs(ctx)
}

func (s *Service) GetChair(ctx context.Context, principal authn.Principal, id string) (Chair, error) {
	if err := requireSchedulingAccess(principal); err != nil {
		return Chair{}, err
	}
	return s.repo.GetChair(ctx, id)
}

func (s *Service) UpdateChair(ctx context.Context, principal authn.Principal, c Chair) (Chair, error) {
	if err := requireAdminAccess(principal); err != nil {
		return Chair{}, err
	}
	if err := ValidateChair(c); err != nil {
		return Chair{}, err
	}
	return s.repo.UpdateChair(ctx, c)
}

// Dentist operations
func (s *Service) CreateDentist(ctx context.Context, principal authn.Principal, d Dentist) (Dentist, error) {
	if err := requireAdminAccess(principal); err != nil {
		return Dentist{}, err
	}
	if err := ValidateDentist(d); err != nil {
		return Dentist{}, err
	}
	return s.repo.CreateDentist(ctx, d)
}

func (s *Service) ListDentists(ctx context.Context, principal authn.Principal) ([]Dentist, error) {
	if err := requireSchedulingAccess(principal); err != nil {
		return nil, err
	}
	return s.repo.ListDentists(ctx)
}

func (s *Service) GetDentist(ctx context.Context, principal authn.Principal, id string) (Dentist, error) {
	if err := requireSchedulingAccess(principal); err != nil {
		return Dentist{}, err
	}
	return s.repo.GetDentist(ctx, id)
}

func (s *Service) UpdateDentist(ctx context.Context, principal authn.Principal, d Dentist) (Dentist, error) {
	if err := requireAdminAccess(principal); err != nil {
		return Dentist{}, err
	}
	if err := ValidateDentist(d); err != nil {
		return Dentist{}, err
	}
	return s.repo.UpdateDentist(ctx, d)
}

// Appointment operations
func (s *Service) CreateAppointment(ctx context.Context, principal authn.Principal, a Appointment) (Appointment, error) {
	if err := requireSchedulingAccess(principal); err != nil {
		return Appointment{}, err
	}
	if a.State == "" {
		a.State = AppointmentStateScheduled
	}
	if err := ValidateAppointment(a, s.now()); err != nil {
		return Appointment{}, err
	}
	created, err := s.repo.CreateAppointment(ctx, a)
	if err != nil {
		if isOverlapError(err) {
			return Appointment{}, ErrChairOverlap
		}
		return Appointment{}, fmt.Errorf("create appointment: %w", err)
	}
	return created, nil
}

func (s *Service) ListAppointments(ctx context.Context, principal authn.Principal, chairID, dentistID, patientID *string, startDate, endDate *string) ([]AppointmentSummary, error) {
	if err := requireSchedulingAccess(principal); err != nil {
		return nil, err
	}
	return s.repo.ListAppointments(ctx, chairID, dentistID, patientID, startDate, endDate)
}

func (s *Service) GetAppointment(ctx context.Context, principal authn.Principal, id string) (Appointment, error) {
	if err := requireSchedulingAccess(principal); err != nil {
		return Appointment{}, err
	}
	return s.repo.GetAppointment(ctx, id)
}

func (s *Service) UpdateAppointment(ctx context.Context, principal authn.Principal, a Appointment) (Appointment, error) {
	if err := requireSchedulingAccess(principal); err != nil {
		return Appointment{}, err
	}
	if err := ValidateAppointment(a, s.now()); err != nil {
		return Appointment{}, err
	}
	existing, err := s.repo.GetAppointment(ctx, a.ID)
	if err != nil {
		return Appointment{}, err
	}
	if !ValidateStateTransition(existing.State, a.State) {
		return Appointment{}, fmt.Errorf("%w: invalid state transition from %s to %s", ErrInvalidAppointment, existing.State, a.State)
	}
	updated, err := s.repo.UpdateAppointment(ctx, a)
	if err != nil {
		if isOverlapError(err) {
			return Appointment{}, ErrChairOverlap
		}
		return Appointment{}, fmt.Errorf("update appointment: %w", err)
	}
	return updated, nil
}

func requireAdminAccess(principal authn.Principal) error {
	return authn.RequireAnyRole(principal, authn.RoleAdminTI)
}

func requireSchedulingAccess(principal authn.Principal) error {
	return authn.RequireAnyRole(principal,
		authn.RoleReceptionistCashier,
		authn.RoleDentist,
		authn.RoleDentalAssistant,
	)
}

func isOverlapError(err error) bool {
	// PostgreSQL EXCLUDE constraint violation
	return err != nil && (err.Error() == "conflicting key value violates exclusion constraint" ||
		err.Error() == "could not serialize access due to concurrent update")
}
