# Container Provisioning — Phase 2: Standby Lifecycle

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a standby container lifecycle that keeps containers alive after task completion for PR review feedback, with explicit teardown when the PR is merged or the task is resolved.

**Architecture:** One new file `internal/container/standby.go` with a `Standby` struct wrapping the `Manager` interface. `StartStandby` creates and starts a container but — unlike `Run` — does not auto-remove it. The caller controls lifecycle via `Teardown` (stop + remove) and `Wait` (block until self-exit). Integration tests cover standby + the provisioning pipeline from Phase 1.

**Tech Stack:** Go 1.26.1, existing `Manager` interface, stdlib `testing`.

**Prerequisites:** Phase 1 must be complete. This phase uses `Provision()`, `ProvisionParams`, `SecretConfig`, `EnvAgentToken`, and `EnvMCPEndpoint` from Phase 1 in integration tests.

**Design spec:** `PRODUCT.md` — see "Docker Isolation Layer" paragraph on standby mode: "After completing work, agents create a PR and enter standby mode — the container stays alive so the agent can address review feedback on its PR/MR. Containers are torn down only after the PR is merged (or the task is otherwise resolved)."

---

## Inherits From

Phase 1 introduced the following that this phase builds on:

- **`internal/container/token.go`** — `GenerateToken() (string, error)` for 256-bit hex tokens
- **`internal/container/secret.go`** — `SecretConfig` type and `ResolveSecrets()` for scoped secret resolution
- **`internal/container/provision.go`** — `Provision(ProvisionParams) (Config, string, error)` that combines token + secrets + MCP endpoint into a `Config.Env` map. Constants: `EnvAgentToken = "ELEPHANT_AGENT_TOKEN"`, `EnvMCPEndpoint = "ELEPHANT_MCP_ENDPOINT"`
- **`internal/container/integration_test.go`** — one new integration test `TestIntegration_ProvisionedContainer` appended at the end of the file

The existing lifecycle code is unchanged:
- **`internal/container/lifecycle.go`** — `Run()` for ephemeral containers (create → start → wait → remove)
- **`internal/container/container.go`** — `Config`, `Mount`, `Container`, `Status` types and `Manager` interface
- **`internal/container/docker.go`** — `DockerManager` implementing `Manager`
- **`internal/container/mock_manager_test.go`** — `mockManager` test double with `On*` function fields and `Calls` slice

The standby lifecycle reuses `stopTimeout` (10 seconds) from `internal/container/lifecycle.go:19` and the `RunResult` type from `internal/container/lifecycle.go:10-15`.

---

### Task 1: Standby Core — StartStandby and ID

**Files:**
- Create: `internal/container/standby.go`
- Create: `internal/container/standby_test.go`

- [ ] **Step 1: Write the failing tests for StartStandby and ID**

Create `internal/container/standby_test.go`:

```go
package container

import (
	"context"
	"fmt"
	"testing"
)

func TestStartStandby_HappyPath(t *testing.T) {
	mgr := &mockManager{}

	s, err := StartStandby(context.Background(), mgr, Config{Image: "alpine"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if s.ID() != "mock-id" {
		t.Fatalf("expected container ID mock-id, got %s", s.ID())
	}

	// Verify Create, Start, and Wait were called — but NOT Remove.
	expected := []string{"Create", "Start", "Wait"}
	if len(mgr.Calls) != len(expected) {
		t.Fatalf("expected %d calls, got %d: %+v", len(expected), len(mgr.Calls), mgr.Calls)
	}
	for i, name := range expected {
		if mgr.Calls[i].Method != name {
			t.Errorf("call %d: expected %s, got %s", i, name, mgr.Calls[i].Method)
		}
	}
}

func TestStartStandby_CreateFailure(t *testing.T) {
	mgr := &mockManager{
		OnCreate: func(ctx context.Context, cfg Config) (*Container, error) {
			return nil, fmt.Errorf("image not found")
		},
	}

	s, err := StartStandby(context.Background(), mgr, Config{Image: "bad"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if s != nil {
		t.Fatalf("expected nil Standby, got %+v", s)
	}
}

func TestStartStandby_StartFailure_CleansUp(t *testing.T) {
	mgr := &mockManager{
		OnStart: func(ctx context.Context, id string) error {
			return fmt.Errorf("start failed")
		},
	}

	s, err := StartStandby(context.Background(), mgr, Config{Image: "alpine"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if s != nil {
		t.Fatalf("expected nil Standby, got %+v", s)
	}

	// Verify Remove was called to clean up the created container.
	hasRemove := false
	for _, c := range mgr.Calls {
		if c.Method == "Remove" {
			hasRemove = true
		}
	}
	if !hasRemove {
		t.Error("expected Remove to be called after Start failure")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/container/ -run "TestStartStandby" -v`
Expected: FAIL — `StartStandby` and `Standby` not defined.

- [ ] **Step 3: Write the implementation**

Create `internal/container/standby.go`:

```go
package container

import (
	"context"
	"fmt"
)

// Standby manages a long-lived container that persists beyond initial task
// completion. Unlike Run, the container is not automatically removed — it stays
// alive for review feedback and is torn down only when explicitly requested.
type Standby struct {
	mgr    Manager
	ctr    *Container
	exitCh <-chan int64
	errCh  <-chan error
}

// StartStandby creates and starts a container that remains running until
// explicitly torn down via Teardown. If Start fails, the created container is
// removed before returning the error.
func StartStandby(ctx context.Context, mgr Manager, cfg Config) (*Standby, error) {
	ctr, err := mgr.Create(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("standby: %w", err)
	}

	if err := mgr.Start(ctx, ctr.ID); err != nil {
		_ = mgr.Remove(context.WithoutCancel(ctx), ctr.ID)
		return nil, fmt.Errorf("standby: %w", err)
	}

	exitCh, errCh := mgr.Wait(ctx, ctr.ID)

	return &Standby{
		mgr:    mgr,
		ctr:    ctr,
		exitCh: exitCh,
		errCh:  errCh,
	}, nil
}

// ID returns the Docker container ID.
func (s *Standby) ID() string {
	return s.ctr.ID
}

// Teardown stops and removes the container. Stop errors are ignored because
// the container may have already exited.
func (s *Standby) Teardown(ctx context.Context) error {
	_ = s.mgr.Stop(ctx, s.ctr.ID, stopTimeout)

	if err := s.mgr.Remove(ctx, s.ctr.ID); err != nil {
		return fmt.Errorf("standby teardown: %w", err)
	}

	return nil
}

// Wait blocks until the container exits on its own and returns the exit result.
// This does not remove the container — call Teardown separately if needed.
func (s *Standby) Wait(ctx context.Context) (*RunResult, error) {
	select {
	case code := <-s.exitCh:
		return &RunResult{ExitCode: code}, nil
	case err := <-s.errCh:
		return nil, fmt.Errorf("standby wait: %w", err)
	case <-ctx.Done():
		return nil, fmt.Errorf("standby wait: %w", ctx.Err())
	}
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/container/ -run "TestStartStandby" -v`
Expected: PASS (all 3 tests).

- [ ] **Step 5: Run linter and commit**

Run: `golangci-lint run ./internal/container/`
Expected: No issues.

```bash
git add internal/container/standby.go internal/container/standby_test.go
git commit -m "feat(container): add standby lifecycle — StartStandby and ID"
```

---

### Task 2: Standby Teardown

**Files:**
- Modify: `internal/container/standby_test.go` (append tests)

`Teardown` is already implemented in Task 1 — this task adds its unit tests.

- [ ] **Step 1: Write the teardown tests**

Append to `internal/container/standby_test.go`:

```go
func TestStandby_Teardown(t *testing.T) {
	mgr := &mockManager{}

	s, err := StartStandby(context.Background(), mgr, Config{Image: "alpine"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Reset calls to isolate teardown behavior.
	mgr.Calls = nil

	if err := s.Teardown(context.Background()); err != nil {
		t.Fatalf("Teardown failed: %v", err)
	}

	// Verify Stop and Remove were called.
	hasStop := false
	hasRemove := false
	for _, c := range mgr.Calls {
		if c.Method == "Stop" {
			hasStop = true
		}
		if c.Method == "Remove" {
			hasRemove = true
		}
	}
	if !hasStop {
		t.Error("expected Stop to be called during Teardown")
	}
	if !hasRemove {
		t.Error("expected Remove to be called during Teardown")
	}
}

func TestStandby_Teardown_RemoveFailure(t *testing.T) {
	mgr := &mockManager{
		OnRemove: func(ctx context.Context, id string) error {
			return fmt.Errorf("device busy")
		},
	}

	s, err := StartStandby(context.Background(), mgr, Config{Image: "alpine"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = s.Teardown(context.Background())
	if err == nil {
		t.Fatal("expected error from Remove failure, got nil")
	}
}
```

- [ ] **Step 2: Run tests to verify they pass**

Run: `go test ./internal/container/ -run "TestStandby_Teardown" -v`
Expected: PASS (both tests).

- [ ] **Step 3: Run linter and commit**

Run: `golangci-lint run ./internal/container/`
Expected: No issues.

```bash
git add internal/container/standby_test.go
git commit -m "test(container): add unit tests for standby Teardown"
```

---

### Task 3: Standby Wait

**Files:**
- Modify: `internal/container/standby_test.go` (append tests)

`Wait` is already implemented in Task 1 — this task adds its unit tests.

- [ ] **Step 1: Write the wait tests**

Append to `internal/container/standby_test.go` (add `"time"` to the import block):

```go
func TestStandby_Wait_ExitCode(t *testing.T) {
	mgr := &mockManager{
		OnWait: func(ctx context.Context, id string) (<-chan int64, <-chan error) {
			exitCh := make(chan int64, 1)
			errCh := make(chan error, 1)
			exitCh <- 0
			return exitCh, errCh
		},
	}

	s, err := StartStandby(context.Background(), mgr, Config{Image: "alpine"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result, err := s.Wait(context.Background())
	if err != nil {
		t.Fatalf("Wait failed: %v", err)
	}
	if result.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", result.ExitCode)
	}
}

func TestStandby_Wait_ContextCancellation(t *testing.T) {
	mgr := &mockManager{
		OnWait: func(ctx context.Context, id string) (<-chan int64, <-chan error) {
			// Never sends — simulates a running container.
			return make(chan int64), make(chan error)
		},
	}

	s, err := StartStandby(context.Background(), mgr, Config{Image: "alpine"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err = s.Wait(ctx)
	if err == nil {
		t.Fatal("expected error from context cancellation, got nil")
	}
}
```

Also add `"time"` to the import block at the top of `standby_test.go`. The final import block should be:

```go
import (
	"context"
	"fmt"
	"testing"
	"time"
)
```

- [ ] **Step 2: Run tests to verify they pass**

Run: `go test ./internal/container/ -run "TestStandby_Wait" -v`
Expected: PASS (both tests).

- [ ] **Step 3: Run all standby unit tests together**

Run: `go test ./internal/container/ -run "TestStartStandby|TestStandby" -v`
Expected: PASS (all 7 tests).

- [ ] **Step 4: Run linter and commit**

Run: `golangci-lint run ./internal/container/`
Expected: No issues.

```bash
git add internal/container/standby_test.go
git commit -m "test(container): add unit tests for standby Wait"
```

---

### Task 4: Standby Integration Tests

**Files:**
- Modify: `internal/container/integration_test.go` (append after the `TestIntegration_ProvisionedContainer` test added in Phase 1)

- [ ] **Step 1: Write integration tests for standby lifecycle**

Append to the end of `internal/container/integration_test.go`:

```go
func TestIntegration_StandbyLifecycle(t *testing.T) {
	mgr := newTestManager(t)

	// Start a long-running container in standby mode.
	s, err := StartStandby(context.Background(), mgr, Config{
		Image: testImage,
		Cmd:   []string{"sleep", "300"},
	})
	if err != nil {
		t.Fatalf("StartStandby failed: %v", err)
	}

	// Verify the container is running.
	status, err := mgr.Status(context.Background(), s.ID())
	if err != nil {
		t.Fatalf("Status failed: %v", err)
	}
	if status != StatusRunning {
		t.Fatalf("expected status %s, got %s", StatusRunning, status)
	}

	// Teardown the container.
	if err := s.Teardown(context.Background()); err != nil {
		t.Fatalf("Teardown failed: %v", err)
	}

	// Verify the container is gone (inspect should fail).
	_, err = mgr.Status(context.Background(), s.ID())
	if err == nil {
		t.Fatal("expected error inspecting removed container, got nil")
	}
}

func TestIntegration_StandbyWaitThenTeardown(t *testing.T) {
	mgr := newTestManager(t)

	// Start a container that exits quickly.
	s, err := StartStandby(context.Background(), mgr, Config{
		Image: testImage,
		Cmd:   []string{"echo", "done"},
	})
	if err != nil {
		t.Fatalf("StartStandby failed: %v", err)
	}

	// Wait for it to exit on its own.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := s.Wait(ctx)
	if err != nil {
		t.Fatalf("Wait failed: %v", err)
	}
	if result.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", result.ExitCode)
	}

	// Teardown should still work (remove the stopped container).
	if err := s.Teardown(context.Background()); err != nil {
		t.Fatalf("Teardown after exit failed: %v", err)
	}
}
```

- [ ] **Step 2: Run the standby integration tests**

Run: `go test -race -tags integration ./internal/container/ -run "TestIntegration_Standby" -v`
Expected: PASS (both tests). Requires Docker daemon running.

- [ ] **Step 3: Run the full test suite**

Run: `go test -race ./internal/container/ -v && go test -race -tags integration ./internal/container/ -v`
Expected: All unit tests and all integration tests pass (including Phase 1's tests).

- [ ] **Step 4: Run linter and commit**

Run: `golangci-lint run ./internal/container/`
Expected: No issues.

```bash
git add internal/container/integration_test.go
git commit -m "test(container): add integration tests for standby lifecycle"
```

---

## User-Visible Behaviors Preserved

All existing behaviors are unaffected by this phase:
- `Run()` still works as an ephemeral lifecycle (create → start → wait → remove)
- `Manager` interface is unchanged
- `DockerManager` is unchanged
- `Config` struct is unchanged
- All existing unit and integration tests continue to pass

## Changes Introduced

**New files:**
- `internal/container/standby.go` — `Standby` struct, `StartStandby()`, `Standby.ID()`, `Standby.Teardown()`, `Standby.Wait()`
- `internal/container/standby_test.go` — 7 unit tests

**Modified files:**
- `internal/container/integration_test.go` — 2 new integration tests appended (`TestIntegration_StandbyLifecycle`, `TestIntegration_StandbyWaitThenTeardown`)

**New exported symbols:**
- `Standby` struct
- `StartStandby(ctx context.Context, mgr Manager, cfg Config) (*Standby, error)`
- `(*Standby).ID() string`
- `(*Standby).Teardown(ctx context.Context) error`
- `(*Standby).Wait(ctx context.Context) (*RunResult, error)`

**Bridge code:** None.

**Dependencies added:** None.
