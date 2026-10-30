CREATE TABLE operatories (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name text NOT NULL CHECK (length(btrim(name)) BETWEEN 1 AND 100),
    description text NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX operatories_name_idx ON operatories (lower(name));

CREATE TABLE chairs (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    operatory_id uuid NOT NULL REFERENCES operatories (id) ON DELETE RESTRICT,
    name text NOT NULL CHECK (length(btrim(name)) BETWEEN 1 AND 100),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX chairs_operatory_idx ON chairs (operatory_id);

CREATE TABLE dentists (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    first_name text NOT NULL CHECK (length(btrim(first_name)) BETWEEN 1 AND 100),
    last_name text NOT NULL CHECK (length(btrim(last_name)) BETWEEN 1 AND 100),
    specialty text NOT NULL DEFAULT '',
    active boolean NOT NULL DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX dentists_name_idx ON dentists (lower(last_name), lower(first_name));
CREATE INDEX dentists_active_idx ON dentists (active) WHERE active = true;

CREATE TABLE appointments (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    chair_id uuid NOT NULL REFERENCES chairs (id) ON DELETE RESTRICT,
    dentist_id uuid NOT NULL REFERENCES dentists (id) ON DELETE RESTRICT,
    patient_id uuid NOT NULL REFERENCES patients (id) ON DELETE RESTRICT,
    start_time timestamptz NOT NULL,
    end_time timestamptz NOT NULL,
    state text NOT NULL DEFAULT 'scheduled' CHECK (state IN ('scheduled', 'confirmed', 'checked_in', 'in_progress', 'completed', 'cancelled', 'no_show')),
    notes text NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CHECK (start_time < end_time)
);

CREATE INDEX appointments_chair_time_idx ON appointments (chair_id, start_time, end_time);
CREATE INDEX appointments_dentist_time_idx ON appointments (dentist_id, start_time, end_time);
CREATE INDEX appointments_patient_time_idx ON appointments (patient_id, start_time, end_time);
CREATE INDEX appointments_state_idx ON appointments (state);

CREATE FUNCTION prevent_chair_overlap() RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM appointments
        WHERE chair_id = NEW.chair_id
          AND id != NEW.id
          AND state NOT IN ('cancelled', 'no_show')
          AND (start_time, end_time) OVERLAPS (NEW.start_time, NEW.end_time)
    ) THEN
        RAISE EXCEPTION 'chair_overlap: Chair %s already has an appointment in this time range', NEW.chair_id;
    END IF;
    RETURN NEW;
END;
$$;

CREATE TRIGGER appointments_prevent_overlap
BEFORE INSERT OR UPDATE ON appointments
FOR EACH ROW EXECUTE FUNCTION prevent_chair_overlap();
