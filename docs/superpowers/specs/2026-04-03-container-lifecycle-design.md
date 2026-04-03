# Container Lifecycle — Design Spec

> Scope: v0.1 Docker Container Management initiative, Container Lifecycle story only.

## Context

Elephant orchestrates AI coding agents in isolated Docker containers. This spec covers the container lifecycle primitive — spawning containers with the codebase mounted, waiting for completion, and tearing them down. It does not cover MCP communication, agent execution logic, or the orchestrator layer.

## Decisions

These were established during brainstorming and are not open for re-evaluation in this spec:

- **Docker Engine API (Go SDK)** — programmatic control via `github.com/docker/docker/client`, no CLI wrapping.
- **Layered image strategy** — Elephant defines a base image, language environments extend it, users extend further. Elephant accepts an image reference string; image building is the user's concern.
- **Bind mounts** — the host project directory is mounted into the container. Agents output PRs/MRs rather than relying on host filesystem changes, which mitigates the security surface of bind mounts.
- **Scoped tokens** — git credentials are injected as environment variables. Other dependency credentials are the user's responsibility.
- **Centralized MCP with per-agent identity** — a single MCP server in Elephant handles all agent communication (covered in a separate initiative).

## Package Structure

All code lives in `internal/container/`.

### Types (`container.go`)

```go
type Config struct {
    Image      string            // required — user-provided image reference
    Mounts     []Mount           // bind mounts (e.g., repo -> /workspace)
    Env        map[string]string // env vars (git token, agent ID, etc.)
    WorkingDir string            // container working directory (default: /workspace)
    Cmd        []string          // command to run (default: image's entrypoint/cmd)
}

type Mount struct {
    Source string // host path
    Target string // container path
}

type Container struct {
    ID string
}

type Status string

const (
    StatusCreated Status = "created"
    StatusRunning Status = "running"
    StatusStopped Status = "stopped"
    StatusRemoved Status = "removed"
)
```

### Manager Interface (`container.go`)

```go
type Manager interface {
    Create(ctx context.Context, cfg Config) (*Container, error)
    Start(ctx context.Context, id string) error
    Stop(ctx context.Context, id string, timeout time.Duration) error
    Remove(ctx context.Context, id string) error
    Wait(ctx context.Context, id string) (<-chan int64, <-chan error)
    Status(ctx context.Context, id string) (Status, error)
}
```

`Manager` is an interface so callers (the future orchestrator, tests) can mock it. The real implementation wraps the Docker SDK client.

### DockerManager (`docker.go`)

```go
type DockerManager struct {
    client *client.Client
}

func NewDockerManager() (*DockerManager, error)
```

**Method behavior:**

- **Create** — calls `client.ContainerCreate`. Translates `Config` to Docker API types. Validates that `Image` is non-empty. Sets label `elephant.managed=true` for identification and cleanup. Does not auto-start.
- **Start** — calls `client.ContainerStart`.
- **Stop** — calls `client.ContainerStop` with the given timeout. Sends SIGTERM, then SIGKILL after timeout.
- **Remove** — calls `client.ContainerRemove` with `force: true` and `removeVolumes: true`.
- **Wait** — calls `client.ContainerWait` with `condition: not-running`. Translates the Docker SDK's `WaitResponse` into a simple exit code channel and error channel.
- **Status** — calls `client.ContainerInspect` and maps Docker's state to the `Status` type.

**Container defaults:**

- Network mode: `bridge` (agents can reach the internet for git push, package installs)
- No privileged mode, no extra capabilities
- Labels for Elephant-managed container identification

### Ephemeral Lifecycle (`lifecycle.go`)

```go
type RunResult struct {
    ExitCode int64
}

func Run(ctx context.Context, mgr Manager, cfg Config) (*RunResult, error)
```

`Run` encapsulates the full ephemeral lifecycle:

1. `Create` with the given config
2. `Start` the container
3. `Wait` for process exit (or context cancellation)
4. `Remove` the container (always — success or failure)
5. Return the exit code

**Context cancellation:** If the caller cancels the context, `Run` calls `Stop` then `Remove`. The container is always cleaned up.

**Error semantics:** Docker API errors (image not found, daemon unreachable) bubble up directly. Non-zero exit code is not an error — it is returned in `RunResult.ExitCode`. The caller decides what a non-zero exit means.

## Testing Strategy

### Unit Tests

Mock the `Manager` interface to test `Run` lifecycle logic without Docker:

- Happy path: create -> start -> wait (exit 0) -> remove
- Non-zero exit: returns exit code, still removes
- Context cancellation: stop -> remove
- Create failure: returns error, no cleanup needed
- Start failure: removes the container
- Remove failure after stop: error surfaces but does not mask the original result

### Integration Tests

Use a real Docker daemon with a lightweight image (e.g., `alpine`):

- Spawn a container that runs `echo hello` and exits 0
- Spawn a container with a bind mount, verify a file is visible inside
- Spawn a container with env vars, verify they are set
- Timeout/cancellation: spawn a `sleep 300` container, cancel context, verify cleanup

Integration tests are tagged with `//go:build integration` so they do not run in CI without Docker.

### What We Do Not Test

`DockerManager` methods are thin wrappers over the Docker SDK. They are covered by the integration tests end-to-end. Unit test effort goes to `Run` and any future orchestration logic.

## Out of Scope

- MCP server implementation (separate initiative)
- Agent execution logic / Claude Code headless invocation (Agent Execution initiative)
- Image building or Dockerfile management (user's responsibility)
- Orchestrator layer that decides when to spawn containers (future work)
