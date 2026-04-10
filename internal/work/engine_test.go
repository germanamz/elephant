package work

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewEngine_DefaultConfig(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "tusk.db")
	engine, err := NewEngine(Config{DBPath: dbPath})
	if err != nil {
		t.Fatalf("NewEngine() error: %v", err)
	}
	defer engine.Close()

	if _, err := os.Stat(dbPath); err != nil {
		t.Errorf("database file not created at %s: %v", dbPath, err)
	}
}

func TestNewEngine_CustomDBPath(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "custom", "test.db")

	engine, err := NewEngine(Config{DBPath: dbPath})
	if err != nil {
		t.Fatalf("NewEngine(custom path) error: %v", err)
	}
	defer engine.Close()

	if _, err := os.Stat(dbPath); err != nil {
		t.Errorf("database file not created at custom path %s: %v", dbPath, err)
	}
}

func TestNewEngine_Close(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "tusk.db")
	engine, err := NewEngine(Config{DBPath: dbPath})
	if err != nil {
		t.Fatalf("NewEngine() error: %v", err)
	}

	// First close should succeed.
	if err := engine.Close(); err != nil {
		t.Fatalf("first Close() error: %v", err)
	}

	// Second close — Tusk/SQLite may return an error; we just verify it doesn't panic.
	_ = engine.Close()
}
