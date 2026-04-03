# Project Scaffolding — Implementation Plan Overview

**Goal:** Set up the Go project foundation for Elephant with module, directory structure, CI/CD, code quality tooling, build/release pipeline, and community files.

**Architecture:** Standard Go project layout with `cmd/elephant/` entry point and `internal/` packages. All tooling configs (golangci-lint, lefthook, conform, goreleaser) ported from germanamz/tusk. GitHub Actions for CI and release.

**Tech Stack:** Go 1.26, Cobra (CLI), Bubbletea (TUI), GoReleaser, golangci-lint, Lefthook, Conform, GitHub Actions

---

## Phases

Execute these phases in order. Each phase has its own detailed plan document.

### Phase 1: Project Foundation
[2026-04-03-scaffolding-phase1-foundation.md](2026-04-03-scaffolding-phase1-foundation.md)

Go module initialization, cobra entry point, stub packages, .gitignore.

**Tasks:**
1. Initialize Go module and cobra entry point
2. Create stub packages (container, agent, tui)
3. Add .gitignore

### Phase 2: Code Quality & Build Tooling
[2026-04-03-scaffolding-phase2-quality-and-build.md](2026-04-03-scaffolding-phase2-quality-and-build.md)

Linter, git hooks, commit enforcement, Makefile, GoReleaser, install script.

**Tasks:**
1. Add code quality configs (golangci-lint, conform, lefthook)
2. Add Makefile
3. Add GoReleaser config
4. Add install script

### Phase 3: CI/CD & Community
[2026-04-03-scaffolding-phase3-cicd-and-community.md](2026-04-03-scaffolding-phase3-cicd-and-community.md)

GitHub Actions workflows and community files.

**Tasks:**
1. Add CI workflow
2. Add release workflow
3. Add community files (LICENSE, CODE_OF_CONDUCT, CONTRIBUTING, README)

---

## File Map (all phases)

| File | Phase | Responsibility |
|------|-------|---------------|
| `go.mod` | 1 | Module definition, dependencies |
| `cmd/elephant/main.go` | 1 | Entry point, cobra root command, version info |
| `internal/container/container.go` | 1 | Stub — Docker container lifecycle |
| `internal/agent/agent.go` | 1 | Stub — headless agent runner |
| `internal/tui/tui.go` | 1 | Stub — bubbletea TUI |
| `.gitignore` | 1 | Git ignore rules |
| `.golangci.yml` | 2 | Linter config |
| `.conform.yaml` | 2 | Conventional commit enforcement |
| `lefthook.yml` | 2 | Git hooks config |
| `Makefile` | 2 | Developer workflow targets |
| `.goreleaser.yaml` | 2 | Release build config |
| `install.sh` | 2 | Quick install script |
| `.github/workflows/ci.yml` | 3 | CI pipeline |
| `.github/workflows/release.yml` | 3 | Release pipeline |
| `LICENSE` | 3 | Apache 2.0 license |
| `CODE_OF_CONDUCT.md` | 3 | Contributor Covenant v2.1 |
| `CONTRIBUTING.md` | 3 | Development guidelines |
| `README.md` | 3 | Project overview and docs |
