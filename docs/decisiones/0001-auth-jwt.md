# Decisión 0001: Autenticación JWT para la revisión inicial

**Estado:** Aplicada por simplificación solicitada para la revisión académica.
**Alcance:** Login y autenticación de endpoints del monolito.

## Decisión

- `POST /auth/login` entrega un JWT HS256 Bearer con `sub`, `role`, `iat` y `exp`.
- El token vence en dos horas. La clave de firma viene de `SIGO_JWT_SECRET` y requiere al menos 32 bytes.
- El middleware valida el token y deja el principal disponible para los handlers; cada servicio de dominio conserva sus verificaciones de autorización.
- No se crea una tabla de sesiones ni endpoint de logout. El cliente elimina el token al salir.

La autenticación stateless no implementa por sí sola el cierre por inactividad de RNF-05. Ese requisito queda pendiente antes de proteger información clínica; este token básico sirve para revisar el flujo inicial.
