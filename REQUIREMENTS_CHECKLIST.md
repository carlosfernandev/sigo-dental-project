# REQUIREMENTS CHECKLIST — SIGO-Dental

> Alcance y fases tomados de `docs/SRS_SIGO-Dental.md`. `AGENTS.md` define la arquitectura y las
> decisiones de implementación del propietario. No añadir tareas de producto que no provengan del
> SRS. Verificar una tarea antes de marcarla y actualizar este estado en el mismo cambio.

## Estado General

| Item | Value |
| --- | --- |
| **Fase actual** | Fase 1 — Core administrativo y turnos |
| **Último trabajo completado** | RNF-05: timeout de inactividad y gestión de usuarios limitada a `admin_ti` |
| **Siguiente requisito** | RF-01: comenzar ficha de paciente con acceso por rol definido por el SRS |
| **Pendiente para el final** | CI, por instrucción del usuario |
| **Arquitectura** | Monolito Go + Fiber, PostgreSQL único, Redis solo caché |

## Base técnica

- [x] Backend Go/Fiber con composición en `backend/cmd/api`.
- [x] Configuración, logging, errores HTTP comunes y PostgreSQL.
- [x] Migraciones SQL forward-only y seed del primer administrador.
- [x] Redis configurado para caché; sus errores no deben hacer indisponible el sistema.
- [x] Estructura por dominio y utilidades compartidas `platform`, `cache` y `middleware`.
- [ ] `GET /health` comprueba PostgreSQL (503 si no responde) y Redis (degradado, manteniendo 200).
- [ ] CI: ejecutar al final del alcance, según instrucción del usuario.

## Fase 1 — Core administrativo y turnos

### Identidad y acceso

- [x] Usuarios y roles del SRS persistidos; seed idempotente de `admin_ti`.
- [x] `POST /auth/login` valida credenciales y emite JWT HS256.
- [x] `middleware.RequireAuth` valida el Bearer token y expone usuario/rol.
- [ ] RNF-05: aplicar la matriz RBAC a los casos de uso de cada dominio cuando se implementen.
- [x] `admin_ti` gestiona usuarios: listar, crear y actualizar rol/estado; no se permite borrar cuentas.
- [x] RBAC en identidad: `/users` solo permite `admin_ti`; otros roles reciben `403`.
- [x] RNF-05: cerrar acceso tras 15 minutos de inactividad; el middleware rechaza las solicitudes
      autenticadas y requiere volver a iniciar sesión. El seguimiento vive en memoria del monolito.
- [x] `authn.RequireAnyRole` aplicado en los casos de uso de administración de usuarios.

### Pacientes y catálogo médico

- [ ] RF-01: ficha, datos de contacto, emergencia, tutor legal para menores y seguro/convenio.
- [ ] RF-02: historial médico obligatorio y alertas críticas en la cabecera del expediente.
- [ ] Catálogo médico requerido por la descripción de Fase 1 del SRS.

### Agenda y turnos

- [ ] RF-09: calendario por recinto, sillón y profesional.
- [ ] RF-10: estados de cita definidos en el SRS.
- [ ] RN-03: impedir citas superpuestas en el mismo sillón, también bajo concurrencia.
- [ ] RF-11: confirmaciones y recordatorios 48 h y 24 h por los canales definidos en el SRS.

### Presupuestos, facturación y portal

- [ ] RF-07: presupuestos por fases/sesiones, con honorarios, laboratorio y descuentos autorizados.
- [ ] RF-08: aprobación total/parcial y transición de ítems aprobados a tareas agendables.
- [ ] RF-12: crear y emitir facturas dentro del sistema y registrar pagos mixtos; sin conexión
      tributaria externa.
- [ ] RN-02: en el cierre de caja, ocultar el total teórico hasta recibir el conteo físico.
- [ ] Portal web responsive para los flujos de esta fase (RNF-07 y RNF-08).

## Fase 2 — Core clínico

- [ ] RF-03: odontograma interactivo con dentición, notación, superficies y catálogo del SRS.
- [ ] RF-04: periodontograma con mediciones y cálculo del nivel de inserción clínica.
- [ ] RF-05: notas SOEP cronológicas, inmutables, con firma profesional y bloqueo tras el plazo
      configurable indicado por el SRS.
- [ ] RF-06: consentimiento informado por procedimiento con firma o validación OTP.
- [ ] Recetas odontológicas, según el objetivo y la descripción de esta fase en el SRS.
- [ ] RN-01: ningún procedimiento realizado se elimina; las correcciones son adendas.
- [ ] RNF-05: conectar RBAC en los casos de uso clínicos y conservar el timeout de 15 minutos.
- [ ] RNF-06: auditar cada CRUD clínico con actor, timestamp ISO-8601 e IP; el registro es inmutable.

## Fase 3 — Finanzas avanzadas, inventarios y reportes

- [ ] RF-13: liquidación configurable de honorarios. Implementar cuando se decida la fórmula.
- [ ] RF-14: descontar insumos al completar procedimientos según la receta estándar configurada.
- [ ] Reportes de inventario y finanzas contemplados para esta fase en el SRS.

## Fase 4 — UAT, seguridad y producción

- [ ] UAT de los flujos de las fases anteriores.
- [ ] ≥80% de cobertura en endpoints críticos del backend (SRS 8.2).
- [ ] DAST/SAST sin vulnerabilidades altas o críticas (SRS 8.2).
- [ ] Documentación API OpenAPI 3.0 (SRS 8.2).
- [ ] Verificar RNF-01 a RNF-08: rendimiento, disponibilidad, concurrencia, cifrado, seguridad,
      responsive y navegadores.
- [ ] Salida a producción y monitoreo.

## Diferimientos decididos

- Integración tributaria externa: fuera del alcance; la factura se crea y emite dentro del sistema.
- MinIO y carga de documentos: diferidos hasta que documentos clínicos estén en alcance.
- App móvil nativa y procesamiento DICOM: fuera de Fase 1 según SRS 2.3.
- Fórmula concreta RF-13: pendiente de decisión.
- CI: se ejecutará al final por instrucción del usuario.

## Reglas de actualización

1. Leer este **Estado General** y el siguiente requisito pendiente antes de empezar.
2. Implementar solo trabajo necesario para los requisitos del SRS y decisiones de `AGENTS.md`.
3. No iniciar una pantalla antes de que exista y esté verificado el endpoint correspondiente.
4. Marcar completado solo después de verificar el criterio aplicable.
5. Mantener `docs/specs/01-fundacion-identidad-auditoria.md` alineado con este checklist.
