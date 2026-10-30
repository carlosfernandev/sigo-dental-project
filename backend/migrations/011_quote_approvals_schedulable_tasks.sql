ALTER TABLE quotes DROP CONSTRAINT quotes_status_check;
ALTER TABLE quotes ADD CONSTRAINT quotes_status_check
    CHECK (status IN ('draft', 'approved', 'partially_approved', 'rejected'));

CREATE TABLE schedulable_tasks (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    patient_id uuid NOT NULL REFERENCES patients (id) ON DELETE RESTRICT,
    quote_id uuid NOT NULL REFERENCES quotes (id) ON DELETE RESTRICT,
    item_id varchar(80) NOT NULL,
    procedure_name text NOT NULL CHECK (length(btrim(procedure_name)) BETWEEN 1 AND 200),
    status text NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'scheduled', 'completed')),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (quote_id, item_id)
);

CREATE INDEX schedulable_tasks_patient_status_idx ON schedulable_tasks (patient_id, status, created_at);

-- Give existing quote items stable identifiers and statuses before clients can approve them.
UPDATE quotes AS q
SET items = COALESCE((
    SELECT jsonb_agg(
        (item.value - 'id' - 'status') || jsonb_build_object(
            'ID', COALESCE(NULLIF(item.value->>'ID', ''), NULLIF(item.value->>'id', ''), gen_random_uuid()::text),
            'Status', CASE WHEN q.status = 'approved' THEN 'approved'
                           ELSE COALESCE(NULLIF(item.value->>'Status', ''), item.value->>'status', 'pending') END
        ) ORDER BY item.ordinality
    )
    FROM jsonb_array_elements(
        CASE WHEN jsonb_typeof(q.items) = 'array' THEN q.items ELSE '[]'::jsonb END
    ) WITH ORDINALITY AS item(value, ordinality)
), '[]'::jsonb);

-- Preserve already approved quotes in the new scheduling task model.
INSERT INTO schedulable_tasks (patient_id, quote_id, item_id, procedure_name)
SELECT q.patient_id,
       q.id,
       COALESCE(NULLIF(item.value->>'ID', ''), NULLIF(item.value->>'id', '')),
       COALESCE(item.value->>'ProcedureName', item.value->>'procedure_name', 'Treatment')
FROM quotes AS q
CROSS JOIN LATERAL jsonb_array_elements(q.items) AS item(value)
WHERE q.status = 'approved'
  AND COALESCE(item.value->>'Status', item.value->>'status') = 'approved'
ON CONFLICT (quote_id, item_id) DO NOTHING;
