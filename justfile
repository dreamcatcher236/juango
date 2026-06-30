build_dir := "/tmp/juango-dev-test"
project_name := "testapp"
module_path := "github.com/test/testapp"
app_port := "18080"
mockoidc_port := "19080"
mockoidc_pidfile := build_dir / "mockoidc.pid"

# Use nvm node if available
export PATH := env_var('HOME') + "/.nvm/versions/node/v22.20.0/bin:" + env_var('PATH')

default:
    @just --list

build:
    go build -o juango .

test:
    go test ./... -short -v

test-integration:
    go test ./integration/... -v -timeout 10m

# Full containerized smoke test: scaffold -> build -> serve -> worker, with Redis
smoke:
    #!/usr/bin/env bash
    set -uo pipefail
    docker compose -f docker-compose.smoke.yml up --build --exit-code-from runner
    code=$?
    docker compose -f docker-compose.smoke.yml down -v
    exit $code

# Tear down the smoke-test stack (also runs automatically after `just smoke`)
smoke-clean:
    docker compose -f docker-compose.smoke.yml down -v

# Full dev-test workflow (no Docker)
dev-test: build dev-test-clean
    #!/usr/bin/env bash
    set -euo pipefail
    echo "=== Setting up dev-test environment ==="
    mkdir -p {{build_dir}}
    echo "=== Scaffolding test project ==="
    ./juango init {{project_name}} -m {{module_path}} -p {{app_port}}
    mv {{project_name}} {{build_dir}}/
    echo "=== Patching config with mock OIDC settings ==="
    cd {{build_dir}}/{{project_name}} && sed \
        -e 's|issuer:.*|issuer: "http://127.0.0.1:{{mockoidc_port}}/oidc"|' \
        -e 's|client_id:.*|client_id: "dev-client"|' \
        -e 's|client_secret:.*|client_secret: "dev-secret"|' \
        -e 's|authentication_key:.*|authentication_key: "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4"|' \
        -e 's|encryption_key:.*|encryption_key: "d4c3b2a1f6e5d4c3b2a1f6e5d4c3b2a1"|' \
        config.example.yml > config.yml
    cd {{justfile_directory()}}
    echo "=== Starting mock OIDC server ==="
    MOCKOIDC_ADDR=127.0.0.1 MOCKOIDC_PORT={{mockoidc_port}} \
    MOCKOIDC_CLIENT_ID=dev-client MOCKOIDC_CLIENT_SECRET=dev-secret \
    MOCKOIDC_USERS='[{"email":"dev@test.local","preferred_username":"dev","name":"Dev User"}]' \
    ./juango mockoidc & echo $! > {{mockoidc_pidfile}}
    sleep 1
    echo ""
    echo "=== Dev-test environment ready ==="
    echo "App:       http://localhost:{{app_port}}"
    echo "Mock OIDC: http://127.0.0.1:{{mockoidc_port}}"
    echo ""
    cd {{build_dir}}/{{project_name}} && {{justfile_directory()}}/juango dev

dev-test-clean:
    #!/usr/bin/env bash
    if [ -f {{mockoidc_pidfile}} ]; then kill $(cat {{mockoidc_pidfile}}) 2>/dev/null || true; rm -f {{mockoidc_pidfile}}; fi
    rm -rf {{build_dir}}
