package container

import (
	"context"
	"fmt"
)

// Standby manages a long-lived container that persists beyond initial task
// completion. Unlike Run, the container is not automatically removed — it stays
// alive for review feedback and is torn down only when explicitly requested.
type Standby struct {
	mgr    Manager
	ctr    *Container
	exitCh <-chan int64
	errCh  <-chan error
}

// StartStandby creates and starts a container that remains running until
// explicitly torn down via Teardown. If Start fails, the created container is
// removed before returning the error.
func StartStandby(ctx context.Context, mgr Manager, cfg Config) (*Standby, error) {
	ctr, err := mgr.Create(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("standby: %w", err)
	}

	if err := mgr.Start(ctx, ctr.ID); err != nil {
		_ = mgr.Remove(context.WithoutCancel(ctx), ctr.ID)
		return nil, fmt.Errorf("standby: %w", err)
	}

	exitCh, errCh := mgr.Wait(ctx, ctr.ID)

	return &Standby{
		mgr:    mgr,
		ctr:    ctr,
		exitCh: exitCh,
		errCh:  errCh,
	}, nil
}

// ID returns the Docker container ID.
func (s *Standby) ID() string {
	return s.ctr.ID
}

// Teardown stops and removes the container. Stop errors are ignored because
// the container may have already exited.
func (s *Standby) Teardown(ctx context.Context) error {
	_ = s.mgr.Stop(ctx, s.ctr.ID, stopTimeout)

	if err := s.mgr.Remove(ctx, s.ctr.ID); err != nil {
		return fmt.Errorf("standby teardown: %w", err)
	}

	return nil
}

// Wait blocks until the container exits on its own and returns the exit result.
// This does not remove the container — call Teardown separately if needed.
func (s *Standby) Wait(ctx context.Context) (*RunResult, error) {
	select {
	case code := <-s.exitCh:
		return &RunResult{ExitCode: code}, nil
	case err := <-s.errCh:
		return nil, fmt.Errorf("standby wait: %w", err)
	case <-ctx.Done():
		return nil, fmt.Errorf("standby wait: %w", ctx.Err())
	}
}
