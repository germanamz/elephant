package container

import "fmt"

// SecretConfig defines which secrets are available per project.
// The special project key "*" defines global secrets available to all projects.
type SecretConfig struct {
	// Projects maps project names to their allowed secret names (env var keys).
	Projects map[string][]string
}

// ResolveSecrets returns the secret env vars for the given project.
// It merges global ("*") secrets with project-specific secrets.
// The lookup function retrieves secret values by name (e.g., os.LookupEnv).
// Returns an error if any configured secret cannot be found via lookup.
func ResolveSecrets(cfg SecretConfig, project string, lookup func(string) (string, bool)) (map[string]string, error) {
	var names []string

	if global, ok := cfg.Projects["*"]; ok {
		names = append(names, global...)
	}

	if project != "*" {
		if proj, ok := cfg.Projects[project]; ok {
			names = append(names, proj...)
		}
	}

	secrets := make(map[string]string, len(names))
	for _, name := range names {
		val, ok := lookup(name)
		if !ok {
			return nil, fmt.Errorf("resolve secrets: secret %q not found", name)
		}
		secrets[name] = val
	}

	return secrets, nil
}
