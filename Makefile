.PHONY: build test dev-test dev-test-clean

BUILD_DIR := /tmp/juango-dev-test
PROJECT_NAME := testapp
MODULE_PATH := github.com/test/testapp
APP_PORT := 18080
MOCKOIDC_PORT := 19080
MOCKOIDC_PIDFILE := $(BUILD_DIR)/mockoidc.pid

# Use nvm node if available
NVM_NODE := $(HOME)/.nvm/versions/node/v22.20.0/bin
PATH := $(NVM_NODE):$(PATH)
export PATH

build:
	go build -o juango .

test:
	go test ./... -short -v

test-integration:
	go test ./integration/... -v -timeout 10m

# Full dev-test workflow (no Docker)
dev-test: build dev-test-clean
	@echo "=== Setting up dev-test environment ==="
	@mkdir -p $(BUILD_DIR)
	@echo "=== Scaffolding test project ==="
	./juango init $(PROJECT_NAME) -m $(MODULE_PATH) -p $(APP_PORT)
	@mv $(PROJECT_NAME) $(BUILD_DIR)/
	@echo "=== Patching config with mock OIDC settings ==="
	cd $(BUILD_DIR)/$(PROJECT_NAME) && sed \
		-e 's|issuer:.*|issuer: "http://127.0.0.1:$(MOCKOIDC_PORT)/oidc"|' \
		-e 's|client_id:.*|client_id: "dev-client"|' \
		-e 's|client_secret:.*|client_secret: "dev-secret"|' \
		-e 's|authentication_key:.*|authentication_key: "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4"|' \
		-e 's|encryption_key:.*|encryption_key: "d4c3b2a1f6e5d4c3b2a1f6e5d4c3b2a1"|' \
		config.example.yml > config.yml
	@echo "=== Starting mock OIDC server ==="
	MOCKOIDC_ADDR=127.0.0.1 MOCKOIDC_PORT=$(MOCKOIDC_PORT) \
	MOCKOIDC_CLIENT_ID=dev-client MOCKOIDC_CLIENT_SECRET=dev-secret \
	MOCKOIDC_USERS='[{"email":"dev@test.local","preferred_username":"dev","name":"Dev User"}]' \
	./juango mockoidc & echo $$! > $(MOCKOIDC_PIDFILE)
	sleep 1
	@echo ""
	@echo "=== Dev-test environment ready ==="
	@echo "App:       http://localhost:$(APP_PORT)"
	@echo "Mock OIDC: http://127.0.0.1:$(MOCKOIDC_PORT)"
	@echo ""
	cd $(BUILD_DIR)/$(PROJECT_NAME) && $(CURDIR)/juango dev

dev-test-clean:
	-@if [ -f $(MOCKOIDC_PIDFILE) ]; then kill $$(cat $(MOCKOIDC_PIDFILE)) 2>/dev/null; rm -f $(MOCKOIDC_PIDFILE); fi
	rm -rf $(BUILD_DIR)
