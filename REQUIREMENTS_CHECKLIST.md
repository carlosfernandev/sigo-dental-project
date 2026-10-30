# REQUIREMENTS CHECKLIST — SIGO-Dental

> Source of truth for implementation progress. Read this at the start of every session.
> Mark `- [x]` when a task is **verified** (not when it is written) and update **Estado General**
> in the same edit.
>
> **Order is strict: backend before frontend.** No UI task starts until the endpoint it consumes
> exists, returns the documented envelope, and has a passing test. The frontend consumes the
> gateway contract, never an imagined one.
>
> **Companion documents:** `docs/SRS_SIGO-Dental.md` (scope), `AGENTS.md` (architecture and
> conventions), `rules/GLOSSARY.md` (naming). If this file contradicts the SRS, the SRS wins and
> this file is wrong.

## Estado General

| Item | Value |
| --- | --- |
| **Current milestone** | Backlog 01 — identidad (CI diferido por instrucción del usuario) |
| **Last completed task** | Backlog 01 `1.1` — roles, users y seed del administrador inicial |
| **Next pending task** | Backlog 01 `1.2` — política de contraseñas con bcrypt según `AGENTS.md` |
| **Blockers** | `0.4` CI queda para el final; las migraciones son forward-only según `AGENTS.md` |
| **Services in flight** | Monolito — identity foundation |

## Target architecture

```text
React SPA (:5173) ──HTTPS/JWT──► gateway (:8080)
                                   ├─► identity   (:8081)
                                   ├─► patient    (:8082)
                                   ├─► scheduling (:8083)
                                   ├─► clinical   (:8084)
                                   ├─► billing    (:8085)
                                   └─► inventory  (:8086)

PostgreSQL 16 (:5432, one database per service) · Redis (:6379) · MinIO (:9000, deferred)
```

## Guiding principles

- **Security by default.** No endpoint is born without auth, input validation, and the error
  envelope. A route added without `RequireAuth` is a defect, not a TODO.
- **Dependency injection at the boundaries.** Interfaces where a service meets I/O; no package
  globals except in the composition root (`cmd/api/main.go`).
- **Pure functions for domain logic.** Money math, availability windows, commission splits, and
  validators are testable without a database.
- **One service, one database.** No cross-service query, ever. If a screen needs two services,
  the frontend makes two calls.
- **Events are contracts.** Every event has a version and a contract test.
- **Verify, then check the box.** A task with a failing test stays unchecked.

## Deferred — do not build these yet

| Item | Unblocks when |
| --- | --- |
| MinIO storage, upload endpoints, presigned URLs | Fase 2 needs consent PDFs or radiographs |
| SRI e-invoicing (stamping, fiscal XML, access keys, certificates) | Core features complete |
| Native mobile app, DICOM processing | Out of scope for Fase 1 per SRS 2.3 |
| RF-13 commission formula choice | Start of Fase 3 |

---

# FASE 0 — BOOTSTRAP (no product behavior)

> Goal: a repository where a new service costs one directory, not one architecture decision.

- [ ] `F0-B1` Layout created: `backend/platform/`, `backend/services/<name>/`, `frontend/`,
      `deploy/`, `migrations/<service>/`
- [ ] `F0-B2` `backend/go.work` joins `platform` and every service module; each service has its
      own `go.mod`
- [ ] `F0-B3` `deploy/docker-compose.yml`: PostgreSQL 16 and Redis with healthchecks and named
      volumes; ports 5432 and 6379
- [ ] `F0-B4` `platform` module: env config loader, `slog` logger, error envelope types, response
      helpers, typed domain-error → HTTP mapping
- [ ] `F0-B5` Migration runner: `migrations/<service>/NNN_name.sql`, applied in order, tracked in
      a `schema_migrations` table per service. Forward-only.
- [ ] `F0-B6` Gateway skeleton: `/health`, service route table, reverse proxy, 503 when a target
      is down
- [ ] `F0-B7` Outbox infrastructure in `platform`: `outbox` table contract, writer helper, relay
      loop, Redis Streams publisher with consumer groups
- [ ] `F0-B8` CI: `go vet`, `go test -race`, `golangci-lint`, frontend `lint` + `tsc --noEmit`
- [ ] `F0-B9` `.gitignore` covers `.env`, binaries, `node_modules/`, `dist/`, coverage output
- [ ] `F0-F1` Frontend scaffold: Vite + React + TypeScript + Tailwind, `app/` `features/`
      `shared/` layout, `shared/lib/http.ts`, TanStack Query provider, route-level lazy loading

---

# FASE 1B — BACKEND: CORE ADMINISTRATIVO Y TURNOS

> SRS Fase 1. Services: `gateway`, `identity`, `patient`, `scheduling`, `billing`.

## Identity

- [ ] `F1-B01` `users`, `roles`, `user_roles` tables; bcrypt hashing on create and password change
- [ ] `F1-B02` `POST /auth/login` → JWT HS256, claims `{sub, role, iat, exp}`, 2 h TTL, secret from
      env. No secret default in code.
- [ ] `F1-B03` Local token validation middleware in every service (`RequireAuth`, `RequireRole`).
      No per-request call to `identity`.
- [ ] `F1-B04` RNF-05 idle timeout: session closes after 15 minutes without activity, tracked
      server-side on the refresh token. A short access-token TTL alone is not an idle timeout —
      it logs out an active clinician mid-procedure.
- [ ] `F1-B05` RNF-05 MFA (TOTP) enforced for roles that reach clinical or financial data
- [ ] `F1-B06` `GET /health` pinging PostgreSQL (503 if down) and Redis (`degraded` if down, 200)

## Patient

- [ ] `F1-B07` Patient CRUD: demographics, emergency contact, `LegalGuardian` for minors,
      `InsurancePlan`
- [ ] `F1-B08` `MedicalHistory`: systemic conditions, drug allergies, surgeries, coagulopathies
- [ ] `F1-B09` `CriticalAlert` derived and returned with the record header (RF-02) — computed
      server-side, never inferred by the UI
- [ ] `F1-B10` RNF-06: every clinical CRUD writes an audit row in the same transaction via the
      outbox (actor, ISO-8601 UTC timestamp, source IP)
- [ ] `F1-B11` `GET /health`

## Scheduling

- [ ] `F1-B12` `Operatory` and `Chair` CRUD
- [ ] `F1-B13` `Appointment` CRUD with the RF-10 state machine: `scheduled`, `confirmed`,
      `checked_in`, `in_progress`, `completed`, `cancelled`, `no_show`. Invalid transitions
      rejected with a stable `code`.
- [ ] `F1-B14` **RN-03 enforced in PostgreSQL**: `btree_gist` + `EXCLUDE USING gist (chair_id
      WITH =, tstzrange(starts_at, ends_at) WITH &&)`. Cancelled appointments excluded. A
      concurrency test with two parallel inserts must prove one fails.
- [ ] `F1-B15` Availability query: free slots by chair and by dentist over a date range (RF-09)
- [ ] `F1-B16` Cache-Aside for availability with a short TTL; cache miss falls through to
      PostgreSQL; Redis failure degrades latency only
- [ ] `F1-B17` `GET /health`

## Billing (basic — no SRI)

- [ ] `F1-B18` `TreatmentQuote` with phases, items, professional fees, lab costs, authorized
      discounts (RF-07)
- [ ] `F1-B19` Total recalculated server-side on every write. A client-supplied total is ignored.
- [ ] `F1-B20` Quote approval, full or partial, storing approved item ids (RF-08)
- [ ] `F1-B21` Emit `billing.quote.approved.v1` through the outbox; `scheduling` consumes it and
      creates agendable tasks. Contract test on both sides.
- [ ] `F1-B22` Payment capture with `PaymentSplit` (cash, card, transfer, insurance) — RF-12
      without fiscal submission
- [ ] `F1-B23` **RN-02 blind cash close**: the receptionist's response omits the theoretical total
      until the physical count is submitted. Verified by a test asserting the field is absent
      from the JSON, not merely hidden in the UI.
- [ ] `F1-B24` `GET /health`

## Gateway

- [ ] `F1-B25` Route table per service prefix, JWT validation, rate limiting, unified error
      envelope for upstream failures
- [ ] `F1-B26` Aggregated OpenAPI 3.0 document served from the gateway; per-service documents
      generated from their handlers

---

# FASE 1F — FRONTEND: CORE ADMINISTRATIVO

> Starts only after the matching `F1-B*` task is checked.

- [ ] `F1-F01` Auth flow: login, token storage, role-based route guards, silent refresh
- [ ] `F1-F02` `shared/types/api.ts` generated from the gateway OpenAPI; regeneration script wired
      into `package.json`
- [ ] `F1-F03` Patients: list with search, detail, create/edit form, Zod validation
- [ ] `F1-F04` Clinical header with `CriticalAlert` badge from the API payload
- [ ] `F1-F05` Scheduling: day/week calendar filterable by chair and by dentist
- [ ] `F1-F06` Appointment board with state transitions limited to legal moves
- [ ] `F1-F07` Quote builder and approval screen with partial selection
- [ ] `F1-F08` Payment capture and cash-close screen; theoretical total absent until count submitted
- [ ] `F1-F09` RNF-07 responsive check at 1920×1080, 1366×768, and tablet ≥10.5"
- [ ] `F1-F10` `GET /health` surfaced in a dev-only status page

---

# FASE 2B — BACKEND: CORE CLÍNICO

- [ ] `F2-B01` `Odontogram` model: 32 permanent + 20 primary teeth, `FdiNotation` and
      `UniversalNotation`, 5 surfaces per tooth (`mesial`, `distal`, `occlusal`/`incisal`,
      `buccal`, `lingual`/`palatal`)
- [ ] `F2-B02` Odontogram state catalog: caries, restoration, endodontic, implant, fixed/removable
      prosthesis, extraction, sealant
- [ ] `F2-B03` `PeriodontalChart`: probing depth (6 points per tooth), gingival margin,
      `ClinicalAttachmentLevel`, furcation, mobility, `BleedingOnProbing`
- [ ] `F2-B04` `ProgressNote` (SOEP) append-only with configurable lock (default 24 h) and
      professional `DigitalSignature` (RF-05)
- [ ] `F2-B05` **RN-01**: no delete path for completed procedures. Corrections are `Addendum`
      rows referencing the original. A test must prove `DELETE` is unavailable at the API level.
- [ ] `F2-B06` `InformedConsent` generation per procedure with signature capture or OTP validation
      (RF-06)
- [ ] `F2-B07` `Prescription` issuance and printing payload
- [ ] `F2-B08` Emit `clinical.procedure.completed.v1` on completion (consumed in Fase 3)
- [ ] `F2-B09` RF-11 notifications: appointment confirmation plus 48 h and 24 h reminders through
      the messaging gateway; delivery failures retried, never blocking the request
- [ ] `F2-B10` Audit trail query endpoint for `Auditor` role
- [ ] `F2-B11` `GET /health`

---

# FASE 2F — FRONTEND: CORE CLÍNICO

- [ ] `F2-F01` Interactive odontogram (SVG): permanent/primary toggle, FDI/Universal toggle,
      per-surface marking
- [ ] `F2-F02` Periodontogram grid with per-tooth entry and computed CAL
- [ ] `F2-F03` SOEP editor with visible lock countdown and read-only state after lock
- [ ] `F2-F04` Addendum flow that never mutates the original note
- [ ] `F2-F05` Informed consent capture with signature pad
- [ ] `F2-F06` Prescription preview and print layout

---

# FASE 3B — BACKEND: FINANZAS AVANZADAS E INVENTARIO

- [ ] `F3-B01` RF-13 `FeeSettlement` with the chosen formula (blocked on the open decision)
- [ ] `F3-B02` `inventory` consumes `clinical.procedure.completed.v1` and deducts
      `SupplyConsumption` per the configured standard recipe
- [ ] `F3-B03` Stock deduction with `SELECT ... FOR UPDATE`; a concurrency test must prove no
      negative stock
- [ ] `F3-B04` `Inventory` CRUD, stock levels per operatory and per professional
- [ ] `F3-B05` `CashReport` and `FeeSettlement` reporting endpoints
- [ ] `F3-B06` BI endpoints for the dashboard (appointments, production, collections)
- [ ] `F3-B07` `GET /health`

---

# FASE 3F — FRONTEND: FINANZAS Y REPORTES

- [ ] `F3-F01` Fee settlement screen with formula configuration
- [ ] `F3-F02` Inventory screen: levels, consumption history, low-stock alerts
- [ ] `F3-F03` Dashboard: production, collections, no-show rate

---

# FASE 4 — UAT, SECURITY, PRODUCTION

- [ ] `F4-01` Coverage ≥80% on critical backend endpoints (SRS 8.2)
- [ ] `F4-02` DAST/SAST clean for High and Critical (OWASP Top 10)
- [ ] `F4-03` Load test at 150 concurrent users without latency degradation (RNF-03)
- [ ] `F4-04` Record read <1.5 s with odontogram render (RNF-01)
- [ ] `F4-05` LOPDP review: data retention, access logging, export controls. No HIPAA claims.
- [ ] `F4-06` Production deploy: TLS 1.3, AES-256 at rest, backups, monitoring on `/health`
- [ ] `F4-07` OpenAPI 3.0 artifact published with the release

---

## Update protocol

1. Read **Estado General**, then the first unchecked task in the current phase.
2. Work one task at a time. A task that reveals a missing prerequisite gets a new task above it,
   not a silent detour.
3. Check a box only after its test passes and the behavior was observed.
4. Update **Estado General** (milestone, last completed, next pending, blockers) in the same edit.
5. Record a decision in `AGENTS.md` if it changes architecture or conventions; record it in
   Engram with `mem_save` if it changes scope.
