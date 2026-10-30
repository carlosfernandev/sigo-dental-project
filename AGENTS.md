# AGENTS.md

## What this repo is

`docs/SRS_SIGO-Dental.md` (v1.0, 2026-10-29) is the requirements source of truth;
`REQUIREMENTS_CHECKLIST.md` tracks implementation progress and task order;
`PROJECT_MAP.md` describes where things live; `rules/GLOSSARY.md` maps its Spanish terms to
code identifiers. The repository contains an initial Go backend foundation; `PROJECT_MAP.md`
describes the current tree. Use the root Makefile for common Go checks.

**Read the SRS before designing or writing anything.** It is in Spanish and is the only
authority on scope.

## Implementation flow

`REQUIREMENTS_CHECKLIST.md` is the working brain: read its **Estado General** first, then take
the first unchecked task in the current phase. Backend before frontend — no UI task starts
before the endpoint it consumes exists and is tested. Update the checklist and its Estado
General in the same edit that completes a task.

## Decided stack

Owner decisions. Do not re-litigate them.

| Concern | Choice |
| --- | --- |
| Frontend | React (SPA, PWA in Fase 2) |
| Backend | Go + Fiber (monolith with domain modules) |
| Database | PostgreSQL (single database, public schema for all domains) |
| Cache | Redis |
| Async messaging | None (synchronous monolith calls) |
| Object storage | MinIO — **deferred, implement last** |

### Reference versions

From the owner's reference project. Re-verify with `go list -m -versions` and `docker pull`
before pinning in `go.mod`; do not copy these blindly.

| Component | Version |
| --- | --- |
| Fiber | v3.x |
| PostgreSQL | 16 |
| pgx | v5 (`pgxpool`) |
| go-redis | v9 |
| golang-jwt | v5, HS256 |
| Passwords | `golang.org/x/crypto/bcrypt` |

MinIO is explicitly last: no upload endpoints, buckets, or presigned-URL flows until clinical
documents (consent PDFs, radiographs) are in scope.

## Architecture: monolith with domain modules

Decided by the owner to reduce operational complexity. One Go binary serving all domains
(identity, patient, scheduling, clinical, billing, inventory) with a single PostgreSQL database.

```text
frontend/  ──HTTPS/JWT──▶  monolith (Go + Fiber)  ──▶  PostgreSQL
                                                └─▶  Redis (cache)
```

| Domain module | Owns |
| --- | --- |
| `identity` | users, roles, sessions, MFA |
| `patient` | patients, legal guardians, medical history, allergies, insurance |
| `scheduling` | chairs, operatory, appointments, appointment states |
| `clinical` | clinical record, odontogram, periodontogram, progress notes, consents |
| `billing` | quotes, invoices, payments, cash close, fee settlement |
| `inventory` | supplies, consumption, stock |

### Non-negotiable monolith rules

- **One database, public schema for all domains.** Domain boundaries are enforced by code
  organization and explicit cross-domain transaction boundaries, not by separate databases.
- **Domain modules are package boundaries.** Import direction is one-way: `api → handler → service → repository`.
  A domain module may import another's repository for cross-domain queries, but never the reverse.
- **Synchronous calls for everything.** No async messaging. Cross-domain flows are function calls
  within the same process, wrapped in transactions when atomicity is required.
- **AuthN at HTTP middleware, AuthZ in the owning domain's service layer.** The JWT carries the role;
  the service layer checks permissions before executing use cases.
- **Single HTTP server.** All endpoints live in one Fiber app with domain-scoped route groups.

### Repository layout

```text
backend/
  platform/                  # shared packages: config, logging, errors, tracing, redis
  cmd/api/                   # composition root: config, pool, routes, start
  internal/
    identity/                # domain module: entity, service, repository, handler, dto
    patient/
    scheduling/
    clinical/
    billing/
    inventory/
  migrations/                # versioned SQL, forward-only, single directory for monolith
frontend/
deploy/                      # docker compose: one postgres, one redis, one monolith container
```

Each domain module is flat and self-contained:

| File | Responsibility |
| --- | --- |
| `entity.go` | Domain types and invariants. Imports no framework, no driver. |
| `service.go` | Use cases. Owns the transaction boundary and RBAC checks. |
| `repository.go` | Port interface + PostgreSQL implementation. |
| `handler.go` | Fiber HTTP mapping only. |
| `dto.go` | Request/response shapes and validation. |

**Do not create top-level `controllers/`, `models/`, `repository/`, or `routes/`.** That
layer-first layout is the anti-pattern this project rejects: one feature edit touches every
folder, and nothing in the tree marks a domain boundary.

Dependency direction is one-way: `handler → service → repository interface`. Domain code never
imports Fiber, pgx, or Redis. The PostgreSQL implementation is the only place that imports a
driver.

### Frontend layout (React)

Organize by **feature**, not by file type.

```text
frontend/
  src/
    app/                   # router, providers, query client
    features/
      patients/{api.ts,components/,pages/,schemas.ts}
      scheduling/
      clinical/
      billing/
    shared/
      ui/                  # presentational primitives, no data fetching
      lib/                 # http client, formatters, guards
      types/api.ts         # GENERATED from the monolith OpenAPI — never hand-edited
```

**Do not create top-level `components/`, `hooks/`, `services/`, or `pages/`.** Cross-feature
imports go through `shared/`, never `features/a → features/b`.

The frontend talks to the monolith API directly. It does not know the internal domain topology.

## Backend practices (Go)

- Money is `int64` minor units (cents). Never `float64` for money. USD only — no currency field.
- Persist `timestamptz` in UTC. Convert to `America/Guayaquil` only at the presentation edge.
- Every domain service and repository method takes `ctx context.Context` first. No
  `context.Background()` inside business code.
- Errors are typed domain values. One Fiber error handler maps them to the response
  envelope and HTTP status; handlers never build error JSON ad hoc.
- The domain service opens transactions and passes a querier to repositories. Repositories never start
  their own transaction.
- SQL uses explicit column lists — no `SELECT *` — and parameters only, never string building.
- No `panic` on the request path. Return an error.
- RBAC and RN-02 data shaping live in the domain service layer. Route middleware alone is not enough:
  RN-02 requires omitting a total from the payload, not hiding it in the UI.
- Table-driven tests. `go test -race ./...` is the gate for the monolith.
- **Redis is cache, never a source of truth.** Every cache read must have a miss path that hits
  PostgreSQL, and a Redis outage must degrade latency, not availability. A cache error is logged
  and ignored; it never becomes a 5xx.
- **The monolith exposes `GET /health`.** It pings PostgreSQL (503 if down) and Redis
  when used (`"degraded"` in the payload if Redis is down, still 200). RNF-02 monitoring uses this endpoint.
- Pessimistic locking (`SELECT ... FOR UPDATE`) is required for RN-02 cash close and for stock
  deduction. An `UPDATE ... WHERE stock >= n` is acceptable only where a single statement fully
  expresses the invariant; otherwise lock, then read, then write.

## Frontend practices (React)

- One HTTP client in `shared/lib/http.ts`. No bare `fetch` in components or feature modules.
- Server state belongs to TanStack Query. Never mirror server data into React context or a
  global store — two sources of truth produce stale clinical screens.
- Business rules live in the backend. Components validate shape, not policy.
- Zod schemas at every boundary (forms and API responses). No `any` crossing a boundary.
- `shared/types/api.ts` is generated from the gateway OpenAPI document. Hand-writing API types
  is a defect, not a shortcut.
- Route-level lazy loading. No barrel `index.ts` re-export chains.

## API contract

- OpenAPI 3.0 is the contract and a required deliverable (SRS 8.2). The monolith owns one
  document; the frontend consumes it directly.
- Every response uses one envelope: `{"data": ...}` on success,
  `{"error": {"code", "message", "details"}}` on failure.
- `code` is a stable machine-readable string (`chair_overlap`, `quote_already_approved`). Never
  branch on `message`; it is for humans and may be translated.
- Existing paths and response shapes are append-only. A breaking change requires a version bump,
  not a silent edit.

## Jurisdiction

Ecuador. "Autoridad tributaria local" / "legislación nacional" in the SRS means the **SRI**
(Servicio de Rentas Internas). "Hora local" is `America/Guayaquil` (UTC−5, no DST) — relevant
to the RNF-02 uptime window (07:00–21:00).

**SRI e-invoicing is deferred** until the main features are covered: no invoice stamping, fiscal
XML, access keys, or certificate handling yet. Fase 1 still needs basic billing and payment
capture; it does not need fiscal submission.

Still open: the three commission formulas in RF-13 (fixed %, fixed amount per procedure, % after
lab/material costs) are not chosen.

Data protection: the governing regime is Ecuador's Ley Orgánica de Protección de Datos
Personales (LOPDP), not the HIPAA reference in SRS 2.2. Do not claim HIPAA compliance in code,
docs, or UI copy.

## Constraints that must shape the data model, not the UI

- **RN-01 — clinical history is append-only.** Nothing recorded as "Realizado" may be deleted.
  Corrections are signed addenda rows. Design immutability up front; retrofitting it after CRUD
  endpoints exist is expensive.
- **RN-02 — blind cash close.** The receptionist must not see the theoretical total before
  entering the physical cash count. Authorization and response shaping, not a hidden field.
- **RN-03 — no chair double-booking.** Two appointments cannot overlap on the same physical
  chair even for different dentists. Enforced inside `scheduling` with a PostgreSQL exclusion
  constraint, not application validation, which loses under concurrency.
- **RNF-06 — immutable audit log** for every CRUD on clinical records: actor, ISO-8601 timestamp,
  source IP.
- **RF-05 — progress notes lock** after a configurable window (default 24 h) and require a
  professional signature.

## Scope fence

The SRS defines Fase 1–4 (section 8.1). Do not build clinical-core features (odontogram,
periodontogram, prescriptions, informed consent) while Fase 1 (patients, scheduling, medical
catalog, basic billing) is incomplete. Out of Fase 1: native mobile apps and DICOM processing —
patients get a PWA in Fase 2, radiology uses external PACS.

Start Fase 1 with the smallest set of domain modules that carries a real flow: `identity`,
`patient`, `scheduling`. Add `billing` and `clinical` when their first flow needs them. Do not
scaffold all modules at once — empty modules are coupling you pay for with no benefit.

## Domain vocabulary

`rules/GLOSSARY.md` is the authoritative SRS-Spanish → English-identifier map. Read it before
naming any table, struct, endpoint, or component. Do not invent synonyms: it is `Chair`, never
`DentalUnit`; `Appointment`, never `Booking`.

## Testing and acceptance gates (SRS 8.2)

- ≥80% unit/integration coverage on critical backend endpoints.
- DAST/SAST clean for High/Critical (OWASP Top 10).
- OpenAPI 3.0 spec delivered as an artifact.

## Development commands

- `make format` formats all Go files with `gofmt`.
- `make lint` checks Go formatting and runs `go vet ./...`.
- `make test` runs backend tests with the race detector.
- `make build` builds the API binary into the ignored `backend/bin/` directory.
- `make migrate-up` applies pending files from `backend/migrations/` in order.
- CI runs `make lint`, `make test`, `make build`, and `govulncheck` on pushes and pull requests.

## Conventions

- Conventional Commits only. Never add `Co-Authored-By` or any AI attribution.
- English for code, comments, identifiers, and commit messages; Spanish only for the SRS and
  end-user clinic copy.
- `.atl/` is auto-generated (`gentle-ai skill-registry refresh --force`). Do not hand-edit.

## Do not run `codegraph init` yet

Global instructions say to initialize CodeGraph when `.codegraph/` is missing. That rule assumes
an existing project; this repo has no code to index. Revisit after the first real source tree.
