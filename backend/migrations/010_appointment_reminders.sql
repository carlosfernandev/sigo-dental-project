CREATE TABLE appointment_reminders (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    appointment_id uuid NOT NULL REFERENCES appointments (id) ON DELETE RESTRICT,
    reminder_type text NOT NULL CHECK (reminder_type IN ('48h', '24h')),
    status text NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'contacted', 'no_response')),
    notes text NOT NULL DEFAULT '' CHECK (length(notes) <= 500),
    handled_by uuid REFERENCES users (id) ON DELETE RESTRICT,
    handled_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (appointment_id, reminder_type),
    CHECK ((status = 'pending' AND handled_by IS NULL AND handled_at IS NULL)
        OR (status <> 'pending' AND handled_by IS NOT NULL AND handled_at IS NOT NULL))
);

CREATE INDEX appointment_reminders_pending_idx
    ON appointment_reminders (status, created_at) WHERE status = 'pending';
