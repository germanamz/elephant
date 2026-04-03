package container

import (
	"context"
	"fmt"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
)

const (
	defaultWorkingDir = "/workspace"
	labelManaged      = "elephant.managed"
)

// Compile-time check that DockerManager implements Manager.
var _ Manager = (*DockerManager)(nil)

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

// Close closes the underlying Docker client connection.
func (m *DockerManager) Close() error {
	return m.client.Close()
}

// Create creates a new container from the given config but does not start it.
func (m *DockerManager) Create(ctx context.Context, cfg Config) (*Container, error) {
	if cfg.Image == "" {
		return nil, fmt.Errorf("create container: image is required")
	}

	workingDir := cfg.WorkingDir
	if workingDir == "" {
		workingDir = defaultWorkingDir
	}

	env := make([]string, 0, len(cfg.Env))
	for k, v := range cfg.Env {
		env = append(env, k+"="+v)
	}

	mounts := make([]mount.Mount, len(cfg.Mounts))
	for i, mt := range cfg.Mounts {
		mounts[i] = mount.Mount{
			Type:   mount.TypeBind,
			Source: mt.Source,
			Target: mt.Target,
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
		Mounts:      mounts,
		NetworkMode: "bridge",
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
