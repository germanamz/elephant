package work

import (
	"context"
	"testing"
)

func TestRegisterAgent_Success(t *testing.T) {
	engine := newTestEngine(t)
	ctx := context.Background()

	player, err := engine.RegisterAgent(ctx, "agent-001")
	if err != nil {
		t.Fatalf("RegisterAgent() error: %v", err)
	}

	if player.ID != "agent-001" {
		t.Errorf("player.ID = %q, want %q", player.ID, "agent-001")
	}
	if player.Type != "agent" {
		t.Errorf("player.Type = %q, want %q", player.Type, "agent")
	}
}

func TestRegisterAgent_Duplicate(t *testing.T) {
	engine := newTestEngine(t)
	ctx := context.Background()

	_, err := engine.RegisterAgent(ctx, "agent-dup")
	if err != nil {
		t.Fatalf("first RegisterAgent() error: %v", err)
	}

	// Tusk returns a conflict error for duplicate player IDs.
	_, err = engine.RegisterAgent(ctx, "agent-dup")
	if err == nil {
		t.Fatal("second RegisterAgent() = nil error, want conflict error")
	}
}

func TestServiceAccessors(t *testing.T) {
	engine := newTestEngine(t)
	ctx := context.Background()

	if engine.Tasks() == nil {
		t.Fatal("Tasks() returned nil")
	}
	if engine.Players() == nil {
		t.Fatal("Players() returned nil")
	}
	if engine.Relations() == nil {
		t.Fatal("Relations() returned nil")
	}

	// Verify basic operations through each accessor.
	tasks, err := engine.Tasks().List(ctx, nil)
	if err != nil {
		t.Fatalf("Tasks().List() error: %v", err)
	}
	if len(tasks) != 0 {
		t.Errorf("Tasks().List() returned %d tasks, want 0", len(tasks))
	}

	players, err := engine.Players().List(ctx)
	if err != nil {
		t.Fatalf("Players().List() error: %v", err)
	}
	if len(players) != 0 {
		t.Errorf("Players().List() returned %d players, want 0", len(players))
	}
}
