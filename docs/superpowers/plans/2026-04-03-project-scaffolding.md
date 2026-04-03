# Project Scaffolding Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Set up the Go project foundation for Elephant with module, directory structure, CI/CD, code quality tooling, build/release pipeline, and community files.

**Architecture:** Standard Go project layout with `cmd/elephant/` entry point and `internal/` packages. All tooling configs (golangci-lint, lefthook, conform, goreleaser) ported from germanamz/tusk. GitHub Actions for CI and release.

**Tech Stack:** Go 1.26, Cobra (CLI), Bubbletea (TUI), GoReleaser, golangci-lint, Lefthook, Conform, GitHub Actions

---

## File Map

| File | Responsibility |
|------|---------------|
| `go.mod` | Module definition, dependencies |
| `cmd/elephant/main.go` | Entry point, cobra root command, version info |
| `internal/container/container.go` | Stub — Docker container lifecycle |
| `internal/agent/agent.go` | Stub — headless agent runner |
| `internal/tui/tui.go` | Stub — bubbletea TUI |
| `.gitignore` | Git ignore rules |
| `.golangci.yml` | Linter config |
| `.conform.yaml` | Conventional commit enforcement |
| `lefthook.yml` | Git hooks config |
| `.goreleaser.yaml` | Release build config |
| `Makefile` | Developer workflow targets |
| `.github/workflows/ci.yml` | CI pipeline |
| `.github/workflows/release.yml` | Release pipeline |
| `install.sh` | Quick install script |
| `LICENSE` | Apache 2.0 license |
| `CODE_OF_CONDUCT.md` | Contributor Covenant v2.1 |
| `CONTRIBUTING.md` | Development guidelines |
| `README.md` | Project overview and docs |

---

### Task 1: Initialize Go Module and Entry Point

**Files:**
- Create: `go.mod`
- Create: `cmd/elephant/main.go`

- [ ] **Step 1: Initialize the Go module**

Run:
```bash
cd /Users/germanamz/projects/elephant
go mod init github.com/germanamz/elephant
```

- [ ] **Step 2: Add cobra and bubbletea dependencies**

Run:
```bash
go get github.com/spf13/cobra@latest
go get github.com/charmbracelet/bubbletea@latest
```

- [ ] **Step 3: Create `cmd/elephant/main.go`**

```go
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	rootCmd := &cobra.Command{
		Use:     "elephant",
		Short:   "AI orchestration platform",
		Version: fmt.Sprintf("%s (%s, %s)", version, commit, date),
	}

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
```

- [ ] **Step 4: Tidy and verify it compiles**

Run:
```bash
go mod tidy
go build ./cmd/elephant/
```

Expected: compiles with no errors, produces `elephant` binary in current directory.

- [ ] **Step 5: Clean up and commit**

Run:
```bash
rm -f elephant
git add go.mod go.sum cmd/elephant/main.go
git commit -m "feat: initialize go module and cobra entry point"
```

---

### Task 2: Create Stub Packages

**Files:**
- Create: `internal/container/container.go`
- Create: `internal/agent/agent.go`
- Create: `internal/tui/tui.go`

- [ ] **Step 1: Create `internal/container/container.go`**

```go
package container
```

- [ ] **Step 2: Create `internal/agent/agent.go`**

```go
package agent
```

- [ ] **Step 3: Create `internal/tui/tui.go`**

```go
package tui
```

- [ ] **Step 4: Verify the project still compiles**

Run:
```bash
go build ./...
```

Expected: compiles with no errors.

- [ ] **Step 5: Commit**

Run:
```bash
git add internal/
git commit -m "feat: add stub packages for container, agent, and tui"
```

---

### Task 3: Add .gitignore

**Files:**
- Create: `.gitignore`

- [ ] **Step 1: Create `.gitignore`**

```gitignore
# Build output
bin/

# Go
*.exe
*.exe~
*.dll
*.so
*.dylib
*.test
*.out

# IDE
.idea/
.vscode/
*.swp
*.swo
*~

# OS
.DS_Store
Thumbs.db

# Dist
dist/
```

- [ ] **Step 2: Commit**

Run:
```bash
git add .gitignore
git commit -m "chore: add gitignore"
```

---

### Task 4: Add Code Quality Configs

**Files:**
- Create: `.golangci.yml`
- Create: `.conform.yaml`
- Create: `lefthook.yml`

- [ ] **Step 1: Create `.golangci.yml`**

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

- [ ] **Step 2: Create `.conform.yaml`**

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

- [ ] **Step 3: Create `lefthook.yml`**

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

- [ ] **Step 4: Verify linter runs cleanly**

Run:
```bash
golangci-lint run ./...
```

Expected: no issues found (or golangci-lint not installed — that's fine, CI will run it).

- [ ] **Step 5: Commit**

Run:
```bash
git add .golangci.yml .conform.yaml lefthook.yml
git commit -m "chore: add golangci-lint, conform, and lefthook configs"
```

---

### Task 5: Add Makefile

**Files:**
- Create: `Makefile`

- [ ] **Step 1: Create `Makefile`**

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

- [ ] **Step 2: Verify `make build` works**

Run:
```bash
make build
```

Expected: compiles and produces `bin/elephant`.

- [ ] **Step 3: Verify `make test` works**

Run:
```bash
make test
```

Expected: passes (no tests yet, but no errors).

- [ ] **Step 4: Clean up and commit**

Run:
```bash
make clean
git add Makefile
git commit -m "build: add makefile with dev workflow targets"
```

---

### Task 6: Add GoReleaser Config

**Files:**
- Create: `.goreleaser.yaml`

- [ ] **Step 1: Create `.goreleaser.yaml`**

```yaml
version: 2

before:
  hooks:
    - go mod tidy

builds:
  - main: ./cmd/elephant
    binary: elephant
    env:
      - CGO_ENABLED=0
    ldflags:
      - -s -w
      - -X main.version={{.Version}}
      - -X main.commit={{.Commit}}
      - -X main.date={{.Date}}
    goos:
      - linux
      - darwin
      - windows
    goarch:
      - amd64
      - arm64

archives:
  - formats:
      - tar.gz
    name_template: >-
      {{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}
    format_overrides:
      - goos: windows
        formats:
          - zip

checksum:
  name_template: checksums.txt

changelog:
  sort: asc
  filters:
    exclude:
      - "^docs:"
      - "^test:"
      - "^ci:"
  groups:
    - title: Features
      regexp: '^feat'
    - title: Bug Fixes
      regexp: '^fix'
    - title: Others
      order: 999

release:
  github:
    owner: germanamz
    name: elephant
  mode: keep-existing
  prerelease: auto
```

- [ ] **Step 2: Verify config is valid (if goreleaser is installed)**

Run:
```bash
goreleaser check 2>/dev/null || echo "goreleaser not installed locally — CI will validate"
```

- [ ] **Step 3: Commit**

Run:
```bash
git add .goreleaser.yaml
git commit -m "build: add goreleaser config for cross-platform releases"
```

---

### Task 7: Add Install Script

**Files:**
- Create: `install.sh`

- [ ] **Step 1: Create `install.sh`**

```bash
#!/bin/sh
set -e

REPO="germanamz/elephant"
INSTALL_DIR="${INSTALL_DIR:-${HOME}/.local/bin}"

# Detect OS
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$OS" in
  linux)  OS="linux" ;;
  darwin) OS="darwin" ;;
  *)      echo "Unsupported OS: $OS" >&2; exit 1 ;;
esac

# Detect architecture
ARCH=$(uname -m)
case "$ARCH" in
  x86_64|amd64)  ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  *)             echo "Unsupported architecture: $ARCH" >&2; exit 1 ;;
esac

# Get latest version
VERSION=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
if [ -z "$VERSION" ]; then
  echo "Failed to fetch latest version" >&2
  exit 1
fi

ARCHIVE="elephant_${VERSION#v}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/download/${VERSION}/${ARCHIVE}"

echo "Installing elephant ${VERSION} (${OS}/${ARCH})..."

TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

curl -fsSL "$URL" -o "${TMPDIR}/${ARCHIVE}"
tar -xzf "${TMPDIR}/${ARCHIVE}" -C "$TMPDIR"

mkdir -p "$INSTALL_DIR"
mv "${TMPDIR}/elephant" "${INSTALL_DIR}/elephant"

echo "elephant ${VERSION} installed to ${INSTALL_DIR}/elephant"
```

- [ ] **Step 2: Make it executable**

Run:
```bash
chmod +x install.sh
```

- [ ] **Step 3: Commit**

Run:
```bash
git add install.sh
git commit -m "build: add install script for quick binary installation"
```

---

### Task 8: Add CI Workflow

**Files:**
- Create: `.github/workflows/ci.yml`

- [ ] **Step 1: Create `.github/workflows/ci.yml`**

```yaml
name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod

      - uses: golangci/golangci-lint-action@v6

  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod

      - run: go test ./...

  build:
    runs-on: ubuntu-latest
    needs: [lint, test]
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod

      - run: go build ./cmd/elephant/
```

- [ ] **Step 2: Commit**

Run:
```bash
git add .github/
git commit -m "ci: add github actions ci workflow"
```

---

### Task 9: Add Release Workflow

**Files:**
- Create: `.github/workflows/release.yml`

- [ ] **Step 1: Create `.github/workflows/release.yml`**

```yaml
name: Release

on:
  release:
    types: [published]

permissions:
  contents: write

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6
        with:
          fetch-depth: 0

      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod

      - uses: goreleaser/goreleaser-action@v6
        with:
          version: "~> v2"
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

- [ ] **Step 2: Commit**

Run:
```bash
git add .github/workflows/release.yml
git commit -m "ci: add github actions release workflow"
```

---

### Task 10: Add Community Files

**Files:**
- Create: `LICENSE`
- Create: `CODE_OF_CONDUCT.md`
- Create: `CONTRIBUTING.md`
- Create: `README.md`

- [ ] **Step 1: Create `LICENSE`**

Copy the Apache 2.0 license text (same as tusk's `LICENSE` file verbatim).

- [ ] **Step 2: Create `CODE_OF_CONDUCT.md`**

```markdown
# Code of Conduct

## Our Pledge

We as contributors and maintainers pledge to make participation in our project a welcoming experience for everyone, regardless of background or identity.

## Our Standards

Examples of behavior that contributes to a positive environment:

- Using welcoming language
- Being respectful of differing viewpoints and experiences
- Gracefully accepting constructive criticism
- Focusing on what is best for the community

Examples of unacceptable behavior:

- Trolling, insulting or derogatory comments, and personal attacks
- Public or private harassment
- Publishing others' private information without explicit permission
- Other conduct which could reasonably be considered inappropriate in a professional setting

## Responsibilities

Project maintainers are responsible for clarifying the standards of acceptable behavior and are expected to take appropriate and fair corrective action in response to any instances of unacceptable behavior.

## Scope

This Code of Conduct applies within all project spaces, including issues, pull requests, and any other communication channels used by the project.

## Enforcement

Instances of unacceptable behavior may be reported by contacting the project maintainers. All complaints will be reviewed and investigated and will result in a response that is deemed necessary and appropriate to the circumstances.

## Attribution

This Code of Conduct is adapted from the [Contributor Covenant](https://www.contributor-covenant.org), version 2.1.
```

- [ ] **Step 3: Create `CONTRIBUTING.md`**

```markdown
# Contributing to Elephant

Thank you for your interest in contributing to Elephant! This document provides guidelines and information to help you get started.

## Getting Started

### Prerequisites

- Go 1.26+
- golangci-lint (for linting)
- lefthook and conform are installed automatically by `make setup-hooks`

### Setup

```bash
git clone https://github.com/germanamz/elephant.git
cd elephant
make setup-hooks
make build
make test
```

## Development Workflow

1. Fork the repository and create a feature branch from `main`.
2. Make your changes following the conventions below.
3. Run tests and linting before submitting.
4. Open a pull request against `main`.

### Running Tests

```bash
make test           # All tests
make test-race      # Tests with race detector
```

### Linting

```bash
make vet
make lint
```

## Architecture

Elephant follows a modular architecture:

```
cmd/elephant/        Entry point (Cobra CLI)
internal/container/  Docker container lifecycle
internal/agent/      Headless agent runner, session management
internal/tui/        Bubbletea TUI
```

## Code Conventions

### Commits

Use conventional commits:

```
feat(agent): add session persistence
fix(container): handle mount path on windows
test: add container lifecycle tests
docs: update README with install instructions
```

### Formatting

Code is formatted with `gofmt`. The pre-commit hook checks this automatically.

## Reporting Issues

When reporting bugs, please include:
- Steps to reproduce
- Expected vs actual behavior
- Go version and OS
- Elephant version or commit hash

## License

By contributing to Elephant, you agree that your contributions will be licensed under the Apache 2.0 License.
```

- [ ] **Step 4: Create `README.md`**

```markdown
# Elephant

AI orchestration platform that lets you define what you want built at a high level, then steps back and lets agents do the work — safely, in parallel, with structured review points.

See [PRODUCT.md](PRODUCT.md) for the full product design and [ROADMAP.md](ROADMAP.md) for the implementation roadmap.

## Installation

### Quick Install

```bash
curl -fsSL https://raw.githubusercontent.com/germanamz/elephant/main/install.sh | sh
```

By default installs to `~/.local/bin`. Override with `INSTALL_DIR`:

```bash
INSTALL_DIR=/usr/local/bin curl -fsSL https://raw.githubusercontent.com/germanamz/elephant/main/install.sh | sh
```

### From Source

Requires Go 1.26+:

```bash
git clone https://github.com/germanamz/elephant.git
cd elephant
make build
```

The binary is compiled to `bin/elephant`.

## Development

```bash
make setup-hooks    # Install git hooks (lefthook + conform)
make build          # Compile to bin/elephant
make test           # Run tests
make test-race      # Tests with race detector
make vet            # go vet
make lint           # golangci-lint
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for development guidelines.

## License

Apache 2.0 — see [LICENSE](LICENSE).
```

- [ ] **Step 5: Commit**

Run:
```bash
git add LICENSE CODE_OF_CONDUCT.md CONTRIBUTING.md README.md
git commit -m "docs: add license, code of conduct, contributing guide, and readme"
```

---

## Task Summary

| Task | Description | Commit |
|------|-------------|--------|
| 1 | Go module + cobra entry point | `feat: initialize go module and cobra entry point` |
| 2 | Stub packages (container, agent, tui) | `feat: add stub packages for container, agent, and tui` |
| 3 | .gitignore | `chore: add gitignore` |
| 4 | Code quality configs (golangci, conform, lefthook) | `chore: add golangci-lint, conform, and lefthook configs` |
| 5 | Makefile | `build: add makefile with dev workflow targets` |
| 6 | GoReleaser config | `build: add goreleaser config for cross-platform releases` |
| 7 | Install script | `build: add install script for quick binary installation` |
| 8 | CI workflow | `ci: add github actions ci workflow` |
| 9 | Release workflow | `ci: add github actions release workflow` |
| 10 | Community files (LICENSE, CoC, CONTRIBUTING, README) | `docs: add license, code of conduct, contributing guide, and readme` |
