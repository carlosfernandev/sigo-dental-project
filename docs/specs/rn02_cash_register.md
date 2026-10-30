# RN-02: Cierre de Caja Ciego
**Dominio a modificar:** `internal/billing/` (o nuevo `internal/cashier/`)
**Estado:** Hecho — el total teórico solo aparece en la respuesta del cierre enviado.

## 1. Base de Datos (Migración requerida)
- Crear tabla `cash_register_closures`:
  - `id` UUID PK
  - `user_id` UUID FK (usuario que hace el cierre)
  - `closure_date` DATE (o timestamptz)
  - `reported_cash` INT64 (lo que el cajero contó)
  - `reported_transfers` INT64 (lo que el cajero sumó en vouchers)
  - `system_cash` INT64 (total calculado por el sistema, oculto)
  - `system_transfers` INT64 (total calculado por el sistema, oculto)
  - `status` VARCHAR (matched, discrepant)

## 2. Entidades y DTOs
- Struct `ClosureRequest`: Recibe únicamente `reported_cash` y `reported_transfers`.

## 3. Endpoints HTTP (`handler.go` - Fiber v3)
- `POST /cash-register/close` -> Recibe el conteo físico, lo compara con las facturas del día (`invoices` filtrado por fecha y estado `paid`), y guarda el registro de cierre. Devuelve el reporte final.

## 4. Reglas de Negocio Estrictas (`service.go`)
- **Cierre Ciego:** El cajero envía el DTO *sin* saber cuánto hay en el sistema. El backend es el responsable de consultar las facturas pagadas hoy, sumar los montos según el método de pago, y comparar contra lo reportado.
- Si hay diferencia (`reported_X` != `system_X`), el `status` es 'discrepant'.
- `reported_cash` y `reported_transfers` son centavos enteros (USD). `system_transfers` suma pagos
  con métodos `card`, `transfer` e `insurance`, reportados como vouchers/valores electrónicos.
- Se toma como fecha del cierre el día local `America/Guayaquil`; una factura cuenta el día en que
  pasa a `paid`. El cierre agrega todos sus componentes de pago por método.
- El cuerpo de la solicitud acepta solo los dos conteos reportados. No existe endpoint de consulta
  previa del cierre ni el servicio devuelve totales teóricos antes de guardar el conteo.
