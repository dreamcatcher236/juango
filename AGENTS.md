# AGENTS.md

Guidance for AI coding agents working in the **juango** repo, or helping a user
bootstrap a brand-new project from it.

> Like Django, but made by Juan in Go.

## What juango is

juango is a scaffolding CLI plus a set of reusable Go libraries for building
Go + Vite/React web apps with batteries included: OIDC authentication, cookie
sessions (SQLite-backed), admin mode with audit logging, user impersonation, an
admin API/UI, a health endpoint, and optional Redis-backed background tasks
(Asynq).

Use it when a user wants to start a new full-stack web app and would otherwise
hand-roll auth, sessions, admin tooling, and frontend/backend wiring. It is *not*
a library you bolt onto an existing app; `juango init` generates a fresh project
that then imports the `github.com/juanfont/juango` libraries.

Current CLI version: see `cli/version.go` (`Version`).

## Prerequisites

- **Go** (to build the CLI and generated projects).
- **Node.js 18+ and npm** — required for scaffolding. `juango init` checks the
  Node major version and runs `npm install` in `frontend/`; it aborts if Node is
  missing or older than 18.

## Build / install the CLI

```bash
# From the juango repo
go build -o juango .
# or
just build

# Or install from anywhere
go install github.com/juanfont/juango@latest
```

## Create a new project

```bash
juango init <name> -m <module-path> [-p <port>] [-d "<description>"]
```

- `<name>` (positional, required): lowercase alphanumeric with hyphens only.
- `-m, --module` (**required**): Go module path, e.g. `github.com/myorg/myapp`.
- `-p, --port` (default `8080`): default port baked into config and templates.
- `-d, --description`: project description.

Concrete example:

```bash
juango init acme-portal -m github.com/acme/acme-portal -p 8080 -d "Acme internal portal"
```

`init` generates the project, runs `npm install` in `frontend/`, then `go mod tidy`.

## Greenfield workflow (after init)

```bash
cd acme-portal

# 1. Create the real config from the example
cp config.example.yml config.yml

# 2. Set OIDC details in config.yml:
#    oidc.issuer, oidc.client_id, oidc.client_secret

# 3. Generate the two session keys (each must be EXACTLY 32 bytes).
#    openssl rand -hex 16 emits a 32-char hex string -> 32 bytes.
openssl rand -hex 16   # -> session.authentication_key
openssl rand -hex 16   # -> session.encryption_key

# 4. Run dev (Vite on :5173 + Go server on the configured port, with proxy)
juango dev
# or build + serve for a production-style run:
just build && ./acme-portal serve -c config.yml
```

`juango dev` must be run from inside the project (it requires `go.mod` and
`frontend/package.json`) and refuses to start if `config.yml` is missing.

## Generated project layout

```
<name>/
├── cmd/<name>/             # Entry point + Cobra CLI (serve, worker, make-admin, db, ...)
├── internal/
│   ├── api/                # HTTP app: route registration + your handlers (app.go)
│   ├── database/           # DB wrapper + sql/schema.sql (squibble-managed)
│   ├── tasks/              # Background task definitions + handler registry
│   └── types/              # Config struct + viper loading
├── frontend/               # Vite/React SPA (pages, components, contexts, lib)
├── app.go                  # Wires config -> DB -> API -> server
├── config.example.yml      # Copy to config.yml
└── justfile                # Project-local dev commands
```

## Where to extend

- **API routes** — `internal/api/app.go`, at the marker
  `// Add your application-specific routes here` inside `registerRoutes()`.
  Wrap handlers with `a.sessionMiddleware.RequireAuth(...)` (and optionally
  `RequireAdmin`/`RequireAdminMode`) as the existing routes do.
- **Background tasks** — define the task type + handler (see
  `internal/tasks/dummy.go` as the template), then register it in
  `internal/tasks/registry.go` at `// Add your application-specific task handlers here`.
  Tasks require `tasks.enabled: true` and a reachable Redis (`redis.addr`) in
  `config.yml`, and a running worker: `juango`/binary `worker` (or `just worker`).
  Enqueue from API handlers via `a.taskClient` (see `internal/api/tasks.go`).
- **DB tables** — edit `internal/database/sql/schema.sql` below
  `-- Add your application-specific tables below`. Keep statements idempotent
  (`CREATE TABLE IF NOT EXISTS`). The schema is squibble-managed; see
  `internal/database/README.md` for greenfield vs. migration-rule workflow.
- **Frontend routes** — `frontend/src/App.tsx`, at the marker
  `{/* Add your application-specific routes here */}` inside the `Layout` route.

## Running tests (juango repo)

```bash
just test               # go test ./... -short -v  (unit tests)
just test-integration   # go test ./integration/... -v -timeout 10m
just smoke              # containerized scaffold -> build -> serve -> worker (needs Docker)
```

Integration and smoke tests scaffold a real project, so they need **Node/npm**.
Browser-based tests need **Chrome/Chromium** (they skip if unavailable), and the
smoke / task tests need **Redis** (the smoke test brings up Redis via Docker
Compose). `just dev-test` spins up a scaffolded app against a mock OIDC server
for manual testing.

## Key gotchas

- **First OIDC login becomes admin.** First-user bootstrap is wired
  (`oidcHandlers.WithUserCounter`): the first user to log in via OIDC is
  auto-promoted to admin. Subsequent users are not.
- **Background tasks are opt-in.** `tasks.enabled` defaults to `false`. When
  disabled, the task API handler returns 503 and `worker` exits with an error.
  Enabling tasks requires Redis.
- **Session keys must be exactly 32 bytes.** `validateConfig` rejects any
  `session.authentication_key` or `session.encryption_key` whose length is not 32.
  `openssl rand -hex 16` produces a correct 32-byte value.
- **Schema drift refuses startup.** squibble compares `schema.sql` against the
  live DB; in greenfield, just edit `schema.sql` and recreate the dev DB. Any
  table a library creates at runtime (e.g. `sessions`) must be declared in
  `schema.sql` too.
