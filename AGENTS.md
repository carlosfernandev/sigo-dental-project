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

Las decisiones vigentes de stack, límites de dominio, API y persistencia, y diseño de frontend están en
[`docs/architecture/`](docs/architecture/). Léelas y mantenlas coherentes con el SRS y el código;
si se actualiza una decisión, actualiza allí su documento y los artefactos afectados. No dupliques
su detalle aquí.

Usa las convenciones de `rules/GLOSSARY.md`: no inventes sinónimos para entidades o estados.
Para trabajo de interfaz, sigue [`docs/design/PLAN_DISENO_FRONTEND.md`](docs/design/PLAN_DISENO_FRONTEND.md) y las skills
`dental-ui-*`.

# Reglas de CRUD: SIGO-Dental

Fuente visual única: PLAN_DISENO_FRONTEND.md. Usa solo sus tokens.
CRUD de referencia: Pacientes. Todo CRUD nuevo copia su estructura.
Usa siempre los componentes compartidos: DataTable, EntityDrawer,
ConfirmDialog, FormField, StatusChip. No crees tablas ni formularios propios.

## Listado
- Cabecera: título del módulo a la izquierda, un único botón primario
  "Nuevo [entidad]" a la derecha.
- Búsqueda arriba de la tabla (atajo `/`), filtros máximo 3 visibles; el resto en "Más filtros".
- Tabla: filas de 48 px, orden por columna, paginación en servidor (25 por página).
- Cada fila: acción principal al hacer clic; acciones secundarias visibles
  al hover, al foco o con la fila seleccionada. Todas accesibles con teclado.
- Estados obligatorios: cargando (skeleton rectangular), vacío (explica qué
  hacer y ofrece "Nuevo [entidad]"), sin resultados de búsqueda, error
  (qué pasó y botón "Reintentar").

## Crear y editar
- Formularios cortos (hasta 8 campos): EntityDrawer de 480 px.
- Formularios largos o con secciones: página completa con tabs.
- No usar modal para formularios; el modal es solo para confirmaciones y firmas.
- Una columna de campos. Etiqueta siempre visible sobre el campo.
  Campos obligatorios marcados con el texto "(obligatorio)", no solo con asterisco.
- Validación: al salir del campo y al guardar. El error va bajo el campo,
  explica cómo corregir y se enlaza con aria-describedby.
  Al guardar con errores, el foco pasa al primer campo inválido.
- Botones: "Guardar [entidad]" (primario) y "Cancelar" (secundario),
  alineados a la izquierda al pie del formulario.
- Avisar antes de cerrar si hay cambios sin guardar.

## Eliminar
- Los registros clínicos nunca se borran (RN-01). La acción se llama
  "Desactivar" o "Anular", pide motivo y queda en auditoría.
- Solo datos sin valor clínico (catálogos, plantillas) permiten "Eliminar".
- Toda acción destructiva usa ConfirmDialog con el nombre de la entidad
  y el verbo exacto en el botón: "Desactivar paciente", no "Aceptar".

## Retroalimentación
- Toast de 8 s con "Deshacer" cuando aplique; role="status".
- El verbo del botón se repite en el toast: "Guardar paciente" → "Paciente guardado".
- Errores de servidor que bloquean: role="alert", no desaparecen solos.

## Permisos y auditoría
- Cada botón y columna declara qué rol puede verlo (recepción, odontólogo,
  asistente, administrador). Si no tiene permiso, no se muestra; no se deshabilita sin explicación.
- Crear, editar y desactivar registran usuario, fecha y valor anterior.

## Estilo
- Solo tokens del plan. Prohibido: hex, clases de paleta por defecto
  (emerald, teal, slate, blue), rounded-*, shadow-*, degradados, azul.
- Estados con color + trama + letra o texto, nunca solo color.

## Antes de entregar, confirma por escrito
1. Usé DataTable, EntityDrawer y FormField sin variantes propias.
2. Existen los cuatro estados del listado.
3. No hay borrado de registros clínicos.
4. Todo el flujo funciona solo con teclado.
5. No hay estilos fuera de los tokens.

## Verificación y comandos

- Usa el Makefile raíz: `make format`, `make lint`, `make test`, `make build`, `make migrate-up` y
  `make seed-admin` según corresponda.
- No marcar una tarea completada sin la verificación pertinente. CI se ejecuta al final del alcance,
  según instrucción del usuario.
- Criterios de aceptación del SRS: ≥80% de cobertura en endpoints críticos, DAST/SAST sin
  vulnerabilidades altas o críticas y OpenAPI 3.0 entregado.

## Context Limit Rules:

- NUNCA leas todo el workspace ni uses herramientas de búsqueda global a menos que se te pida explícitamente.

- Este proyecto usa Vertical Slicing en Go. Cada dominio vive aislado en internal/<dominio>/ con sus propios archivos entity.go, dto.go, repository.go, service.go y handler.go.

- Al implementar un requerimiento, lee ÚNICAMENTE el archivo específico en docs/specs/ y edita ÚNICAMENTE los archivos de la carpeta del dominio correspondiente.

- No generes explicaciones largas ni texto de relleno; devuelve solo código listo para producción.

- Manejo de base de datos: Usa PostgreSQL y migraciones SQL puras (forward-only).

- Framework HTTP: Usa Fiber v2.

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
