# devx

A developer productivity CLI with interactive TUI for managing local dev environments, scaffolding projects, monitoring services, and aggregating logs.

## Features

- **Project Management** — Scaffold, register, and manage development projects
- **Service Monitor** — Track and health-check running local services in real-time
- **Log Aggregator** — Tail files, full-text search across stored log entries
- **Environment Manager** — Switch between .env profiles per project
- **Interactive Dashboard** — Full-screen TUI combining projects, services, and logs

## Installation

### From Source

```bash
git clone https://github.com/norfrt6-lab/go-dev-cli.git
cd go-dev-cli
make build
# Binary is at ./devx
```

### From Releases

Download the latest binary from [GitHub Releases](https://github.com/norfrt6-lab/go-dev-cli/releases).

## Quick Start

```bash
# Scaffold a new Go CLI project
devx project init --template go-cli --name my-tool

# Register an existing project
devx project add --name my-api --path ./my-api --lang go

# List projects
devx project list

# Monitor a service
devx service add api --host 127.0.0.1 --port 3000
devx service list
devx service health api

# Tail and search logs
devx logs tail /var/log/app.log -f
devx logs search "connection refused"

# Manage environment profiles
devx env list
devx env show development
devx env switch production
devx env diff development production

# Launch interactive dashboard
devx dashboard
```

## Commands

```
devx — Developer productivity CLI

Available Commands:
  project     Manage development projects
  service     Monitor local services
  logs        Aggregate and search logs
  env         Manage environment profiles
  dashboard   Interactive TUI dashboard
  info        Show system information
  version     Print version info
  completion  Generate shell completions
```

### Project Commands

| Command | Description |
|---------|-------------|
| `devx project init` | Scaffold from template (go-cli, go-api, node-api, react-app, python-api) |
| `devx project add` | Register an existing project |
| `devx project list` | List all registered projects |
| `devx project remove <name>` | Unregister a project |
| `devx project open <name>` | Open in $EDITOR |
| `devx project info <name>` | Show project details |

### Service Commands

| Command | Description |
|---------|-------------|
| `devx service scan` | Auto-detect services on common ports |
| `devx service add <name>` | Register a service to monitor |
| `devx service list` | List all services with status |
| `devx service health [name]` | Run health checks |
| `devx service remove <name>` | Unregister a service |
| `devx service watch` | Full-screen live service monitor |

### Log Commands

| Command | Description |
|---------|-------------|
| `devx logs tail <file>` | Tail a log file (`-f` to follow) |
| `devx logs search <query>` | Full-text search stored entries |
| `devx logs query` | Query with filters (source, level, since) |
| `devx logs sources` | List all log sources |
| `devx logs clear` | Clear stored entries |
| `devx logs view` | Full-screen log viewer |

### Environment Commands

| Command | Description |
|---------|-------------|
| `devx env list` | List environment profiles |
| `devx env show <profile>` | Display variables (masked by default) |
| `devx env switch <profile>` | Activate a profile as `.env` |
| `devx env diff <a> <b>` | Compare two profiles |
| `devx env export <profile>` | Print `export` statements |

## Architecture

- **Single binary** — No runtime dependencies, cross-platform
- **SQLite** — Local persistent storage via pure Go driver (no CGO)
- **Cobra + Viper** — CLI framework with config file support
- **Bubble Tea** — Interactive TUI with Elm architecture
- **FTS5** — Full-text search for log entries

## Configuration

Configuration is stored at `~/.devx/config.yaml`:

```yaml
editor: code
default_health_path: /health
log_retention_days: 30
theme: dark
scan_ports: [3000, 5173, 5432, 6379, 8080, 9090]
```

## Development

```bash
make build       # Build binary
make test        # Run tests with race detection
make lint        # Run golangci-lint
make fmt         # Format code
make vet         # Run go vet
```

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Language | Go 1.24+ |
| CLI | Cobra + Viper |
| TUI | Bubble Tea + Lip Gloss |
| Database | SQLite (modernc.org/sqlite) |
| Testing | testify |
| Linting | golangci-lint |
| Releases | GoReleaser |

## License

MIT
