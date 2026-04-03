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
