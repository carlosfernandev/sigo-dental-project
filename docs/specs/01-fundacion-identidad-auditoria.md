# Plan de implementación alineado al SRS

Este plan organiza el trabajo del backend y la identidad del SIGO-Dental según
[`docs/SRS_SIGO-Dental.md`](../SRS_SIGO-Dental.md). No agrega requisitos de producto. Si este plan
contradice el SRS, prevalece el SRS; las decisiones de arquitectura aprobadas en `AGENTS.md`
siguen definiendo cómo se implementa.

## Alcance

- Una aplicación Go + Fiber organizada en módulos de dominio y una base PostgreSQL.
- React SPA para el portal web responsive.
- JWT Bearer para autenticar solicitudes, de acuerdo con la arquitectura sugerida por el SRS.
- Redis solo para caché; MinIO queda diferido según `AGENTS.md`.
- Ecuador y LOPDP como jurisdicción aplicable. No se declara cumplimiento HIPAA.

El JWT vence en dos horas y el middleware además cierra el acceso tras 15 minutos sin actividad.
El seguimiento es en memoria para el monolito de un solo proceso; al reiniciar, la inactividad del
primer acceso se calcula desde `iat`.

## Identidad y seguridad requeridas por el SRS

Roles del SRS: `admin_ti`, `recepcionista_cajero`, `odontologo`, `asistente_dental` y
`auditor_contador`.

- `POST /auth/login` y middleware JWT Bearer: implementados.
- Helper `authn.RequireAnyRole` disponible; conectar los permisos RBAC definidos por el SRS a cada
  caso de uso conforme se implementen los endpoints de dominio. `GET /users` ya exige `admin_ti`.
- Gestión mínima de usuarios por `admin_ti`: listar, crear y modificar rol/estado. Las cuentas se
  desactivan en vez de borrarse.
- Cierre tras 15 minutos de inactividad en estaciones clínicas: implementado por el middleware.
- Auditoría inmutable de cada CRUD clínico con actor, timestamp ISO-8601 e IP: pendiente para
  Fase 2, cuando existan operaciones clínicas.

No se detallan sesiones por cookie, refresh tokens, bloqueo de cuentas ni un CRUD administrativo
más amplio que la gestión de usuarios requerida por el perfil `admin_ti`.

## Fase 1 — Core administrativo y turnos

Según SRS 8.1: pacientes, agendamiento, catálogo médico y facturación básica.

- `RF-01`: ficha del paciente, contacto de emergencia, tutor legal para menores y seguro/convenio.
- `RF-02`: historial médico general y alertas críticas visibles en la cabecera del expediente.
- `RF-09` y `RF-10`: agenda por recinto/sillón/profesional y estados de turno definidos por el SRS.
- `RN-03`: impedir citas superpuestas en el mismo sillón.
- `RF-07` y `RF-08`: presupuestos por fases, costos y descuentos; aprobación total o parcial.
- `RF-12`: crear y emitir facturas dentro del sistema y registrar pagos mixtos; sin conexión
  tributaria externa.
- `RN-02`: el cierre de caja no revela el total teórico antes del conteo físico.
- Portal web responsive para los flujos anteriores (RNF-07 y RNF-08).

## Fase 2 — Core clínico

Según SRS 8.1:

- `RF-03`: odontograma 2D con dentición, notación y catálogo indicados en el SRS.
- `RF-04`: periodontograma y cálculo del nivel de inserción clínica.
- `RF-05`: notas SOEP cronológicas, inmutables, firmadas y bloqueadas según el plazo indicado.
- `RF-06`: consentimiento informado y captura de firma o validación OTP.
- Recetas odontológicas, como parte del objetivo del sistema y del core clínico.
- `RN-01`: no borrar procedimientos realizados; las correcciones se registran como adendas.
- `RNF-05`: el cierre por inactividad de 15 minutos debe proteger el acceso a este módulo.
- `RNF-06`: auditoría inmutable para cada CRUD clínico con actor, fecha/hora e IP.

## Fase 3 — Finanzas avanzadas, inventarios y reportes

Según SRS 8.1:

- `RF-13`: cálculo configurable de comisiones. La fórmula concreta sigue abierta hasta su decisión.
- `RF-14`: descargo de insumos al completar procedimientos, según receta estándar de consumo.
- Reportes de inventario, finanzas y operación para los perfiles indicados en el SRS.

## Fase 4 — UAT, seguridad y salida a producción

Según SRS 8.1 y 8.2:

- UAT y pruebas de los flujos de las fases anteriores.
- Cobertura mínima del 80% en endpoints críticos.
- DAST/SAST sin vulnerabilidades altas o críticas.
- Documentación API OpenAPI 3.0.
- Verificación de rendimiento, concurrencia, disponibilidad, cifrado, navegadores y responsive
  según RNF-01 a RNF-08.
- Despliegue de producción con monitoreo y respaldos.

## Fuera del alcance actual o diferido

- Aplicación móvil nativa y procesamiento DICOM: fuera de Fase 1 según SRS 2.3.
- Integraciones tributarias externas, archivos clínicos y MinIO: fuera del alcance acordado.
- Fórmula de comisiones RF-13: pendiente de decisión.
- CI: queda para el final por instrucción del usuario.

## Estado actual

- Base Go/Fiber, configuración, migraciones, PostgreSQL, Redis de caché y `/health` inicial: en el
  repositorio.
- Login JWT y `RequireAuth`: implementados.
- Timeout de 15 minutos implementado. El RBAC debe aplicarse dentro de cada servicio de dominio;
  no exponer información clínica hasta conectar allí los permisos del SRS.
