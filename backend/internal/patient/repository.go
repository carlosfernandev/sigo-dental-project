package patient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/carlosfpatino/sigo-dental/platform/audit"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	Create(ctx context.Context, actorID, sourceIP string, patient Patient) (Patient, error)
	List(ctx context.Context, actorID, sourceIP string) ([]PatientSummary, error)
	Get(ctx context.Context, actorID, sourceIP, patientID string) (Patient, error)
	Update(ctx context.Context, actorID, sourceIP string, patient Patient) (Patient, error)
}

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Create(ctx context.Context, actorID, sourceIP string, patient Patient) (Patient, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return Patient{}, fmt.Errorf("begin patient create transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	args := patientColumns(patient)
	err = tx.QueryRow(ctx, `
		INSERT INTO patients (
			first_name, last_name, date_of_birth, email, phone, address,
			emergency_contact_name, emergency_contact_phone,
			legal_guardian_name, legal_guardian_identification, legal_guardian_phone, insurance_plan
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		RETURNING id::text, created_at, updated_at
	`, args...).Scan(&patient.ID, &patient.CreatedAt, &patient.UpdatedAt)
	if err != nil {
		return Patient{}, fmt.Errorf("insert patient: %w", err)
	}
	if err := insertMedicalHistory(ctx, tx, patient.ID, patient.MedicalHistory); err != nil {
		return Patient{}, err
	}
	if err := audit.Record(ctx, tx, audit.Event{
		ActorID: actorID, Entity: "patient", EntityID: patient.ID, Action: "created",
		Fields: []string{"demographics", "contact", "emergency_contact", "guardian", "insurance"}, SourceIP: sourceIP,
	}); err != nil {
		return Patient{}, err
	}
	if err := audit.Record(ctx, tx, audit.Event{
		ActorID: actorID, Entity: "medical_history", EntityID: patient.ID, Action: "created",
		Fields: []string{"systemic_diseases", "allergies", "surgical_history", "coagulopathies"}, SourceIP: sourceIP,
	}); err != nil {
		return Patient{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Patient{}, fmt.Errorf("commit patient create transaction: %w", err)
	}
	return patient, nil
}

func (r *PostgresRepository) List(ctx context.Context, actorID, sourceIP string) ([]PatientSummary, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin patient list transaction: %w", err)
	}
	defer tx.Rollback(ctx)
	rows, err := tx.Query(ctx, `
		SELECT id::text, first_name, last_name, date_of_birth, phone
		FROM patients
		ORDER BY lower(last_name), lower(first_name), id
	`)
	if err != nil {
		return nil, fmt.Errorf("list patients: %w", err)
	}
	defer rows.Close()

	patients := make([]PatientSummary, 0)
	for rows.Next() {
		var patient PatientSummary
		if err := rows.Scan(&patient.ID, &patient.FirstName, &patient.LastName, &patient.DateOfBirth, &patient.Phone); err != nil {
			return nil, fmt.Errorf("scan patient summary: %w", err)
		}
		patients = append(patients, patient)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read patient summaries: %w", err)
	}
	rows.Close()
	for _, patient := range patients {
		if err := audit.Record(ctx, tx, audit.Event{
			ActorID: actorID, Entity: "patient", EntityID: patient.ID, Action: "read",
			Fields: []string{"patient_summary"}, SourceIP: sourceIP,
		}); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit patient list transaction: %w", err)
	}
	return patients, nil
}

func (r *PostgresRepository) Get(ctx context.Context, actorID, sourceIP, patientID string) (Patient, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return Patient{}, fmt.Errorf("begin patient read transaction: %w", err)
	}
	defer tx.Rollback(ctx)
	patient, err := scanPatient(tx.QueryRow(ctx, patientQuery+` WHERE p.id = $1::uuid`, patientID))
	if errors.Is(err, pgx.ErrNoRows) {
		return Patient{}, ErrPatientNotFound
	}
	if err != nil {
		return Patient{}, fmt.Errorf("get patient: %w", err)
	}
	for _, entity := range []string{"patient", "medical_history"} {
		if err := audit.Record(ctx, tx, audit.Event{
			ActorID: actorID, Entity: entity, EntityID: patient.ID, Action: "read",
			Fields: []string{"record"}, SourceIP: sourceIP,
		}); err != nil {
			return Patient{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return Patient{}, fmt.Errorf("commit patient read transaction: %w", err)
	}
	return patient, nil
}

func (r *PostgresRepository) Update(ctx context.Context, actorID, sourceIP string, patient Patient) (Patient, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return Patient{}, fmt.Errorf("begin patient update transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	args := patientColumns(patient)
	args = append(args, patient.ID)
	tag, err := tx.Exec(ctx, `
		UPDATE patients SET
			first_name=$1, last_name=$2, date_of_birth=$3, email=$4, phone=$5, address=$6,
			emergency_contact_name=$7, emergency_contact_phone=$8,
			legal_guardian_name=$9, legal_guardian_identification=$10, legal_guardian_phone=$11,
			insurance_plan=$12, updated_at=now()
		WHERE id=$13::uuid
	`, args...)
	if err != nil {
		return Patient{}, fmt.Errorf("update patient: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return Patient{}, ErrPatientNotFound
	}
	if _, err := tx.Exec(ctx, `
		UPDATE patient_medical_histories SET systemic_diseases=$2, allergies=$3,
			surgical_history=$4, coagulopathies=$5, updated_at=now()
		WHERE patient_id=$1::uuid
	`, patient.ID, jsonValue(patient.MedicalHistory.SystemicDiseases), jsonValue(patient.MedicalHistory.Allergies),
		jsonValue(patient.MedicalHistory.SurgicalHistory), jsonValue(patient.MedicalHistory.Coagulopathies)); err != nil {
		return Patient{}, fmt.Errorf("update patient medical history: %w", err)
	}
	if err := audit.Record(ctx, tx, audit.Event{
		ActorID: actorID, Entity: "patient", EntityID: patient.ID, Action: "updated",
		Fields: []string{"demographics", "contact", "emergency_contact", "guardian", "insurance"}, SourceIP: sourceIP,
	}); err != nil {
		return Patient{}, err
	}
	if err := audit.Record(ctx, tx, audit.Event{
		ActorID: actorID, Entity: "medical_history", EntityID: patient.ID, Action: "updated",
		Fields: []string{"systemic_diseases", "allergies", "surgical_history", "coagulopathies"}, SourceIP: sourceIP,
	}); err != nil {
		return Patient{}, err
	}
	updated, err := scanPatient(tx.QueryRow(ctx, patientQuery+` WHERE p.id = $1::uuid`, patient.ID))
	if err != nil {
		return Patient{}, fmt.Errorf("read updated patient: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return Patient{}, fmt.Errorf("commit patient update transaction: %w", err)
	}
	return updated, nil
}

const patientQuery = `
	SELECT p.id::text, p.first_name, p.last_name, p.date_of_birth, p.email, p.phone, p.address,
	       p.emergency_contact_name, p.emergency_contact_phone, p.legal_guardian_name,
	       p.legal_guardian_identification, p.legal_guardian_phone, p.insurance_plan,
	       p.created_at, p.updated_at, mh.systemic_diseases, mh.allergies,
	       mh.surgical_history, mh.coagulopathies
	FROM patients p
	JOIN patient_medical_histories mh ON mh.patient_id = p.id
`

type rowScanner interface {
	Scan(dest ...any) error
}

func scanPatient(row rowScanner) (Patient, error) {
	var patient Patient
	var guardianName, guardianID, guardianPhone *string
	var systemicDiseases, allergies, surgicalHistory, coagulopathies []byte
	if err := row.Scan(
		&patient.ID, &patient.FirstName, &patient.LastName, &patient.DateOfBirth, &patient.Email, &patient.Phone,
		&patient.Address, &patient.EmergencyContactName, &patient.EmergencyContactPhone,
		&guardianName, &guardianID, &guardianPhone, &patient.InsurancePlan, &patient.CreatedAt, &patient.UpdatedAt,
		&systemicDiseases, &allergies, &surgicalHistory, &coagulopathies,
	); err != nil {
		return Patient{}, err
	}
	if guardianName != nil {
		patient.LegalGuardian = &Guardian{Name: *guardianName, Identification: *guardianID, Phone: *guardianPhone}
	}
	if err := json.Unmarshal(systemicDiseases, &patient.MedicalHistory.SystemicDiseases); err != nil {
		return Patient{}, fmt.Errorf("decode systemic diseases: %w", err)
	}
	if err := json.Unmarshal(allergies, &patient.MedicalHistory.Allergies); err != nil {
		return Patient{}, fmt.Errorf("decode allergies: %w", err)
	}
	if err := json.Unmarshal(surgicalHistory, &patient.MedicalHistory.SurgicalHistory); err != nil {
		return Patient{}, fmt.Errorf("decode surgical history: %w", err)
	}
	if err := json.Unmarshal(coagulopathies, &patient.MedicalHistory.Coagulopathies); err != nil {
		return Patient{}, fmt.Errorf("decode coagulopathies: %w", err)
	}
	return patient, nil
}

func patientColumns(patient Patient) []any {
	var guardianName, guardianID, guardianPhone any
	if patient.LegalGuardian != nil {
		guardianName = patient.LegalGuardian.Name
		guardianID = patient.LegalGuardian.Identification
		guardianPhone = patient.LegalGuardian.Phone
	}
	return []any{
		patient.FirstName, patient.LastName, patient.DateOfBirth, patient.Email, patient.Phone,
		patient.Address, patient.EmergencyContactName, patient.EmergencyContactPhone,
		guardianName, guardianID, guardianPhone, patient.InsurancePlan,
	}
}

func insertMedicalHistory(ctx context.Context, tx pgx.Tx, patientID string, history MedicalHistory) error {
	if _, err := tx.Exec(ctx, `
		INSERT INTO patient_medical_histories (patient_id, systemic_diseases, allergies, surgical_history, coagulopathies)
		VALUES ($1::uuid, $2, $3, $4, $5)
	`, patientID, jsonValue(history.SystemicDiseases), jsonValue(history.Allergies),
		jsonValue(history.SurgicalHistory), jsonValue(history.Coagulopathies)); err != nil {
		return fmt.Errorf("insert patient medical history: %w", err)
	}
	return nil
}

func jsonValue(value any) []byte {
	encoded, _ := json.Marshal(value)
	return encoded
}
