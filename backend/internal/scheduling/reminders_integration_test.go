package scheduling

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/carlosfpatino/sigo-dental/platform/authn"
	"github.com/carlosfpatino/sigo-dental/platform/migrate"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestReminderRepositoryPostgresIntegration(t *testing.T) {
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
		t.Fatalf("generate isolated schema name: %v", err)
	}
	schema := "scheduling_test_" + hex.EncodeToString(random[:])
	if _, err := adminPool.Exec(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatalf("create isolated schema: %v", err)
	}
	t.Cleanup(func() {
		if _, err := adminPool.Exec(context.Background(), "DROP SCHEMA "+schema+" CASCADE"); err != nil {
			t.Errorf("drop isolated schema: %v", err)
		}
	})
	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatalf("parse integration DSN: %v", err)
	}
	poolConfig.ConnConfig.RuntimeParams["search_path"] = schema
	db, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		t.Fatalf("connect to isolated schema: %v", err)
	}
	t.Cleanup(db.Close)
	if err := migrate.Up(ctx, db, os.DirFS("../../migrations")); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}

	var userID string
	err = db.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, role_id)
		SELECT 'reminder-reception@example.com', 'test-hash', id FROM roles WHERE name = $1
		RETURNING id::text
	`, authn.RoleReceptionistCashier).Scan(&userID)
	if err != nil {
		t.Fatalf("insert receptionist: %v", err)
	}
	var patientID, operatoryID, chairID, dentistID string
	err = db.QueryRow(ctx, `
		INSERT INTO patients (first_name, last_name, date_of_birth, phone, emergency_contact_name, emergency_contact_phone)
		VALUES ('Ada', 'Lovelace', '1990-01-01', '+593991234567', 'Grace Lovelace', '+593991234568')
		RETURNING id::text
	`).Scan(&patientID)
	if err != nil {
		t.Fatalf("insert patient: %v", err)
	}
	if err := db.QueryRow(ctx, `INSERT INTO operatories (name) VALUES ('North') RETURNING id::text`).Scan(&operatoryID); err != nil {
		t.Fatalf("insert operatory: %v", err)
	}
	if err := db.QueryRow(ctx, `INSERT INTO chairs (operatory_id, name) VALUES ($1::uuid, 'Chair 1') RETURNING id::text`, operatoryID).Scan(&chairID); err != nil {
		t.Fatalf("insert chair: %v", err)
	}
	if err := db.QueryRow(ctx, `INSERT INTO dentists (first_name, last_name) VALUES ('Grace', 'Hopper') RETURNING id::text`).Scan(&dentistID); err != nil {
		t.Fatalf("insert dentist: %v", err)
	}

	now := time.Now().UTC().Truncate(time.Second)
	for _, start := range []time.Time{now.Add(47 * time.Hour), now.Add(23 * time.Hour)} {
		if _, err := db.Exec(ctx, `
			INSERT INTO appointments (chair_id, dentist_id, patient_id, start_time, end_time, state)
			VALUES ($1::uuid, $2::uuid, $3::uuid, $4, $4 + interval '30 minutes', 'confirmed')
		`, chairID, dentistID, patientID, start); err != nil {
			t.Fatalf("insert confirmed appointment: %v", err)
		}
	}

	service := NewService(NewPostgresRepository(db))
	service.now = func() time.Time { return now.In(time.FixedZone("America/Guayaquil", -5*60*60)) }
	principal := authn.Principal{UserID: userID, Role: authn.RoleReceptionistCashier}
	reminders, err := service.ListPendingReminders(ctx, principal)
	if err != nil || len(reminders) != 2 {
		t.Fatalf("ListPendingReminders() = %d records, error %v", len(reminders), err)
	}
	types := map[string]bool{}
	for _, reminder := range reminders {
		types[reminder.ReminderType] = true
		if reminder.PatientName != "Ada Lovelace" || reminder.PatientPhone != "+593991234567" {
			t.Fatalf("reminder patient details = %+v", reminder)
		}
	}
	if !types["48h"] || !types["24h"] {
		t.Fatalf("generated reminder types = %v", types)
	}
	if err := service.GeneratePendingReminders(ctx); err != nil {
		t.Fatalf("repeat generation: %v", err)
	}
	reminders, err = service.ListPendingReminders(ctx, principal)
	if err != nil || len(reminders) != 2 {
		t.Fatalf("repeated generation produced %d records, error %v", len(reminders), err)
	}

	notes := "Contacted by phone"
	updated, err := service.UpdateReminder(ctx, principal, reminders[0].ID, updateReminderRequest{Status: "contacted", Notes: &notes})
	if err != nil || updated.Status != "contacted" || updated.HandledBy == nil || *updated.HandledBy != userID || updated.HandledAt == nil {
		t.Fatalf("UpdateReminder() = %+v, error %v", updated, err)
	}
	if _, err := service.UpdateReminder(ctx, principal, reminders[0].ID, updateReminderRequest{Status: "no_response"}); !errors.Is(err, ErrReminderHandled) {
		t.Fatalf("second update error = %v, want %v", err, ErrReminderHandled)
	}
}
