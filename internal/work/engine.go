package work

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/germanamz/tusk"
	"github.com/germanamz/tusk/config"
	"github.com/germanamz/tusk/service"
)

const (
	DefaultDBPath   = "~/.local/share/elephant/tusk.db"
	DefaultWorkflow = "wbs"
	DefaultProject  = "default"
)

// Config holds the configuration for creating a new Engine.
type Config struct {
	// DBPath is the path to the SQLite database file.
	// If empty, defaults to DefaultDBPath.
	DBPath string
}

// Engine manages work breakdown structure tasks via a Tusk client.
type Engine struct {
	client *tusk.Client
}

// NewEngine creates a new Engine backed by a Tusk client.
func NewEngine(cfg Config) (*Engine, error) {
	dbPath := cfg.DBPath
	if dbPath == "" {
		dbPath = DefaultDBPath
	}

	if strings.HasPrefix(dbPath, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("work: new engine: %w", err)
		}
		dbPath = filepath.Join(home, dbPath[2:])
	}

	tuskCfg := tusk.Config{
		DBPath: dbPath,
		Workflows: map[string]config.WorkflowConfig{
			DefaultWorkflow: {
				Statuses: []string{"pending", "active", "completed", "deleted"},
				Transitions: []config.WorkflowTransitionConfig{
					{From: "pending", To: "active"},
					{From: "active", To: "completed"},
					{From: "completed", To: "pending"},
					{From: "pending", To: "deleted"},
					{From: "active", To: "deleted"},
					{From: "completed", To: "deleted"},
				},
			},
		},
		Projects: map[string]config.ProjectConfig{
			DefaultProject: {
				Workflow: DefaultWorkflow,
			},
		},
	}

	client, err := tusk.NewClient(tuskCfg)
	if err != nil {
		return nil, fmt.Errorf("work: new engine: %w", err)
	}

	return &Engine{client: client}, nil
}

// Tasks returns the underlying Tusk TaskService for operations
// that don't need Elephant-specific wrapping (claim, complete, pop, annotate, etc.).
func (e *Engine) Tasks() *service.TaskService { return e.client.Tasks }

// Players returns the underlying Tusk PlayerService.
func (e *Engine) Players() *service.PlayerService { return e.client.Players }

// Relations returns the underlying Tusk RelationService.
func (e *Engine) Relations() *service.RelationService { return e.client.Relations }

// Close releases the underlying Tusk client and database connection.
func (e *Engine) Close() error {
	if err := e.client.Close(); err != nil {
		return fmt.Errorf("work: close: %w", err)
	}
	return nil
}
