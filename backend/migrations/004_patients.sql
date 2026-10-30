CREATE TABLE patients (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    first_name text NOT NULL CHECK (length(btrim(first_name)) BETWEEN 1 AND 100),
    last_name text NOT NULL CHECK (length(btrim(last_name)) BETWEEN 1 AND 100),
    date_of_birth date NOT NULL,
    email text NOT NULL DEFAULT '',
    phone text NOT NULL,
    address text NOT NULL DEFAULT '',
    emergency_contact_name text NOT NULL,
    emergency_contact_phone text NOT NULL,
    legal_guardian_name text,
    legal_guardian_identification text,
    legal_guardian_phone text,
    insurance_plan text NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CHECK (
        (legal_guardian_name IS NULL AND legal_guardian_identification IS NULL AND legal_guardian_phone IS NULL)
        OR
        (legal_guardian_name IS NOT NULL AND legal_guardian_identification IS NOT NULL AND legal_guardian_phone IS NOT NULL)
    )
);

CREATE INDEX patients_name_idx ON patients (lower(last_name), lower(first_name));
CREATE INDEX patients_phone_idx ON patients (phone);

CREATE TABLE patient_medical_histories (
    patient_id uuid PRIMARY KEY REFERENCES patients (id) ON DELETE RESTRICT,
    systemic_diseases jsonb NOT NULL DEFAULT '[]'::jsonb,
    allergies jsonb NOT NULL DEFAULT '[]'::jsonb,
    surgical_history jsonb NOT NULL DEFAULT '[]'::jsonb,
    coagulopathies jsonb NOT NULL DEFAULT '[]'::jsonb,
    updated_at timestamptz NOT NULL DEFAULT now(),
    CHECK (jsonb_typeof(systemic_diseases) = 'array'),
    CHECK (jsonb_typeof(allergies) = 'array'),
    CHECK (jsonb_typeof(surgical_history) = 'array'),
    CHECK (jsonb_typeof(coagulopathies) = 'array')
);
