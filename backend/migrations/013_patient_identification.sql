ALTER TABLE patients
    ADD COLUMN identification text NOT NULL DEFAULT ''
    CHECK (identification = '' OR identification ~ '^[0-9]{10}$');

CREATE UNIQUE INDEX patients_identification_idx
    ON patients (identification)
    WHERE identification <> '';
