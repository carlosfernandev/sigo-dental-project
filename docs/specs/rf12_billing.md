# RF-12: Facturación Interna y Pagos

**Dominio a modificar:** `internal/billing/`

## 1. Base de Datos (Migraciones)

- **Tabla `invoices` (Facturas):**
  - `id` UUID PK
  - `patient_id` UUID FK
  - `total_amount` DECIMAL(10,2)
  - `status` VARCHAR (ej: 'pending', 'paid', 'voided')
  - `created_at`, `updated_at`
- **Tabla `invoice_items` (Detalle de factura):**
  - `id` UUID PK
  - `invoice_id` UUID FK
  - `description` VARCHAR
  - `unit_price` DECIMAL(10,2)
  - `quantity` INT
  - `subtotal` DECIMAL(10,2)

## 2. Entidades y DTOs (`entity.go`, `dto.go`)

- Reflejar las tablas anteriores en structs de Go.
- DTO para crear factura: `CreateInvoiceRequest` (con slice de ítems).
- DTO para registrar pago: `RegisterPaymentRequest` (monto, método: 'tarjeta', 'transferencia', 'efectivo', y referencia).

## 3. Endpoints HTTP (`handler.go`)

- `POST /api/v1/invoices` -> Crea la factura con sus ítems (transaccional).
- `PATCH /api/v1/invoices/:id/pay` -> Registra el pago.
- `GET /api/v1/invoices/patient/:patient_id` -> Lista facturas de un paciente.

## 4. Reglas de Negocio Estrictas (`service.go` y `repository.go`)

- **Restricción MVP:** Cero conexión tributaria externa. Las facturas son solo registros internos.
- Los pagos con tarjeta, transferencia y seguro se registran únicamente con un número de referencia (texto), no procesan transacciones reales.
- **Transaccionalidad:** La creación de la cabecera de factura y sus ítems debe insertarse en una sola transacción SQL.
