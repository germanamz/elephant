package container

import "testing"

func TestProvision_InjectsTokenAndEndpoint(t *testing.T) {
	params := ProvisionParams{
		Base:        Config{Image: "agent:latest"},
		MCPEndpoint: "http://host.docker.internal:9090",
		Project:     "myproj",
		Secrets:     SecretConfig{},
		LookupSecret: func(name string) (string, bool) {
			return "", false
		},
	}

	cfg, token, err := Provision(params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if token == "" {
		t.Fatal("expected non-empty token")
	}
	if len(token) != 64 {
		t.Fatalf("expected 64-char token, got %d", len(token))
	}

	if cfg.Env[EnvAgentToken] != token {
		t.Fatalf("expected env %s=%s, got %q", EnvAgentToken, token, cfg.Env[EnvAgentToken])
	}
	if cfg.Env[EnvMCPEndpoint] != "http://host.docker.internal:9090" {
		t.Fatalf("expected env %s=http://host.docker.internal:9090, got %q", EnvMCPEndpoint, cfg.Env[EnvMCPEndpoint])
	}
}

func TestProvision_InjectsSecrets(t *testing.T) {
	params := ProvisionParams{
		Base:    Config{Image: "agent:latest"},
		Project: "frontend",
		Secrets: SecretConfig{
			Projects: map[string][]string{
				"*":        {"GITHUB_TOKEN"},
				"frontend": {"NPM_TOKEN"},
			},
		},
		LookupSecret: func(name string) (string, bool) {
			switch name {
			case "GITHUB_TOKEN":
				return "ghp_abc", true
			case "NPM_TOKEN":
				return "npm_xyz", true
			}
			return "", false
		},
	}

	cfg, _, err := Provision(params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Env["GITHUB_TOKEN"] != "ghp_abc" {
		t.Fatalf("expected GITHUB_TOKEN=ghp_abc, got %q", cfg.Env["GITHUB_TOKEN"])
	}
	if cfg.Env["NPM_TOKEN"] != "npm_xyz" {
		t.Fatalf("expected NPM_TOKEN=npm_xyz, got %q", cfg.Env["NPM_TOKEN"])
	}
}

func TestProvision_PreservesBaseEnv(t *testing.T) {
	params := ProvisionParams{
		Base: Config{
			Image: "agent:latest",
			Env:   map[string]string{"EXISTING": "value"},
		},
		MCPEndpoint: "http://localhost:9090",
		Project:     "proj",
		Secrets:     SecretConfig{},
		LookupSecret: func(name string) (string, bool) {
			return "", false
		},
	}

	cfg, _, err := Provision(params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Env["EXISTING"] != "value" {
		t.Fatalf("expected EXISTING=value preserved, got %q", cfg.Env["EXISTING"])
	}
	if cfg.Env[EnvAgentToken] == "" {
		t.Fatal("expected agent token to be injected")
	}
}

func TestProvision_PreservesBaseConfig(t *testing.T) {
	params := ProvisionParams{
		Base: Config{
			Image:      "agent:latest",
			WorkingDir: "/custom",
			Cmd:        []string{"bash"},
			Mounts:     []Mount{{Source: "/host", Target: "/ctr"}},
		},
		MCPEndpoint: "http://localhost:9090",
		Project:     "proj",
		Secrets:     SecretConfig{},
		LookupSecret: func(name string) (string, bool) {
			return "", false
		},
	}

	cfg, _, err := Provision(params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Image != "agent:latest" {
		t.Fatalf("expected image agent:latest, got %s", cfg.Image)
	}
	if cfg.WorkingDir != "/custom" {
		t.Fatalf("expected WorkingDir /custom, got %s", cfg.WorkingDir)
	}
	if len(cfg.Cmd) != 1 || cfg.Cmd[0] != "bash" {
		t.Fatalf("expected Cmd [bash], got %v", cfg.Cmd)
	}
	if len(cfg.Mounts) != 1 || cfg.Mounts[0].Source != "/host" {
		t.Fatalf("expected mount preserved, got %v", cfg.Mounts)
	}
}

func TestProvision_DoesNotMutateBaseEnv(t *testing.T) {
	base := Config{
		Image: "agent:latest",
		Env:   map[string]string{"EXISTING": "value"},
	}
	params := ProvisionParams{
		Base:        base,
		MCPEndpoint: "http://localhost:9090",
		Project:     "proj",
		Secrets:     SecretConfig{},
		LookupSecret: func(name string) (string, bool) {
			return "", false
		},
	}

	_, _, err := Provision(params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := base.Env[EnvAgentToken]; ok {
		t.Fatal("Provision mutated the caller's Base.Env map")
	}
	if len(base.Env) != 1 {
		t.Fatalf("expected Base.Env to still have 1 entry, got %d", len(base.Env))
	}
}

func TestProvision_ReservedEnvVarsCannotBeOverwritten(t *testing.T) {
	params := ProvisionParams{
		Base:        Config{Image: "agent:latest"},
		MCPEndpoint: "http://localhost:9090",
		Project:     "proj",
		Secrets: SecretConfig{
			Projects: map[string][]string{
				"proj": {EnvAgentToken, EnvMCPEndpoint},
			},
		},
		LookupSecret: func(name string) (string, bool) {
			return "malicious-value", true
		},
	}

	cfg, token, err := Provision(params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Env[EnvAgentToken] != token {
		t.Fatalf("expected %s to be the generated token, got %q", EnvAgentToken, cfg.Env[EnvAgentToken])
	}
	if cfg.Env[EnvMCPEndpoint] != "http://localhost:9090" {
		t.Fatalf("expected %s=http://localhost:9090, got %q", EnvMCPEndpoint, cfg.Env[EnvMCPEndpoint])
	}
}

func TestProvision_FailsOnMissingSecret(t *testing.T) {
	params := ProvisionParams{
		Base:    Config{Image: "agent:latest"},
		Project: "proj",
		Secrets: SecretConfig{
			Projects: map[string][]string{
				"proj": {"MISSING"},
			},
		},
		LookupSecret: func(name string) (string, bool) {
			return "", false
		},
	}

	_, _, err := Provision(params)
	if err == nil {
		t.Fatal("expected error for missing secret, got nil")
	}
}
