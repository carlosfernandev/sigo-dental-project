# RF-11: Recordatorios de Citas 48h / 24h
**Dominio a modificar:** `internal/scheduling/`

## 1. Base de Datos (Migración requerida)
- Crear tabla `appointment_reminders`:
  - `id` UUID PK
  - `appointment_id` UUID FK
  - `reminder_type` VARCHAR ('48h', '24h')
  - `status` VARCHAR ('pending', 'contacted', 'no_response')
  - `created_at`, `updated_at`

## 2. Entidades y DTOs
- Struct `UpdateReminderRequest`: Recibe el nuevo `status` ('contacted', 'no_response') y `notes` opcionales.

## 3. Endpoints HTTP (`handler.go` - Fiber v3)
- `GET /reminders/pending` -> Devuelve la cola de recepción (recordatorios con status='pending'). Incluye datos de la cita y paciente (JOIN o precargado).
- `PATCH /reminders/:id` -> Actualiza el estado del recordatorio.

## 4. Reglas de Negocio Estrictas (`service.go`)
- **Cola Interna MVP:** No hay proveedor externo de mensajería (Twilio, WhatsApp). Todo vive en el sistema.
- **Worker (Opcional/Cron):** Crear una función `GeneratePendingReminders()` que busque citas confirmadas a 48h o 24h de distancia y que no tengan ya un recordatorio generado, insertándolas en la tabla con estado 'pending'.
