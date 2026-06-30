# juango

> Like Django, but made by Juan in Go.

A scaffolding CLI and reusable libraries for building Go + Vite/React web applications with batteries included: OIDC authentication, session management, admin mode, audit logging, and more.

## Disclaimer

This project has been created with [Claude](https://claude.ai) (Anthropic's AI assistant). I had several internal projects that shared the same patterns - OIDC auth, session management, admin mode, etc. - but the code was just copy-pasted between them with no actual shared library. I used Claude to extract the minimum common denominator from these projects and refactor it into this reusable framework. The patterns come from real-world production code, but the extraction, refactoring, and documentation were done through AI pair programming.

## Features

- **CLI scaffolding** - `juango init myapp` creates a complete project structure
- **OIDC authentication** - Plug-and-play with any OIDC provider (Entra ID, Okta, Auth0, etc.)
- **First-user bootstrap** - First OIDC login is automatically promoted to admin
- **Session management** - Secure cookie-based sessions with SQLite backend
- **Admin mode** - Time-limited elevated privileges with audit logging
- **User impersonation** - Debug issues as another user (admin only)
- **Admin API** - User management, stats, and audit log querying with filtering/pagination
- **Audit logging** - Track all sensitive operations with actor/impersonator awareness
- **Health endpoint** - `GET /api/health` with database connectivity check
- **Frontend serving** - Dev proxy to Vite, embedded SPA in production
- **SQLite with WAL** - Simple, fast, single-file database
- **Background tasks** - Asynq integration for Redis-backed job queues
- **React UI library** - Pre-built admin pages, components with shadcn/ui styling
- **Generated CLI commands** - `make-admin`, `list-admins`, `remove-admin`, `db backup`, `db stats`

## Installation

```bash
go install github.com/juanfont/juango@latest
```

## Quick Start

```bash
# Create a new project
juango init myapp
cd myapp

# Configure OIDC (edit config.yml with your provider details)
cp config.example.yml config.yml
vim config.yml

# Start development (runs Vite + Go concurrently)
juango dev

# Visit http://localhost:8080
```

> **Using an AI coding agent?** See [`AGENTS.md`](AGENTS.md) for a concise guide to
> bootstrapping and extending a project with juango. Every generated project also
> ships its own `AGENTS.md` tailored to that app.

## Project Structure

A generated project looks like this:

```
myapp/
├── cmd/myapp/
│   ├── myapp.go              # Entry point
│   └── cli/
│       ├── root.go           # CLI setup
│       ├── serve.go          # Server command
│       ├── make_admin.go     # Promote user to admin
│       ├── list_admins.go    # List admin users
│       ├── remove_admin.go   # Revoke admin from user
│       ├── worker.go         # Background task worker
│       └── db.go             # Database utilities (backup, stats)
├── internal/
│   ├── api/
│   │   ├── app.go            # API routes and handlers
│   │   └── tasks.go          # Task-enqueueing endpoints
│   ├── database/
│   │   ├── db.go             # Database wrapper
│   │   ├── README.md         # Schema & squibble migration guide
│   │   └── sql/
│   │       └── schema.sql    # Your tables
│   ├── tasks/
│   │   ├── registry.go       # Task handler registration
│   │   └── dummy.go          # Example task
│   └── types/
│       └── config.go         # Configuration struct
├── frontend/
│   ├── src/
│   │   ├── App.tsx           # React app
│   │   ├── pages/            # Your pages (including admin/)
│   │   ├── components/       # Your components
│   │   ├── contexts/         # Auth, Admin, Breadcrumb contexts
│   │   └── lib/              # API client, types, utilities
│   ├── src/test/            # Vitest setup
│   ├── package.json
│   └── vite.config.ts
├── app.go                    # Main application
├── justfile                  # Dev commands (dev, build, serve, worker, test)
├── AGENTS.md                 # Guidance for AI coding agents
├── go.mod
├── config.example.yml
└── .gitignore
```

## CLI Commands

### `juango init <project-name>`

Creates a new project with all the scaffolding.

```bash
juango init myapp                           # Uses github.com/user/myapp
juango init myapp -m github.com/org/myapp   # Custom module path
```

### `juango dev`

Starts the development environment:
- Vite dev server with HMR on port 5173
- Go server on port 8080 (proxies to Vite for frontend)

```bash
cd myapp
juango dev
```

### `juango version`

Shows version information.

## Generated Project CLI Commands

Every scaffolded project includes these commands out of the box:

### `myapp serve`

Starts the HTTP server.

```bash
myapp serve -c config.yml
```

### `myapp make-admin <email>`

Promotes a user to admin by email address.

```bash
myapp make-admin -c config.yml user@example.com
```

### `myapp list-admins`

Lists all admin users.

```bash
myapp list-admins -c config.yml
```

### `myapp remove-admin <email>`

Revokes admin privileges from a user. Refuses to remove the last admin.

```bash
myapp remove-admin -c config.yml user@example.com
```

### `myapp db backup <path>`

Creates a database backup using SQLite `VACUUM INTO`.

```bash
myapp db backup -c config.yml /backups/myapp.db
```

### `myapp db stats`

Shows database statistics: user count, admin count, audit log entries, and file size.

```bash
myapp db stats -c config.yml
```

### `myapp worker`

Runs the background task worker (Asynq). Requires `tasks.enabled: true` in config
and a reachable Redis. Register handlers in `internal/tasks/registry.go`.

```bash
myapp worker -c config.yml
```

## Go Library

The `github.com/juanfont/juango` module provides reusable packages:

### `juango/frontend`

SPA serving with automatic dev/prod detection.

```go
import "github.com/juanfont/juango/frontend"

//go:embed frontend/dist
var frontendFS embed.FS

func main() {
    router := mux.NewRouter()
    frontend.Setup(router, frontendFS, "frontend/dist")
}
```

### `juango/auth`

OIDC authentication and session middleware.

```go
import "github.com/juanfont/juango/auth"

// Create OIDC provider
provider, _ := auth.NewOIDCProvider(ctx, auth.OIDCProviderConfig{
    ServerURL: "http://localhost:8080",
    OIDCConfig: types.OIDCConfig{
        Issuer:       "https://login.microsoftonline.com/...",
        ClientID:     "your-client-id",
        ClientSecret: "your-client-secret",
    },
})

// Setup handlers
handlers := auth.NewOIDCHandlers(provider, sessionStore, "session", userStore, auditLogger)
router.HandleFunc("/api/auth/login", handlers.LoginHandler)
router.HandleFunc("/api/auth/logout", handlers.LogoutHandler).Methods("POST")
router.HandleFunc(provider.CallbackPath(), handlers.CallbackHandler)

// Enable first-user bootstrap (first OIDC login gets auto-promoted to admin)
handlers.WithUserCounter(database)

// Protect routes
middleware := auth.NewSessionMiddleware(sessionStore, "session", userStore, auditLogger, 30*time.Minute)
router.HandleFunc("/api/protected", middleware.RequireAuth(myHandler))
router.HandleFunc("/api/admin-only", middleware.RequireAuth(middleware.RequireAdmin(adminHandler)))
```

**Interfaces**: Your database layer implements these to integrate with the auth system:

- `auth.UserStore` - `CreateOrUpdateUserFromClaim`, `GetUserByID`
- `auth.AuditLogger` - `CreateAuditLog`
- `auth.AdminUserStore` - Extends UserStore with `GetAllUsers`, `GetUserByEmail`, `UpdateUser`, `CountUsers`, `CountAdminUsers`, `CountActiveUsers`, `CountNewUsers`
- `auth.AdminAuditStore` - Extends AuditLogger with `GetAuditLogs`, `CountAuditLogs`, `GetDistinctActions`, `GetDistinctResourceTypes`
- `auth.UserCounter` - `CountUsers` (used for first-user bootstrap)

### `juango/admin`

Admin mode, impersonation, user management API, audit log API, and health check.

```go
import "github.com/juanfont/juango/admin"

// Admin mode and impersonation
handlers := admin.NewHandlers(sessionStore, "session", userStore, auditLogger, 30*time.Minute)

router.HandleFunc("/api/admin/mode/enable", middleware.RequireAdmin(handlers.AdminModeEnableHandler)).Methods("POST")
router.HandleFunc("/api/admin/mode/disable", middleware.RequireAdmin(handlers.AdminModeDisableHandler)).Methods("POST")
router.HandleFunc("/api/admin/mode/status", handlers.AdminModeStatusHandler).Methods("GET")

router.HandleFunc("/api/admin/impersonate/start", middleware.RequireAdminMode(handlers.ImpersonationStartHandler)).Methods("POST")
router.HandleFunc("/api/admin/impersonate/stop", handlers.ImpersonationStopHandler).Methods("POST")

// Admin API (user management + audit logs)
adminAPI := admin.NewAdminAPIHandlers(adminUserStore, adminAuditStore)

router.HandleFunc("/api/admin/users/stats", middleware.RequireAdminMode(adminAPI.GetUserStatsHandler)).Methods("GET")
router.HandleFunc("/api/admin/users", middleware.RequireAdminMode(adminAPI.ListUsersHandler)).Methods("GET")
router.HandleFunc("/api/admin/users/{id}", middleware.RequireAdminMode(adminAPI.GetUserHandler)).Methods("GET")
router.HandleFunc("/api/admin/users/{id}", middleware.RequireAdminMode(adminAPI.UpdateUserHandler)).Methods("PATCH")

router.HandleFunc("/api/admin/audit-logs/stats", middleware.RequireAdminMode(adminAPI.GetAuditLogStatsHandler)).Methods("GET")
router.HandleFunc("/api/admin/audit-logs", middleware.RequireAdminMode(adminAPI.ListAuditLogsHandler)).Methods("GET")

// Health check (no auth required)
healthHandler := admin.NewHealthHandler(database) // database implements admin.DBPinger
router.HandleFunc("/api/health", healthHandler.ServeHTTP).Methods("GET")
```

### `juango/middleware`

Common HTTP middleware.

```go
import "github.com/juanfont/juango/middleware"

router.Use(middleware.Logging(logger))    // Request logging with zerolog
router.Use(middleware.Metrics())          // Prometheus metrics
router.Use(middleware.Recovery())         // Panic recovery
router.Use(middleware.CORS(nil))          // CORS (nil = permissive defaults)
```

### `juango/database`

SQLite helpers with WAL mode and migrations.

```go
import "github.com/juanfont/juango/database"

//go:embed sql/schema.sql
var schema string

db, _ := database.New("app.db", schema)
defer db.Close()

// Transactions
db.WithTx(ctx, func(tx *sqlx.Tx) error {
    // ...
    return nil
})
```

### `juango/tasks`

Asynq task queue wrappers.

```go
import "github.com/juanfont/juango/tasks"

// Client (enqueue tasks): NewClient(redisAddr, redisPassword, redisDB)
client := tasks.NewClient("localhost:6379", "", 0)
defer client.Close()
client.Enqueue("email:send", payload) // returns (*asynq.TaskInfo, error)

// Server (process tasks)
cfg := tasks.DefaultServerConfig("localhost:6379", "", 0)
server := tasks.NewServer(cfg)
server.HandleFunc("email:send", handleSendEmail)
server.Run() // blocks until shutdown
```

### `juango/types`

Common types used across packages.

```go
import "github.com/juanfont/juango/types"

// Core types
// types.User, types.OIDCClaims, types.OIDCConfig
// types.AdminModeState, types.ImpersonationState

// Audit logging
// types.AuditLog, types.AuditLogFilter

// Admin API
// types.UserStats, types.AuditLogStats
// types.UpdateUserRequest, types.HealthResponse
```

## Frontend Components

Scaffolded projects include a complete React frontend with:

- **Contexts**: `AuthContext`, `AdminContext`, `BreadcrumbContext`
- **Components**: `ProtectedRoute`, `Layout`, `AdminGuard`, `AdminModeToggle`, `ImpersonationBanner`, `AuditLogDetailModal`
- **Admin pages**: `admin/Users` (user management with sorting/pagination), `admin/AuditLogs` (filterable logs with CSV export)
- **UI components**: Button, Card, Dialog, Table, Badge, Select, Input, Tooltip, etc. (shadcn/ui)
- **Utilities**: `ApiClient` for backend communication, `cn()` for class names

### Example App.tsx

```tsx
import { Routes, Route } from 'react-router-dom'
import { AuthProvider } from './contexts/AuthContext'
import { AdminProvider } from './contexts/AdminContext'
import { ProtectedRoute } from './components/ProtectedRoute'
import Layout from './components/Layout'

export default function App() {
  return (
    <AuthProvider>
      <AdminProvider>
        <Routes>
          <Route path="/login" element={<Login />} />
          <Route path="/" element={<ProtectedRoute><Layout /></ProtectedRoute>}>
            <Route index element={<Home />} />
          </Route>
        </Routes>
      </AdminProvider>
    </AuthProvider>
  )
}
```

## Configuration

Example `config.yml`:

```yaml
listen_addr: ":8080"
advertise_url: "http://localhost:8080"
admin_mode_timeout: 30m

database:
  path: "myapp.db"

session:
  cookie_name: "myapp_session"
  cookie_expiry: 24h
  authentication_key: "32-byte-key-for-authentication!!"  # exactly 32 bytes
  encryption_key: "32-byte-key-for-encryption-here"      # exactly 32 bytes

oidc:
  issuer: "https://login.microsoftonline.com/{tenant}/v2.0"
  client_id: "your-client-id"
  client_secret: "your-client-secret"
  scopes:
    - openid
    - profile
    - email

# Background tasks (opt-in). When enabled, the app connects to Redis and the
# `worker` command processes queued jobs.
tasks:
  enabled: false

redis:
  addr: "localhost:6379"

logging:
  level: info
  format: text
```

Generate secure keys:

```bash
openssl rand -hex 16  # generates 32-character hex string
```

## API Endpoints

Every scaffolded project comes with these endpoints pre-wired:

| Endpoint | Method | Auth | Description |
|---|---|---|---|
| `/api/health` | GET | None | Health check with database status |
| `/api/auth/login` | GET | None | Initiates OIDC login flow |
| `/api/auth/logout` | POST | None | Destroys session |
| `/api/auth/session` | GET | None | Returns current session info |
| `/api/admin/mode/status` | GET | Auth | Admin mode status |
| `/api/admin/mode/enable` | POST | Admin | Enable admin mode (with reason) |
| `/api/admin/mode/disable` | POST | Admin | Disable admin mode |
| `/api/admin/impersonate/start` | POST | Admin Mode | Start impersonating a user |
| `/api/admin/impersonate/stop` | POST | Auth | Stop impersonation |
| `/api/admin/impersonate/status` | GET | Auth | Impersonation status |
| `/api/admin/users` | GET | Admin Mode | List users (with search) |
| `/api/admin/users/stats` | GET | Admin Mode | User statistics |
| `/api/admin/users/{id}` | GET | Admin Mode | Get user details |
| `/api/admin/users/{id}` | PATCH | Admin Mode | Update user (admin flag, active status) |
| `/api/admin/audit-logs` | GET | Admin Mode | List audit logs (with filtering/pagination) |
| `/api/admin/audit-logs/stats` | GET | Admin Mode | Audit log statistics |
| `/api/admin/tasks/dummy` | POST | Admin Mode | Enqueue the example background task |

## Testing

The repo uses [`just`](https://github.com/casey/just) for common tasks:

```bash
just test              # unit tests (go test -short)
just test-integration  # full integration suite (needs Node; Chrome/Redis optional)
just smoke             # containerized end-to-end smoke test (Docker)
```

Under the hood:

```bash
# Unit tests only (skips integration)
go test ./... -short

# Integration suite (scaffolds a project, builds it, runs serve + worker).
# Browser tests skip without Chrome; the worker test skips without Redis
# (set REDIS_ADDR to point at one).
go test ./integration/... -v

# Fully hermetic run in Docker (Go + Node + headless Chromium + Redis)
just smoke
```

The scaffolded frontend ships with [Vitest](https://vitest.dev) tests; run them in
a generated project with `just test` (or `npm test` in `frontend/`).

### Test Coverage

- **API Tests**: Session management, OIDC flow (verified authenticated), protected endpoints, logout, health endpoint, first-user bootstrap
- **Browser Tests**: Real headless-Chrome login flow, asserting an authenticated session
- **Scaffolding Tests**: Project structure, module path handling, frontend deps/imports, full build, frontend Vitest run, mock OIDC end-to-end
- **Background Tasks**: Worker processes an enqueued task (against Redis)
- **Smoke (CI + local)**: The whole path in one container — scaffold → build → serve → worker

## Building for Production

```bash
# Build frontend
cd frontend && npm run build && cd ..

# Build Go binary (embeds frontend)
go build -o myapp ./cmd/myapp

# Run
./myapp serve -c config.yml
```

Or use GoReleaser:

```bash
goreleaser release --snapshot --clean
```

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        Browser                               │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                      Go HTTP Server                          │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │  Middleware │  │    Auth     │  │      Frontend       │  │
│  │  - Logging  │  │  - OIDC     │  │  - Dev: Vite proxy  │  │
│  │  - Metrics  │  │  - Session  │  │  - Prod: Embedded   │  │
│  │  - Recovery │  │  - Admin    │  │                     │  │
│  └─────────────┘  └─────────────┘  └─────────────────────┘  │
│                              │                               │
│                              ▼                               │
│  ┌─────────────────────────────────────────────────────────┐│
│  │                     Your API Handlers                    ││
│  └─────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────┘
                              │
              ┌───────────────┼───────────────┐
              ▼               ▼               ▼
        ┌──────────┐   ┌──────────┐   ┌──────────┐
        │  SQLite  │   │  Redis   │   │   OIDC   │
        │   (WAL)  │   │ (tasks)  │   │ Provider │
        └──────────┘   └──────────┘   └──────────┘
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see [LICENSE](LICENSE) for details.

---

Built with Claude and caffeine.
