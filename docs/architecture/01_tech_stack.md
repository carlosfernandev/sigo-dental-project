# Stack Tecnológico y Arquitectura — SIGO-Dental

## Core

- **Lenguaje:** Go (Golang).
- **Framework HTTP:** Fiber v2.
- **Arquitectura:** Monolito modular con Vertical Slicing (Domain-Driven Design a nivel de carpetas).
- **Estructura de Dominio:** Cada módulo vive en `internal/<nombre_dominio>/` y contiene exclusivamemte sus propios: `dto.go`, `entity.go`, `repository.go`, `service.go` y `handler.go`.
- **Rutas:** Centralizadas o montadas desde cada dominio en la inicialización (según convención de `backend/cmd/api`).

## Infraestructura y Datos

- **Base de Datos Principal:** PostgreSQL (fuente única de verdad).
- **Caché y Sesiones:** Redis (solo caché y rate limiting; si Redis cae, el sistema debe degradarse elegantemente, no fallar).
- **Restricción MVP:** Cero dependencias externas. El sistema debe poder operar 100% offline o en red local. No hay integraciones con pasarelas de pago, autoridades tributarias, ni APIs de mensajería (Twilio, WhatsApp). Todo se simula con registros internos.
