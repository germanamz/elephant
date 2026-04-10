package work

import (
	"context"
	"fmt"

	"github.com/germanamz/tusk/domain"
	"github.com/google/uuid"
)

// CreateTaskParams holds the inputs for creating a WBS-shaped task.
type CreateTaskParams struct {
	Title       string
	Description string
	Level       string     // must pass ValidLevel()
	ParentID    *uuid.UUID // optional parent for hierarchy
}

// CreateTask creates a task in Tusk with the correct WBS level UDA and project assignment.
func (e *Engine) CreateTask(ctx context.Context, params CreateTaskParams) (*domain.Task, error) {
	if !ValidLevel(params.Level) {
		return nil, fmt.Errorf("work: create task: invalid WBS level %q", params.Level)
	}

	task := domain.Task{
		Title:       params.Title,
		Description: params.Description,
		ProjectID:   DefaultProject,
		ParentID:    params.ParentID,
		UDA:         map[string]any{"wbs_level": params.Level},
	}

	if err := e.client.Tasks.Create(ctx, &task); err != nil {
		return nil, fmt.Errorf("work: create task: %w", err)
	}

	return &task, nil
}
