package container

import "testing"

func TestResolveSecrets_GlobalOnly(t *testing.T) {
	cfg := SecretConfig{
		Projects: map[string][]string{
			"*": {"GITHUB_TOKEN"},
		},
	}
	lookup := func(name string) (string, bool) {
		if name == "GITHUB_TOKEN" {
			return "ghp_abc123", true
		}
		return "", false
	}

	secrets, err := ResolveSecrets(cfg, "any-project", lookup)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if secrets["GITHUB_TOKEN"] != "ghp_abc123" {
		t.Fatalf("expected GITHUB_TOKEN=ghp_abc123, got %q", secrets["GITHUB_TOKEN"])
	}
	if len(secrets) != 1 {
		t.Fatalf("expected 1 secret, got %d", len(secrets))
	}
}

func TestResolveSecrets_ProjectSpecific(t *testing.T) {
	cfg := SecretConfig{
		Projects: map[string][]string{
			"frontend": {"NPM_TOKEN"},
		},
	}
	lookup := func(name string) (string, bool) {
		if name == "NPM_TOKEN" {
			return "npm_xyz", true
		}
		return "", false
	}

	secrets, err := ResolveSecrets(cfg, "frontend", lookup)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if secrets["NPM_TOKEN"] != "npm_xyz" {
		t.Fatalf("expected NPM_TOKEN=npm_xyz, got %q", secrets["NPM_TOKEN"])
	}
}

func TestResolveSecrets_MergesGlobalAndProject(t *testing.T) {
	cfg := SecretConfig{
		Projects: map[string][]string{
			"*":        {"GITHUB_TOKEN"},
			"frontend": {"NPM_TOKEN"},
		},
	}
	lookup := func(name string) (string, bool) {
		switch name {
		case "GITHUB_TOKEN":
			return "ghp_abc", true
		case "NPM_TOKEN":
			return "npm_xyz", true
		}
		return "", false
	}

	secrets, err := ResolveSecrets(cfg, "frontend", lookup)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(secrets) != 2 {
		t.Fatalf("expected 2 secrets, got %d", len(secrets))
	}
	if secrets["GITHUB_TOKEN"] != "ghp_abc" {
		t.Fatalf("expected GITHUB_TOKEN=ghp_abc, got %q", secrets["GITHUB_TOKEN"])
	}
	if secrets["NPM_TOKEN"] != "npm_xyz" {
		t.Fatalf("expected NPM_TOKEN=npm_xyz, got %q", secrets["NPM_TOKEN"])
	}
}

func TestResolveSecrets_MissingSecret(t *testing.T) {
	cfg := SecretConfig{
		Projects: map[string][]string{
			"*": {"MISSING_VAR"},
		},
	}
	lookup := func(name string) (string, bool) {
		return "", false
	}

	_, err := ResolveSecrets(cfg, "proj", lookup)
	if err == nil {
		t.Fatal("expected error for missing secret, got nil")
	}
}

func TestResolveSecrets_NoConfig(t *testing.T) {
	cfg := SecretConfig{}
	lookup := func(name string) (string, bool) {
		return "val", true
	}

	secrets, err := ResolveSecrets(cfg, "proj", lookup)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(secrets) != 0 {
		t.Fatalf("expected 0 secrets, got %d", len(secrets))
	}
}

func TestResolveSecrets_UnknownProjectGetsGlobalOnly(t *testing.T) {
	cfg := SecretConfig{
		Projects: map[string][]string{
			"*":        {"GLOBAL_SECRET"},
			"frontend": {"NPM_TOKEN"},
		},
	}
	lookup := func(name string) (string, bool) {
		if name == "GLOBAL_SECRET" {
			return "global_val", true
		}
		return "", false
	}

	secrets, err := ResolveSecrets(cfg, "backend", lookup)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(secrets) != 1 {
		t.Fatalf("expected 1 secret, got %d", len(secrets))
	}
	if secrets["GLOBAL_SECRET"] != "global_val" {
		t.Fatalf("expected GLOBAL_SECRET=global_val, got %q", secrets["GLOBAL_SECRET"])
	}
}
