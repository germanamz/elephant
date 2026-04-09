package container

import "fmt"

const (
	// EnvAgentToken is the env var name for the agent's auth token.
	EnvAgentToken = "ELEPHANT_AGENT_TOKEN"

	// EnvMCPEndpoint is the env var name for the Elephant MCP server address.
	EnvMCPEndpoint = "ELEPHANT_MCP_ENDPOINT"
)

// ProvisionParams holds the inputs needed to provision a container for an agent.
type ProvisionParams struct {
	// Base is the base container config (image, mounts, working dir, cmd).
	Base Config

	// MCPEndpoint is the network address of Elephant's MCP server.
	MCPEndpoint string

	// Project is the project name, used to scope secret injection.
	Project string

	// Secrets defines which secrets are available per project.
	Secrets SecretConfig

	// LookupSecret retrieves a secret value by name.
	// In production, pass os.LookupEnv.
	LookupSecret func(string) (string, bool)
}

// Provision creates a container Config with auth token, scoped secrets, and MCP
// endpoint injected as environment variables. It returns the populated Config and
// the generated token (so the caller can store it for later MCP auth validation).
func Provision(params ProvisionParams) (Config, string, error) {
	token, err := GenerateToken()
	if err != nil {
		return Config{}, "", fmt.Errorf("provision: %w", err)
	}

	secrets, err := ResolveSecrets(params.Secrets, params.Project, params.LookupSecret)
	if err != nil {
		return Config{}, "", fmt.Errorf("provision: %w", err)
	}

	cfg := params.Base

	env := make(map[string]string, len(params.Base.Env)+len(secrets)+2)
	for k, v := range params.Base.Env {
		env[k] = v
	}
	for k, v := range secrets {
		env[k] = v
	}

	// Write reserved vars last so secrets cannot overwrite them.
	env[EnvAgentToken] = token
	env[EnvMCPEndpoint] = params.MCPEndpoint
	cfg.Env = env

	return cfg, token, nil
}
