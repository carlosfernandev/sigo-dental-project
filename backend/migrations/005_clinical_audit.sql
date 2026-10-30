CREATE TABLE clinical_audit_events (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_user_id uuid NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    entity_type text NOT NULL CHECK (entity_type IN ('patient', 'medical_history', 'clinical_record')),
    entity_id uuid NOT NULL,
    action text NOT NULL CHECK (action IN ('read', 'created', 'updated', 'exported')),
    fields text[] NOT NULL DEFAULT '{}',
    occurred_at timestamptz NOT NULL DEFAULT now(),
    source_ip inet NOT NULL
);

CREATE INDEX clinical_audit_entity_idx ON clinical_audit_events (entity_type, entity_id, occurred_at DESC);
CREATE INDEX clinical_audit_actor_idx ON clinical_audit_events (actor_user_id, occurred_at DESC);

CREATE FUNCTION reject_clinical_audit_mutation() RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    RAISE EXCEPTION 'clinical audit events are immutable';
END;
$$;

CREATE TRIGGER clinical_audit_events_immutable
BEFORE UPDATE OR DELETE ON clinical_audit_events
FOR EACH ROW EXECUTE FUNCTION reject_clinical_audit_mutation();
