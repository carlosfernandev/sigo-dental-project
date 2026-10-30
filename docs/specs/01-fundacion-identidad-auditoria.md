# Backlog 01: Cimientos, Identidad, MFA, RBAC y Auditoría

Ubicación sugerida: `docs/specs/01-fundacion-identidad-auditoria.md`

Cubre: RNF-04, RNF-05, RNF-06 y la sección 3 del SRS (roles).
Orden obligatorio: Bloque 0 → 1 → 2 → 3. Dentro de cada bloque, en orden numérico.

---

## Cómo usar este documento

1. Abre una rama por tarea (`git checkout -b tarea/1.3-login`).
2. Pega al agente el **Prompt base** y luego el **prompt de la tarea**.
3. Revisa el diff tú mismo, corre los tests y haz commit solo si todo pasa.
4. Marca la tarea como hecha en la tabla de seguimiento al final.

### Prompt base (pegar siempre al inicio de cada sesión)

```text
Trabajas en SIGO-Dental, un sistema de gestión odontológica con datos de salud.
Antes de hacer nada, lee AGENTS.md y docs/SRS_SIGO-Dental.md.

Reglas generales:
- Usa exactamente el stack y las convenciones definidos en AGENTS.md. No introduzcas dependencias nuevas sin avisarme.
- Haz SOLO la tarea que te pido. No adelantes otras tareas ni refactorices código ajeno.
- Escribe tests para todo lo que implementes y córrelos antes de terminar.
- Nunca registres contraseñas, tokens, secretos TOTP ni datos clínicos en logs.
- Seguridad por defecto: denegar si no hay permiso explícito.
- Al terminar, resume qué archivos cambiaste, qué tests corriste y qué quedó pendiente o dudoso.
- Si algo del requerimiento es ambiguo, pregunta antes de asumir.
```

---

## Decisiones previas (confirmar antes de la tarea 1.4)

| Tema | Decisión recomendada |
|---|---|
| Sesión | Cookie `httpOnly` + `Secure` + `SameSite=Lax/Strict`, sesión almacenada en servidor (Redis o tabla). Sin JWT en `localStorage`. |
| Hash de contraseñas | argon2id |
| MFA | TOTP (RFC 6238) con códigos de recuperación. Sin SMS como segundo factor. |
| Timeout por inactividad | 15 minutos (RNF-05), configurable por variable de entorno |

Registra la decisión final en `docs/decisiones/0001-auth-sesion-mfa.md`.

---

# BLOQUE 0: Cimientos

### Tarea 0.1: Repo, estructura, AGENTS.md, lint y formato

**Prompt:**
```text
Tarea 0.1: prepara la estructura base del repositorio.

Haz:
- Crea la estructura de carpetas acordada en el plan (backend, frontend, docs, scripts).
- Configura linter y formateador del stack, con scripts `lint`, `format` y `test` ejecutables con un solo comando.
- Verifica que AGENTS.md existe y que sus comandos de lint, test y build son correctos y ejecutables.
- Añade .gitignore y .editorconfig. Crea .env.example sin valores reales.

No hagas: lógica de negocio, endpoints ni base de datos.

Criterio de aceptación: `lint` y `test` corren en limpio (aunque haya un solo test de ejemplo).
```

### Tarea 0.2: Backend base con `/health`

**Prompt:**
```text
Tarea 0.2: crea el backend mínimo.

Haz:
- Servidor con configuración por variables de entorno (sin valores hardcodeados).
- Endpoint GET /health que responde 200 con { "status": "ok" }.
- Manejo de errores centralizado con formato de error consistente.
- Logging estructurado que NO incluya cuerpos de request ni cabeceras sensibles.
- Un test automatizado del endpoint.

Criterio de aceptación: el servidor arranca con un comando documentado y el test de /health pasa.
```

### Tarea 0.3: PostgreSQL con migraciones

**Prompt:**
```text
Tarea 0.3: base de datos y migraciones.

Haz:
- docker-compose.yml para desarrollo con PostgreSQL (y Redis si la decisión de sesión lo requiere).
- Herramienta de migraciones del stack, con comandos `migrate up` y `migrate down`.
- Una migración inicial vacía o con una tabla `schema_info` para validar el flujo.
- Conexión a la base de datos leída desde variables de entorno.
- Test de integración que conecte a la base de datos de pruebas.

Criterio de aceptación: desde cero, `docker compose up`, `migrate up` y `migrate down` funcionan sin errores.
```

### Tarea 0.4: CI

**Prompt:**
```text
Tarea 0.4: pipeline de integración continua.

Haz:
- Pipeline (GitHub Actions u otro definido en AGENTS.md) que corra en cada push y PR: instalar dependencias, lint, tests (con base de datos en servicio) y build.
- Cachea dependencias para acelerar.
- Añade un job de auditoría de dependencias (npm audit / pip-audit / equivalente) que reporte vulnerabilidades altas o críticas.

Criterio de aceptación: el pipeline falla si lint o tests fallan. Demuéstralo con un commit de prueba que rompa un test (y luego revierte).
```

---

# BLOQUE 1: Identidad y sesión

### Tarea 1.1: Modelo `users` y `roles` + seed del admin

**Prompt:**
```text
Tarea 1.1: modelo de usuarios y roles.

Haz:
- Migración con tablas `roles` y `users`.
- Roles según la sección 3 del SRS: admin_ti, recepcionista_cajero, odontologo, asistente_dental, auditor_contador.
- `users`: id (UUID), email único (normalizado a minúsculas), password_hash, role_id, is_active, failed_attempts, locked_until, mfa_enabled, created_at, updated_at.
- Script de seed que cree el primer admin_ti leyendo email y contraseña inicial desde variables de entorno (nunca hardcodeadas). Debe forzar cambio de contraseña en el primer login (campo `must_change_password`).
- Tests de la migración y del seed.

No hagas: login, endpoints ni hashing todavía (deja el campo password_hash listo).

Criterio de aceptación: migración aplicada y admin creado por script; el seed es idempotente.
```

### Tarea 1.2: Hash argon2id y política de contraseñas

**Prompt:**
```text
Tarea 1.2: módulo de contraseñas.

Haz:
- Módulo aislado `passwords` con `hash(password)` y `verify(password, hash)` usando argon2id con parámetros razonables (documenta los elegidos).
- Política de contraseñas: mínimo 12 caracteres, rechazo de contraseñas comunes (lista corta embebida o librería) y de contraseñas que contengan el email.
- Rehash automático si los parámetros del hash almacenado quedan obsoletos.
- Tests: hash/verify correcto, verify con contraseña errónea, rechazo de contraseñas débiles, rehash.
- Conecta el seed de la tarea 1.1 para que use este módulo.

Criterio de aceptación: tests pasan y no existe ningún lugar del código que guarde contraseñas en texto plano.
```

### Tarea 1.3: Endpoint de login

**Prompt:**
```text
Tarea 1.3: POST /auth/login (aún sin MFA ni sesión final).

Haz:
- Recibe email y contraseña, valida con el módulo de la tarea 1.2.
- Respuesta de error GENÉRICA e idéntica ("Credenciales inválidas") tanto si el usuario no existe, como si la contraseña es incorrecta o el usuario está inactivo.
- Evita diferencias de tiempo observables: ejecuta una verificación de hash ficticia cuando el usuario no existe.
- Valida entrada (formato de email, longitud máxima) y rechaza cuerpos malformados con 400.
- Si las credenciales son válidas, devuelve un resultado interno "autenticado" (la emisión de sesión se hace en la tarea 1.4).
- Tests: usuario válido, contraseña errónea, usuario inexistente, usuario inactivo (las tres respuestas de error deben ser indistinguibles).

Criterio de aceptación: no se puede deducir desde la respuesta si un email existe.
```

### Tarea 1.4: Sesión segura y logout

**Prompt:**
```text
Tarea 1.4: sesiones en servidor.

Haz:
- Al autenticar, crea una sesión en el almacén definido (Redis o tabla) con ID aleatorio criptográficamente seguro (mínimo 128 bits).
- Cookie `httpOnly`, `Secure`, `SameSite` según la decisión, con nombre y path definidos. Rota el ID de sesión en el login.
- POST /auth/logout: invalida la sesión en el servidor y borra la cookie.
- GET /auth/me: devuelve usuario y rol de la sesión activa; 401 si no hay sesión.
- Protección CSRF para métodos que cambian estado (token o doble envío de cookie, según el stack).
- Middleware `requireSession` reutilizable.
- Tests: sesión creada, /auth/me con y sin sesión, logout invalida (reusar la cookie anterior da 401), cookie con los flags correctos.

Criterio de aceptación: tras logout, la cookie antigua ya no sirve aunque el atacante la conserve.
```

### Tarea 1.5: Rate limit y bloqueo por intentos fallidos

**Prompt:**
```text
Tarea 1.5: protección contra fuerza bruta.

Haz:
- Contador de intentos fallidos por usuario: tras 5 fallos consecutivos, `locked_until` = ahora + 15 minutos (valores configurables por entorno).
- Rate limit por IP en /auth/login (por ejemplo 20 intentos por 15 minutos).
- Un login exitoso reinicia el contador.
- El bloqueo debe devolver el mismo error genérico de la tarea 1.3 (no revelar que la cuenta está bloqueada).
- Tests con reloj simulado: bloqueo a los N fallos, desbloqueo tras el tiempo, reinicio con éxito, rate limit por IP.

Criterio de aceptación: no se puede probar más de N contraseñas por cuenta en la ventana definida.
```

### Tarea 1.6: Pantalla de login (frontend)

**Prompt:**
```text
Tarea 1.6: frontend de login.

Haz:
- Pantalla de login (email y contraseña) con validación básica, estado de carga y mensaje de error genérico.
- Cliente HTTP que envía cookies de sesión (credentials) y el token CSRF.
- Rutas protegidas: si /auth/me devuelve 401, redirige al login.
- Botón de logout visible en la cabecera.
- Pantalla de cambio de contraseña obligatorio cuando `must_change_password` sea true (requiere un endpoint POST /auth/change-password, impleméntalo con la política de la tarea 1.2).
- Diseño responsive (escritorio y tablet, RNF-07). Accesible: labels, foco y navegación por teclado.
- Tests de componentes y un test E2E del flujo login, ruta protegida y logout.

Criterio de aceptación: flujo completo en el navegador sin guardar nada sensible en localStorage/sessionStorage.
```

### Tarea 1.7: CRUD de usuarios (solo Admin TI)

**Prompt:**
```text
Tarea 1.7: administración de usuarios.

Haz:
- Endpoints (solo rol admin_ti): crear usuario (con contraseña temporal y must_change_password=true), listar, ver, editar (nombre, rol), desactivar/reactivar y resetear contraseña.
- No existe borrado físico de usuarios: solo desactivar.
- Al desactivar un usuario o cambiarle el rol, invalida todas sus sesiones activas.
- Un admin no puede desactivarse ni quitarse el rol a sí mismo.
- Pantalla de administración en el frontend (tabla, formulario, acciones).
- Tests: solo admin_ti accede (otros roles reciben 403), usuario desactivado no puede entrar, sesiones invalidadas al desactivar.

Criterio de aceptación: un usuario desactivado pierde el acceso de inmediato, incluso con sesión abierta.
```

---

# BLOQUE 2: MFA y timeout (RNF-05)

### Tarea 2.1: Enrolamiento TOTP

**Prompt:**
```text
Tarea 2.1: alta de MFA con TOTP.

Haz:
- Endpoints para usuario autenticado: POST /auth/mfa/setup (genera secreto y devuelve URI otpauth y QR) y POST /auth/mfa/verify-setup (valida el primer código y activa MFA).
- El secreto TOTP se guarda CIFRADO en la base de datos (AES-256-GCM, clave desde variable de entorno o KMS; documenta la rotación de clave).
- Tolerancia de ±1 ventana de 30 s; protege contra reutilización del mismo código en la misma ventana.
- Pantalla de enrolamiento en el frontend con QR y campo de verificación.
- Tests: secreto cifrado en reposo (no legible en la tabla), código correcto activa MFA, código incorrecto no.

Criterio de aceptación: ningún secreto TOTP aparece en texto plano en base de datos ni en logs.
```

### Tarea 2.2: Segundo paso en el login

**Prompt:**
```text
Tarea 2.2: login en dos pasos.

Haz:
- Si el usuario tiene MFA activo, /auth/login valida la contraseña y devuelve un estado "mfa_required" con un token temporal de corta vida (5 min) ligado a ese intento, SIN crear la sesión completa.
- POST /auth/mfa/challenge: recibe token temporal + código TOTP; solo si es válido se crea la sesión.
- Aplica el mismo bloqueo por intentos fallidos de la tarea 1.5 al paso MFA.
- Frontend: segunda pantalla para ingresar el código.
- Tests: sin código válido no hay sesión, token temporal expirado, reuso del token, bloqueo por fallos.

Criterio de aceptación: es imposible obtener una sesión completa saltándose el segundo factor.
```

### Tarea 2.3: Códigos de recuperación

**Prompt:**
```text
Tarea 2.3: códigos de recuperación de MFA.

Haz:
- Al activar MFA, genera 10 códigos de un solo uso, se muestran UNA sola vez; en base de datos se guardan solo sus hashes.
- /auth/mfa/challenge acepta un código de recuperación en lugar del TOTP; al usarlo queda consumido.
- Endpoint para regenerar códigos (requiere reautenticación con contraseña + TOTP), lo que invalida los anteriores.
- Endpoint de admin_ti para resetear el MFA de un usuario que perdió acceso (queda para auditoría en el bloque 3).
- Tests: cada código funciona una sola vez, regenerar invalida los anteriores.

Criterio de aceptación: un código de recuperación usado no vuelve a funcionar.
```

### Tarea 2.4: Cierre por inactividad (15 min)

**Prompt:**
```text
Tarea 2.4: timeout de sesión.

Haz:
- Servidor: la sesión expira tras 15 min sin actividad (configurable por entorno) y tiene un límite absoluto (por ejemplo 12 h). Cada request autenticado renueva la ventana de inactividad.
- Cliente: detecta inactividad (ratón, teclado, toque), muestra un aviso 1 minuto antes con opción "Seguir conectado" y cierra sesión al agotarse el tiempo, limpiando el estado en memoria.
- Las respuestas 401 por expiración llevan al login con mensaje claro.
- Tests de servidor con reloj simulado (expira a los 15 min, renueva con actividad, límite absoluto) y de cliente para el aviso.

Criterio de aceptación: 15 min sin actividad cierran la sesión, tanto en el servidor como en la interfaz.
```

### Tarea 2.5: Reglas de MFA obligatorio

**Prompt:**
```text
Tarea 2.5: forzar MFA donde corresponde.

Haz:
- Middleware `requireMfa` que devuelva 403 si la sesión no se completó con MFA.
- Política: MFA obligatorio para todos los roles que acceden a expedientes clínicos o información financiera. En la práctica, todos los roles del SRS lo requieren; implementa la política como configuración por rol para poder ajustarla.
- Un usuario sin MFA activo puede iniciar sesión, pero solo puede acceder a la pantalla de enrolamiento (2.1) hasta activarlo.
- Marca en la sesión si el MFA fue verificado.
- Tests: rutas protegidas devuelven 403 sin MFA y 200 con MFA, usuario sin MFA queda confinado al enrolamiento.

Criterio de aceptación: no hay manera de llegar a datos clínicos o financieros sin MFA completado.
```

---

# BLOQUE 3: RBAC y auditoría (RNF-05/06)

### Tarea 3.1: Matriz de permisos por rol

**Prompt:**
```text
Tarea 3.1: define los permisos.

Haz:
- Lista de permisos granulares con formato `recurso:accion` (ej. `patient:read`, `patient:write`, `appointment:write`, `audit:read`, `user:manage`, `cash:close`).
- Matriz rol → permisos basada en la sección 3 del SRS, almacenada en código/configuración versionada, con migración o seed a base de datos si el diseño lo requiere.
- Documenta la matriz en docs/decisiones/0002-matriz-permisos.md en formato tabla.
- Aplica el principio de menor privilegio: si el SRS no otorga explícitamente un permiso a un rol, no lo tiene. Lista aparte las dudas que encuentres para que yo las decida.
- Tests que verifiquen la matriz (p. ej. recepcionista no tiene `clinical_record:write`, asistente no puede cerrar caja).

Criterio de aceptación: existe una única fuente de verdad de permisos y está documentada.
```

### Tarea 3.2: Middleware de autorización (denegado por defecto)

**Prompt:**
```text
Tarea 3.2: autorización por permiso.

Haz:
- Middleware/decorador `requirePermission('recurso:accion')` que consulta la matriz de la tarea 3.1.
- Denegado por defecto: cualquier endpoint sin declaración de permiso explícita debe fallar el arranque o devolver 403 (elige lo más seguro y documéntalo).
- Aplica el middleware a los endpoints existentes (usuarios, auth/admin) reemplazando verificaciones de rol sueltas.
- Evita IDOR: deja un helper/patrón para validar propiedad o alcance del recurso (se usará con pacientes).
- Tests parametrizados: por cada rol y endpoint protegido, comprobar permitido/denegado.

Criterio de aceptación: no existe un endpoint accesible sin declaración de permiso.
```

### Tarea 3.3: Tabla de auditoría append-only

**Prompt:**
```text
Tarea 3.3: almacenamiento inmutable de auditoría (RNF-06).

Haz:
- Migración de la tabla `audit_log`: id, occurred_at (timestamptz, ISO-8601 al exponer), actor_user_id (nullable para eventos anónimos), actor_role, action, resource_type, resource_id, outcome (success/failure), ip_address, user_agent, metadata (JSON sin datos sensibles).
- Inmutabilidad a nivel de base de datos: el rol de la aplicación solo tiene INSERT y SELECT sobre la tabla; además, triggers que rechacen UPDATE y DELETE.
- Opcional recomendado: encadenado por hash (cada fila guarda el hash de la anterior) para detectar manipulación; impleméntalo si es razonable en el stack y documenta cómo verificarlo.
- Test de integración: intentar UPDATE y DELETE con las credenciales de la app debe fallar.

Criterio de aceptación: ni la aplicación puede modificar ni borrar filas del log.
```

### Tarea 3.4: Servicio de auditoría

**Prompt:**
```text
Tarea 3.4: servicio único para registrar eventos.

Haz:
- Servicio `audit.record({ actor, action, resource, outcome, metadata })` como ÚNICO punto de escritura en audit_log.
- Captura automática de IP y user agent desde el contexto de la request (respetando la configuración de proxy de confianza).
- Sanitiza `metadata`: nunca contraseñas, tokens, secretos ni contenido clínico, solo identificadores y tipo de acción.
- Si falla la escritura de auditoría en una operación sensible, la operación debe fallar (no auditar en silencio).
- Catálogo de acciones definido en un enum (LOGIN_SUCCESS, LOGIN_FAILURE, LOGOUT, ROLE_CHANGED, USER_CREATED, USER_DEACTIVATED, MFA_ENABLED, MFA_RESET, RECORD_READ, RECORD_UPDATED, RECORD_EXPORTED, ...).
- Tests: registro correcto, sanitización de metadata, fallo de auditoría aborta la operación.

Criterio de aceptación: nadie escribe en audit_log excepto este servicio.
```

### Tarea 3.5: Auditar eventos de autenticación y administración

**Prompt:**
```text
Tarea 3.5: conectar auditoría a lo ya construido.

Haz:
- Registra con el servicio de la tarea 3.4: login exitoso y fallido (sin guardar la contraseña, con el email intentado solo si es seguro), logout, expiración de sesión, bloqueo de cuenta, enrolamiento y reset de MFA, uso de código de recuperación, creación/desactivación/reactivación de usuarios, cambios de rol y de contraseña, y accesos denegados (403).
- Cada evento con actor, IP y resultado.
- Tests: por cada evento listado, verificar que aparece exactamente una fila de auditoría.

Criterio de aceptación: cada evento de autenticación y administración deja rastro verificable.
```

### Tarea 3.6: Consulta de auditoría

**Prompt:**
```text
Tarea 3.6: visor de auditoría.

Haz:
- Endpoint GET /audit con filtros (usuario, acción, rango de fechas, recurso, resultado) y paginación. Requiere `audit:read` (admin_ti y auditor_contador).
- La consulta de auditoría también queda auditada (AUDIT_VIEWED).
- Exportación a CSV limitada y auditada (RECORD_EXPORTED), con tope de filas.
- Pantalla en el frontend: tabla filtrable, fechas en ISO-8601 o formato local claro.
- Solo lectura: ninguna acción de edición o borrado en la UI ni en la API.
- Tests: permisos por rol, filtros, paginación, la consulta queda registrada.

Criterio de aceptación: Admin TI y Auditor consultan el log con filtros; nadie puede modificarlo.
```

---

## Checklist de cierre del bloque (antes de pasar a RF-01)

- [ ] Login con contraseña + TOTP funciona de punta a punta.
- [ ] Sin MFA no se llega a rutas protegidas.
- [ ] 15 min de inactividad cierran la sesión.
- [ ] Cada rol solo accede a lo permitido (tests por rol en verde).
- [ ] UPDATE/DELETE sobre audit_log fallan con las credenciales de la app.
- [ ] Todos los eventos de auth quedan registrados.
- [ ] CI en verde y sin vulnerabilidades altas o críticas en dependencias.
- [ ] Revisión manual hecha de las tareas 1.3, 1.4, 1.5, 2.2 y 3.3.

## Seguimiento

| Tarea | Estado | Rama / PR | Notas |
|---|---|---|---|
| 0.1 | ☑ | | Estructura, EditorConfig y comandos Make; lint, tests y build verificados. |
| 0.2 | ☑ | | `/health` implementado y probado; error envelope centralizado. |
| 0.3 | ☑ | | Migración inicial aplicada dos veces en PostgreSQL local; test aislado de integración añadido. `down` excluido por la política forward-only de AGENTS.md. |
| 0.4 | ◷ | | Diferido por instrucción del usuario hasta el final del spec. Workflow ya está preparado. |
| 1.1 | ☑ | | Migración aplicada; test de integración confirmó roles, hash bcrypt, cambio obligatorio y seed idempotente. |
| 1.2 | ☑ | | Política bcrypt, hash, verify y rehash implementados; Argon2id del backlog queda reemplazado por decisión del repositorio. |
| 1.3 | ☐ | | |
| 1.4 | ☐ | | |
| 1.5 | ☐ | | |
| 1.6 | ☐ | | |
| 1.7 | ☐ | | |
| 2.1 | ☐ | | |
| 2.2 | ☐ | | |
| 2.3 | ☐ | | |
| 2.4 | ☐ | | |
| 2.5 | ☐ | | |
| 3.1 | ☐ | | |
| 3.2 | ☐ | | |
| 3.3 | ☐ | | |
| 3.4 | ☐ | | |
| 3.5 | ☐ | | |
| 3.6 | ☐ | | |
