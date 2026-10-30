# RF-11: Recordatorios de Citas 48h / 24h
**Dominio a modificar:** `internal/scheduling/`
**Estado:** Hecho — generación periódica local y cola de recepción disponibles; sin proveedor externo.

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
- El monolito ejecuta el generador al iniciar y cada 15 minutos. También lo ejecuta al cargar la
  cola para recuperar recordatorios pendientes. Las ventanas capturan recordatorios atrasados por
  reinicios sin duplicarlos: 48h cuando faltan entre 24 y 48 horas; 24h cuando falta hasta un día.
- El endpoint de actualización solo permite pasar de `pending` a `contacted` o `no_response`; guarda
  el usuario responsable, la hora de gestión y notas opcionales de hasta 500 caracteres.
