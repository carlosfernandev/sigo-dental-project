# RF-12: Facturación Interna y Pagos

**Dominio a modificar:** `internal/billing/`

## 1. Base de Datos (Migraciones)

- **Tabla `invoices` (Facturas):**
  - `id` UUID PK
  - `patient_id` UUID FK
  - `total_amount` BIGINT: total en centavos (USD), nunca decimal ni punto flotante
  - `status` VARCHAR (ej: 'pending', 'paid', 'voided')
  - `created_at`, `updated_at`
- **Tabla `invoice_items` (Detalle de factura):**
  - `id` UUID PK
  - `invoice_id` UUID FK
  - `description` VARCHAR
  - `unit_price` BIGINT: precio unitario en centavos (USD)
  - `quantity` INT: cantidad entera positiva
  - `subtotal` BIGINT: subtotal en centavos, calculado como `unit_price * quantity`
- **Tabla `payments` (Pagos):** cada componente de un pago mixto se registra por separado.
  - `id` UUID PK
  - `invoice_id` UUID FK
  - `amount` BIGINT: monto en centavos (USD)
  - `method` VARCHAR (efectivo, tarjeta, transferencia o seguro)
  - `reference` VARCHAR nullable; requerida para tarjeta, transferencia y seguro
  - `created_at`

En Go, todos los campos monetarios (`total_amount`, `unit_price`, `subtotal` y `amount`) son
`int64` en centavos. No convertirlos a `float64` ni usar tipos decimales para cálculos. Los importes
que recibe o devuelve la API también se expresan como enteros en centavos.

## 2. Entidades y DTOs (`entity.go`, `dto.go`)

- Reflejar las tablas anteriores en structs de Go.
- DTO para crear factura: `CreateInvoiceRequest` (con slice de ítems).
- DTO para registrar un componente de pago: `RegisterPaymentRequest` (`amount` en centavos como
  entero, `method` y `reference` cuando aplique). Se aceptan varios componentes para una factura
  hasta cubrir el saldo, lo que permite pagos mixtos.

## 3. Endpoints HTTP (`handler.go`)

- `POST /api/v1/invoices` -> Crea la factura con sus ítems (transaccional).
- `PATCH /api/v1/invoices/:id/pay` -> Registra el pago.
- `GET /api/v1/invoices/patient/:patient_id` -> Lista facturas de un paciente.

## 4. Reglas de Negocio Estrictas (`service.go` y `repository.go`)

- **Restricción MVP:** Cero conexión tributaria externa. Las facturas son solo registros internos.
- Los pagos con tarjeta, transferencia y seguro se registran únicamente con un número de referencia (texto), no procesan transacciones reales.
- **Transaccionalidad:** La creación de la cabecera de factura y sus ítems debe insertarse en una sola transacción SQL.
- Los importes y cantidades deben ser positivos; validar que subtotales y total no excedan el rango
  de `BIGINT` y que el pago no supere el saldo pendiente de la factura.
