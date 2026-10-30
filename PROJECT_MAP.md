# PROJECT MAP — SIGO-Dental

> Architectural map. Read this **before** exploring code.
>
> **Rule: this file describes what EXISTS.** Every path listed here must be readable today.
> Planned paths are listed only in the clearly marked section below, because an agent that opens
> a nonexistent file wastes a turn and then guesses the design from the filename.

## Current state (2026-10-30)

**Fase 1 backend flows are in progress.** The Go/Fiber monolith includes identity and user
administration, patients, scheduling and appointments, RF-11's internal reminder queue and
15-minute generator, medical catalog, RF-08 quote approvals and schedulable tasks, RF-12 invoices and
mixed payments, and RN-02 cash close. Migrations are forward-only through
`011_quote_approvals_schedulable_tasks.sql`. Domain RBAC still needs
to be completed across all use cases; the frontend scaffold now includes login and base routing,
while the remaining portal flows, clinical and inventory flows, production deployment, and CI
execution remain pending.

| Path | Content |
| --- | --- |
| `docs/SRS_SIGO-Dental.md` | Requirements and scope authority. Spanish. |
| `docs/specs/01-fundacion-identidad-auditoria.md` | Concise implementation plan aligned with SRS scope and phases. |
| `docs/decisiones/0001-auth-jwt.md` | JWT Bearer login and in-process idle-timeout decision. |
| `REQUIREMENTS_CHECKLIST.md` | Progress, task IDs, and order. Backend before frontend. |
| `AGENTS.md` | Stack, architecture, conventions, Go/React practices. |
| `rules/GLOSSARY.md` | SRS-Spanish → English identifier map. |
| `Makefile` | `lint`, `format`, `test`, and `build` commands. |
| `.editorconfig` | Shared text and Go formatting defaults. |
| `backend/platform/config/` | Environment, PostgreSQL, and Redis configuration. |
| `backend/platform/postgres/` | PostgreSQL pool construction. |
| `backend/platform/logging/` | Structured logger setup. |
| `backend/platform/migrate/` | Forward-only transactional PostgreSQL migration runner and tests. |
| `backend/platform/httpx/` | Typed HTTP errors mapped by the central Fiber error handler. |
| `backend/cache/` | Redis Cache-Aside adapter; Redis failures return cache misses and are logged. |
| `backend/middleware/` | HS256 Bearer JWT authentication middleware and authenticated principal context. |
| `backend/internal/identity/` | Identity persistence and first-admin seed use case. |
| `backend/internal/identity/passwords/` | Password policy, bcrypt hashing, verification, and rehash. |
| `backend/internal/identity/auth.go` | Login authentication use case and credential verification. |
| `backend/platform/authn/` | JWT creation/validation, principal, role check, and idle activity tracker. |
| `backend/internal/identity/handler.go` | `POST /auth/login` HTTP validation and response mapping. |
| `backend/internal/identity/users.go` | Admin-only user listing, creation, and update use cases with service-layer authorization. |
| `backend/internal/identity/users_handler.go` | `/users` HTTP mapping. |
| `backend/internal/identity/routes/` | Identity HTTP route declarations, separate from handler logic. |
| `backend/internal/scheduling/` | Operatories, chairs, dentists, appointments, RF-11 internal reminders, and RF-08 schedulable tasks. |
| `backend/internal/scheduling/routes/` | Scheduling and reminder route declarations. |
| `backend/internal/billing/` | Medical catalog, RF-07/08 quotes and approvals, RF-12 invoices/payments, and RN-02 cash close. |
| `backend/internal/billing/routes/` | Billing and invoice route declarations. |
| `backend/cmd/api/main.go` | Monolith composition root and Fiber bootstrap. |
| `backend/cmd/migrate/main.go` | `go run ./cmd/migrate up` migration command. |
| `backend/cmd/seed-admin/main.go` | `make seed-admin` initial SystemAdmin seed command. |
| `backend/migrations/` | Versioned SQL migrations for the monolith. |
| `backend/migrations/008_billing_invoice_lines_payments.sql` | Normalized invoice lines and payment records for RF-12. |
| `backend/migrations/009_cash_register_closures.sql` | Daily blind cash register closure records for RN-02. |
| `backend/migrations/010_appointment_reminders.sql` | Deduplicated 48h/24h appointment reminder queue for RF-11. |
| `backend/migrations/011_quote_approvals_schedulable_tasks.sql` | Partial/total quote approvals and schedulable tasks for RF-08. |
| `frontend/src/app/` | React application router, public/private layouts, and authentication provider. |
| `frontend/src/features/auth/` | Login screen, API schemas, and authentication requests. |
| `frontend/src/features/dashboard/pages/` | Inicio con accesos directos a los módulos de Fase 1. |
| `frontend/src/features/patients/pages/` | Estado inicial de la pantalla de pacientes. |
| `frontend/src/features/scheduling/pages/` | Estado inicial de la pantalla de agenda. |
| `frontend/src/features/billing/pages/` | Estado inicial de la pantalla de presupuestos. |
| `frontend/src/features/navigation/` | Authenticated navigation shell and module links. |
| `frontend/src/shared/ui/` | Presentational module empty state reused by feature pages. |
| `frontend/src/shared/lib/http.ts` | Shared HTTP client with Bearer token injection and API error mapping. |
| `frontend/src/styles.css` | Frontend tokens and responsive styles aligned with the design plan. |
| `docs/decisiones/0002-password-hashing.md` | Owner-approved bcrypt cost and password policy. |
| `deploy/docker-compose.yml` | Development PostgreSQL and Redis containers. |
| `.github/workflows/ci.yml` | GitHub Actions checks, PostgreSQL integration service, vulnerability scan. |
| `rules/ahorro-tokens.md` | Token discipline. Loaded through `opencode.json`. |
| `opencode.json` | Registers `rules/ahorro-tokens.md` as instructions. |
| `.atl/skill-registry.md` | Auto-generated skill index. Do not hand-edit. |

**Not implemented yet:** domain RBAC enforcement for all use cases, clinical and inventory flows,
frontend application scaffold, production deployment configuration, and CI execution. CI workflow
configuration is present; the user requested running it last.

## Document precedence

When two documents disagree, the higher one wins:

1. **SRS** — scope and requirements. The SRS always wins.
2. **`AGENTS.md`** — architecture, stack, and conventions.
3. **`rules/GLOSSARY.md`** — naming.
4. **This map** — where things live. If it disagrees with the filesystem, the filesystem is right
   and this file is stale: fix it before continuing.
5. **`REQUIREMENTS_CHECKLIST.md`** — what to do next. It records intent, not truth.

## Planned layout (read paths only after they exist)

The target follows the owner-approved monolith architecture in `AGENTS.md`.

```text
backend/
  platform/                     # config, logging, errors, tracing, redis, migrations
  cmd/api/                      # one server and composition root
  cmd/migrate/                  # forward-only migration command
  cmd/seed-admin/               # initial admin seed command
  internal/identity/            # identity domain module
  migrations/                   # one versioned SQL directory for the monolith
frontend/
  src/app/  src/features/  src/shared/
deploy/                         # PostgreSQL, Redis, and monolith container
```

## Ports

| Component | Port |
| --- | --- |
| Frontend (Vite) | 5173 |
| Monolith API | 8080 |
| PostgreSQL 16 | 5432 |
| Redis | 6379 |
| MinIO (deferred) | 9000 |

## API contracts

- **Response envelope:** `{"data": ...}` on success, `{"error": {"code", "message", "details"}}`
  on failure.
- **Error codes** are stable `snake_case` strings (`chair_overlap`, `quote_already_approved`).
  Never branch on `message`.
- **Events** are `snake_case` and versioned: `billing.quote.approved.v1`.
- **Identifiers** come from `rules/GLOSSARY.md`. It is `Chair`, never `DentalUnit`.

## Refreshing this map

Regenerate the tree, then update the tables above:

```sh
find . -path ./.git -prune -o -path ./node_modules -prune -o -type f -print | sort
```

Update this file in the same commit that creates a service, a top-level directory, or a new
cross-service contract. A map updated "later" is a map nobody trusts.
