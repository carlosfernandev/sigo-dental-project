# Fase 1 — Core administrativo y turnos

Este resumen usa únicamente los requisitos de la Fase 1 del SRS. El alcance funcional se define en
[`docs/SRS_SIGO-Dental.md`](../SRS_SIGO-Dental.md); el orden de trabajo está en
[`REQUIREMENTS_CHECKLIST.md`](../../REQUIREMENTS_CHECKLIST.md). `AGENTS.md` define la arquitectura
monolítica y sus decisiones técnicas.

## Incluido

- `RF-01` y `RF-02`: ficha del paciente, tutor legal, seguro, historial médico y alertas críticas.
- Catálogo médico mencionado en SRS 8.1.
- `RF-09` y `RF-10`: agenda y estados de turnos.
- `RN-03`: evitar citas simultáneas en el mismo sillón.
- `RF-07` y `RF-08`: presupuestos y su aprobación total o parcial.
- `RF-12`: crear y emitir facturas en el sistema y registrar pagos mixtos, sin conexión tributaria
  externa.
- `RN-02`: cierre de caja ciego.
- Portal web responsive (RNF-07 y RNF-08).
- Roles, RBAC y cierre tras 15 minutos de inactividad (RNF-05), según la versión vigente del SRS.

## Fuera de esta fase

- Odontograma, periodontograma, notas SOEP y consentimiento (`RF-03` a `RF-06`): Fase 2.
- Descargo de inventario y liquidación de honorarios (`RF-13` y `RF-14`): Fase 3.
- Integración tributaria externa y almacenamiento MinIO: fuera del alcance acordado.
- Aplicación móvil nativa y procesamiento DICOM: fuera del alcance indicado por el SRS.

## Organización

La Fase 1 usa una sola aplicación Go + Fiber y una base PostgreSQL. El trabajo se agrupa en los
módulos de dominio que ya sostienen los flujos: `identity`, `patient`, `scheduling` y, cuando se
inicie facturación básica, `billing`. Las rutas HTTP quedan dentro del dominio correspondiente;
`middleware` y `platform` contienen utilidades compartidas.
