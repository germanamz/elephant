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
