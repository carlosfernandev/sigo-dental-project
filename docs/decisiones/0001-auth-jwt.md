# Decisión 0001: Autenticación JWT para la revisión inicial

**Estado:** Aplicada por simplificación solicitada para la revisión académica.
**Alcance:** Login y autenticación de endpoints del monolito.

## Decisión

- `POST /auth/login` entrega un JWT HS256 Bearer con `sub`, `role`, `jti`, `iat` y `exp`.
- El token vence en dos horas. La clave de firma viene de `SIGO_JWT_SECRET` y requiere al menos 32 bytes.
- El middleware valida el token, aplica 15 minutos de inactividad y deja el principal disponible; cada servicio de dominio conserva sus verificaciones RBAC.
- No se crea una tabla de sesiones ni endpoint de logout. El cliente elimina el token al salir.

La última actividad se guarda en memoria en el proceso único del monolito. Al reiniciar, el primer acceso de cada token se mide desde su `iat`; esto puede pedir un nuevo login a usuarios con tokens antiguos. Cuando se implementen los servicios de dominio, cada caso de uso deberá aplicar `authn.RequireAnyRole` con los roles autorizados por el SRS.
