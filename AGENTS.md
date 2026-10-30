# AGENTS.md

## Fuentes y flujo de trabajo

- `docs/SRS_SIGO-Dental.md` (actualmente v1.3) es la autoridad de alcance; léelo antes de diseñar
  o implementar.
- `REQUIREMENTS_CHECKLIST.md` define el estado y orden: lee primero **Estado General** y toma la
  primera tarea sin marcar de la fase actual. Verifica el trabajo y actualiza tarea y estado en el
  mismo cambio.
- `PROJECT_MAP.md` describe el árbol; `rules/GLOSSARY.md` es la autoridad para nombres de dominio.
- Backend antes que frontend: no iniciar una pantalla sin que exista y esté verificado su endpoint.
- Respeta el límite de fase del SRS. No iniciar el core clínico antes de completar Fase 1; MinIO
  queda para el final del alcance.

## Arquitectura y prácticas

Las decisiones vigentes de stack, límites de dominio, API y persistencia están en
[`docs/architecture/`](docs/architecture/). Léelas y mantenlas coherentes con el SRS y el código;
si se actualiza una decisión, actualiza allí su documento y los artefactos afectados. No dupliques
su detalle aquí.

Usa las convenciones de `rules/GLOSSARY.md`: no inventes sinónimos para entidades o estados.
Para trabajo de interfaz, sigue [`docs/design/PLAN_DISENO_FRONTEND.md`](docs/design/PLAN_DISENO_FRONTEND.md) y las skills
`dental-ui-*`.

## Verificación y comandos

- Usa el Makefile raíz: `make format`, `make lint`, `make test`, `make build`, `make migrate-up` y
  `make seed-admin` según corresponda.
- No marcar una tarea completada sin la verificación pertinente. CI se ejecuta al final del alcance,
  según instrucción del usuario.
- Criterios de aceptación del SRS: ≥80% de cobertura en endpoints críticos, DAST/SAST sin
  vulnerabilidades altas o críticas y OpenAPI 3.0 entregado.

## Convenciones del repositorio

- Conventional Commits. No agregar `Co-Authored-By` ni atribución de IA.
- Código, comentarios, identificadores y commits en inglés; español solo para el SRS y el texto
  dirigido al personal/pacientes de la clínica.
- `.atl/` es generado con `gentle-ai skill-registry refresh --force`; no editarlo a mano.
- `opencode.json` y `.gentle-ai-default-agent.json` se generan por máquina con `gentle-ai sync` y
  no se versionan. Cada desarrollador ejecuta su propio sync.
- Lee `rules/ahorro-tokens.md` antes de navegar, consultar memoria o ejecutar pruebas.

## Ejecución de Comandos Go (Sandbox Fix)
- El entorno de ejecución está restringido. NUNCA intentes escribir en la caché global de Go.
- Para ejecutar tests o compilar, SIEMPRE antepón las variables de entorno para usar la carpeta `/tmp/` o una ruta local.
- Comando obligatorio para tests: `GOCACHE=/tmp/go-cache GOMODCACHE=/tmp/go-modcache go test ./...`
- Para comandos como este y otros, consulta el Makefile en la raiz del proyecto.
## CodeGraph

No ejecutar `codegraph init` todavía. Este repositorio aún no tiene un árbol de código suficiente
para indexar; reconsiderarlo cuando exista una base de código real.
