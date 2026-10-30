CREATE TABLE identity_user_audit_events (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_user_id uuid NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    target_user_id uuid NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    action text NOT NULL CHECK (action IN ('created', 'updated', 'deactivated')),
    reason text,
    previous_values jsonb NOT NULL DEFAULT '{}'::jsonb,
    new_values jsonb NOT NULL DEFAULT '{}'::jsonb,
    occurred_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX identity_user_audit_target_idx
    ON identity_user_audit_events (target_user_id, occurred_at DESC);

CREATE INDEX identity_user_audit_actor_idx
    ON identity_user_audit_events (actor_user_id, occurred_at DESC);

CREATE FUNCTION reject_identity_user_audit_mutation() RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    RAISE EXCEPTION 'identity user audit events are immutable';
END;
$$;

CREATE TRIGGER identity_user_audit_events_immutable
BEFORE UPDATE OR DELETE ON identity_user_audit_events
FOR EACH ROW EXECUTE FUNCTION reject_identity_user_audit_mutation();
