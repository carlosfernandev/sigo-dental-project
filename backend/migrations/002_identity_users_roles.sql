CREATE TABLE roles (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name text NOT NULL UNIQUE CHECK (name IN (
        'admin_ti',
        'recepcionista_cajero',
        'odontologo',
        'asistente_dental',
        'auditor_contador'
    )),
    created_at timestamptz NOT NULL DEFAULT now()
);

INSERT INTO roles (name) VALUES
    ('admin_ti'),
    ('recepcionista_cajero'),
    ('odontologo'),
    ('asistente_dental'),
    ('auditor_contador');

CREATE TABLE users (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    email text NOT NULL UNIQUE CHECK (email = lower(btrim(email))),
    password_hash text NOT NULL CHECK (password_hash <> ''),
    role_id uuid NOT NULL REFERENCES roles (id) ON DELETE RESTRICT,
    is_active boolean NOT NULL DEFAULT true,
    failed_attempts integer NOT NULL DEFAULT 0 CHECK (failed_attempts >= 0),
    locked_until timestamptz,
    mfa_enabled boolean NOT NULL DEFAULT false,
    must_change_password boolean NOT NULL DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);
