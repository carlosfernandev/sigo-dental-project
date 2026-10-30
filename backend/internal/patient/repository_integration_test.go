package patient

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"os"
	"testing"
	"time"

	"github.com/carlosfpatino/sigo-dental/platform/migrate"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestPatientRepositoryPostgresIntegration(t *testing.T) {
	dsn := os.Getenv("SIGO_TEST_PG_DSN")
	if dsn == "" {
		t.Skip("SIGO_TEST_PG_DSN is not set")
	}
	ctx := context.Background()
	adminPool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("connect to integration database: %v", err)
	}
	t.Cleanup(adminPool.Close)

	var random [12]byte
	if _, err := rand.Read(random[:]); err != nil {
		t.Fatalf("generate test schema name: %v", err)
	}
	schema := "patient_test_" + hex.EncodeToString(random[:])
	if _, err := adminPool.Exec(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatalf("create isolated test schema: %v", err)
	}
	t.Cleanup(func() {
		if _, err := adminPool.Exec(context.Background(), "DROP SCHEMA "+schema+" CASCADE"); err != nil {
			t.Errorf("drop isolated test schema: %v", err)
		}
	})
	p_config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatalf("parse integration database DSN: %v", err)
	}
	p_config.ConnConfig.RuntimeParams["search_path"] = schema
	db, err := pgxpool.NewWithConfig(ctx, p_config)
	if err != nil {
		t.Fatalf("connect to isolated test schema: %v", err)
	}
	t.Cleanup(db.Close)
	if err := migrate.Up(ctx, db, os.DirFS("../../migrations")); err != nil {
		t.Fatalf("apply monolith migrations: %v", err)
	}
	var actorID string
	if err := db.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, role_id)
		SELECT 'patient-auditor@example.com', 'test-hash', id FROM roles WHERE name = 'odontologo'
		RETURNING id::text
	`).Scan(&actorID); err != nil {
		t.Fatalf("insert audit actor: %v", err)
	}

	repository := NewPostgresRepository(db)
	patient := Patient{
		FirstName: "Ada", LastName: "Lovelace", DateOfBirth: time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
		Email: "ada@example.com", Phone: "+593991234567", Address: "Quito",
		EmergencyContactName: "Grace Lovelace", EmergencyContactPhone: "+593991234568",
		MedicalHistory: MedicalHistory{
			SystemicDiseases: []string{"Asthma"}, Allergies: []Allergy{{Name: "Penicillin", IsCritical: true}},
			SurgicalHistory: []string{}, Coagulopathies: []string{},
		},
	}
	created, err := repository.Create(ctx, actorID, "203.0.113.9", patient)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	loaded, err := repository.Get(ctx, actorID, "203.0.113.9", created.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if loaded.MedicalHistory.Allergies[0].Name != "Penicillin" || loaded.FirstName != "Ada" {
		t.Fatalf("Get() = %+v, patient fields/history did not round trip", loaded)
	}
	loaded.Phone = "+593991234569"
	loaded.MedicalHistory.Coagulopathies = []string{"Anticoagulant treatment"}
	updated, err := repository.Update(ctx, actorID, "203.0.113.9", loaded)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if updated.Phone != "+593991234569" || len(updated.MedicalHistory.CriticalAlerts()) != 2 {
		t.Fatalf("Update() = %+v", updated)
	}
	list, err := repository.List(ctx, actorID, "203.0.113.9")
	if err != nil || len(list) != 1 {
		t.Fatalf("List() = %d rows, error = %v", len(list), err)
	}
	var eventCount int
	if err := db.QueryRow(ctx, "SELECT count(*) FROM clinical_audit_events WHERE entity_id = $1::uuid", created.ID).Scan(&eventCount); err != nil {
		t.Fatalf("count audit events: %v", err)
	}
	if eventCount != 7 {
		t.Fatalf("audit event count = %d, want 7", eventCount)
	}
	if _, err := db.Exec(ctx, "DELETE FROM clinical_audit_events WHERE entity_id = $1::uuid", created.ID); err == nil {
		t.Fatal("audit deletion succeeded, want immutable audit log")
	}
}
