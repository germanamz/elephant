# Project Scaffolding — Design Spec

> Initiative from ROADMAP.md v0.1 — Foundation

## Overview

Set up the Go project foundation for Elephant: module, directory structure, CI/CD, code quality tooling, build/release pipeline, and community files. All configs and patterns are ported from [germanamz/tusk](https://github.com/germanamz/tusk) for consistency across projects.

## Project Structure

```
elephant/
├── cmd/
│   └── elephant/
│       └── main.go              # Cobra root command, version via ldflags
├── internal/
│   ├── container/
│   │   └── container.go         # Stub — Docker container lifecycle
│   ├── agent/
│   │   └── agent.go             # Stub — headless agent runner, session management
│   └── tui/
│       └── tui.go               # Stub — bubbletea TUI
├── .github/
│   └── workflows/
│       ├── ci.yml               # Lint, test, build on push/PR
│       └── release.yml          # GoReleaser on release published
├── .golangci.yml
├── .conform.yaml
├── .goreleaser.yaml
├── lefthook.yml
├── Makefile
├── install.sh
├── go.mod
├── go.sum
├── .gitignore
├── README.md
├── CONTRIBUTING.md
├── CODE_OF_CONDUCT.md
├── LICENSE                      # Apache 2.0
├── PRODUCT.md                   # (existing)
└── ROADMAP.md                   # (existing)
```

## Go Module

- **Module path:** `github.com/germanamz/elephant`
- **Go version:** 1.26
- **Dependencies:**
  - `github.com/spf13/cobra` — CLI framework
  - `github.com/charmbracelet/bubbletea` — TUI framework (added now, used in later initiatives)

## Entry Point — `cmd/elephant/main.go`

Minimal cobra root command that prints version info. Version, commit, and date are injected via ldflags at build time:

```go
var (
    version = "dev"
    commit  = "none"
    date    = "unknown"
)
```

The root command prints `elephant <version> (<commit>, <date>)` when invoked with no subcommands or with `--version`.

## Stub Packages

Each stub package contains a single file with just the package declaration. These exist to establish the directory structure for later initiatives:

- **`internal/container/`** — Docker container lifecycle (v0.1: Docker Container Management)
- **`internal/agent/`** — Headless agent runner, session persistence (v0.1: Agent Execution)
- **`internal/tui/`** — Bubbletea TUI (v0.1: Basic TUI)

## CI Pipeline — `.github/workflows/ci.yml`

Triggered on push to `main` and pull requests.

**Jobs (parallel, then build):**

| Job | Runs on | Steps |
|-----|---------|-------|
| `lint` | `ubuntu-latest` | checkout, setup Go 1.26 (via `go-version-file: go.mod`), `golangci-lint-action` |
| `test` | `ubuntu-latest` | checkout, setup Go 1.26, `go test ./...` |
| `build` | `ubuntu-latest` | checkout, setup Go 1.26, `go build ./cmd/elephant/` |

`lint` and `test` run in parallel. `build` depends on both passing.

## Release Pipeline — `.github/workflows/release.yml`

Ported from tusk. Triggered on GitHub Release published.

```yaml
on:
  release:
    types: [published]

permissions:
  contents: write
```

Steps: checkout (fetch-depth: 0), setup Go (from go.mod), GoReleaser action v2 with `release --clean`.

## GoReleaser — `.goreleaser.yaml`

Ported from tusk, adapted for elephant:

- **Pre-hook:** `go mod tidy`
- **Build:** `./cmd/elephant`, binary name `elephant`, CGO disabled
- **Ldflags:** `-s -w -X main.version={{.Version}} -X main.commit={{.Commit}} -X main.date={{.Date}}`
- **Targets:** linux (amd64, arm64), darwin (amd64, arm64), windows (amd64, arm64)
- **Archives:** tar.gz (linux/darwin), zip (windows), named `{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}`
- **Checksum:** `checksums.txt`
- **Changelog:** sorted asc, grouped by Features (feat), Bug Fixes (fix), Others; excludes docs/test/ci commits
- **Release:** `germanamz/elephant`, keep-existing mode, auto prerelease

## Install Script — `install.sh`

Adapted from tusk. Shell script for quick installation:

- Detects OS (linux, darwin) and architecture (amd64, arm64)
- Fetches latest release version from GitHub API
- Downloads archive from GitHub Releases
- Extracts `elephant` binary to `INSTALL_DIR` (default: `~/.local/bin`)
- Usage: `curl -fsSL https://raw.githubusercontent.com/germanamz/elephant/main/install.sh | sh`

## Code Quality Configs

All ported from tusk without modification (except project-specific paths).

### golangci-lint — `.golangci.yml`

```yaml
version: "2"

linters:
  exclusions:
    rules:
      - linters: [errcheck]
        path: _test\.go
      - linters: [errcheck]
        text: 'Error return value of `.*\.Close` is not checked'
```

### Lefthook — `lefthook.yml`

```yaml
pre-commit:
  parallel: true
  commands:
    gofmt:
      glob: "*.go"
      run: test -z "$(gofmt -l {staged_files})"
      fail_text: "Run 'gofmt -w .' to fix formatting"
    vet:
      glob: "*.go"
      run: go vet ./...
    lint:
      glob: "*.go"
      run: golangci-lint run ./...
    test:
      glob: "*.go"
      run: go test ./internal/... ./cmd/...

commit-msg:
  commands:
    conform:
      run: conform enforce --commit-msg-file {1}
```

### Conform — `.conform.yaml`

```yaml
policies:
  - type: commit
    spec:
      header:
        length: 89
        imperative: true
        case: lower
        invalidLastCharacters: .
      body:
        required: false
      gpg:
        required: false
      conventional:
        types:
          - feat
          - fix
          - docs
          - style
          - refactor
          - perf
          - test
          - build
          - ci
          - chore
          - revert
        scopes:
          - ".*"
```

## Makefile

Adapted from tusk:

```makefile
BINARY_NAME := elephant
BUILD_DIR := bin
GO := go
GOFLAGS := -v

.PHONY: all build clean test test-race vet lint run install setup-hooks

all: build

build:
	$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/elephant

clean:
	rm -rf $(BUILD_DIR)

test:
	$(GO) test $(GOFLAGS) ./...

test-race:
	$(GO) test $(GOFLAGS) -race ./...

vet:
	$(GO) vet ./...

lint:
	golangci-lint run ./...

run: build
	./$(BUILD_DIR)/$(BINARY_NAME)

install:
	$(GO) install $(GOFLAGS) ./cmd/elephant

setup-hooks:
	go install github.com/evilmartians/lefthook@latest
	go install github.com/siderolabs/conform/cmd/conform@latest
	lefthook install
	@echo "Git hooks installed via lefthook"
```

## Community Files

### README.md

Adapted for Elephant:
- Project description (AI orchestration platform)
- Installation section (install.sh, from source)
- Development section (make targets, setup-hooks)
- Links to PRODUCT.md and ROADMAP.md
- License (Apache 2.0)

### CONTRIBUTING.md

Adapted for Elephant:
- Prerequisites: Go 1.26+, golangci-lint
- Setup via `make setup-hooks`
- Development workflow (fork, branch, test, PR)
- Commit conventions (conventional commits)
- Architecture overview (`internal/container`, `internal/agent`, `internal/tui`)

### CODE_OF_CONDUCT.md

Contributor Covenant v2.1 — same as tusk.

### LICENSE

Apache 2.0 — same as tusk.

### .gitignore

Standard Go ignores plus project-specific:
- `bin/` build output
- OS files (`.DS_Store`, `Thumbs.db`)
- IDE files (`.idea/`, `.vscode/`, `*.swp`)
- Go build artifacts
