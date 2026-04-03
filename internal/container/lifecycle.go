package container

import (
	"context"
	"errors"
)

// RunResult holds the outcome of an ephemeral container execution.
type RunResult struct {
	// ExitCode is the exit code of the container's main process.
	// A non-zero exit code is not treated as an error — the caller
	// decides what it means.
	ExitCode int64
}

// Run executes the full ephemeral container lifecycle: create, start, wait
// for exit, and remove. The container is always removed, even on failure.
// If ctx is cancelled, the container is stopped then removed.
func Run(ctx context.Context, mgr Manager, cfg Config) (*RunResult, error) {
	return nil, errors.New("not implemented")
}
