package work

import (
	"context"
	"fmt"

	"github.com/germanamz/tusk/domain"
)

// RegisterAgent registers an agent as a Tusk player.
func (e *Engine) RegisterAgent(ctx context.Context, agentID string) (*domain.Player, error) {
	player, err := e.client.Players.Register(ctx, agentID, "agent")
	if err != nil {
		return nil, fmt.Errorf("work: register agent: %w", err)
	}

	return player, nil
}
