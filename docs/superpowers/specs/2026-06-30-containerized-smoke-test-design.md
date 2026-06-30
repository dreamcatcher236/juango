# Containerized Smoke Test (Approach B)

## Goal

Run the whole juango "create-a-project" path as a single hermetic smoke test that
works identically on a developer machine and in CI:

1. Build the `juango` CLI.
2. Scaffold a fresh project (`juango init`).
3. Build it (Go backend + frontend).
4. Start `serve` and exercise auth/OIDC/health/bootstrap (and a browser login flow).
5. Start the background `worker` and confirm it processes an enqueued task.

The test must not depend on host-installed Node, Chrome, or Redis, and must not be
brittle about toolchain versions (the previous `justfile`/`Makefile` pinned a node
version that did not exist on the machine).

## Why Docker

- **Redis**: the new background-tasks feature needs Redis. CI has none today, so the
  worker path is untestable there. A `redis` service provides it.
- **Toolchain**: a runner image bakes in Go 1.24 + Node 22 + headless Chromium, so
  the npm/nvm PATH friction disappears and runs are reproducible.

## Architecture

`docker-compose.smoke.yml` with two services:

- `redis` — `redis:7-alpine`, healthchecked.
- `runner` — built from `smoke/Dockerfile`, mounts the repo, depends on `redis`
  (healthy), and runs `go test ./integration/... -timeout 15m`. The runner has Go,
  Node, and Chromium. `REDIS_ADDR=redis:6379` is passed via env.

mockoidc stays **embedded** in the test process (`mockoidc.Run()`, the
`github.com/oauth2-proxy/mockoidc` library already in use). Since the test harness,
the scaffolded app, and mockoidc all run inside the single `runner` container, they
communicate over localhost — only Redis needs to be a networked service.

A Go module cache named volume keeps repeat runs fast.

## Test changes

- `setup()` reads `REDIS_ADDR` from env (default `localhost:6379`) and writes the
  scaffolded config with `tasks.enabled: true`. `serve` connects to Redis lazily, so
  this does not break the existing serve-path tests when Redis is absent.
- New `TestTasksWorkerProcessesDummyTask`:
  - Skips if Redis is not reachable (TCP dial), mirroring how browser tests skip when
    Chrome is absent.
  - Starts the scaffolded `worker` subprocess.
  - Enqueues a `dummy:hello` task via juango's `tasks.Client`.
  - Polls the worker's stdout for the completion log within a timeout, then stops it.

## Runner command surface

- `just smoke` — `docker compose -f docker-compose.smoke.yml up --build
  --abort-on-container-exit --exit-code-from runner`, then `down -v`.
- CI gains a `smoke` job that runs `just smoke` (or the compose directly) on PRs to
  `main`, giving the tasks path real coverage.

## Out of scope (follow-ups)

- "Real published path" test (installing a tagged module the way a user would) —
  belongs in a release-gated workflow since it cannot run against unreleased code.
- Pushing the runner image to a registry for faster CI pulls.
