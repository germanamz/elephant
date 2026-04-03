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
