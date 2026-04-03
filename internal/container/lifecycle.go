package container

import (
	"context"
	"fmt"
	"time"
)

// RunResult holds the outcome of an ephemeral container execution.
type RunResult struct {
	// ExitCode is the exit code of the container's main process.
	// A non-zero exit code is not treated as an error — the caller
	// decides what it means.
	ExitCode int64
}

// stopTimeout is how long Stop waits (SIGTERM → SIGKILL) when the context
// is cancelled.
const stopTimeout = 10 * time.Second

// Run executes the full ephemeral container lifecycle: create, start, wait
// for exit, and remove. The container is always removed, even on failure.
// If ctx is cancelled, the container is stopped then removed.
func Run(ctx context.Context, mgr Manager, cfg Config) (result *RunResult, retErr error) {
	ctr, err := mgr.Create(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("run: %w", err)
	}

	// From this point on, always remove the container.
	defer func() {
		removeErr := mgr.Remove(context.WithoutCancel(ctx), ctr.ID)
		if removeErr != nil && retErr == nil {
			retErr = fmt.Errorf("run: container finished but remove failed: %w", removeErr)
		}
	}()

	if err := mgr.Start(ctx, ctr.ID); err != nil {
		return nil, fmt.Errorf("run: %w", err)
	}

	exitCh, waitErrCh := mgr.Wait(ctx, ctr.ID)

	select {
	case code := <-exitCh:
		return &RunResult{ExitCode: code}, nil
	case err := <-waitErrCh:
		return nil, fmt.Errorf("run: %w", err)
	case <-ctx.Done():
		// Context cancelled — stop the container before removal.
		_ = mgr.Stop(context.WithoutCancel(ctx), ctr.ID, stopTimeout)
		return nil, fmt.Errorf("run: %w", ctx.Err())
	}
}
