# Tusk Integration — Design Spec

## Context

Elephant is an AI orchestration platform that runs agents in isolated Docker containers. The container management layer (`internal/container/`) is complete. The next step is embedding the Tusk SDK (`github.com/germanamz/tusk`) as Elephant's task/work management engine, enabling hierarchical task tracking, workflow enforcement, and agent-to-task coordination.

This spec covers the "Tusk Integration" initiative from the v0.1 roadmap:
- Add Tusk SDK dependency
- Initialize Tusk Client with Elephant-managed config
- Configure WBS-aligned task hierarchy
- Register agents as Tusk players on container spawn

## Architecture

### New package: `internal/work/`

A dedicated package that wraps the Tusk SDK and owns all work management semantics for Elephant. It mirrors how `internal/container/` encapsulates Docker — the two packages remain independent, composed by the caller.

The package exposes an `Engine` type that owns the Tusk client lifecycle.

### Engine

```go
type Engine struct {
    client *tusk.Client
}
```

**`NewEngine(cfg Config) (*Engine, error)`** — Creates and initializes the Tusk client. Manages the underlying SQLite database.

**`Close() error`** — Releases the Tusk client and database connection.

**Config:**

```go
type Config struct {
    DBPath string // defaults to ~/.local/share/elephant/tusk.db
}
```

- If `DBPath` is empty, defaults to `~/.local/share/elephant/tusk.db`.
- Workflow and project configuration are hardcoded — Elephant owns these semantics.

### Hardcoded Tusk Configuration

**Workflow: `"wbs"`**

| Status | Description |
|--------|-------------|
| `pending` | Initial state, not yet started |
| `active` | Work in progress |
| `completed` | Done |
| `deleted` | Soft-deleted |

**Transitions:**
- `pending` → `active`
- `active` → `completed`
- `completed` → `pending` (reopen)
- any → `deleted`

**Project: `"default"`**
- Uses the `"wbs"` workflow.
- No custom project settings initially.

### WBS Levels

Six constants representing the Work Breakdown Structure hierarchy:

```go
const (
    LevelVision     = "vision"
    LevelRoadmap    = "roadmap"
    LevelInitiative = "initiative"
    LevelStory      = "story"
    LevelTask       = "task"
    LevelSubtask    = "subtask"
)
```

Stored as a Tusk UDA field (`wbs_level`) on every task created through the work package. Parent-child nesting in Tusk naturally mirrors the WBS hierarchy — a "story" task's parent is an "initiative" task, and so on.

A validation function ensures a given string is a known WBS level:

```go
func ValidLevel(level string) bool
```

### Task Creation

```go
type CreateTaskParams struct {
    Title       string
    Description string
    Level       string     // must be a valid WBS level
    ParentID    *uuid.UUID // optional, for hierarchy
}

func (e *Engine) CreateTask(ctx context.Context, params CreateTaskParams) (*domain.Task, error)
```

- Validates `Level` is a known WBS level.
- Sets `uda["wbs_level"]` on the Tusk task.
- Assigns to the `"default"` project.
- Delegates to `tusk.Client.Tasks.Create()`.
- Returns the created task with its generated IDs.

### Player Registration

```go
func (e *Engine) RegisterAgent(ctx context.Context, agentID string) (*domain.Player, error)
```

- Wraps `tusk.Client.Players.Register(ctx, agentID, "agent")`.
- Called by the orchestration layer before container provisioning.
- The caller passes the agent/player ID into `ProvisionParams` for env var injection — the `work` package has no awareness of containers.

### Delegation to Tusk

The `Engine` provides direct access to the underlying Tusk client's services for operations that don't need Elephant-specific wrapping:

```go
func (e *Engine) Tasks() *service.TaskService
func (e *Engine) Players() *service.PlayerService
func (e *Engine) Relations() *service.RelationService
```

This avoids wrapping every Tusk method with a pass-through. Callers use these for operations like `Claim`, `Complete`, `Pop`, `Annotate`, `Start`, `Release`, etc.

## Data Flow

```
Caller (cmd/ or future orchestrator)
  |
  |-- work.NewEngine(cfg)           -> initializes Tusk client + SQLite DB
  |-- work.RegisterAgent(agentID)   -> registers player in Tusk
  |-- work.CreateTask(params)       -> creates WBS-shaped task
  |
  |-- container.Provision(params)   -> provisions container (player ID as env var)
  +-- container.Run / StartStandby  -> executes agent
```

The `work` and `container` packages are independent. The caller composes them.

## What Stays Unchanged

- **`internal/container/`** — No Tusk awareness. No changes to `Provision()`, `Run()`, `StartStandby()`, or any existing code.
- **`ProvisionParams`** — The caller passes the player ID as an additional env var in the base `Config.Env`, same as any other env var.

## Testing Strategy

- **Unit tests use a real Tusk client** with a temp SQLite DB (created via `t.TempDir()`). SQLite is fast enough that mocking adds complexity without value.
- **Test coverage:**
  - Engine initialization with default and custom DB paths.
  - WBS level validation (all valid levels, invalid levels rejected).
  - Task creation with correct UDA fields, project assignment, and parent linkage.
  - Player registration (success, duplicate registration behavior).
  - Engine close/cleanup.
- **No integration tests** needed at this layer — no Docker, no network.

## Error Handling

- All errors wrapped with context: `fmt.Errorf("work: <operation>: %w", err)`.
- Tusk sentinel errors (`domain.ErrNotFound`, `domain.ErrConflict`, `domain.ErrInvalidTransition`, etc.) are preserved in the chain so callers can use `errors.Is()`.

## Dependencies

- `github.com/germanamz/tusk` — Tusk SDK (new dependency)
- `github.com/google/uuid` — already a transitive dependency via Tusk
