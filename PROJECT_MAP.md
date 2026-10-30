# PROJECT MAP — SIGO-Dental

> Architectural map. Read this **before** exploring code.
>
> **Rule: this file describes what EXISTS.** Every path listed here must be readable today.
> Planned paths are listed only in the clearly marked section below, because an agent that opens
> a nonexistent file wastes a turn and then guesses the design from the filename.

## Current state (2026-10-29)

**Backend foundation and identity schema exist.** The repository has Go configuration,
logging, PostgreSQL pool setup, a Fiber API bootstrap, a forward-only migration runner, the
identity roles/users schema, and an idempotent first-admin seed. Login, sessions, MFA, RBAC,
and audit behavior are still pending.

| Path | Content |
| --- | --- |
| `docs/SRS_SIGO-Dental.md` | Requirements and scope authority. Spanish. |
| `docs/specs/01-fundacion-identidad-auditoria.md` | Ordered backlog for foundation, identity, MFA, RBAC, and audit work. |
| `REQUIREMENTS_CHECKLIST.md` | Progress, task IDs, and order. Backend before frontend. |
| `AGENTS.md` | Stack, architecture, conventions, Go/React practices. |
| `rules/GLOSSARY.md` | SRS-Spanish → English identifier map. |
| `Makefile` | `lint`, `format`, `test`, and `build` commands. |
| `.editorconfig` | Shared text and Go formatting defaults. |
| `backend/platform/config/` | Environment, PostgreSQL, and Redis configuration. |
| `backend/platform/postgres/` | PostgreSQL pool construction. |
| `backend/platform/logging/` | Structured logger setup. |
| `backend/platform/migrate/` | Forward-only transactional PostgreSQL migration runner and tests. |
| `backend/internal/identity/` | Identity persistence and first-admin seed use case. |
| `backend/internal/identity/passwords/` | Password policy, bcrypt hashing, verification, and rehash. |
| `backend/cmd/api/main.go` | Monolith composition root and Fiber bootstrap. |
| `backend/cmd/migrate/main.go` | `go run ./cmd/migrate up` migration command. |
| `backend/cmd/seed-admin/main.go` | `make seed-admin` initial SystemAdmin seed command. |
| `backend/migrations/` | Versioned SQL migrations for the monolith. |
| `docs/decisiones/0002-password-hashing.md` | Owner-approved bcrypt cost and password policy. |
| `deploy/docker-compose.yml` | Development PostgreSQL and Redis containers. |
| `.github/workflows/ci.yml` | GitHub Actions checks, PostgreSQL integration service, vulnerability scan. |
| `rules/ahorro-tokens.md` | Token discipline. Loaded through `opencode.json`. |
| `opencode.json` | Registers `rules/ahorro-tokens.md` as instructions. |
| `.atl/skill-registry.md` | Auto-generated skill index. Do not hand-edit. |

**Not implemented yet:** identity HTTP flows, sessions, MFA, RBAC, audit service, other domain
modules, frontend scaffold, production deployment configuration, and CI execution. CI workflow
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
