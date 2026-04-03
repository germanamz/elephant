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
make test           # Unit tests
make test-race      # Tests with race detector
```

### Linting

```bash
make vet
make lint
```

## Architecture

- `cmd/elephant/` — Entry point (Cobra CLI)
- `internal/container/` — Docker container lifecycle
- `internal/agent/` — Headless agent runner, session management
- `internal/tui/` — Bubbletea TUI

## Code Conventions

### Commits

Use conventional commits with scope:

```
feat(cli): add tree view command
fix(container): handle null parent_id in query
test(agent): add filter syntax scenarios
docs: update README with examples
```

## Reporting Issues

When reporting bugs, please include:
- Steps to reproduce
- Expected vs actual behavior
- Go version and OS
- Elephant version or commit hash

## License

By contributing to Elephant, you agree that your contributions will be licensed under the Apache 2.0 License.
