# Phase 2: Code Quality & Build Tooling

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add code quality tooling (linter, git hooks, commit enforcement), a Makefile for developer workflow, GoReleaser for cross-platform releases, and an install script.

**Architecture:** All config files ported from [germanamz/tusk](https://github.com/germanamz/tusk) for consistency across projects. The Makefile wraps common Go commands. GoReleaser builds cross-platform binaries with version injection via ldflags.

**Tech Stack:** golangci-lint v2, Lefthook, Conform, GoReleaser v2, Make

**Prerequisite:** Phase 1 completed — Go module initialized, `cmd/elephant/main.go` exists with cobra root command and version ldflags, stub packages in `internal/`, `.gitignore` in place.

---

## File Map

| File | Action | Responsibility |
|------|--------|---------------|
| `.golangci.yml` | Create | golangci-lint v2 config — errcheck exclusions for tests and `.Close()` |
| `.conform.yaml` | Create | Conventional commit enforcement — types, header rules |
| `lefthook.yml` | Create | Git hooks — pre-commit (fmt, vet, lint, test), commit-msg (conform) |
| `Makefile` | Create | Developer targets: build, test, lint, vet, setup-hooks, etc. |
| `.goreleaser.yaml` | Create | Cross-platform binary builds, changelog, GitHub release |
| `install.sh` | Create | Shell script to download and install latest release binary |

---

### Task 1: Add Code Quality Configs

**What this does:** Adds three config files that enforce code quality:
1. **golangci-lint** — runs multiple Go linters in one tool. Our config excludes `errcheck` in test files (tests often ignore errors intentionally) and for `.Close()` calls (a common Go pattern where the error is non-actionable).
2. **Conform** — validates commit messages follow conventional commit format. This ensures consistent, parseable commit history that GoReleaser uses to generate changelogs.
3. **Lefthook** — runs checks automatically on git hooks so issues are caught before code is pushed.

**Files:**
- Create: `.golangci.yml`
- Create: `.conform.yaml`
- Create: `lefthook.yml`

- [ ] **Step 1: Create `.golangci.yml`**

Create the file `.golangci.yml` in the project root with this exact content:

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

**What this config does:**
- `version: "2"` — uses golangci-lint v2 config format.
- First exclusion rule: disables `errcheck` linter for any file matching `_test.go`. Tests often call functions without checking errors when the test will fail anyway.
- Second exclusion rule: disables `errcheck` for any `.Close()` call. In Go, closing file handles/connections can return errors, but there's rarely anything useful to do with them.
- All other linters run with their defaults (errcheck, govet, staticcheck, etc.).

- [ ] **Step 2: Create `.conform.yaml`**

Create the file `.conform.yaml` in the project root with this exact content:

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

**What this config does:**
- Enforces [Conventional Commits](https://www.conventionalcommits.org/) format: `type(scope): description`
- Header rules: max 89 characters, must start with imperative verb, lowercase, can't end with a period
- `types` — allowed commit types. Examples: `feat: add session persistence`, `fix: handle nil pointer in agent`, `ci: update go version in workflow`
- `scopes` — `".*"` means any scope is allowed (including no scope). Examples: `feat(agent): ...`, `fix(tui): ...`, `docs: ...`
- Body and GPG signature are optional

- [ ] **Step 3: Create `lefthook.yml`**

Create the file `lefthook.yml` in the project root with this exact content:

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

**What this config does:**
- **pre-commit** hook runs 4 commands in parallel whenever you `git commit`:
  - `gofmt` — checks that all staged `.go` files are formatted. If any aren't, it fails with a message telling you to run `gofmt -w .`
  - `vet` — runs `go vet` which catches suspicious constructs (e.g., unreachable code, bad format strings)
  - `lint` — runs golangci-lint with the `.golangci.yml` config
  - `test` — runs tests in `internal/` and `cmd/` packages
- **commit-msg** hook runs conform to validate the commit message format. `{1}` is lefthook's variable for the commit message file path.
- `glob: "*.go"` means the commands only trigger when `.go` files are staged (skip if you're only changing docs).

- [ ] **Step 4: Verify the linter runs cleanly (if installed)**

Run:
```bash
golangci-lint run ./...
```

If golangci-lint is installed, expected output: no issues found (empty output, exit code 0).
If not installed, you'll get `command not found` — that's fine, the CI pipeline will run it. You can install it with:
```bash
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

- [ ] **Step 5: Commit**

Run:
```bash
git add .golangci.yml .conform.yaml lefthook.yml
git commit -m "chore: add golangci-lint, conform, and lefthook configs"
```

**Note:** If lefthook hooks are already installed (from a previous `make setup-hooks`), this commit will be validated by conform. If not, that's fine — the hooks will be installed when someone runs `make setup-hooks` (Task 2).

---

### Task 2: Add Makefile

**What this does:** Creates a Makefile with standard targets for building, testing, linting, and setting up git hooks. This gives every developer the same commands regardless of their setup.

**Files:**
- Create: `Makefile`

- [ ] **Step 1: Create `Makefile`**

Create the file `Makefile` in the project root with this exact content:

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

**IMPORTANT:** Makefile rules require actual tab characters for indentation, not spaces. If your editor converts tabs to spaces, the Makefile will not work. Make sure each indented line under a target starts with a real tab (`\t`).

**What each target does:**
- `make` or `make build` — compiles the binary to `bin/elephant` (the `bin/` directory is gitignored)
- `make clean` — removes the `bin/` directory
- `make test` — runs all tests with verbose output
- `make test-race` — runs tests with Go's race detector enabled (catches concurrent access bugs)
- `make vet` — runs `go vet` for static analysis
- `make lint` — runs golangci-lint (must be installed)
- `make run` — builds then runs the binary
- `make install` — installs the binary to your `$GOPATH/bin`
- `make setup-hooks` — installs lefthook and conform, then sets up the git hooks. Run this once after cloning the repo.

- [ ] **Step 2: Verify `make build` works**

Run:
```bash
make build
```

Expected output:
```
go build -v -o bin/elephant ./cmd/elephant
```

Verify the binary exists:
```bash
./bin/elephant --version
```

Expected output:
```
elephant version dev (none, unknown)
```

- [ ] **Step 3: Verify `make test` works**

Run:
```bash
make test
```

Expected: passes with no errors. There are no tests yet, so it will just report the packages with `[no test files]`.

- [ ] **Step 4: Clean up and commit**

Run:
```bash
make clean
git add Makefile
git commit -m "build: add makefile with dev workflow targets"
```

`make clean` removes the `bin/` directory we just created during verification.

---

### Task 3: Add GoReleaser Config

**What this does:** Configures GoReleaser to build cross-platform binaries and create GitHub Releases with changelogs. When you publish a GitHub Release, the release workflow (Phase 3) runs GoReleaser to build binaries for all platforms and attach them to the release.

**Files:**
- Create: `.goreleaser.yaml`

- [ ] **Step 1: Create `.goreleaser.yaml`**

Create the file `.goreleaser.yaml` in the project root with this exact content:

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

**What each section does:**

- `before.hooks` — runs `go mod tidy` before building to ensure dependencies are clean.
- `builds` section:
  - `main: ./cmd/elephant` — path to the main package to build
  - `CGO_ENABLED=0` — disables CGO for fully static binaries (no external C dependencies needed)
  - `ldflags`:
    - `-s -w` — strips debug info and symbol table, reducing binary size
    - `-X main.version={{.Version}}` — injects the git tag as the version into the `version` variable in `cmd/elephant/main.go`
    - Same for `commit` and `date`
  - `goos` + `goarch` — builds for Linux, macOS, and Windows on both amd64 and arm64 architectures (6 binaries total)
- `archives` — packages binaries as `.tar.gz` (Linux/macOS) or `.zip` (Windows). Named like `elephant_1.0.0_darwin_arm64.tar.gz`
- `checksum` — generates a `checksums.txt` file with SHA-256 hashes of all archives
- `changelog` — auto-generates release notes from git commits. Groups by feat/fix, excludes docs/test/ci commits
- `release` — publishes to `germanamz/elephant` on GitHub. `keep-existing` means it won't overwrite release notes you've already written. `prerelease: auto` marks releases with `-rc`, `-beta`, etc. as pre-releases automatically.

- [ ] **Step 2: Verify the config is valid (optional)**

Run:
```bash
goreleaser check 2>/dev/null || echo "goreleaser not installed locally — CI will validate"
```

If goreleaser is installed, expected output: no errors. If not installed, that's fine — the release workflow in CI will run it.

- [ ] **Step 3: Commit**

Run:
```bash
git add .goreleaser.yaml
git commit -m "build: add goreleaser config for cross-platform releases"
```

---

### Task 4: Add Install Script

**What this does:** Creates a shell script that users can pipe to `sh` to install the latest Elephant binary. It detects the user's OS and architecture, downloads the correct archive from GitHub Releases, and extracts it.

**Files:**
- Create: `install.sh`

- [ ] **Step 1: Create `install.sh`**

Create the file `install.sh` in the project root with this exact content:

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

**How this script works, line by line:**

1. `set -e` — exit immediately if any command fails
2. `INSTALL_DIR` — defaults to `~/.local/bin` but the user can override with `INSTALL_DIR=/usr/local/bin curl ... | sh`
3. OS detection — `uname -s` returns `Darwin` or `Linux`, converted to lowercase to match GoReleaser archive names
4. Architecture detection — `uname -m` returns `x86_64` or `arm64`, normalized to Go's naming (`amd64`/`arm64`)
5. Version fetch — queries the GitHub API for the latest release tag (e.g., `v0.1.0`)
6. `${VERSION#v}` — strips the `v` prefix from the tag because GoReleaser archives use `0.1.0` not `v0.1.0`
7. Downloads to a temp directory, extracts, moves binary to install dir
8. `trap 'rm -rf "$TMPDIR"' EXIT` — ensures the temp directory is cleaned up even if the script fails

**Usage:** `curl -fsSL https://raw.githubusercontent.com/germanamz/elephant/main/install.sh | sh`

- [ ] **Step 2: Make the script executable**

Run:
```bash
chmod +x install.sh
```

This sets the executable permission so the script can be run directly with `./install.sh`. This permission is tracked by git.

- [ ] **Step 3: Commit**

Run:
```bash
git add install.sh
git commit -m "build: add install script for quick binary installation"
```

---

## Verification

After completing all 4 tasks, verify the project state:

```bash
# All config files should exist
ls .golangci.yml .conform.yaml lefthook.yml Makefile .goreleaser.yaml install.sh

# Makefile should build successfully
make build
./bin/elephant --version
# Expected: elephant version dev (none, unknown)
make clean

# Linter should pass (if installed)
golangci-lint run ./... 2>/dev/null || echo "golangci-lint not installed — ok"

# install.sh should be executable
test -x install.sh && echo "install.sh is executable" || echo "ERROR: install.sh not executable"

# Git log should show 4 new commits from this phase
git log --oneline -6
```
