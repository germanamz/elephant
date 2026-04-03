# Phase 1: Project Foundation

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Initialize the Go module, create the cobra CLI entry point with version injection, set up stub packages for future initiatives, and add a .gitignore.

**Architecture:** Standard Go project layout — `cmd/elephant/` for the binary entry point, `internal/` for private packages. The entry point uses cobra for CLI and accepts version/commit/date via ldflags at build time.

**Tech Stack:** Go 1.26, Cobra, Bubbletea (dependency only — no code using it yet)

**Prerequisite:** Go 1.26+ installed. The project repo is at `/Users/germanamz/projects/elephant` with only `PRODUCT.md` and `ROADMAP.md` existing.

---

## File Map

| File | Action | Responsibility |
|------|--------|---------------|
| `go.mod` | Create | Module definition (`github.com/germanamz/elephant`), Go 1.26, cobra + bubbletea deps |
| `go.sum` | Auto-generated | Dependency checksums (created by `go mod tidy`) |
| `cmd/elephant/main.go` | Create | Cobra root command, version string via ldflags |
| `internal/container/container.go` | Create | Stub — package declaration only |
| `internal/agent/agent.go` | Create | Stub — package declaration only |
| `internal/tui/tui.go` | Create | Stub — package declaration only |
| `.gitignore` | Create | Ignore build output, IDE files, OS files |

---

### Task 1: Initialize Go Module and Cobra Entry Point

**What this does:** Creates the Go module, adds cobra and bubbletea as dependencies, and creates the CLI entry point that prints version info.

**Why version injection matters:** GoReleaser injects the version, commit hash, and build date at compile time using `-ldflags`. The variables in `main.go` have default values (`dev`, `none`, `unknown`) so the binary works during local development without ldflags.

**Files:**
- Create: `go.mod` (via `go mod init`)
- Create: `go.sum` (via `go mod tidy`)
- Create: `cmd/elephant/main.go`

- [ ] **Step 1: Initialize the Go module**

Run this from the project root (`/Users/germanamz/projects/elephant`):

```bash
go mod init github.com/germanamz/elephant
```

Expected output:
```
go: creating new go.mod: module github.com/germanamz/elephant
```

This creates a `go.mod` file with the module path and Go version.

- [ ] **Step 2: Add cobra and bubbletea dependencies**

Run:
```bash
go get github.com/spf13/cobra@latest
go get github.com/charmbracelet/bubbletea@latest
```

This downloads both libraries and adds them to `go.mod`. Cobra is used in this task for the CLI. Bubbletea is added now so it's available for the TUI initiative later — no code uses it yet.

- [ ] **Step 3: Create the directory for the entry point**

Run:
```bash
mkdir -p cmd/elephant
```

- [ ] **Step 4: Create `cmd/elephant/main.go`**

Create the file `cmd/elephant/main.go` with this exact content:

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

**What each part does:**
- `version`, `commit`, `date` — variables that GoReleaser overrides at build time via ldflags. During local dev they default to `"dev"`, `"none"`, `"unknown"`.
- `rootCmd` — the top-level cobra command. `Use` is the binary name. `Short` is the one-line description shown in help. `Version` is printed when the user runs `elephant --version`.
- `rootCmd.Execute()` — runs the command. If it fails, we exit with code 1.

- [ ] **Step 5: Tidy dependencies and verify it compiles**

Run:
```bash
go mod tidy
go build ./cmd/elephant/
```

`go mod tidy` removes any unused dependencies and ensures `go.sum` is correct. `go build` compiles the binary. If successful, it produces an `elephant` binary in the current directory with no output.

Verify it runs:
```bash
./elephant --version
```

Expected output:
```
elephant version dev (none, unknown)
```

- [ ] **Step 6: Clean up the binary and commit**

Run:
```bash
rm -f elephant
git add go.mod go.sum cmd/elephant/main.go
git commit -m "feat: initialize go module and cobra entry point"
```

The `elephant` binary was just for verification — we don't commit it. The Makefile (Phase 2) will build to `bin/` which will be gitignored.

---

### Task 2: Create Stub Packages

**What this does:** Creates empty Go packages under `internal/` to establish the directory structure for future v0.1 initiatives. Each package maps to an initiative in the roadmap.

**Why stubs:** These give future initiatives a clear landing zone. The `internal/` directory in Go means these packages are private to this module — they can't be imported by external projects.

**Files:**
- Create: `internal/container/container.go`
- Create: `internal/agent/agent.go`
- Create: `internal/tui/tui.go`

- [ ] **Step 1: Create the directories**

Run:
```bash
mkdir -p internal/container internal/agent internal/tui
```

- [ ] **Step 2: Create `internal/container/container.go`**

Create the file `internal/container/container.go` with this exact content:

```go
package container
```

This stub is for the Docker Container Management initiative — it will handle spawning containers with codebase mounted, teardown, and ephemeral lifecycle.

- [ ] **Step 3: Create `internal/agent/agent.go`**

Create the file `internal/agent/agent.go` with this exact content:

```go
package agent
```

This stub is for the Agent Execution initiative — it will handle running Claude Code in headless mode, parsing output, and session persistence/resumption.

- [ ] **Step 4: Create `internal/tui/tui.go`**

Create the file `internal/tui/tui.go` with this exact content:

```go
package tui
```

This stub is for the Basic TUI initiative — it will use bubbletea for agent status monitoring and log streaming.

- [ ] **Step 5: Verify the whole project still compiles**

Run:
```bash
go build ./...
```

The `./...` pattern means "build all packages in this module." Expected: no output (success). This confirms the stub packages don't break anything.

- [ ] **Step 6: Commit**

Run:
```bash
git add internal/
git commit -m "feat: add stub packages for container, agent, and tui"
```

---

### Task 3: Add .gitignore

**What this does:** Creates a `.gitignore` file so build artifacts, IDE files, and OS files don't get accidentally committed.

**Files:**
- Create: `.gitignore`

- [ ] **Step 1: Create `.gitignore`**

Create the file `.gitignore` with this exact content:

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

# Dist (goreleaser output)
dist/
```

**What each section ignores:**
- `bin/` — where `make build` puts the compiled binary (Makefile created in Phase 2)
- Go section — compiled binaries and test artifacts on all platforms
- IDE section — JetBrains (`.idea/`), VS Code (`.vscode/`), and vim swap files
- OS section — macOS `.DS_Store` and Windows `Thumbs.db`
- `dist/` — GoReleaser's default output directory when running locally

- [ ] **Step 2: Commit**

Run:
```bash
git add .gitignore
git commit -m "chore: add gitignore"
```

---

## Verification

After completing all 3 tasks, verify the project state:

```bash
# Project should compile cleanly
go build ./...

# Entry point should show version
go run ./cmd/elephant/ --version
# Expected: elephant version dev (none, unknown)

# Directory structure should look like this
find . -not -path './.git/*' -not -path './.git' | sort
# Expected:
# .
# ./.gitignore
# ./PRODUCT.md
# ./ROADMAP.md
# ./cmd
# ./cmd/elephant
# ./cmd/elephant/main.go
# ./go.mod
# ./go.sum
# ./internal
# ./internal/agent
# ./internal/agent/agent.go
# ./internal/container
# ./internal/container/container.go
# ./internal/tui
# ./internal/tui/tui.go

# Git log should show 3 new commits
git log --oneline -5
```
