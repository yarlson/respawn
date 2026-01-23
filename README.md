# Turbine

Turbine is a Go-based CLI that autonomously drives coding agents through a resilient "Decompose → Execute → Verify → Commit" loop to implement a PRD end-to-end without human interaction.

- **PRD decomposition** - Breaks down PRD files into actionable tasks with dependencies
- **Autonomous task execution** - Reads tasks from `.turbine/tasks.yaml` and executes them sequentially
- **Multiple backend support** - Works with Claude and OpenCode agents
- **Resilient execution** - Implements 3x3 rotation/stroke retry policy for reliable task completion
- **State persistence** - Resumes interrupted runs from saved state

## Install

```bash
go install github.com/yarlson/turbine@latest
```

Or build from source:

```bash
go build -o turbine .
```

## Quickstart

1. Create a configuration file at `~/.config/turbine/turbine.yaml` (optional - defaults are provided)
2. Load your PRD into the task manifest:
   ```bash
   turbine load --prd path/to/your-prd.md
   ```
3. Spin up the turbine:
   ```bash
   turbine
   ```

## Usage

### Run Tasks Autonomously

```bash
turbine [flags]
```

Reads tasks from `.turbine/tasks.yaml` and executes each one using the configured backend.

### Load PRD into Task Manifest

```bash
turbine load --prd <path> [flags]
```

Takes a PRD file and breaks it down into actionable tasks in `.turbine/tasks.yaml`.

### Flags

| Flag        | Description                         |
| ----------- | ----------------------------------- |
| `--backend` | AI backend (`opencode` or `claude`) |
| `--model`   | Model name for the backend          |
| `--variant` | Variant configuration               |
| `--yes`     | Skip confirmation prompts           |
| `--verbose` | Show detailed output                |
| `--debug`   | Show debug logs                     |

## Configuration

### Config File

Global configuration is stored at:

- `~/.config/turbine/turbine.yaml`
- or `${XDG_CONFIG_HOME}/turbine/turbine.yaml`

If missing, Turbine uses default configuration with `opencode` backend and 3x3 retry policy.

### Environment Variables

| Variable          | Required | Description                       |
| ----------------- | -------- | --------------------------------- |
| `XDG_CONFIG_HOME` | No       | Override default config directory |
| `NO_COLOR`        | No       | Disable colored terminal output   |

### Tasks File

Task definitions are stored at:

- `./.turbine/tasks.yaml` (tracked in git)

Each task includes:

- `id` - Unique identifier
- `title` - Task title
- `status` - `todo`, `done`, or `failed`
- `deps` - Task dependencies (optional)
- `description` - Detailed description
- `acceptance` - Acceptance criteria
- `verify` - Verification commands
- `commit_message` - Git commit message

### Run Artifacts

Execution logs and state are stored under:

- `./.turbine/runs/` (gitignored)
- `./.turbine/state/` (gitignored)

## Troubleshooting

| Symptom                                        | Solution                                                                                                                                                           |
| ---------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| Task execution may fail during agent execution | Turbine implements 3x3 rotation/stroke policy: same-session strokes and new-session rotations. Check `cmd/turbine/run.go` and `internal/run/retry.go` for details. |
| State persistence across restarts              | `.turbine/state/` is used for resume information (gitignored), `.turbine/runs/` for artifacts. Both directories are managed automatically.                         |
| Configuration not found                        | Turbine falls back to default configuration if `~/.config/turbine/turbine.yaml` is missing. Defaults use `opencode` backend with 3x3 retry policy.                 |

## Development

### Prerequisites

- Go 1.25.5 or later

### Build

```bash
go build -o turbine .
```

### Test

```bash
go test ./...
```

## License

MIT License. See [LICENSE](LICENSE) for details.
