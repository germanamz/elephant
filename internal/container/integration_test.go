//go:build integration

package container

import (
	"context"
	"os"
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
