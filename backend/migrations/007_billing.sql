CREATE TABLE procedures (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    code text NOT NULL CHECK (length(btrim(code)) BETWEEN 1 AND 50),
    name text NOT NULL CHECK (length(btrim(name)) BETWEEN 1 AND 200),
    description text NOT NULL DEFAULT '',
    category text NOT NULL DEFAULT '',
    base_price bigint NOT NULL CHECK (base_price >= 0 AND base_price <= 100000000),
    active boolean NOT NULL DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (code)
);

CREATE INDEX procedures_code_idx ON procedures (lower(code));
CREATE INDEX procedures_category_idx ON procedures (category);
CREATE INDEX procedures_active_idx ON procedures (active) WHERE active = true;

CREATE TABLE quotes (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    patient_id uuid NOT NULL REFERENCES patients (id) ON DELETE RESTRICT,
    items jsonb NOT NULL DEFAULT '[]'::jsonb,
    subtotal bigint NOT NULL DEFAULT 0 CHECK (subtotal >= 0),
    discount bigint NOT NULL DEFAULT 0 CHECK (discount >= 0),
    total bigint NOT NULL DEFAULT 0 CHECK (total >= 0),
    status text NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'approved', 'rejected')),
    approved_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX quotes_patient_idx ON quotes (patient_id);
CREATE INDEX quotes_status_idx ON quotes (status);

CREATE TABLE invoices (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    patient_id uuid NOT NULL REFERENCES patients (id) ON DELETE RESTRICT,
    quote_id uuid REFERENCES quotes (id) ON DELETE SET NULL,
    items jsonb NOT NULL DEFAULT '[]'::jsonb,
    subtotal bigint NOT NULL DEFAULT 0 CHECK (subtotal >= 0),
    tax bigint NOT NULL DEFAULT 0 CHECK (tax >= 0),
    total bigint NOT NULL DEFAULT 0 CHECK (total >= 0),
    payments jsonb NOT NULL DEFAULT '[]'::jsonb,
    status text NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'issued', 'paid', 'cancelled')),
    issued_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX invoices_patient_idx ON invoices (patient_id);
CREATE INDEX invoices_quote_idx ON invoices (quote_id);
CREATE INDEX invoices_status_idx ON invoices (status);
