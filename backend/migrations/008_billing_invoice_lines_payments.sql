CREATE TABLE invoice_items (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_id uuid NOT NULL REFERENCES invoices (id) ON DELETE RESTRICT,
    description text NOT NULL CHECK (length(btrim(description)) BETWEEN 1 AND 500),
    unit_price bigint NOT NULL CHECK (unit_price > 0),
    quantity integer NOT NULL CHECK (quantity BETWEEN 1 AND 100),
    subtotal bigint NOT NULL CHECK (subtotal > 0),
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX invoice_items_invoice_idx ON invoice_items (invoice_id, id);

-- Preserve any invoice lines written by the earlier JSONB schema.
INSERT INTO invoice_items (invoice_id, description, unit_price, quantity, subtotal)
SELECT i.id,
       COALESCE(item->>'Description', item->>'description', item->>'ProcedureName', item->>'procedure_name', 'Item'),
       COALESCE(item->>'UnitPrice', item->>'unit_price', '0')::bigint,
       COALESCE(item->>'Quantity', item->>'quantity', '1')::integer,
       COALESCE(item->>'Subtotal', item->>'subtotal', item->>'Total', item->>'total', '0')::bigint
FROM invoices AS i
CROSS JOIN LATERAL jsonb_array_elements(
    CASE WHEN jsonb_typeof(i.items) = 'array' THEN i.items ELSE '[]'::jsonb END
) AS item
WHERE COALESCE(item->>'UnitPrice', item->>'unit_price', '0')::bigint > 0
  AND COALESCE(item->>'Quantity', item->>'quantity', '1')::integer BETWEEN 1 AND 100
  AND COALESCE(item->>'Subtotal', item->>'subtotal', item->>'Total', item->>'total', '0')::bigint > 0;

CREATE TABLE payments (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_id uuid NOT NULL REFERENCES invoices (id) ON DELETE RESTRICT,
    amount bigint NOT NULL CHECK (amount > 0),
    method text NOT NULL CHECK (method IN ('cash', 'card', 'transfer', 'insurance')),
    reference text,
    created_at timestamptz NOT NULL DEFAULT now(),
    CHECK (reference IS NULL OR length(btrim(reference)) BETWEEN 1 AND 150),
    CHECK (method = 'cash' OR reference IS NOT NULL)
);

CREATE INDEX payments_invoice_idx ON payments (invoice_id, created_at, id);
