CREATE TABLE cash_register_closures (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    closure_date date NOT NULL,
    reported_cash bigint NOT NULL CHECK (reported_cash >= 0),
    reported_transfers bigint NOT NULL CHECK (reported_transfers >= 0),
    system_cash bigint NOT NULL CHECK (system_cash >= 0),
    system_transfers bigint NOT NULL CHECK (system_transfers >= 0),
    status text NOT NULL CHECK (status IN ('matched', 'discrepant')),
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (user_id, closure_date)
);
