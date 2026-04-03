# Container Lifecycle Phase 3: Integration Tests

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Verify the container lifecycle works against a real Docker daemon. These tests exercise `DockerManager` and `Run` end-to-end using lightweight Alpine containers.

**Architecture:** Integration tests live alongside the package code but are gated behind the `//go:build integration` build tag so they only run when Docker is available. A shared test helper handles setup (Docker client creation, cleanup).

**Tech Stack:** Go 1.26.1, Docker Engine (must be running locally)

**Reference spec:** `docs/superpowers/specs/2026-04-03-container-lifecycle-design.md`

**Prerequisite:** Phase 1 and Phase 2 must be completed (`DockerManager` and `Run` exist and unit tests pass).

**Important:** These tests require a running Docker daemon. Before starting, verify Docker is available:

```bash
docker info > /dev/null 2>&1 && echo "Docker is running" || echo "Docker is NOT running"
```

---

### Task 1: Test Helpers and Basic Spawn/Exit Test

**Files:**
- Create: `internal/container/integration_test.go`

This file contains the build tag, shared helpers, and the first integration test.

- [ ] **Step 1: Pull the Alpine image**

The integration tests use `alpine:latest`. Pull it once so tests don't depend on network speed:

```bash
docker pull alpine:latest
```

- [ ] **Step 2: Write the test helpers and first test**

Create `internal/container/integration_test.go`:

```go
//go:build integration

package container

import (
	"context"
	"testing"
	"time"
)

const testImage = "alpine:latest"

// newTestManager creates a DockerManager for integration tests.
// It calls t.Fatal if the Docker daemon is unreachable.
func newTestManager(t *testing.T) *DockerManager {
	t.Helper()
	mgr, err := NewDockerManager()
	if err != nil {
		t.Fatalf("failed to create DockerManager: %v", err)
	}
	return mgr
}

// removeContainer is a cleanup helper that removes a container by ID,
// ignoring errors (the container may already be removed).
func removeContainer(t *testing.T, mgr *DockerManager, id string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = mgr.Remove(ctx, id)
}

func TestIntegration_SpawnAndExit(t *testing.T) {
	mgr := newTestManager(t)
	ctx := context.Background()

	// Create a container that runs "echo hello" and exits 0.
	ctr, err := mgr.Create(ctx, Config{
		Image: testImage,
		Cmd:   []string{"echo", "hello"},
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	defer removeContainer(t, mgr, ctr.ID)

	// Verify it was created.
	status, err := mgr.Status(ctx, ctr.ID)
	if err != nil {
		t.Fatalf("Status failed: %v", err)
	}
	if status != StatusCreated {
		t.Fatalf("expected status %s, got %s", StatusCreated, status)
	}

	// Start it.
	if err := mgr.Start(ctx, ctr.ID); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Wait for it to finish.
	exitCh, errCh := mgr.Wait(ctx, ctr.ID)
	select {
	case code := <-exitCh:
		if code != 0 {
			t.Fatalf("expected exit code 0, got %d", code)
		}
	case err := <-errCh:
		t.Fatalf("Wait error: %v", err)
	case <-time.After(30 * time.Second):
		t.Fatal("timed out waiting for container to exit")
	}

	// Verify it stopped.
	status, err = mgr.Status(ctx, ctr.ID)
	if err != nil {
		t.Fatalf("Status after exit failed: %v", err)
	}
	if status != StatusStopped {
		t.Fatalf("expected status %s after exit, got %s", StatusStopped, status)
	}
}
```

- [ ] **Step 3: Run the test**

Run:

```bash
go test -v -tags integration -run TestIntegration_SpawnAndExit ./internal/container/...
```

Expected: PASS. The container runs `echo hello`, exits 0, and status transitions from `created` → `running` → `stopped`.

- [ ] **Step 4: Commit**

```bash
git add internal/container/integration_test.go
git commit -m "test(container): add integration test for basic container spawn and exit"
```

---

### Task 2: Test Bind Mounts and Environment Variables

**Files:**
- Modify: `internal/container/integration_test.go`

These tests verify that bind mounts and environment variables are correctly passed to containers.

- [ ] **Step 1: Add the bind mount test**

Append to `internal/container/integration_test.go`:

```go
func TestIntegration_BindMount(t *testing.T) {
	mgr := newTestManager(t)
	ctx := context.Background()

	// Create a temp directory with a file to mount into the container.
	tmpDir := t.TempDir()
	testFile := tmpDir + "/testfile.txt"
	if err := os.WriteFile(testFile, []byte("elephant"), 0o644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Run a container that reads the mounted file and exits with code 0
	// if the content matches, or code 1 if it doesn't.
	// "grep -q" exits 0 on match, 1 on no match.
	ctr, err := mgr.Create(ctx, Config{
		Image: testImage,
		Cmd:   []string{"grep", "-q", "elephant", "/workspace/testfile.txt"},
		Mounts: []Mount{
			{Source: tmpDir, Target: "/workspace"},
		},
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	defer removeContainer(t, mgr, ctr.ID)

	if err := mgr.Start(ctx, ctr.ID); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	exitCh, errCh := mgr.Wait(ctx, ctr.ID)
	select {
	case code := <-exitCh:
		if code != 0 {
			t.Fatalf("expected exit code 0 (file content matched), got %d", code)
		}
	case err := <-errCh:
		t.Fatalf("Wait error: %v", err)
	case <-time.After(30 * time.Second):
		t.Fatal("timed out waiting for container")
	}
}

func TestIntegration_EnvVars(t *testing.T) {
	mgr := newTestManager(t)
	ctx := context.Background()

	// Run a container that checks if the env var is set.
	// "test" exits 0 if the condition is true, 1 otherwise.
	ctr, err := mgr.Create(ctx, Config{
		Image: testImage,
		Cmd:   []string{"sh", "-c", "test \"$MY_TOKEN\" = \"secret123\""},
		Env: map[string]string{
			"MY_TOKEN": "secret123",
		},
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	defer removeContainer(t, mgr, ctr.ID)

	if err := mgr.Start(ctx, ctr.ID); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	exitCh, errCh := mgr.Wait(ctx, ctr.ID)
	select {
	case code := <-exitCh:
		if code != 0 {
			t.Fatalf("expected exit code 0 (env var matched), got %d", code)
		}
	case err := <-errCh:
		t.Fatalf("Wait error: %v", err)
	case <-time.After(30 * time.Second):
		t.Fatal("timed out waiting for container")
	}
}
```

- [ ] **Step 2: Add the `os` import**

Add `"os"` to the import block at the top of `internal/container/integration_test.go`:

```go
import (
	"context"
	"os"
	"testing"
	"time"
)
```

- [ ] **Step 3: Run the tests**

Run:

```bash
go test -v -tags integration -run "TestIntegration_BindMount|TestIntegration_EnvVars" ./internal/container/...
```

Expected: both tests PASS.

- [ ] **Step 4: Commit**

```bash
git add internal/container/integration_test.go
git commit -m "test(container): add integration tests for bind mounts and env vars"
```

---

### Task 3: Test Context Cancellation and Run End-to-End

**Files:**
- Modify: `internal/container/integration_test.go`

These tests verify cleanup on cancellation and the full `Run` function end-to-end.

- [ ] **Step 1: Add the cancellation and Run tests**

Append to `internal/container/integration_test.go`:

```go
func TestIntegration_ContextCancellation(t *testing.T) {
	mgr := newTestManager(t)
	ctx, cancel := context.WithCancel(context.Background())

	// Start a long-running container (sleep 300 seconds).
	ctr, err := mgr.Create(ctx, Config{
		Image: testImage,
		Cmd:   []string{"sleep", "300"},
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	containerID := ctr.ID

	if err := mgr.Start(ctx, containerID); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Verify it's running.
	status, err := mgr.Status(ctx, containerID)
	if err != nil {
		t.Fatalf("Status failed: %v", err)
	}
	if status != StatusRunning {
		t.Fatalf("expected status %s, got %s", StatusRunning, status)
	}

	// Cancel the context, then stop and remove manually (simulating what Run does).
	cancel()

	stopCtx, stopCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer stopCancel()

	if err := mgr.Stop(stopCtx, containerID, 5*time.Second); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}

	// Verify it stopped.
	status, err = mgr.Status(stopCtx, containerID)
	if err != nil {
		t.Fatalf("Status after stop failed: %v", err)
	}
	if status != StatusStopped {
		t.Fatalf("expected status %s after stop, got %s", StatusStopped, status)
	}

	// Clean up.
	if err := mgr.Remove(stopCtx, containerID); err != nil {
		t.Fatalf("Remove failed: %v", err)
	}
}

func TestIntegration_RunEndToEnd(t *testing.T) {
	mgr := newTestManager(t)

	// Use Run to execute a container that exits with code 0.
	result, err := Run(context.Background(), mgr, Config{
		Image: testImage,
		Cmd:   []string{"echo", "hello from Run"},
	})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if result.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", result.ExitCode)
	}
}

func TestIntegration_RunNonZeroExit(t *testing.T) {
	mgr := newTestManager(t)

	// Use Run to execute a container that exits with code 1.
	result, err := Run(context.Background(), mgr, Config{
		Image: testImage,
		Cmd:   []string{"sh", "-c", "exit 1"},
	})
	if err != nil {
		t.Fatalf("Run should not error on non-zero exit, got: %v", err)
	}

	if result.ExitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", result.ExitCode)
	}
}
```

- [ ] **Step 2: Run all integration tests**

Run:

```bash
go test -v -tags integration ./internal/container/...
```

Expected: all 6 integration tests pass (SpawnAndExit, BindMount, EnvVars, ContextCancellation, RunEndToEnd, RunNonZeroExit).

- [ ] **Step 3: Run the linter**

Run:

```bash
make lint
```

Expected: no errors.

- [ ] **Step 4: Commit**

```bash
git add internal/container/integration_test.go
git commit -m "test(container): add integration tests for cancellation and Run end-to-end"
```

---

### Task 4: Add Integration Test Make Target

**Files:**
- Modify: `Makefile`

Add a dedicated make target so developers can easily run integration tests.

- [ ] **Step 1: Add the `test-integration` target**

Add the following target to the `Makefile`, after the existing `test-race` target:

```makefile
test-integration:
	$(GO) test $(GOFLAGS) -tags integration ./...
```

Also update the `.PHONY` line to include it:

```makefile
.PHONY: all build clean test test-race test-integration vet lint run install setup-hooks
```

- [ ] **Step 2: Verify the target works**

Run:

```bash
make test-integration
```

Expected: all unit tests AND integration tests pass.

- [ ] **Step 3: Verify regular `make test` still skips integration tests**

Run:

```bash
make test
```

Expected: only unit tests run (the 6 `TestRun_*` tests from Phase 2). No integration tests should appear in the output.

- [ ] **Step 4: Commit**

```bash
git add Makefile
git commit -m "build: add make target for integration tests"
```
