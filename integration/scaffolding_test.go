package integration_test

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

// TestScaffoldingRequiresModuleFlag tests that juango init requires -m flag
func TestScaffoldingRequiresModuleFlag(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tmpDir, err := os.MkdirTemp("", "juango-scaffold-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Build juango CLI
	juangoBin := filepath.Join(tmpDir, "juango")
	buildCmd := exec.Command("go", "build", "-o", juangoBin, ".")
	buildCmd.Dir = projectRoot()
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build juango: %v\n%s", err, output)
	}

	// Try init without -m flag - should fail
	initCmd := exec.Command(juangoBin, "init", "myapp")
	initCmd.Dir = tmpDir
	output, err := initCmd.CombinedOutput()

	if err == nil {
		t.Fatal("Expected init to fail without -m flag, but it succeeded")
	}

	if !strings.Contains(string(output), "required flag") || !strings.Contains(string(output), "module") {
		t.Errorf("Expected error about required module flag, got: %s", output)
	}
}

// TestScaffoldingRequiresNpm tests that juango init checks for npm
func TestScaffoldingRequiresNpm(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tmpDir, err := os.MkdirTemp("", "juango-scaffold-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Build juango CLI
	juangoBin := filepath.Join(tmpDir, "juango")
	buildCmd := exec.Command("go", "build", "-o", juangoBin, ".")
	buildCmd.Dir = projectRoot()
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build juango: %v\n%s", err, output)
	}

	// Run init with empty PATH (no npm available)
	initCmd := exec.Command(juangoBin, "init", "myapp", "-m", "github.com/test/myapp")
	initCmd.Dir = tmpDir
	initCmd.Env = []string{"PATH="} // Empty PATH

	output, err := initCmd.CombinedOutput()

	if err == nil {
		t.Fatal("Expected init to fail without npm in PATH, but it succeeded")
	}

	if !strings.Contains(string(output), "npm not found") {
		t.Errorf("Expected error about npm not found, got: %s", output)
	}
}

// TestScaffoldingModulePath tests that the provided module path is used correctly
func TestScaffoldingModulePath(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Skip if npm not available
	if _, err := exec.LookPath("npm"); err != nil {
		t.Skip("npm not available, skipping test")
	}

	tests := []struct {
		name           string
		projectName    string
		moduleFlag     string
		expectedModule string
	}{
		{
			name:           "github module path",
			projectName:    "myapp",
			moduleFlag:     "github.com/myorg/myapp",
			expectedModule: "github.com/myorg/myapp",
		},
		{
			name:           "gitlab module path",
			projectName:    "myapp",
			moduleFlag:     "gitlab.com/myuser/myapp",
			expectedModule: "gitlab.com/myuser/myapp",
		},
		{
			name:           "custom domain",
			projectName:    "myapp",
			moduleFlag:     "git.example.com/team/myapp",
			expectedModule: "git.example.com/team/myapp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, err := os.MkdirTemp("", "juango-scaffold-test-*")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tmpDir)

			// Build juango CLI
			juangoBin := filepath.Join(tmpDir, "juango")
			buildCmd := exec.Command("go", "build", "-o", juangoBin, ".")
			buildCmd.Dir = projectRoot()
			if output, err := buildCmd.CombinedOutput(); err != nil {
				t.Fatalf("Failed to build juango: %v\n%s", err, output)
			}

			// Run init with module flag
			initCmd := exec.Command(juangoBin, "init", tt.projectName, "-m", tt.moduleFlag)
			initCmd.Dir = tmpDir
			if output, err := initCmd.CombinedOutput(); err != nil {
				t.Fatalf("juango init failed: %v\n%s", err, output)
			}

			// Check the generated go.mod
			goModPath := filepath.Join(tmpDir, tt.projectName, "go.mod")
			goModContent, err := os.ReadFile(goModPath)
			if err != nil {
				t.Fatalf("Failed to read go.mod: %v", err)
			}

			// Extract module path from go.mod
			moduleRegex := regexp.MustCompile(`module\s+(\S+)`)
			matches := moduleRegex.FindStringSubmatch(string(goModContent))
			if len(matches) < 2 {
				t.Fatalf("Could not find module path in go.mod:\n%s", goModContent)
			}

			actualModule := matches[1]
			if actualModule != tt.expectedModule {
				t.Errorf("Module path mismatch:\n  expected: %s\n  actual:   %s", tt.expectedModule, actualModule)
			}
		})
	}
}

// TestScaffoldingFrontendDependencies tests that the frontend has no invalid npm dependencies
func TestScaffoldingFrontendDependencies(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Skip if npm not available
	if _, err := exec.LookPath("npm"); err != nil {
		t.Skip("npm not available, skipping test")
	}

	tmpDir, err := os.MkdirTemp("", "juango-scaffold-frontend-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Build and run juango init
	juangoBin := filepath.Join(tmpDir, "juango")
	buildCmd := exec.Command("go", "build", "-o", juangoBin, ".")
	buildCmd.Dir = projectRoot()
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build juango: %v\n%s", err, output)
	}

	initCmd := exec.Command(juangoBin, "init", "testapp", "-m", "github.com/test/testapp")
	initCmd.Dir = tmpDir
	if output, err := initCmd.CombinedOutput(); err != nil {
		t.Fatalf("juango init failed: %v\n%s", err, output)
	}

	// Read and validate package.json
	packageJSONPath := filepath.Join(tmpDir, "testapp", "frontend", "package.json")
	packageJSONContent, err := os.ReadFile(packageJSONPath)
	if err != nil {
		t.Fatalf("Failed to read package.json: %v", err)
	}

	var packageJSON map[string]interface{}
	if err := json.Unmarshal(packageJSONContent, &packageJSON); err != nil {
		t.Fatalf("Failed to parse package.json: %v", err)
	}

	// Check dependencies
	deps, ok := packageJSON["dependencies"].(map[string]interface{})
	if !ok {
		t.Fatal("Could not find dependencies in package.json")
	}

	// List of dependencies that should NOT exist (unpublished packages)
	forbiddenDeps := []string{
		"@juango/ui",
	}

	for _, forbidden := range forbiddenDeps {
		if _, exists := deps[forbidden]; exists {
			t.Errorf("package.json contains unpublished dependency: %s", forbidden)
		}
	}

	// List of dependencies that SHOULD exist
	requiredDeps := []string{
		"react",
		"react-dom",
		"react-router-dom",
		"@radix-ui/react-slot",
	}

	for _, required := range requiredDeps {
		if _, exists := deps[required]; !exists {
			t.Errorf("package.json missing required dependency: %s", required)
		}
	}
}

// TestScaffoldingFrontendImports tests that frontend files use valid local imports
func TestScaffoldingFrontendImports(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Skip if npm not available
	if _, err := exec.LookPath("npm"); err != nil {
		t.Skip("npm not available, skipping test")
	}

	tmpDir, err := os.MkdirTemp("", "juango-scaffold-imports-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Build and run juango init
	juangoBin := filepath.Join(tmpDir, "juango")
	buildCmd := exec.Command("go", "build", "-o", juangoBin, ".")
	buildCmd.Dir = projectRoot()
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build juango: %v\n%s", err, output)
	}

	initCmd := exec.Command(juangoBin, "init", "testapp", "-m", "github.com/test/testapp")
	initCmd.Dir = tmpDir
	if output, err := initCmd.CombinedOutput(); err != nil {
		t.Fatalf("juango init failed: %v\n%s", err, output)
	}

	// Files to check for @juango/ui imports
	filesToCheck := []string{
		"frontend/src/App.tsx",
		"frontend/src/pages/Home.tsx",
		"frontend/src/pages/Login.tsx",
		"frontend/src/components/Layout.tsx",
	}

	for _, file := range filesToCheck {
		filePath := filepath.Join(tmpDir, "testapp", file)
		content, err := os.ReadFile(filePath)
		if err != nil {
			t.Errorf("Failed to read %s: %v", file, err)
			continue
		}

		if strings.Contains(string(content), "@juango/ui") {
			t.Errorf("%s contains import from @juango/ui (unpublished package)", file)
		}
	}

	// Verify required files exist
	requiredFiles := []string{
		"frontend/src/contexts/AuthContext.tsx",
		"frontend/src/contexts/AdminContext.tsx",
		"frontend/src/contexts/BreadcrumbContext.tsx",
		"frontend/src/components/ProtectedRoute.tsx",
		"frontend/src/components/ui/button.tsx",
		"frontend/src/components/ui/card.tsx",
		"frontend/src/lib/api.ts",
		"frontend/src/lib/types.ts",
		"frontend/src/lib/utils.ts",
	}

	for _, file := range requiredFiles {
		filePath := filepath.Join(tmpDir, "testapp", file)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Errorf("Required file missing: %s", file)
		}
	}
}

// TestScaffoldingProjectStructure validates the overall project structure
func TestScaffoldingProjectStructure(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Skip if npm not available
	if _, err := exec.LookPath("npm"); err != nil {
		t.Skip("npm not available, skipping test")
	}

	tmpDir, err := os.MkdirTemp("", "juango-scaffold-structure-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Build and run juango init
	juangoBin := filepath.Join(tmpDir, "juango")
	buildCmd := exec.Command("go", "build", "-o", juangoBin, ".")
	buildCmd.Dir = projectRoot()
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build juango: %v\n%s", err, output)
	}

	initCmd := exec.Command(juangoBin, "init", "myproject", "-m", "github.com/test/myproject")
	initCmd.Dir = tmpDir
	if output, err := initCmd.CombinedOutput(); err != nil {
		t.Fatalf("juango init failed: %v\n%s", err, output)
	}

	projectDir := filepath.Join(tmpDir, "myproject")

	// Check required directories
	requiredDirs := []string{
		"cmd/myproject",
		"cmd/myproject/cli",
		"internal/api",
		"internal/database",
		"internal/database/sql",
		"internal/tasks",
		"internal/types",
		"frontend/src",
		"frontend/src/components",
		"frontend/src/pages",
		"frontend/src/lib",
		"scripts",
	}

	for _, dir := range requiredDirs {
		dirPath := filepath.Join(projectDir, dir)
		if info, err := os.Stat(dirPath); os.IsNotExist(err) || !info.IsDir() {
			t.Errorf("Required directory missing: %s", dir)
		}
	}

	// Check required files
	requiredFiles := []string{
		"go.mod",
		"app.go",
		"config.example.yml",
		".gitignore",
		".goreleaser.yml",
		"justfile",
		"cmd/myproject/myproject.go",
		"cmd/myproject/cli/root.go",
		"cmd/myproject/cli/serve.go",
		"cmd/myproject/cli/make_admin.go",
		"cmd/myproject/cli/list_admins.go",
		"cmd/myproject/cli/remove_admin.go",
		"cmd/myproject/cli/db.go",
		"cmd/myproject/cli/worker.go",
		"internal/api/app.go",
		"internal/api/tasks.go",
		"internal/database/db.go",
		"internal/database/sql/schema.sql",
		"internal/tasks/registry.go",
		"internal/tasks/dummy.go",
		"internal/types/config.go",
		"frontend/package.json",
		"frontend/vite.config.ts",
		"frontend/tsconfig.json",
		"frontend/index.html",
		"frontend/src/App.tsx",
		"frontend/src/main.tsx",
		"scripts/build_frontend.sh",
	}

	for _, file := range requiredFiles {
		filePath := filepath.Join(projectDir, file)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Errorf("Required file missing: %s", file)
		}
	}
}

// TestScaffoldingFullBuild tests that scaffolded project can build completely
func TestScaffoldingFullBuild(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Skip if npm not available
	if _, err := exec.LookPath("npm"); err != nil {
		t.Skip("npm not available, skipping test")
	}

	tmpDir, err := os.MkdirTemp("", "juango-scaffold-build-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Build juango CLI
	juangoBin := filepath.Join(tmpDir, "juango")
	buildCmd := exec.Command("go", "build", "-o", juangoBin, ".")
	buildCmd.Dir = projectRoot()
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build juango: %v\n%s", err, output)
	}

	// Create scaffolded project
	initCmd := exec.Command(juangoBin, "init", "testapp", "-m", "github.com/test/testapp")
	initCmd.Dir = tmpDir
	if output, err := initCmd.CombinedOutput(); err != nil {
		t.Fatalf("juango init failed: %v\n%s", err, output)
	}

	// Add replace directive for local juango
	projectDir := filepath.Join(tmpDir, "testapp")
	goModPath := filepath.Join(projectDir, "go.mod")
	goModContent, err := os.ReadFile(goModPath)
	if err != nil {
		t.Fatalf("Failed to read go.mod: %v", err)
	}
	replaceDirective := "\nreplace github.com/juanfont/juango => " + projectRoot() + "\n"
	if err := os.WriteFile(goModPath, append(goModContent, []byte(replaceDirective)...), 0644); err != nil {
		t.Fatalf("Failed to update go.mod: %v", err)
	}

	// Run go mod tidy
	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = projectDir
	if output, err := tidyCmd.CombinedOutput(); err != nil {
		t.Fatalf("go mod tidy failed: %v\n%s", err, output)
	}

	// Build frontend
	npmBuildCmd := exec.Command("npm", "run", "build")
	npmBuildCmd.Dir = filepath.Join(tmpDir, "testapp", "frontend")
	if output, err := npmBuildCmd.CombinedOutput(); err != nil {
		t.Fatalf("npm run build failed: %v\n%s", err, output)
	}

	// Verify dist exists
	distPath := filepath.Join(tmpDir, "testapp", "frontend", "dist", "index.html")
	if _, err := os.Stat(distPath); os.IsNotExist(err) {
		t.Error("dist/index.html not created after build")
	}

	// Build Go backend (needs frontend dist)
	goBuildCmd := exec.Command("go", "build", "-o", "testapp", "./cmd/testapp")
	goBuildCmd.Dir = projectDir
	if output, err := goBuildCmd.CombinedOutput(); err != nil {
		t.Fatalf("go build failed: %v\n%s", err, output)
	}

	// Verify binary exists
	binaryPath := filepath.Join(tmpDir, "testapp", "testapp")
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		t.Error("testapp binary not created after build")
	}
}

// TestScaffoldingWithMockOIDC tests that scaffolded project runs with mock OIDC
func TestScaffoldingWithMockOIDC(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Skip if npm not available
	if _, err := exec.LookPath("npm"); err != nil {
		t.Skip("npm not available, skipping test")
	}

	tmpDir, err := os.MkdirTemp("", "juango-scaffold-oidc-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	projectDir := filepath.Join(tmpDir, "testapp")

	// Build juango CLI
	juangoBin := filepath.Join(tmpDir, "juango")
	buildCmd := exec.Command("go", "build", "-o", juangoBin, ".")
	buildCmd.Dir = projectRoot()
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build juango: %v\n%s", err, output)
	}

	// Create scaffolded project
	initCmd := exec.Command(juangoBin, "init", "testapp", "-m", "github.com/test/testapp")
	initCmd.Dir = tmpDir
	if output, err := initCmd.CombinedOutput(); err != nil {
		t.Fatalf("juango init failed: %v\n%s", err, output)
	}

	// Build frontend
	npmBuildCmd := exec.Command("npm", "run", "build")
	npmBuildCmd.Dir = filepath.Join(projectDir, "frontend")
	if output, err := npmBuildCmd.CombinedOutput(); err != nil {
		t.Fatalf("npm run build failed: %v\n%s", err, output)
	}

	// Add replace directive for local juango
	goModPath := filepath.Join(projectDir, "go.mod")
	goModContent, err := os.ReadFile(goModPath)
	if err != nil {
		t.Fatalf("Failed to read go.mod: %v", err)
	}
	replaceDirective := "\nreplace github.com/juanfont/juango => " + projectRoot() + "\n"
	if err := os.WriteFile(goModPath, append(goModContent, []byte(replaceDirective)...), 0644); err != nil {
		t.Fatalf("Failed to update go.mod: %v", err)
	}

	// Run go mod tidy
	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = projectDir
	if output, err := tidyCmd.CombinedOutput(); err != nil {
		t.Fatalf("go mod tidy failed: %v\n%s", err, output)
	}

	// Build the app
	appBuildCmd := exec.Command("go", "build", "-o", "testapp", "./cmd/testapp")
	appBuildCmd.Dir = projectDir
	if output, err := appBuildCmd.CombinedOutput(); err != nil {
		t.Fatalf("go build failed: %v\n%s", err, output)
	}

	// Start mock OIDC server using juango mockoidc command
	mockOIDCPort := "19080"
	mockOIDCUsers := `[{"subject":"test-user","email":"test@example.com","preferred_username":"testuser","email_verified":true}]`

	mockOIDCCmd := exec.Command(juangoBin, "mockoidc")
	mockOIDCCmd.Env = append(os.Environ(),
		"MOCKOIDC_CLIENT_ID=test-client",
		"MOCKOIDC_CLIENT_SECRET=test-secret",
		"MOCKOIDC_ADDR=localhost",
		"MOCKOIDC_PORT="+mockOIDCPort,
		"MOCKOIDC_USERS="+mockOIDCUsers,
	)
	var mockOIDCOutput strings.Builder
	mockOIDCCmd.Stdout = &mockOIDCOutput
	mockOIDCCmd.Stderr = &mockOIDCOutput
	if err := mockOIDCCmd.Start(); err != nil {
		t.Fatalf("Failed to start mock OIDC: %v", err)
	}
	defer func() {
		mockOIDCCmd.Process.Kill()
		mockOIDCCmd.Wait()
	}()

	// Wait for mock OIDC to be ready
	// mockoidc serves the OIDC endpoints under /oidc path
	mockOIDCBaseURL := "http://127.0.0.1:" + mockOIDCPort
	wellKnownURL := mockOIDCBaseURL + "/oidc/.well-known/openid-configuration"
	t.Logf("Waiting for mock OIDC at %s", wellKnownURL)
	if !waitForEndpoint(wellKnownURL, 10*time.Second) {
		t.Fatalf("Mock OIDC server did not become ready. Output:\n%s", mockOIDCOutput.String())
	}

	// Fetch the actual issuer from the well-known config
	resp, err := http.Get(wellKnownURL)
	if err != nil {
		t.Fatalf("Failed to fetch well-known config: %v", err)
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		t.Fatalf("Failed to read well-known config body: %v", err)
	}
	t.Logf("Well-known config response: %s", string(bodyBytes))

	var oidcConfig map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &oidcConfig); err != nil {
		t.Fatalf("Failed to decode well-known config: %v\nBody: %s", err, string(bodyBytes))
	}
	mockOIDCIssuer, ok := oidcConfig["issuer"].(string)
	if !ok {
		t.Fatalf("Issuer not found or not a string in well-known config: %v", oidcConfig)
	}
	t.Logf("Mock OIDC ready, issuer: %s", mockOIDCIssuer)

	// Create config file
	appPort := "19081"
	configPath := filepath.Join(projectDir, "config.yml")
	config := `
listen_addr: ":` + appPort + `"
advertise_url: "http://localhost:` + appPort + `"
admin_mode_timeout: 30m

database:
  path: "` + projectDir + `/test.db"

session:
  cookie_name: "testapp_session"
  cookie_expiry: 24h
  authentication_key: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
  encryption_key: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"

oidc:
  issuer: "` + mockOIDCIssuer + `"
  client_id: "test-client"
  client_secret: "test-secret"
  scopes:
    - openid
    - profile
    - email

redis:
  addr: "localhost:6379"

logging:
  level: debug
  format: text
`
	if err := os.WriteFile(configPath, []byte(config), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Start the app
	appCmd := exec.Command(filepath.Join(projectDir, "testapp"), "serve", "-c", configPath)
	appCmd.Dir = projectDir
	var appOutput strings.Builder
	appCmd.Stdout = &appOutput
	appCmd.Stderr = &appOutput
	if err := appCmd.Start(); err != nil {
		t.Fatalf("Failed to start app: %v", err)
	}
	defer func() {
		appCmd.Process.Kill()
		appCmd.Wait()
	}()

	// Wait for app to be ready (session endpoint returns 401 when unauthenticated)
	appURL := "http://localhost:" + appPort
	if !waitForEndpointAnyStatus(appURL+"/api/auth/session", 10*time.Second) {
		t.Fatalf("App server did not become ready. Output:\n%s", appOutput.String())
	}

	// Test unauthenticated session
	sessionResp, err := http.Get(appURL + "/api/auth/session")
	if err != nil {
		t.Fatalf("Failed to get session: %v", err)
	}
	defer sessionResp.Body.Close()

	var sessionData map[string]interface{}
	if err := json.NewDecoder(sessionResp.Body).Decode(&sessionData); err != nil {
		t.Fatalf("Failed to decode session: %v", err)
	}

	if sessionData["authenticated"] != false {
		t.Errorf("Expected unauthenticated session, got: %v", sessionData)
	}

	// Test health endpoint
	healthResp, err := http.Get(appURL + "/api/health")
	if err != nil {
		t.Fatalf("Failed to get health: %v", err)
	}
	defer healthResp.Body.Close()

	if healthResp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 from health endpoint, got: %d", healthResp.StatusCode)
	}

	var healthData map[string]interface{}
	if err := json.NewDecoder(healthResp.Body).Decode(&healthData); err != nil {
		t.Fatalf("Failed to decode health response: %v", err)
	}

	if healthData["status"] != "ok" {
		t.Errorf("Expected health status 'ok', got: %v", healthData["status"])
	}

	if healthData["database"] != "ok" {
		t.Errorf("Expected database status 'ok', got: %v", healthData["database"])
	}

	t.Log("Scaffolded app is running with mock OIDC!")
}

func waitForEndpoint(url string, timeout time.Duration) bool {
	return waitForEndpointWithStatus(url, timeout, func(status int) bool {
		return status >= 200 && status < 300
	})
}

func waitForEndpointAnyStatus(url string, timeout time.Duration) bool {
	return waitForEndpointWithStatus(url, timeout, func(status int) bool {
		return status > 0
	})
}

func waitForEndpointWithStatus(url string, timeout time.Duration, acceptStatus func(int) bool) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := http.Get(url)
		if err == nil {
			status := resp.StatusCode
			resp.Body.Close()
			if acceptStatus(status) {
				return true
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	return false
}
