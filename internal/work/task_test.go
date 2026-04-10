package work

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
)

func newTestEngine(t *testing.T) *Engine {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "tusk.db")
	engine, err := NewEngine(Config{DBPath: dbPath})
	if err != nil {
		t.Fatalf("NewEngine() error: %v", err)
	}
	t.Cleanup(func() { engine.Close() })
	return engine
}

func TestCreateTask_ValidLevel(t *testing.T) {
	engine := newTestEngine(t)
	ctx := context.Background()

	levels := []string{
		LevelVision,
		LevelRoadmap,
		LevelInitiative,
		LevelStory,
		LevelTask,
		LevelSubtask,
	}

	for _, level := range levels {
		t.Run(level, func(t *testing.T) {
			task, err := engine.CreateTask(ctx, CreateTaskParams{
				Title:       "Test " + level,
				Description: "A test task at " + level + " level",
				Level:       level,
			})
			if err != nil {
				t.Fatalf("CreateTask(%q) error: %v", level, err)
			}

			if task.UDA["wbs_level"] != level {
				t.Errorf("UDA[wbs_level] = %v, want %q", task.UDA["wbs_level"], level)
			}
			if task.ProjectID != DefaultProject {
				t.Errorf("ProjectID = %q, want %q", task.ProjectID, DefaultProject)
			}
			if task.ID == (uuid.UUID{}) {
				t.Error("ID is zero UUID, want populated")
			}
			if task.ShortID == "" {
				t.Error("ShortID is empty, want populated")
			}
		})
	}
}

func TestCreateTask_InvalidLevel(t *testing.T) {
	engine := newTestEngine(t)
	ctx := context.Background()

	task, err := engine.CreateTask(ctx, CreateTaskParams{
		Title: "Bad level",
		Level: "milestone",
	})
	if err == nil {
		t.Fatal("CreateTask(invalid level) = nil error, want error")
	}
	if task != nil {
		t.Error("CreateTask(invalid level) returned non-nil task, want nil")
	}

	// Verify no task was created.
	tasks, err := engine.Tasks().List(ctx, nil)
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if len(tasks) != 0 {
		t.Errorf("List() returned %d tasks, want 0", len(tasks))
	}
}

func TestCreateTask_WithParent(t *testing.T) {
	engine := newTestEngine(t)
	ctx := context.Background()

	parent, err := engine.CreateTask(ctx, CreateTaskParams{
		Title: "Parent",
		Level: LevelInitiative,
	})
	if err != nil {
		t.Fatalf("CreateTask(parent) error: %v", err)
	}

	child, err := engine.CreateTask(ctx, CreateTaskParams{
		Title:    "Child",
		Level:    LevelStory,
		ParentID: &parent.ID,
	})
	if err != nil {
		t.Fatalf("CreateTask(child) error: %v", err)
	}

	if child.ParentID == nil {
		t.Fatal("child.ParentID is nil, want parent ID")
	}
	if *child.ParentID != parent.ID {
		t.Errorf("child.ParentID = %v, want %v", *child.ParentID, parent.ID)
	}
}

func TestCreateTask_EmptyTitle(t *testing.T) {
	engine := newTestEngine(t)
	ctx := context.Background()

	// Tusk requires a non-empty title; verify it returns an error.
	_, err := engine.CreateTask(ctx, CreateTaskParams{
		Title: "",
		Level: LevelTask,
	})
	if err == nil {
		t.Fatal("CreateTask(empty title) = nil error, want error")
	}
}
