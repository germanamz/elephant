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
