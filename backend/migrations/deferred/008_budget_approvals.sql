-- PARKED: not applied by `go run ./cmd/migrate up`.
--
-- This migration ALTERs budget_items, a table no earlier migration creates:
-- 007_billing.sql defines quotes/invoices/procedures instead. Applying it aborts
-- the whole single-transaction run, so the schema stays empty.
--
-- Unblock by designing the TreatmentQuote tables per rules/GLOSSARY.md, then
-- moving this file back to backend/migrations/ with the next free version.
--
-- Actualización para presupuestos (RF-08)

-- Añadir el estado de aprobación al ítem del presupuesto
ALTER TABLE budget_items ADD COLUMN status VARCHAR(50) DEFAULT 'pending';
-- 'pending', 'approved', 'rejected'

-- Crear tabla para las tareas agendables (transición de ítem aprobado a turno)
CREATE TABLE schedulable_tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    patient_id UUID NOT NULL REFERENCES patients(id) ON DELETE CASCADE,
    budget_item_id UUID NOT NULL UNIQUE REFERENCES budget_items(id) ON DELETE CASCADE,
    status VARCHAR(50) DEFAULT 'pending_scheduling', -- 'pending_scheduling', 'scheduled', 'completed'
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Índices útiles para búsquedas rápidas
CREATE INDEX idx_schedulable_tasks_patient_id ON schedulable_tasks(patient_id);
CREATE INDEX idx_schedulable_tasks_status ON schedulable_tasks(status);