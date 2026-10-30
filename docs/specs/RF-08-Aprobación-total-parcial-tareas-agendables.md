# RF-08: Aprobación total/parcial y tareas agendables
**Dominios a modificar:** `internal/billing/` (presupuestos) y `internal/scheduling/` (transición).

**Estado:** Implementado el 2026-10-30.

## 1. Base de Datos (Migración requerida)
- **Tabla `quotes` (Existente en billing):**
  - Actualizar el constraint o lógica del campo `status` para que soporte: `draft`, `approved`, `partially_approved`, `rejected`.
- **Tabla `schedulable_tasks` (NUEVA en scheduling):**
  - `id` UUID PRIMARY KEY DEFAULT gen_random_uuid()
  - `patient_id` UUID NOT NULL (FK a patients)
  - `quote_id` UUID NOT NULL (FK a quotes)
  - `item_id` VARCHAR NOT NULL (Identificador del ítem dentro del JSONB aprobado)
  - `procedure_name` VARCHAR NOT NULL
  - `status` VARCHAR DEFAULT 'pending' (pending, scheduled, completed)
  - `created_at`, `updated_at` TIMESTAMP WITH TIME ZONE

## 2. Entidades y DTOs (`dto.go`, `entity.go`)
- **En `billing`:**
  - Struct `ApproveQuoteRequest`: `{ "approved_item_ids": ["id-1", "id-2"] }`.
  - Asegurar que el struct interno que representa los ítems del `jsonb` de `quotes` tenga un campo `Status` (string).
- **En `scheduling`:**
  - Struct `SchedulableTask` que refleje la nueva tabla.

## 3. Endpoints HTTP (`handler.go` - Fiber v3)
- `PATCH /quotes/:id/approve` -> Recibe el array de IDs de ítems aprobados.
- **Respuesta:** Usar el envelope estricto del proyecto: `{"data": ...}` o `{"error": {code, message, details}}`.

## 4. Reglas de Negocio Estrictas (`service.go` y `repository.go`)
- **Manejo del JSONB:** NO usar funciones complejas de SQL (como `jsonb_set`). El repositorio debe:
  1. Extraer el registro `Quote` completo.
  2. En la capa de servicio (Go), iterar sobre el array de `items`. Si el ID del ítem está en el request, cambiar su estado interno a `approved`.
  3. Si todos los ítems fueron aprobados, cambiar el status del Quote a `approved`. Si solo fueron algunos, a `partially_approved`.
  4. Actualizar toda la fila en la BD (`UPDATE quotes SET items = $1, status = $2 WHERE id = $3`).
- **Transaccionalidad (Crucial):** El UPDATE de la tabla `quotes` y los INSERTs múltiples en la tabla `schedulable_tasks` (uno por cada ítem aprobado) DEBEN ejecutarse dentro de la misma transacción SQL (`tx, err := db.BeginTxx(...)`).
- **Acoplamiento:** El endpoint vive en `billing`, pero al insertar las tareas agendables, el repositorio de `billing` ejecutará un query hacia la tabla `schedulable_tasks` (o se comunicará con el servicio de `scheduling`, según el estándar que sigas en tu monolito para comunicación inter-dominios).

## Implementación
- Migración `011_quote_approvals_schedulable_tasks.sql`: estados de presupuesto, IDs/estados estables por ítem y tabla de tareas agendables; conserva presupuestos aprobados existentes.
- `PATCH /billing/quotes/:id/approve`: aprobación de IDs explícitos, parcial o total, protegida por RBAC y con bloqueo del presupuesto.
- Las tareas aprobadas se insertan mediante el servicio de Scheduling dentro de la misma transacción que actualiza el presupuesto.
- Los presupuestos dejan de poder editarse tras iniciar su aprobación; la actualización condicionada a `draft` evita una carrera con la aprobación.
- Verificación ejecutada: `go test ./internal/billing ./internal/scheduling` pasó. La prueba PostgreSQL de integración se omite si `SIGO_TEST_PG_DSN` no está configurado.
