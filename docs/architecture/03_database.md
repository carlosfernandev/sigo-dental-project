# Convenciones de Base de Datos (PostgreSQL)

## Migraciones

- Puras en SQL (archivos `.up.sql` y `.down.sql`).
- Estrategia _Forward-only_ (nunca modificar migraciones ya aplicadas, crear una nueva para alterar tablas).

## Diseño de Tablas

- **Claves Primarias:** UUID v4 (`id UUID PRIMARY KEY DEFAULT gen_random_uuid()`).
- **Auditoría Básica:** Todas las tablas principales deben tener `created_at` y `updated_at` (TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP).
- **Soft Deletes:** (Regla RN-01) En el core clínico no se elimina nada, se inactiva o se añaden adendas.

## Transacciones (Repositorio)

- Usar transacciones SQL explícitas (`tx, err := db.BeginTxx(...)`) cuando un caso de uso modifique más de una tabla.
- Asegurar siempre el `defer tx.Rollback()` antes del `Commit()`.
- La capa de repositorio no debe saber nada de HTTP ni de Fiber.
