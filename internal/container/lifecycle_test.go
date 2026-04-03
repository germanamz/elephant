package container

import (
	"context"
	"fmt"
	"testing"
)

func TestRun_HappyPath(t *testing.T) {
	mgr := &mockManager{}
	cfg := Config{Image: "alpine:latest"}

	result, err := Run(context.Background(), mgr, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", result.ExitCode)
	}

	// Verify the exact call sequence: Create -> Start -> Wait -> Remove
	expected := []string{"Create", "Start", "Wait", "Remove"}
	if len(mgr.Calls) != len(expected) {
		t.Fatalf("expected %d calls, got %d: %+v", len(expected), len(mgr.Calls), mgr.Calls)
	}
	for i, name := range expected {
		if mgr.Calls[i].Method != name {
			t.Errorf("call %d: expected %s, got %s", i, name, mgr.Calls[i].Method)
		}
	}

	// Verify Start, Wait, Remove were called with the container ID from Create
	for _, c := range mgr.Calls[1:] {
		if c.ID != "mock-id" {
			t.Errorf("expected call %s with ID 'mock-id', got '%s'", c.Method, c.ID)
		}
	}
}

func TestRun_NonZeroExitCode(t *testing.T) {
	mgr := &mockManager{
		OnWait: func(ctx context.Context, id string) (<-chan int64, <-chan error) {
			exitCh := make(chan int64, 1)
			errCh := make(chan error, 1)
			exitCh <- 42
			return exitCh, errCh
		},
	}

	result, err := Run(context.Background(), mgr, Config{Image: "alpine"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ExitCode != 42 {
		t.Fatalf("expected exit code 42, got %d", result.ExitCode)
	}

	// Container should still be removed
	lastCall := mgr.Calls[len(mgr.Calls)-1]
	if lastCall.Method != "Remove" {
		t.Errorf("expected last call to be Remove, got %s", lastCall.Method)
	}
}

func TestRun_CreateFailure(t *testing.T) {
	mgr := &mockManager{
		OnCreate: func(ctx context.Context, cfg Config) (*Container, error) {
			return nil, fmt.Errorf("image not found")
		},
	}

	result, err := Run(context.Background(), mgr, Config{Image: "nonexistent"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if result != nil {
		t.Fatalf("expected nil result, got %+v", result)
	}

	// No Remove call — nothing was created
	for _, c := range mgr.Calls {
		if c.Method == "Remove" {
			t.Error("Remove should not be called when Create fails")
		}
	}
}

func TestRun_StartFailure(t *testing.T) {
	mgr := &mockManager{
		OnStart: func(ctx context.Context, id string) error {
			return fmt.Errorf("container start failed")
		},
	}

	_, err := Run(context.Background(), mgr, Config{Image: "alpine"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// Container must still be removed after Start failure
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

func TestRun_WaitError(t *testing.T) {
	mgr := &mockManager{
		OnWait: func(ctx context.Context, id string) (<-chan int64, <-chan error) {
			exitCh := make(chan int64, 1)
			errCh := make(chan error, 1)
			errCh <- fmt.Errorf("wait failed")
			return exitCh, errCh
		},
	}

	_, err := Run(context.Background(), mgr, Config{Image: "alpine"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// Container must still be removed
	hasRemove := false
	for _, c := range mgr.Calls {
		if c.Method == "Remove" {
			hasRemove = true
		}
	}
	if !hasRemove {
		t.Error("expected Remove to be called after Wait error")
	}
}

func TestRun_RemoveFailure(t *testing.T) {
	mgr := &mockManager{
		OnRemove: func(ctx context.Context, id string) error {
			return fmt.Errorf("remove failed: device busy")
		},
	}

	result, err := Run(context.Background(), mgr, Config{Image: "alpine"})

	// The result should still be returned with the exit code.
	if result == nil {
		t.Fatal("expected result even when Remove fails")
	}
	if result.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", result.ExitCode)
	}

	// An error should be surfaced about the remove failure.
	if err == nil {
		t.Fatal("expected error from Remove failure, got nil")
	}
}

func TestRun_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	mgr := &mockManager{
		OnWait: func(ctx context.Context, id string) (<-chan int64, <-chan error) {
			// Simulate a long-running container — never sends on either channel.
			// The context cancellation should trigger Stop.
			exitCh := make(chan int64)
			errCh := make(chan error)
			return exitCh, errCh
		},
	}

	// Cancel the context immediately after Wait is called.
	originalOnWait := mgr.OnWait
	mgr.OnWait = func(ctx context.Context, id string) (<-chan int64, <-chan error) {
		exitCh, errCh := originalOnWait(ctx, id)
		cancel()
		return exitCh, errCh
	}

	_, err := Run(ctx, mgr, Config{Image: "alpine"})
	if err == nil {
		t.Fatal("expected error from context cancellation, got nil")
	}

	// Verify Stop was called
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
		t.Error("expected Stop to be called on context cancellation")
	}
	if !hasRemove {
		t.Error("expected Remove to be called on context cancellation")
	}
}
