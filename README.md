# Elephant

AI orchestration platform that lets you define what you want built at a high level, then steps back and lets agents do the work — safely, in parallel, with structured review points.

- [PRODUCT.md](PRODUCT.md)
- [ROADMAP.md](ROADMAP.md)

## Installation

### Quick Install

```bash
curl -fsSL https://raw.githubusercontent.com/germanamz/elephant/main/install.sh | sh
```

Set `INSTALL_DIR` to override the default installation directory:

```bash
INSTALL_DIR=/usr/local/bin curl -fsSL https://raw.githubusercontent.com/germanamz/elephant/main/install.sh | sh
```

### From Source

```bash
git clone https://github.com/germanamz/elephant.git
cd elephant
make build
```

The binary is placed at `bin/elephant`.

## Development

| Target | Description |
|--------|-------------|
| `make setup-hooks` | Install git hooks (lefthook + conform) |
| `make build` | Build the binary |
| `make test` | Run unit tests |
| `make test-race` | Run tests with race detector |
| `make vet` | Run go vet |
| `make lint` | Run golangci-lint |

See [CONTRIBUTING.md](CONTRIBUTING.md) for full contribution guidelines.

## License

Apache 2.0 — see [LICENSE](LICENSE).
