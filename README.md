# Turbine

Turbine is a Go-based CLI that autonomously drives coding agents through a resilient "Decompose → Execute → Verify → Commit" loop to implement a PRD end-to-end without human interaction.

- **Autonomous task execution** - Reads tasks from `.turbine/tasks.yaml` and executes them
- **PRD decomposition** - Breaks down PRD files into actionable tasks
- **Multiple backend support** - Works with Claude and OpenCode agents
- **Resilient execution** - Implements retry/reset policies for reliable task completion
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

1. Create a configuration file at `~/.config/turbine/turbine.yaml`
2. Decompose your PRD into tasks:
   ```bash
   turbine decompose --prd path/to/your-prd.md
   ```
3. Run the tasks:
   ```bash
   turbine
   ```

## Usage

### Run Tasks Autonomously

```bash
turbine [flags]
```

Reads tasks from `.turbine/tasks.yaml` and executes them using the configured backend.

### Decompose PRD into Tasks

```bash
turbine decompose --prd <path> [flags]
```

Takes a PRD file and breaks it down into actionable tasks in `.turbine/tasks.yaml`.

### Flags

| Flag        | Description                                              |
| ----------- | -------------------------------------------------------- |
| `--backend` | Select the coding agent backend (`opencode` or `claude`) |
| `--model`   | Specify the model name                                   |
| `--variant` | Specify the reasoning effort (OpenCode only)             |
| `--yes`     | Auto-accept prompts (e.g., .gitignore updates)           |
| `--verbose` | Stream full backend output live                          |
| `--debug`   | Include gate stdout/stderr inline in terminal output     |

## Configuration

### Config File

Global configuration is stored at:

- `~/.config/turbine/turbine.yaml`
- or `${XDG_CONFIG_HOME}/turbine/turbine.yaml`

| Variable          | Required | Description                       |
| ----------------- | -------- | --------------------------------- |
| `XDG_CONFIG_HOME` | No       | Override default config directory |

### Tasks File

Task definitions are stored at:

- `./.turbine/tasks.yaml` (tracked in git)

### Run Artifacts

Execution logs and state are stored under:

- `./.turbine/runs/` (gitignored)
- `./.turbine/state/` (gitignored)

## Troubleshooting

| Symptom                                        | Solution                                                                                                                                                             |
| ---------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Task execution may fail during agent execution | Turbine implements 3x3 retry/reset policy: same-session retries and new-session resets. Check `cmd/turbine/run.go` and `internal/run/retryandresume.go` for details. |
| State persistence across restarts              | `.turbine/state/` is used for resume information (gitignored), `.turbine/runs/` for artifacts.                                                                       |

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
