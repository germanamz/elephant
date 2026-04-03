# Container Lifecycle Phase 1: Types, Interface & DockerManager

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Define the container package's public API (types + interface) and implement the `DockerManager` that wraps the Docker Engine SDK.

**Architecture:** The `internal/container` package exposes a `Manager` interface for container lifecycle operations. `DockerManager` is the concrete implementation wrapping `github.com/docker/docker/client`. All Docker SDK types are translated to Elephant's own types at the package boundary.

**Tech Stack:** Go 1.26.1, Docker Engine API Go SDK v28.5.2 (`github.com/docker/docker`)

**Reference spec:** `docs/superpowers/specs/2026-04-03-container-lifecycle-design.md`

---

### Task 1: Add Docker SDK Dependency

**Files:**
- Modify: `go.mod`
- Modify: `go.sum`

- [ ] **Step 1: Add the Docker SDK module**

Run:

```bash
go get github.com/docker/docker@v28.5.2
```

This adds the Docker Engine API client library to the project.

- [ ] **Step 2: Verify the dependency was added**

Run:

```bash
grep 'docker/docker' go.mod
```

Expected output should contain:

```
github.com/docker/docker v28.5.2
```

- [ ] **Step 3: Verify the project still builds**

Run:

```bash
make build
```

Expected: build succeeds with no errors.

- [ ] **Step 4: Commit**

```bash
git add go.mod go.sum
git commit -m "build: add docker engine api sdk dependency"
```

---

### Task 2: Define Types and Manager Interface

**Files:**
- Create: `internal/container/container.go` (replace existing stub)

This file defines all the public types and the `Manager` interface. No implementation yet — just the contract.

- [ ] **Step 1: Write the types and interface**

Replace the contents of `internal/container/container.go` with:

```go
package container

import (
	"context"
	"time"
)

// Config holds the configuration for creating a new container.
type Config struct {
	// Image is the Docker image to use (e.g., "elephant-base:latest").
	// Required.
	Image string

	// Mounts defines bind mounts from the host into the container.
	Mounts []Mount

	// Env holds environment variables to inject into the container
	// as key-value pairs (e.g., {"GIT_TOKEN": "ghp_xxx"}).
	Env map[string]string

	// WorkingDir is the working directory inside the container.
	// Defaults to "/workspace" if empty.
	WorkingDir string

	// Cmd is the command to run inside the container.
	// If empty, the image's default entrypoint/cmd is used.
	Cmd []string
}

// Mount defines a bind mount from the host filesystem into the container.
type Mount struct {
	// Source is the absolute path on the host (e.g., "/home/user/project").
	Source string

	// Target is the path inside the container (e.g., "/workspace").
	Target string
}

// Container represents a managed Docker container.
type Container struct {
	// ID is the Docker container ID.
	ID string
}

// Status represents the lifecycle state of a container.
type Status string

const (
	StatusCreated Status = "created"
	StatusRunning Status = "running"
	StatusStopped Status = "stopped"
	StatusRemoved Status = "removed"
)

// Manager defines the interface for container lifecycle operations.
// Implementations wrap a container runtime (e.g., Docker Engine API).
type Manager interface {
	// Create creates a new container from the given config but does not start it.
	Create(ctx context.Context, cfg Config) (*Container, error)

	// Start starts a previously created container.
	Start(ctx context.Context, id string) error

	// Stop stops a running container. It sends SIGTERM and waits up to the
	// given timeout before sending SIGKILL.
	Stop(ctx context.Context, id string, timeout time.Duration) error

	// Remove removes a container. The container is forcefully removed even if
	// it is still running.
	Remove(ctx context.Context, id string) error

	// Wait blocks until the container exits and returns a channel that
	// receives the exit code and a channel that receives any error.
	Wait(ctx context.Context, id string) (<-chan int64, <-chan error)

	// Status returns the current lifecycle status of a container.
	Status(ctx context.Context, id string) (Status, error)
}
```

- [ ] **Step 2: Verify the file compiles**

Run:

```bash
go vet ./internal/container/...
```

Expected: no errors.

- [ ] **Step 3: Run the linter**

Run:

```bash
make lint
```

Expected: no errors.

- [ ] **Step 4: Commit**

```bash
git add internal/container/container.go
git commit -m "feat(container): define types and manager interface"
```

---

### Task 3: Implement DockerManager — Create, Start, Remove, Status

**Files:**
- Create: `internal/container/docker.go`

This file implements the `DockerManager` struct and four of its six methods. The remaining two (`Stop`, `Wait`) are in Task 4.

- [ ] **Step 1: Implement NewDockerManager, Create, Start, Remove, and Status**

Create `internal/container/docker.go` with:

```go
package container

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
)

const (
	defaultWorkingDir = "/workspace"
	labelManaged      = "elephant.managed"
	labelContainerID  = "elephant.container-id"
)

// DockerManager implements Manager using the Docker Engine API.
type DockerManager struct {
	client *client.Client
}

// NewDockerManager creates a new DockerManager connected to the local Docker daemon.
// It uses the DOCKER_HOST environment variable if set, otherwise the default socket.
func NewDockerManager() (*DockerManager, error) {
	c, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("create docker client: %w", err)
	}

	return &DockerManager{client: c}, nil
}

// Create creates a new container from the given config but does not start it.
func (m *DockerManager) Create(ctx context.Context, cfg Config) (*Container, error) {
	workingDir := cfg.WorkingDir
	if workingDir == "" {
		workingDir = defaultWorkingDir
	}

	env := make([]string, 0, len(cfg.Env))
	for k, v := range cfg.Env {
		env = append(env, k+"="+v)
	}

	mounts := make([]mount.Mount, len(cfg.Mounts))
	for i, m := range cfg.Mounts {
		mounts[i] = mount.Mount{
			Type:   mount.TypeBind,
			Source: m.Source,
			Target: m.Target,
		}
	}

	labels := map[string]string{
		labelManaged: "true",
	}

	containerCfg := &container.Config{
		Image:      cfg.Image,
		Env:        env,
		WorkingDir: workingDir,
		Labels:     labels,
		Cmd:        cfg.Cmd,
	}

	hostCfg := &container.HostConfig{
		Mounts: mounts,
	}

	resp, err := m.client.ContainerCreate(ctx, containerCfg, hostCfg, nil, nil, "")
	if err != nil {
		return nil, fmt.Errorf("create container: %w", err)
	}

	return &Container{ID: resp.ID}, nil
}

// Start starts a previously created container.
func (m *DockerManager) Start(ctx context.Context, id string) error {
	if err := m.client.ContainerStart(ctx, id, container.StartOptions{}); err != nil {
		return fmt.Errorf("start container %s: %w", id, err)
	}

	return nil
}

// Remove forcefully removes a container and its anonymous volumes.
func (m *DockerManager) Remove(ctx context.Context, id string) error {
	err := m.client.ContainerRemove(ctx, id, container.RemoveOptions{
		Force:         true,
		RemoveVolumes: true,
	})
	if err != nil {
		return fmt.Errorf("remove container %s: %w", id, err)
	}

	return nil
}

// Status returns the current lifecycle status of a container.
func (m *DockerManager) Status(ctx context.Context, id string) (Status, error) {
	info, err := m.client.ContainerInspect(ctx, id)
	if err != nil {
		return "", fmt.Errorf("inspect container %s: %w", id, err)
	}

	switch info.State.Status {
	case "created":
		return StatusCreated, nil
	case "running":
		return StatusRunning, nil
	case "exited", "dead":
		return StatusStopped, nil
	default:
		return Status(info.State.Status), nil
	}
}
```

- [ ] **Step 2: Verify the file compiles**

Run:

```bash
go vet ./internal/container/...
```

Expected: no errors.

- [ ] **Step 3: Run the linter**

Run:

```bash
make lint
```

Expected: no errors.

- [ ] **Step 4: Commit**

```bash
git add internal/container/docker.go
git commit -m "feat(container): implement DockerManager create, start, remove, status"
```

---

### Task 4: Implement DockerManager — Stop and Wait

**Files:**
- Modify: `internal/container/docker.go`

Add the remaining two methods to `DockerManager`.

- [ ] **Step 1: Add the Stop and Wait methods**

Add the following to the end of `internal/container/docker.go`:

```go
// Stop stops a running container. It sends SIGTERM and waits up to the given
// timeout before sending SIGKILL.
func (m *DockerManager) Stop(ctx context.Context, id string, timeout time.Duration) error {
	timeoutSeconds := int(timeout.Seconds())

	err := m.client.ContainerStop(ctx, id, container.StopOptions{
		Timeout: &timeoutSeconds,
	})
	if err != nil {
		return fmt.Errorf("stop container %s: %w", id, err)
	}

	return nil
}

// Wait blocks until the container stops and returns channels for the exit code
// and any error. The exit code channel receives exactly one value. The error
// channel receives a value only if the wait itself fails.
func (m *DockerManager) Wait(ctx context.Context, id string) (<-chan int64, <-chan error) {
	exitCh := make(chan int64, 1)
	errCh := make(chan error, 1)

	statusCh, dockerErrCh := m.client.ContainerWait(ctx, id, container.WaitConditionNotRunning)

	go func() {
		defer close(exitCh)
		defer close(errCh)

		select {
		case status := <-statusCh:
			if status.Error != nil {
				errCh <- fmt.Errorf("container wait error: %s", status.Error.Message)
				return
			}
			exitCh <- status.StatusCode
		case err := <-dockerErrCh:
			errCh <- fmt.Errorf("wait for container %s: %w", id, err)
		case <-ctx.Done():
			errCh <- ctx.Err()
		}
	}()

	return exitCh, errCh
}
```

- [ ] **Step 2: Add the `time` import**

The `time` import is needed for `Stop`. Add `"time"` to the import block at the top of `internal/container/docker.go`:

```go
import (
	"context"
	"fmt"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
)
```

- [ ] **Step 3: Verify the file compiles**

Run:

```bash
go vet ./internal/container/...
```

Expected: no errors.

- [ ] **Step 4: Verify the interface is satisfied**

Add a compile-time interface check. Add this line at the top of `internal/container/docker.go`, right after the `const` block:

```go
// Compile-time check that DockerManager implements Manager.
var _ Manager = (*DockerManager)(nil)
```

Run:

```bash
go vet ./internal/container/...
```

Expected: no errors (confirms `DockerManager` satisfies the `Manager` interface).

- [ ] **Step 5: Run the linter**

Run:

```bash
make lint
```

Expected: no errors.

- [ ] **Step 6: Commit**

```bash
git add internal/container/docker.go
git commit -m "feat(container): implement DockerManager stop and wait"
```
