// Package audit provides the append-only persistence port for clinical audit events.
package audit

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type Event struct {
	ActorID  string
	Entity   string
	EntityID string
	Action   string
	Fields   []string
	SourceIP string
}

// Record appends an audit event in the caller's transaction.
func Record(ctx context.Context, tx pgx.Tx, event Event) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO clinical_audit_events (actor_user_id, entity_type, entity_id, action, fields, source_ip)
		VALUES ($1::uuid, $2, $3::uuid, $4, $5, $6::inet)
	`, event.ActorID, event.Entity, event.EntityID, event.Action, event.Fields, event.SourceIP)
	if err != nil {
		return fmt.Errorf("append clinical audit event: %w", err)
	}
	return nil
}
