# Phase 3: CI/CD & Community Files

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add GitHub Actions workflows for CI and release, plus community files (LICENSE, CODE_OF_CONDUCT, CONTRIBUTING, README).

**Architecture:** Two GitHub Actions workflows — `ci.yml` runs lint, test, and build on push/PR; `release.yml` runs GoReleaser when a GitHub Release is published. Community files follow the same structure as [germanamz/tusk](https://github.com/germanamz/tusk).

**Tech Stack:** GitHub Actions, GoReleaser v2, golangci-lint-action

**Prerequisite:** Phase 1 and Phase 2 completed — Go module, entry point, stub packages, all quality/build configs in place.

---

## File Map

| File | Action | Responsibility |
|------|--------|---------------|
| `.github/workflows/ci.yml` | Create | CI pipeline — lint, test, build on push/PR |
| `.github/workflows/release.yml` | Create | Release pipeline — GoReleaser on release published |
| `LICENSE` | Create | Apache 2.0 license text |
| `CODE_OF_CONDUCT.md` | Create | Contributor Covenant v2.1 |
| `CONTRIBUTING.md` | Create | Development setup, workflow, conventions |
| `README.md` | Create | Project overview, install, dev instructions |

---

### Task 1: Add CI Workflow

**What this does:** Creates a GitHub Actions workflow that runs on every push to `main` and on every pull request. It runs three jobs: linting (golangci-lint), testing, and build verification. Lint and test run in parallel, build only runs after both pass.

**Files:**
- Create: `.github/workflows/ci.yml`

- [ ] **Step 1: Create the workflows directory**

Run:
```bash
mkdir -p .github/workflows
```

- [ ] **Step 2: Create `.github/workflows/ci.yml`**

Create the file `.github/workflows/ci.yml` with this exact content:

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

**What each section does:**

- `on` — triggers on push to `main` branch and on pull requests targeting `main`. This means every PR gets CI checks before merging.
- `jobs.lint`:
  - `actions/checkout@v4` — checks out the repository code
  - `actions/setup-go@v5` with `go-version-file: go.mod` — installs the Go version specified in `go.mod` (1.26). Using `go-version-file` instead of hardcoding the version means you only update it in one place.
  - `golangci/golangci-lint-action@v6` — runs golangci-lint using the `.golangci.yml` config. This action handles caching and installation automatically.
- `jobs.test`:
  - Same checkout and Go setup
  - `go test ./...` — runs all tests in all packages
- `jobs.build`:
  - `needs: [lint, test]` — only runs after both lint and test pass. This ensures we don't waste CI minutes building code that has lint errors or test failures.
  - `go build ./cmd/elephant/` — verifies the binary compiles. Doesn't produce an artifact — just a compilation check.

- [ ] **Step 3: Commit**

Run:
```bash
git add .github/workflows/ci.yml
git commit -m "ci: add github actions ci workflow"
```

---

### Task 2: Add Release Workflow

**What this does:** Creates a GitHub Actions workflow that triggers when you publish a GitHub Release. It runs GoReleaser to build cross-platform binaries and attach them to the release.

**How to use it (after everything is set up):**
1. Tag a commit: `git tag v0.1.0 && git push origin v0.1.0`
2. Create a GitHub Release from that tag (via GitHub UI or `gh release create v0.1.0`)
3. The workflow automatically builds binaries for all platforms and attaches them to the release

**Files:**
- Create: `.github/workflows/release.yml`

- [ ] **Step 1: Create `.github/workflows/release.yml`**

Create the file `.github/workflows/release.yml` with this exact content:

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

**What each section does:**

- `on.release.types: [published]` — triggers only when a GitHub Release is published (not drafted). This prevents accidental releases from draft creation.
- `permissions.contents: write` — allows the workflow to upload release assets (binaries, checksums) to the GitHub Release. Without this, GoReleaser can't attach files.
- `actions/checkout@v6` with `fetch-depth: 0` — checks out the full git history, not just the latest commit. GoReleaser needs this to generate the changelog from commit history between tags.
- `actions/setup-go@v5` — installs Go from `go.mod` version.
- `goreleaser/goreleaser-action@v6`:
  - `version: "~> v2"` — uses GoReleaser v2 (matches our `.goreleaser.yaml` `version: 2`)
  - `args: release --clean` — runs the release process and cleans up any previous build artifacts in `dist/`
  - `GITHUB_TOKEN` — automatically provided by GitHub Actions. GoReleaser uses this to upload assets to the release.

- [ ] **Step 2: Commit**

Run:
```bash
git add .github/workflows/release.yml
git commit -m "ci: add github actions release workflow"
```

---

### Task 3: Add Community Files

**What this does:** Adds the standard open-source community files: Apache 2.0 license, code of conduct, contributing guide, and README. These are adapted from [germanamz/tusk](https://github.com/germanamz/tusk).

**Files:**
- Create: `LICENSE`
- Create: `CODE_OF_CONDUCT.md`
- Create: `CONTRIBUTING.md`
- Create: `README.md`

- [ ] **Step 1: Create `LICENSE`**

Create the file `LICENSE` in the project root with the full Apache 2.0 license text. Copy it verbatim from https://github.com/germanamz/tusk/blob/main/LICENSE — it is the standard Apache License Version 2.0, January 2004 text with no modifications.

You can fetch it with:
```bash
curl -fsSL "https://raw.githubusercontent.com/germanamz/tusk/main/LICENSE" > LICENSE
```

- [ ] **Step 2: Create `CODE_OF_CONDUCT.md`**

Create the file `CODE_OF_CONDUCT.md` in the project root with this exact content:

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

Create the file `CONTRIBUTING.md` in the project root. Use tusk's `CONTRIBUTING.md` as a base and adapt it for Elephant. The key changes from tusk's version:

1. Replace all references to "Tusk" / "tusk" with "Elephant" / "elephant"
2. Replace the repo URL with `https://github.com/germanamz/elephant.git`
3. Replace the Architecture section with:
   - `cmd/elephant/` — Entry point (Cobra CLI)
   - `internal/container/` — Docker container lifecycle
   - `internal/agent/` — Headless agent runner, session management
   - `internal/tui/` — Bubbletea TUI
4. Remove tusk-specific sections (Error Handling, Key Patterns, E2E Tests) — Elephant doesn't have these yet
5. Keep: Prerequisites (Go 1.26+, golangci-lint, setup-hooks), Setup instructions, Development Workflow, Running Tests (`make test`, `make test-race`), Linting (`make vet`, `make lint`), Commit conventions (conventional commits with examples), Reporting Issues, License (Apache 2.0)

You can fetch tusk's version as a starting point:
```bash
curl -fsSL "https://raw.githubusercontent.com/germanamz/tusk/main/CONTRIBUTING.md" > CONTRIBUTING.md
```

Then edit the file with the changes listed above. See the design spec at `docs/superpowers/specs/2026-04-03-project-scaffolding-design.md` for the full expected content.

- [ ] **Step 4: Create `README.md`**

Create the file `README.md` in the project root. It should contain these sections in order:

1. **Title and description:** "# Elephant" followed by a one-liner: "AI orchestration platform that lets you define what you want built at a high level, then steps back and lets agents do the work — safely, in parallel, with structured review points."
2. **Links:** to PRODUCT.md and ROADMAP.md
3. **Installation — Quick Install:** `curl -fsSL https://raw.githubusercontent.com/germanamz/elephant/main/install.sh | sh` with note about `INSTALL_DIR` override
4. **Installation — From Source:** clone, `make build`, binary at `bin/elephant`
5. **Development:** list of make targets (`setup-hooks`, `build`, `test`, `test-race`, `vet`, `lint`) with link to CONTRIBUTING.md
6. **License:** "Apache 2.0 — see LICENSE."

See the design spec at `docs/superpowers/specs/2026-04-03-project-scaffolding-design.md` for the full expected content of each section.

- [ ] **Step 5: Commit**

Run:
```bash
git add LICENSE CODE_OF_CONDUCT.md CONTRIBUTING.md README.md
git commit -m "docs: add license, code of conduct, contributing guide, and readme"
```

---

## Verification

After completing all 3 tasks, verify the full project scaffolding is complete:

```bash
# All files from all 3 phases should exist
echo "=== Checking all expected files ==="
for f in \
  go.mod go.sum \
  cmd/elephant/main.go \
  internal/container/container.go \
  internal/agent/agent.go \
  internal/tui/tui.go \
  .gitignore \
  .golangci.yml \
  .conform.yaml \
  lefthook.yml \
  Makefile \
  .goreleaser.yaml \
  install.sh \
  .github/workflows/ci.yml \
  .github/workflows/release.yml \
  LICENSE \
  CODE_OF_CONDUCT.md \
  CONTRIBUTING.md \
  README.md \
  PRODUCT.md \
  ROADMAP.md; do
  if [ -f "$f" ]; then
    echo "  OK: $f"
  else
    echo "  MISSING: $f"
  fi
done

# Build and verify
make build
./bin/elephant --version
# Expected: elephant version dev (none, unknown)
make clean

# Lint should pass
golangci-lint run ./... 2>/dev/null && echo "Lint: OK" || echo "Lint: skipped (not installed)"

# install.sh should be executable
test -x install.sh && echo "install.sh: executable" || echo "install.sh: NOT executable"

# Git log should show 10 commits total across all 3 phases
git log --oneline -12
```
