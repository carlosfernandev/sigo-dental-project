# RF-XX: [Nombre del Requerimiento]

**Dominio a modificar:** `internal/<dominio>/`

## 1. Base de Datos (Migración requerida)

- [ ] Tabla nueva o alteraciones:
  - Columna A (tipo)
  - Columna B (tipo)

## 2. Entidades y DTOs

- Struct `EntityName`: [Campos mapeados a DB]
- Struct `RequestDTO`: [Payload esperado del cliente]

## 3. Endpoints (Fiber)

- `METODO /api/v1/ruta` -> [Qué hace]

## 4. Reglas de Negocio Estrictas

- [ ] Regla 1 (ej: Si el monto es 0, lanzar error 400).
- [ ] Regla 2 (ej: La actualización debe ser dentro de una transacción).
