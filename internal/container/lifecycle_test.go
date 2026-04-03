package container

import (
	"context"
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
