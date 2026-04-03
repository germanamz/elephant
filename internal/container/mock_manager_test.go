package container

import (
	"context"
	"time"
)

// call records a method invocation on mockManager.
type call struct {
	Method string
	ID     string
}

// mockManager is a test double for the Manager interface.
// Each method field can be set to control behavior. If nil, the method
// returns zero values. All calls are recorded in the Calls slice.
type mockManager struct {
	Calls []call

	OnCreate func(ctx context.Context, cfg Config) (*Container, error)
	OnStart  func(ctx context.Context, id string) error
	OnStop   func(ctx context.Context, id string, timeout time.Duration) error
	OnRemove func(ctx context.Context, id string) error
	OnWait   func(ctx context.Context, id string) (<-chan int64, <-chan error)
	OnStatus func(ctx context.Context, id string) (Status, error)
}

func (m *mockManager) Create(ctx context.Context, cfg Config) (*Container, error) {
	m.Calls = append(m.Calls, call{Method: "Create"})
	if m.OnCreate != nil {
		return m.OnCreate(ctx, cfg)
	}
	return &Container{ID: "mock-id"}, nil
}

func (m *mockManager) Start(ctx context.Context, id string) error {
	m.Calls = append(m.Calls, call{Method: "Start", ID: id})
	if m.OnStart != nil {
		return m.OnStart(ctx, id)
	}
	return nil
}

func (m *mockManager) Stop(ctx context.Context, id string, timeout time.Duration) error {
	m.Calls = append(m.Calls, call{Method: "Stop", ID: id})
	if m.OnStop != nil {
		return m.OnStop(ctx, id, timeout)
	}
	return nil
}

func (m *mockManager) Remove(ctx context.Context, id string) error {
	m.Calls = append(m.Calls, call{Method: "Remove", ID: id})
	if m.OnRemove != nil {
		return m.OnRemove(ctx, id)
	}
	return nil
}

func (m *mockManager) Wait(ctx context.Context, id string) (<-chan int64, <-chan error) {
	m.Calls = append(m.Calls, call{Method: "Wait", ID: id})
	if m.OnWait != nil {
		return m.OnWait(ctx, id)
	}
	exitCh := make(chan int64, 1)
	errCh := make(chan error, 1)
	exitCh <- 0
	return exitCh, errCh
}

func (m *mockManager) Status(ctx context.Context, id string) (Status, error) {
	m.Calls = append(m.Calls, call{Method: "Status", ID: id})
	if m.OnStatus != nil {
		return m.OnStatus(ctx, id)
	}
	return StatusRunning, nil
}
