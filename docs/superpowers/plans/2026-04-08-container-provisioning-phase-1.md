# Container Provisioning — Phase 1: Provisioning Infrastructure

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add token generation, scoped secret resolution, and a provisioner that combines them with an MCP endpoint into a ready-to-use container `Config`.

**Architecture:** Three new files in `internal/container/`: `token.go` (crypto token generation), `secret.go` (config-driven secret resolution), and `provision.go` (combines token + secrets + MCP endpoint into `Config.Env`). Each has a `_test.go`. No existing files are modified — these are purely additive. All higher-level code (container Manager, lifecycle, DockerManager) is untouched.

**Tech Stack:** Go 1.26.1, `crypto/rand`, stdlib `testing`.

**Prerequisites:** None. This phase operates on the base codebase as it exists on `main`.

**Design spec:** `PRODUCT.md` — see "Docker Isolation Layer" and "Elephant MCP Server > Agent authentication" and "Secret management" sections.

---

## Codebase Context

The implementer needs to know these existing types (in `internal/container/container.go`):

```go
type Config struct {
    Image      string
    Mounts     []Mount
    Env        map[string]string   // <-- this is what we populate
    WorkingDir string
    Cmd        []string
}
```

The `Env` field is already used by `DockerManager.Create` (in `internal/container/docker.go:53-55`) to convert `map[string]string` into Docker's `[]string{"K=V"}` format. No changes to the existing code are needed — the provisioner just builds the `Env` map that flows through the existing pipeline.

Existing test patterns (see `internal/container/lifecycle_test.go` and `internal/container/mock_manager_test.go`):
- Package: `container` (white-box, same package)
- Each scenario is a separate `Test<Function>_<Scenario>` function (not table-driven)
- Assertions: stdlib `t.Fatalf` / `t.Errorf` only — no external libraries
- Hand-written mocks with `On*` function fields and a `Calls` slice

---

### Task 1: Token Generation

**Files:**
- Create: `internal/container/token.go`
- Create: `internal/container/token_test.go`

- [ ] **Step 1: Write the failing tests**

Create `internal/container/token_test.go`:

```go
package container

import (
	"encoding/hex"
	"testing"
)

func TestGenerateToken_Length(t *testing.T) {
	token, err := GenerateToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 32 bytes = 64 hex characters.
	if len(token) != 64 {
		t.Fatalf("expected 64-char token, got %d chars: %s", len(token), token)
	}
}

func TestGenerateToken_ValidHex(t *testing.T) {
	token, err := GenerateToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := hex.DecodeString(token); err != nil {
		t.Fatalf("token is not valid hex: %v", err)
	}
}

func TestGenerateToken_Unique(t *testing.T) {
	token1, err := GenerateToken()
	if err != nil {
		t.Fatalf("unexpected error generating token 1: %v", err)
	}

	token2, err := GenerateToken()
	if err != nil {
		t.Fatalf("unexpected error generating token 2: %v", err)
	}

	if token1 == token2 {
		t.Fatal("expected two different tokens, got identical values")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/container/ -run TestGenerateToken -v`
Expected: FAIL — `GenerateToken` not defined.

- [ ] **Step 3: Write the implementation**

Create `internal/container/token.go`:

```go
package container

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

const tokenBytes = 32

// GenerateToken returns a cryptographically random 256-bit hex-encoded token.
func GenerateToken() (string, error) {
	b := make([]byte, tokenBytes)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}

	return hex.EncodeToString(b), nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/container/ -run TestGenerateToken -v`
Expected: PASS (all 3 tests).

- [ ] **Step 5: Run linter and commit**

Run: `golangci-lint run ./internal/container/`
Expected: No issues.

```bash
git add internal/container/token.go internal/container/token_test.go
git commit -m "feat(container): add cryptographic agent token generation"
```

---

### Task 2: Secret Resolution

**Files:**
- Create: `internal/container/secret.go`
- Create: `internal/container/secret_test.go`

- [ ] **Step 1: Write the failing tests**

Create `internal/container/secret_test.go`:

```go
package container

import "testing"

func TestResolveSecrets_GlobalOnly(t *testing.T) {
	cfg := SecretConfig{
		Projects: map[string][]string{
			"*": {"GITHUB_TOKEN"},
		},
	}
	lookup := func(name string) (string, bool) {
		if name == "GITHUB_TOKEN" {
			return "ghp_abc123", true
		}
		return "", false
	}

	secrets, err := ResolveSecrets(cfg, "any-project", lookup)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if secrets["GITHUB_TOKEN"] != "ghp_abc123" {
		t.Fatalf("expected GITHUB_TOKEN=ghp_abc123, got %q", secrets["GITHUB_TOKEN"])
	}
	if len(secrets) != 1 {
		t.Fatalf("expected 1 secret, got %d", len(secrets))
	}
}

func TestResolveSecrets_ProjectSpecific(t *testing.T) {
	cfg := SecretConfig{
		Projects: map[string][]string{
			"frontend": {"NPM_TOKEN"},
		},
	}
	lookup := func(name string) (string, bool) {
		if name == "NPM_TOKEN" {
			return "npm_xyz", true
		}
		return "", false
	}

	secrets, err := ResolveSecrets(cfg, "frontend", lookup)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if secrets["NPM_TOKEN"] != "npm_xyz" {
		t.Fatalf("expected NPM_TOKEN=npm_xyz, got %q", secrets["NPM_TOKEN"])
	}
}

func TestResolveSecrets_MergesGlobalAndProject(t *testing.T) {
	cfg := SecretConfig{
		Projects: map[string][]string{
			"*":        {"GITHUB_TOKEN"},
			"frontend": {"NPM_TOKEN"},
		},
	}
	lookup := func(name string) (string, bool) {
		switch name {
		case "GITHUB_TOKEN":
			return "ghp_abc", true
		case "NPM_TOKEN":
			return "npm_xyz", true
		}
		return "", false
	}

	secrets, err := ResolveSecrets(cfg, "frontend", lookup)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(secrets) != 2 {
		t.Fatalf("expected 2 secrets, got %d", len(secrets))
	}
	if secrets["GITHUB_TOKEN"] != "ghp_abc" {
		t.Fatalf("expected GITHUB_TOKEN=ghp_abc, got %q", secrets["GITHUB_TOKEN"])
	}
	if secrets["NPM_TOKEN"] != "npm_xyz" {
		t.Fatalf("expected NPM_TOKEN=npm_xyz, got %q", secrets["NPM_TOKEN"])
	}
}

func TestResolveSecrets_MissingSecret(t *testing.T) {
	cfg := SecretConfig{
		Projects: map[string][]string{
			"*": {"MISSING_VAR"},
		},
	}
	lookup := func(name string) (string, bool) {
		return "", false
	}

	_, err := ResolveSecrets(cfg, "proj", lookup)
	if err == nil {
		t.Fatal("expected error for missing secret, got nil")
	}
}

func TestResolveSecrets_NoConfig(t *testing.T) {
	cfg := SecretConfig{}
	lookup := func(name string) (string, bool) {
		return "val", true
	}

	secrets, err := ResolveSecrets(cfg, "proj", lookup)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(secrets) != 0 {
		t.Fatalf("expected 0 secrets, got %d", len(secrets))
	}
}

func TestResolveSecrets_UnknownProjectGetsGlobalOnly(t *testing.T) {
	cfg := SecretConfig{
		Projects: map[string][]string{
			"*":        {"GLOBAL_SECRET"},
			"frontend": {"NPM_TOKEN"},
		},
	}
	lookup := func(name string) (string, bool) {
		if name == "GLOBAL_SECRET" {
			return "global_val", true
		}
		return "", false
	}

	secrets, err := ResolveSecrets(cfg, "backend", lookup)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(secrets) != 1 {
		t.Fatalf("expected 1 secret, got %d", len(secrets))
	}
	if secrets["GLOBAL_SECRET"] != "global_val" {
		t.Fatalf("expected GLOBAL_SECRET=global_val, got %q", secrets["GLOBAL_SECRET"])
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/container/ -run TestResolveSecrets -v`
Expected: FAIL — `SecretConfig` and `ResolveSecrets` not defined.

- [ ] **Step 3: Write the implementation**

Create `internal/container/secret.go`:

```go
package container

import "fmt"

// SecretConfig defines which secrets are available per project.
// The special project key "*" defines global secrets available to all projects.
type SecretConfig struct {
	// Projects maps project names to their allowed secret names (env var keys).
	Projects map[string][]string
}

// ResolveSecrets returns the secret env vars for the given project.
// It merges global ("*") secrets with project-specific secrets.
// The lookup function retrieves secret values by name (e.g., os.LookupEnv).
// Returns an error if any configured secret cannot be found via lookup.
func ResolveSecrets(cfg SecretConfig, project string, lookup func(string) (string, bool)) (map[string]string, error) {
	var names []string

	if global, ok := cfg.Projects["*"]; ok {
		names = append(names, global...)
	}

	if project != "*" {
		if proj, ok := cfg.Projects[project]; ok {
			names = append(names, proj...)
		}
	}

	secrets := make(map[string]string, len(names))
	for _, name := range names {
		val, ok := lookup(name)
		if !ok {
			return nil, fmt.Errorf("resolve secrets: secret %q not found", name)
		}
		secrets[name] = val
	}

	return secrets, nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/container/ -run TestResolveSecrets -v`
Expected: PASS (all 6 tests).

- [ ] **Step 5: Run linter and commit**

Run: `golangci-lint run ./internal/container/`
Expected: No issues.

```bash
git add internal/container/secret.go internal/container/secret_test.go
git commit -m "feat(container): add scoped secret resolution from config"
```

---

### Task 3: Provisioner

**Files:**
- Create: `internal/container/provision.go`
- Create: `internal/container/provision_test.go`

This task depends on Task 1 (`GenerateToken` in `token.go`) and Task 2 (`SecretConfig`, `ResolveSecrets` in `secret.go`).

- [ ] **Step 1: Write the failing tests**

Create `internal/container/provision_test.go`:

```go
package container

import "testing"

func TestProvision_InjectsTokenAndEndpoint(t *testing.T) {
	params := ProvisionParams{
		Base:        Config{Image: "agent:latest"},
		MCPEndpoint: "http://host.docker.internal:9090",
		Project:     "myproj",
		Secrets:     SecretConfig{},
		LookupSecret: func(name string) (string, bool) {
			return "", false
		},
	}

	cfg, token, err := Provision(params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if token == "" {
		t.Fatal("expected non-empty token")
	}
	if len(token) != 64 {
		t.Fatalf("expected 64-char token, got %d", len(token))
	}

	if cfg.Env[EnvAgentToken] != token {
		t.Fatalf("expected env %s=%s, got %q", EnvAgentToken, token, cfg.Env[EnvAgentToken])
	}
	if cfg.Env[EnvMCPEndpoint] != "http://host.docker.internal:9090" {
		t.Fatalf("expected env %s=http://host.docker.internal:9090, got %q", EnvMCPEndpoint, cfg.Env[EnvMCPEndpoint])
	}
}

func TestProvision_InjectsSecrets(t *testing.T) {
	params := ProvisionParams{
		Base:    Config{Image: "agent:latest"},
		Project: "frontend",
		Secrets: SecretConfig{
			Projects: map[string][]string{
				"*":        {"GITHUB_TOKEN"},
				"frontend": {"NPM_TOKEN"},
			},
		},
		LookupSecret: func(name string) (string, bool) {
			switch name {
			case "GITHUB_TOKEN":
				return "ghp_abc", true
			case "NPM_TOKEN":
				return "npm_xyz", true
			}
			return "", false
		},
	}

	cfg, _, err := Provision(params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Env["GITHUB_TOKEN"] != "ghp_abc" {
		t.Fatalf("expected GITHUB_TOKEN=ghp_abc, got %q", cfg.Env["GITHUB_TOKEN"])
	}
	if cfg.Env["NPM_TOKEN"] != "npm_xyz" {
		t.Fatalf("expected NPM_TOKEN=npm_xyz, got %q", cfg.Env["NPM_TOKEN"])
	}
}

func TestProvision_PreservesBaseEnv(t *testing.T) {
	params := ProvisionParams{
		Base: Config{
			Image: "agent:latest",
			Env:   map[string]string{"EXISTING": "value"},
		},
		MCPEndpoint: "http://localhost:9090",
		Project:     "proj",
		Secrets:     SecretConfig{},
		LookupSecret: func(name string) (string, bool) {
			return "", false
		},
	}

	cfg, _, err := Provision(params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Env["EXISTING"] != "value" {
		t.Fatalf("expected EXISTING=value preserved, got %q", cfg.Env["EXISTING"])
	}
	if cfg.Env[EnvAgentToken] == "" {
		t.Fatal("expected agent token to be injected")
	}
}

func TestProvision_PreservesBaseConfig(t *testing.T) {
	params := ProvisionParams{
		Base: Config{
			Image:      "agent:latest",
			WorkingDir: "/custom",
			Cmd:        []string{"bash"},
			Mounts:     []Mount{{Source: "/host", Target: "/ctr"}},
		},
		MCPEndpoint: "http://localhost:9090",
		Project:     "proj",
		Secrets:     SecretConfig{},
		LookupSecret: func(name string) (string, bool) {
			return "", false
		},
	}

	cfg, _, err := Provision(params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Image != "agent:latest" {
		t.Fatalf("expected image agent:latest, got %s", cfg.Image)
	}
	if cfg.WorkingDir != "/custom" {
		t.Fatalf("expected WorkingDir /custom, got %s", cfg.WorkingDir)
	}
	if len(cfg.Cmd) != 1 || cfg.Cmd[0] != "bash" {
		t.Fatalf("expected Cmd [bash], got %v", cfg.Cmd)
	}
	if len(cfg.Mounts) != 1 || cfg.Mounts[0].Source != "/host" {
		t.Fatalf("expected mount preserved, got %v", cfg.Mounts)
	}
}

func TestProvision_FailsOnMissingSecret(t *testing.T) {
	params := ProvisionParams{
		Base:    Config{Image: "agent:latest"},
		Project: "proj",
		Secrets: SecretConfig{
			Projects: map[string][]string{
				"proj": {"MISSING"},
			},
		},
		LookupSecret: func(name string) (string, bool) {
			return "", false
		},
	}

	_, _, err := Provision(params)
	if err == nil {
		t.Fatal("expected error for missing secret, got nil")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/container/ -run TestProvision -v`
Expected: FAIL — `ProvisionParams`, `Provision`, `EnvAgentToken`, `EnvMCPEndpoint` not defined.

- [ ] **Step 3: Write the implementation**

Create `internal/container/provision.go`:

```go
package container

import "fmt"

const (
	// EnvAgentToken is the env var name for the agent's auth token.
	EnvAgentToken = "ELEPHANT_AGENT_TOKEN"

	// EnvMCPEndpoint is the env var name for the Elephant MCP server address.
	EnvMCPEndpoint = "ELEPHANT_MCP_ENDPOINT"
)

// ProvisionParams holds the inputs needed to provision a container for an agent.
type ProvisionParams struct {
	// Base is the base container config (image, mounts, working dir, cmd).
	Base Config

	// MCPEndpoint is the network address of Elephant's MCP server.
	MCPEndpoint string

	// Project is the project name, used to scope secret injection.
	Project string

	// Secrets defines which secrets are available per project.
	Secrets SecretConfig

	// LookupSecret retrieves a secret value by name.
	// In production, pass os.LookupEnv.
	LookupSecret func(string) (string, bool)
}

// Provision creates a container Config with auth token, scoped secrets, and MCP
// endpoint injected as environment variables. It returns the populated Config and
// the generated token (so the caller can store it for later MCP auth validation).
func Provision(params ProvisionParams) (Config, string, error) {
	token, err := GenerateToken()
	if err != nil {
		return Config{}, "", fmt.Errorf("provision: %w", err)
	}

	secrets, err := ResolveSecrets(params.Secrets, params.Project, params.LookupSecret)
	if err != nil {
		return Config{}, "", fmt.Errorf("provision: %w", err)
	}

	cfg := params.Base
	if cfg.Env == nil {
		cfg.Env = make(map[string]string)
	}

	cfg.Env[EnvAgentToken] = token
	cfg.Env[EnvMCPEndpoint] = params.MCPEndpoint

	for k, v := range secrets {
		cfg.Env[k] = v
	}

	return cfg, token, nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/container/ -run TestProvision -v`
Expected: PASS (all 5 tests).

- [ ] **Step 5: Run linter and commit**

Run: `golangci-lint run ./internal/container/`
Expected: No issues.

```bash
git add internal/container/provision.go internal/container/provision_test.go
git commit -m "feat(container): add provisioner for token, secrets, and MCP endpoint injection"
```

---

### Task 4: Provisioning Integration Test

**Files:**
- Modify: `internal/container/integration_test.go` (append after line 248)

This test exercises the full provisioning pipeline through a real Docker container: generate token, resolve secrets, build Config, run container, verify all env vars are present inside the container.

- [ ] **Step 1: Write the integration test**

Append to the end of `internal/container/integration_test.go`:

```go
func TestIntegration_ProvisionedContainer(t *testing.T) {
	mgr := newTestManager(t)

	params := ProvisionParams{
		Base: Config{
			Image: testImage,
			Cmd:   []string{"sh", "-c", "test -n \"$ELEPHANT_AGENT_TOKEN\" && test \"$ELEPHANT_MCP_ENDPOINT\" = \"http://localhost:9090\" && test \"$MY_SECRET\" = \"s3cret\""},
		},
		MCPEndpoint: "http://localhost:9090",
		Project:     "testproj",
		Secrets: SecretConfig{
			Projects: map[string][]string{
				"testproj": {"MY_SECRET"},
			},
		},
		LookupSecret: func(name string) (string, bool) {
			if name == "MY_SECRET" {
				return "s3cret", true
			}
			return "", false
		},
	}

	cfg, token, err := Provision(params)
	if err != nil {
		t.Fatalf("Provision failed: %v", err)
	}

	if token == "" {
		t.Fatal("expected non-empty token from Provision")
	}

	result, err := Run(context.Background(), mgr, cfg)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if result.ExitCode != 0 {
		t.Fatalf("expected exit code 0 (all env vars present and correct), got %d", result.ExitCode)
	}
}
```

- [ ] **Step 2: Run the integration test**

Run: `go test -race -tags integration ./internal/container/ -run TestIntegration_ProvisionedContainer -v`
Expected: PASS. Requires Docker daemon running.

- [ ] **Step 3: Run the full test suite**

Run: `go test -race ./internal/container/ -v && go test -race -tags integration ./internal/container/ -v`
Expected: All unit tests and integration tests pass.

- [ ] **Step 4: Run linter and commit**

Run: `golangci-lint run ./internal/container/`
Expected: No issues.

```bash
git add internal/container/integration_test.go
git commit -m "test(container): add integration test for provisioned container"
```

---

## Changes Introduced

**New files:**
- `internal/container/token.go` — `GenerateToken() (string, error)`, `tokenBytes` constant
- `internal/container/token_test.go` — 3 unit tests
- `internal/container/secret.go` — `SecretConfig` type, `ResolveSecrets()` function
- `internal/container/secret_test.go` — 6 unit tests
- `internal/container/provision.go` — `ProvisionParams` type, `Provision()` function, `EnvAgentToken` and `EnvMCPEndpoint` constants
- `internal/container/provision_test.go` — 5 unit tests

**Modified files:**
- `internal/container/integration_test.go` — 1 new integration test appended

**New environment variables (injected into containers):**
- `ELEPHANT_AGENT_TOKEN` — unique per-agent auth token (256-bit hex)
- `ELEPHANT_MCP_ENDPOINT` — Elephant MCP server address

**New exported symbols:**
- `GenerateToken() (string, error)`
- `SecretConfig` struct
- `ResolveSecrets(cfg SecretConfig, project string, lookup func(string) (string, bool)) (map[string]string, error)`
- `ProvisionParams` struct
- `Provision(params ProvisionParams) (Config, string, error)`
- `EnvAgentToken` constant (`"ELEPHANT_AGENT_TOKEN"`)
- `EnvMCPEndpoint` constant (`"ELEPHANT_MCP_ENDPOINT"`)

**Bridge code:** None.

**Dependencies added:** None (uses only stdlib `crypto/rand`, `encoding/hex`, `fmt`).
