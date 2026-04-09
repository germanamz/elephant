# Tusk Integration — Phase 1: SDK Dependency & Engine Core

## Prerequisites

- None beyond the current codebase on `main`.

## Goal

Add the Tusk SDK dependency and create the `internal/work/` package with the `Engine` type, configuration, WBS level constants, and level validation. At the end of this phase, the `Engine` can be instantiated, creates a Tusk client with the hardcoded workflow and project, and exposes WBS level validation. No task creation or player registration yet — those come in Phase 2.

## Reference Material

- Design spec: `docs/superpowers/specs/2026-04-09-tusk-integration-design.md`
- Tusk SDK: `github.com/germanamz/tusk` — see `tusk.Config`, `tusk.NewClient()`, `tusk.Client.Close()`
- Tusk domain types: `github.com/germanamz/tusk/domain` — `Workflow`, `WorkflowTransition`, `Project`, `ProjectSettings`
- Existing pattern to follow: `internal/container/container.go` for type definitions, `internal/container/docker.go` for constructor patterns

## Tasks

### Task 1: Add Tusk SDK dependency

Run `go get github.com/germanamz/tusk` to add the dependency to `go.mod` and `go.sum`. Verify the module resolves and `go mod tidy` succeeds.

**Acceptance:** `go.mod` lists `github.com/germanamz/tusk` as a direct dependency. `go mod tidy` exits cleanly.

### Task 2: Create `internal/work/` package with WBS level constants and validation

Create file `internal/work/level.go` with:

```go
package work

// WBS level constants
const (
    LevelVision     = "vision"
    LevelRoadmap    = "roadmap"
    LevelInitiative = "initiative"
    LevelStory      = "story"
    LevelTask       = "task"
    LevelSubtask    = "subtask"
)

// ValidLevel reports whether level is a known WBS level.
func ValidLevel(level string) bool
```

Implementation: use a switch statement or a set lookup over the six constants.

**Acceptance:** `ValidLevel` returns `true` for all six levels and `false` for empty string, unknown strings, and mixed-case variants.

### Task 3: Create Engine type with Config and constructor

Create file `internal/work/engine.go` with:

```go
package work

import (
    "fmt"
    "github.com/germanamz/tusk"
    "github.com/germanamz/tusk/domain"
)

const (
    DefaultDBPath     = "~/.local/share/elephant/tusk.db"
    DefaultWorkflow   = "wbs"
    DefaultProject    = "default"
)

type Config struct {
    // DBPath is the path to the SQLite database file.
    // If empty, defaults to DefaultDBPath.
    DBPath string
}

type Engine struct {
    client *tusk.Client
}

// NewEngine creates a new Engine backed by a Tusk client.
func NewEngine(cfg Config) (*Engine, error)

// Close releases the underlying Tusk client and database connection.
func (e *Engine) Close() error
```

**Constructor behavior:**
1. If `cfg.DBPath` is empty, expand `DefaultDBPath` (resolve `~` to the user's home directory using `os.UserHomeDir()`).
2. Create a `tusk.Config` with:
   - `DBPath`: the resolved path
   - `Workflows`: a single workflow named `"wbs"` with statuses `["pending", "active", "completed", "deleted"]` and transitions: `pending→active`, `active→completed`, `completed→pending`, plus `pending→deleted`, `active→deleted`, `completed→deleted`.
   - `Projects`: a single project `"default"` using the `"wbs"` workflow with no custom settings.
3. Call `tusk.NewClient(tuskCfg)` and wrap the result in `Engine`.
4. On error, return `fmt.Errorf("work: new engine: %w", err)`.

**`Close()`:** delegates to `e.client.Close()`, wrapping errors with `fmt.Errorf("work: close: %w", err)`.

**Acceptance:** `NewEngine` with an empty config resolves the default DB path. `NewEngine` with a custom `DBPath` uses that path. `Close()` cleans up without error. The Tusk client is initialized with the correct workflow and project.

### Task 4: Unit tests for WBS levels

Create file `internal/work/level_test.go` with tests for `ValidLevel`:

- All six valid levels return `true`.
- Empty string returns `false`.
- Unknown string (e.g., `"milestone"`) returns `false`.
- Case-sensitive: `"Vision"` returns `false`.

**Acceptance:** All tests pass. No mocks needed.

### Task 5: Unit tests for Engine lifecycle

Create file `internal/work/engine_test.go` with tests using a temp directory (`t.TempDir()`) for the DB path:

- `TestNewEngine_DefaultConfig`: verify engine initializes with default workflow/project. After init, call `engine.Close()` — no error.
- `TestNewEngine_CustomDBPath`: pass a custom `DBPath` pointing to temp dir, verify initialization succeeds and creates the DB file.
- `TestNewEngine_Close`: verify `Close()` is idempotent-safe (or returns appropriate error on double close — match Tusk's behavior).

**Acceptance:** All tests pass. Tests use real Tusk client + temp SQLite DB (no mocks).

## Changes Introduced

| Category | Detail |
|----------|--------|
| New dependency | `github.com/germanamz/tusk` in `go.mod` |
| New file | `internal/work/level.go` — WBS level constants + `ValidLevel()` |
| New file | `internal/work/engine.go` — `Config`, `Engine`, `NewEngine()`, `Close()` |
| New file | `internal/work/level_test.go` — level validation tests |
| New file | `internal/work/engine_test.go` — engine lifecycle tests |
| No bridge code | None introduced |
| No modified files | `internal/container/` is untouched |
