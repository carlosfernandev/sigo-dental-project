// Package scheduling owns operatory, chair, dentist, and appointment management.
package scheduling

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	// Operatory
	CreateOperatory(ctx context.Context, o Operatory) (Operatory, error)
	ListOperatories(ctx context.Context) ([]Operatory, error)
	GetOperatory(ctx context.Context, id string) (Operatory, error)
	UpdateOperatory(ctx context.Context, o Operatory) (Operatory, error)

	// Chair
	CreateChair(ctx context.Context, c Chair) (Chair, error)
	ListChairs(ctx context.Context) ([]Chair, error)
	GetChair(ctx context.Context, id string) (Chair, error)
	UpdateChair(ctx context.Context, c Chair) (Chair, error)

	// Dentist
	CreateDentist(ctx context.Context, d Dentist) (Dentist, error)
	ListDentists(ctx context.Context) ([]Dentist, error)
	GetDentist(ctx context.Context, id string) (Dentist, error)
	UpdateDentist(ctx context.Context, d Dentist) (Dentist, error)

	// Appointment
	CreateAppointment(ctx context.Context, a Appointment) (Appointment, error)
	ListAppointments(ctx context.Context, chairID, dentistID, patientID *string, startDate, endDate *string) ([]AppointmentSummary, error)
	GetAppointment(ctx context.Context, id string) (Appointment, error)
	UpdateAppointment(ctx context.Context, a Appointment) (Appointment, error)

	GeneratePendingReminders(ctx context.Context, now time.Time) error
	ListPendingReminders(ctx context.Context) ([]Reminder, error)
	GetReminder(ctx context.Context, id string) (Reminder, error)
	UpdateReminder(ctx context.Context, id, status, actorID string, notes *string) (Reminder, error)
	CreateSchedulableTasks(ctx context.Context, executor Executor, tasks []SchedulableTask) error
}

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// Operatory operations
func (r *PostgresRepository) CreateOperatory(ctx context.Context, o Operatory) (Operatory, error) {
	err := r.db.QueryRow(ctx, `
		INSERT INTO operatories (name, description)
		VALUES ($1, $2)
		RETURNING id::text, created_at, updated_at
	`, o.Name, o.Description).Scan(&o.ID, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		return Operatory{}, fmt.Errorf("insert operatory: %w", err)
	}
	return o, nil
}

func (r *PostgresRepository) ListOperatories(ctx context.Context) ([]Operatory, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id::text, name, description, created_at, updated_at
		FROM operatories
		ORDER BY lower(name)
	`)
	if err != nil {
		return nil, fmt.Errorf("list operatories: %w", err)
	}
	defer rows.Close()

	var operatories []Operatory
	for rows.Next() {
		var o Operatory
		if err := rows.Scan(&o.ID, &o.Name, &o.Description, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan operatory: %w", err)
		}
		operatories = append(operatories, o)
	}
	return operatories, nil
}

func (r *PostgresRepository) GetOperatory(ctx context.Context, id string) (Operatory, error) {
	var o Operatory
	err := r.db.QueryRow(ctx, `
		SELECT id::text, name, description, created_at, updated_at
		FROM operatories
		WHERE id = $1::uuid
	`, id).Scan(&o.ID, &o.Name, &o.Description, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return Operatory{}, ErrInvalidOperatory
		}
		return Operatory{}, fmt.Errorf("get operatory: %w", err)
	}
	return o, nil
}

func (r *PostgresRepository) UpdateOperatory(ctx context.Context, o Operatory) (Operatory, error) {
	tag, err := r.db.Exec(ctx, `
		UPDATE operatories SET name=$1, description=$2, updated_at=now()
		WHERE id=$3::uuid
	`, o.Name, o.Description, o.ID)
	if err != nil {
		return Operatory{}, fmt.Errorf("update operatory: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return Operatory{}, ErrInvalidOperatory
	}
	return r.GetOperatory(ctx, o.ID)
}

// Chair operations
func (r *PostgresRepository) CreateChair(ctx context.Context, c Chair) (Chair, error) {
	err := r.db.QueryRow(ctx, `
		INSERT INTO chairs (operatory_id, name)
		VALUES ($1::uuid, $2)
		RETURNING id::text, created_at, updated_at
	`, c.OperatoryID, c.Name).Scan(&c.ID, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return Chair{}, fmt.Errorf("insert chair: %w", err)
	}
	return c, nil
}

func (r *PostgresRepository) ListChairs(ctx context.Context) ([]Chair, error) {
	rows, err := r.db.Query(ctx, `
		SELECT c.id::text, c.operatory_id::text, c.name, c.created_at, c.updated_at
		FROM chairs c
		ORDER BY c.operatory_id, lower(c.name)
	`)
	if err != nil {
		return nil, fmt.Errorf("list chairs: %w", err)
	}
	defer rows.Close()

	var chairs []Chair
	for rows.Next() {
		var c Chair
		if err := rows.Scan(&c.ID, &c.OperatoryID, &c.Name, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan chair: %w", err)
		}
		chairs = append(chairs, c)
	}
	return chairs, nil
}

func (r *PostgresRepository) GetChair(ctx context.Context, id string) (Chair, error) {
	var c Chair
	err := r.db.QueryRow(ctx, `
		SELECT id::text, operatory_id::text, name, created_at, updated_at
		FROM chairs
		WHERE id = $1::uuid
	`, id).Scan(&c.ID, &c.OperatoryID, &c.Name, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return Chair{}, ErrInvalidChair
		}
		return Chair{}, fmt.Errorf("get chair: %w", err)
	}
	return c, nil
}

func (r *PostgresRepository) UpdateChair(ctx context.Context, c Chair) (Chair, error) {
	tag, err := r.db.Exec(ctx, `
		UPDATE chairs SET operatory_id=$1::uuid, name=$2, updated_at=now()
		WHERE id=$3::uuid
	`, c.OperatoryID, c.Name, c.ID)
	if err != nil {
		return Chair{}, fmt.Errorf("update chair: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return Chair{}, ErrInvalidChair
	}
	return r.GetChair(ctx, c.ID)
}

// Dentist operations
func (r *PostgresRepository) CreateDentist(ctx context.Context, d Dentist) (Dentist, error) {
	err := r.db.QueryRow(ctx, `
		INSERT INTO dentists (first_name, last_name, specialty, active)
		VALUES ($1, $2, $3, $4)
		RETURNING id::text, created_at, updated_at
	`, d.FirstName, d.LastName, d.Specialty, d.Active).Scan(&d.ID, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		return Dentist{}, fmt.Errorf("insert dentist: %w", err)
	}
	return d, nil
}

func (r *PostgresRepository) ListDentists(ctx context.Context) ([]Dentist, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id::text, first_name, last_name, specialty, active, created_at, updated_at
		FROM dentists
		ORDER BY lower(last_name), lower(first_name)
	`)
	if err != nil {
		return nil, fmt.Errorf("list dentists: %w", err)
	}
	defer rows.Close()

	var dentists []Dentist
	for rows.Next() {
		var d Dentist
		if err := rows.Scan(&d.ID, &d.FirstName, &d.LastName, &d.Specialty, &d.Active, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan dentist: %w", err)
		}
		dentists = append(dentists, d)
	}
	return dentists, nil
}

func (r *PostgresRepository) GetDentist(ctx context.Context, id string) (Dentist, error) {
	var d Dentist
	err := r.db.QueryRow(ctx, `
		SELECT id::text, first_name, last_name, specialty, active, created_at, updated_at
		FROM dentists
		WHERE id = $1::uuid
	`, id).Scan(&d.ID, &d.FirstName, &d.LastName, &d.Specialty, &d.Active, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return Dentist{}, ErrInvalidDentist
		}
		return Dentist{}, fmt.Errorf("get dentist: %w", err)
	}
	return d, nil
}

func (r *PostgresRepository) UpdateDentist(ctx context.Context, d Dentist) (Dentist, error) {
	tag, err := r.db.Exec(ctx, `
		UPDATE dentists SET first_name=$1, last_name=$2, specialty=$3, active=$4, updated_at=now()
		WHERE id=$5::uuid
	`, d.FirstName, d.LastName, d.Specialty, d.Active, d.ID)
	if err != nil {
		return Dentist{}, fmt.Errorf("update dentist: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return Dentist{}, ErrInvalidDentist
	}
	return r.GetDentist(ctx, d.ID)
}

// Appointment operations
func (r *PostgresRepository) CreateAppointment(ctx context.Context, a Appointment) (Appointment, error) {
	err := r.db.QueryRow(ctx, `
		INSERT INTO appointments (chair_id, dentist_id, patient_id, start_time, end_time, state, notes)
		VALUES ($1::uuid, $2::uuid, $3::uuid, $4, $5, $6, $7)
		RETURNING id::text, created_at, updated_at
	`, a.ChairID, a.DentistID, a.PatientID, a.StartTime, a.EndTime, a.State, a.Notes).Scan(&a.ID, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return Appointment{}, fmt.Errorf("insert appointment: %w", err)
	}
	return a, nil
}

func (r *PostgresRepository) ListAppointments(ctx context.Context, chairID, dentistID, patientID *string, startDate, endDate *string) ([]AppointmentSummary, error) {
	query := `
		SELECT a.id::text, a.chair_id::text, c.name, a.dentist_id::text,
		       CONCAT(d.first_name, ' ', d.last_name), a.patient_id::text,
		       CONCAT(p.first_name, ' ', p.last_name), a.start_time, a.end_time, a.state
		FROM appointments a
		JOIN chairs c ON c.id = a.chair_id
		JOIN dentists d ON d.id = a.dentist_id
		JOIN patients p ON p.id = a.patient_id
		WHERE 1=1
	`
	args := []any{}
	argIdx := 1

	if chairID != nil {
		query += fmt.Sprintf(" AND a.chair_id = $%d::uuid", argIdx)
		args = append(args, *chairID)
		argIdx++
	}
	if dentistID != nil {
		query += fmt.Sprintf(" AND a.dentist_id = $%d::uuid", argIdx)
		args = append(args, *dentistID)
		argIdx++
	}
	if patientID != nil {
		query += fmt.Sprintf(" AND a.patient_id = $%d::uuid", argIdx)
		args = append(args, *patientID)
		argIdx++
	}
	if startDate != nil {
		query += fmt.Sprintf(" AND a.start_time >= $%d::timestamptz", argIdx)
		args = append(args, *startDate)
		argIdx++
	}
	if endDate != nil {
		query += fmt.Sprintf(" AND a.end_time <= $%d::timestamptz", argIdx)
		args = append(args, *endDate)
		argIdx++
	}

	query += " ORDER BY a.start_time"

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list appointments: %w", err)
	}
	defer rows.Close()

	var appointments []AppointmentSummary
	for rows.Next() {
		var a AppointmentSummary
		if err := rows.Scan(&a.ID, &a.ChairID, &a.ChairName, &a.DentistID, &a.DentistName,
			&a.PatientID, &a.PatientName, &a.StartTime, &a.EndTime, &a.State); err != nil {
			return nil, fmt.Errorf("scan appointment: %w", err)
		}
		appointments = append(appointments, a)
	}
	return appointments, nil
}

func (r *PostgresRepository) GetAppointment(ctx context.Context, id string) (Appointment, error) {
	var a Appointment
	err := r.db.QueryRow(ctx, `
		SELECT id::text, chair_id::text, dentist_id::text, patient_id::text,
		       start_time, end_time, state, notes, created_at, updated_at
		FROM appointments
		WHERE id = $1::uuid
	`, id).Scan(&a.ID, &a.ChairID, &a.DentistID, &a.PatientID, &a.StartTime, &a.EndTime, &a.State, &a.Notes, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return Appointment{}, ErrAppointmentNotFound
		}
		return Appointment{}, fmt.Errorf("get appointment: %w", err)
	}
	return a, nil
}

func (r *PostgresRepository) UpdateAppointment(ctx context.Context, a Appointment) (Appointment, error) {
	tag, err := r.db.Exec(ctx, `
		UPDATE appointments SET chair_id=$1::uuid, dentist_id=$2::uuid, patient_id=$3::uuid,
		                        start_time=$4, end_time=$5, state=$6, notes=$7, updated_at=now()
		WHERE id=$8::uuid
	`, a.ChairID, a.DentistID, a.PatientID, a.StartTime, a.EndTime, a.State, a.Notes, a.ID)
	if err != nil {
		return Appointment{}, fmt.Errorf("update appointment: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return Appointment{}, ErrAppointmentNotFound
	}
	return r.GetAppointment(ctx, a.ID)
}

func (r *PostgresRepository) GeneratePendingReminders(ctx context.Context, now time.Time) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO appointment_reminders (appointment_id, reminder_type)
		SELECT a.id, reminder_window.reminder_type
		FROM appointments AS a
		CROSS JOIN (VALUES
			('48h'::text, $1::timestamptz + interval '24 hours', $1::timestamptz + interval '48 hours'),
			('24h'::text, $1::timestamptz, $1::timestamptz + interval '24 hours')
		) AS reminder_window (reminder_type, lower_bound, upper_bound)
		WHERE a.state = 'confirmed'
		  AND a.start_time > reminder_window.lower_bound
		  AND a.start_time <= reminder_window.upper_bound
		ON CONFLICT (appointment_id, reminder_type) DO NOTHING
	`, now.UTC())
	if err != nil {
		return fmt.Errorf("generate appointment reminders: %w", err)
	}
	return nil
}

func (r *PostgresRepository) ListPendingReminders(ctx context.Context) ([]Reminder, error) {
	rows, err := r.db.Query(ctx, reminderQuery+`
		WHERE r.status = 'pending' AND a.state = 'confirmed'
		ORDER BY a.start_time, r.reminder_type, r.id
	`)
	if err != nil {
		return nil, fmt.Errorf("list pending reminders: %w", err)
	}
	defer rows.Close()
	reminders := make([]Reminder, 0)
	for rows.Next() {
		reminder, err := scanReminder(rows)
		if err != nil {
			return nil, fmt.Errorf("scan pending reminder: %w", err)
		}
		reminders = append(reminders, reminder)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read pending reminders: %w", err)
	}
	return reminders, nil
}

func (r *PostgresRepository) GetReminder(ctx context.Context, id string) (Reminder, error) {
	reminder, err := scanReminder(r.db.QueryRow(ctx, reminderQuery+` WHERE r.id = $1::uuid`, id))
	if err != nil {
		if err == pgx.ErrNoRows {
			return Reminder{}, ErrReminderNotFound
		}
		return Reminder{}, fmt.Errorf("get reminder: %w", err)
	}
	return reminder, nil
}

func (r *PostgresRepository) UpdateReminder(ctx context.Context, id, status, actorID string, notes *string) (Reminder, error) {
	tag, err := r.db.Exec(ctx, `
		UPDATE appointment_reminders
		SET status = $1, notes = COALESCE($2, notes), handled_by = $3::uuid,
		    handled_at = now(), updated_at = now()
		WHERE id = $4::uuid AND status = 'pending'
	`, status, notes, actorID, id)
	if err != nil {
		return Reminder{}, fmt.Errorf("update reminder: %w", err)
	}
	if tag.RowsAffected() == 0 {
		if _, err := r.GetReminder(ctx, id); err != nil {
			return Reminder{}, err
		}
		return Reminder{}, ErrReminderHandled
	}
	reminder, err := r.GetReminder(ctx, id)
	if err != nil {
		return Reminder{}, err
	}
	return reminder, nil
}

func (r *PostgresRepository) CreateSchedulableTasks(ctx context.Context, executor Executor, tasks []SchedulableTask) error {
	for _, task := range tasks {
		if err := executor.Exec(ctx, `
			INSERT INTO schedulable_tasks (patient_id, quote_id, item_id, procedure_name)
			VALUES ($1::uuid, $2::uuid, $3, $4)
		`, task.PatientID, task.QuoteID, task.ItemID, task.ProcedureName); err != nil {
			return fmt.Errorf("insert schedulable task: %w", err)
		}
	}
	return nil
}

const reminderQuery = `
	SELECT r.id::text, r.appointment_id::text, r.reminder_type, r.status, r.notes,
	       a.start_time, a.end_time, p.id::text, p.first_name || ' ' || p.last_name,
	       p.phone, d.first_name || ' ' || d.last_name, c.name,
	       r.handled_by::text, r.handled_at, r.created_at, r.updated_at
	FROM appointment_reminders AS r
	JOIN appointments AS a ON a.id = r.appointment_id
	JOIN patients AS p ON p.id = a.patient_id
	JOIN dentists AS d ON d.id = a.dentist_id
	JOIN chairs AS c ON c.id = a.chair_id
`

type reminderScanner interface {
	Scan(dest ...any) error
}

func scanReminder(row reminderScanner) (Reminder, error) {
	var reminder Reminder
	err := row.Scan(
		&reminder.ID, &reminder.AppointmentID, &reminder.ReminderType, &reminder.Status,
		&reminder.Notes, &reminder.AppointmentStartTime, &reminder.AppointmentEndTime,
		&reminder.PatientID, &reminder.PatientName, &reminder.PatientPhone, &reminder.DentistName,
		&reminder.ChairName, &reminder.HandledBy, &reminder.HandledAt, &reminder.CreatedAt,
		&reminder.UpdatedAt,
	)
	return reminder, err
}
