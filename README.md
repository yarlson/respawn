# Turbine

Turbine is a Go-based CLI that autonomously drives coding agents through a resilient "Plan → Execute → Verify → Commit" loop to implement a PRD end-to-end without human interaction.

<div align="center">
  <img src="assets/turbine.jpeg" alt="Turbine" width="100%" />
</div>

- **Task planning** - Plans the next task from PRD + progress without a full upfront DAG
- **AGENTS.md generation** - Creates project guidelines with progressive disclosure and appropriate development methodologies (TDD for backend, UI validation for frontend)
- **Autonomous task execution** - Plans and executes the next task from `.turbine/task.yaml` using PRD + progress
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
2. Generate project guidelines from your PRD:
   ```bash
   turbine agents --prd path/to/your-prd.md
   ```
3. Spin up the turbine:
   ```bash
   turbine --prd path/to/your-prd.md
   ```

## Usage

### Run Tasks Autonomously

```bash
turbine [flags]
```

Plans the next task from PRD + progress, writes `.turbine/task.yaml`, and executes it using the configured backend.

Each command prints the backend and model being used:

```
Using backend: claude, model: claude-3-5-sonnet-latest
```

### Generate Project Guidelines

```bash
turbine agents --prd <path> [flags]
```

Generates `AGENTS.md` and supporting documentation in `docs/` with progressive disclosure. Automatically selects the appropriate development methodology based on project type:

- **Backend/API/Library** → Test-Driven Development (TDD)
- **Frontend/UI** → Browser/UI validation patterns
- **CLI tools** → Output verification patterns

Also creates a `CLAUDE.md` symlink pointing to `AGENTS.md` for tool compatibility.

### Flags

| Flag        | Description                         |
| ----------- | ----------------------------------- |
| `--backend` | AI backend (`opencode` or `claude`) |
| `--model`   | Model name for the backend          |
| `--variant` | Variant configuration               |
| `--prd`     | Path to the PRD file                |
| `--yes`     | Skip confirmation prompts           |

## Configuration

See [Configuration Guide](docs/CONFIGURATION.md) for complete configuration options and examples.

### Config File Location

Global configuration is stored at:

- `~/.config/turbine/turbine.yaml`
- or `${XDG_CONFIG_HOME}/turbine/turbine.yaml`

If missing, Turbine uses default configuration with `opencode` backend and 3x3 retry policy.

### Environment Variables

| Variable          | Required | Description                       |
| ----------------- | -------- | --------------------------------- |
| `XDG_CONFIG_HOME` | No       | Override default config directory |
| `NO_COLOR`        | No       | Disable colored terminal output   |

### Model Strategy (Fast vs Slow)

Turbine uses different models for different operations to optimize cost and quality:

**Slow Model** (e.g., Claude Opus 4.5):

- Used for one-time, high-stakes operations
- `turbine agents` - Generates AGENTS.md with methodology selection
- `turbine` (planning phase) - Plans the next task from PRD + progress
- Better quality for planning and architecture

**Fast Model** (e.g., Claude Sonnet):

- Used for repetitive implementation work
- Task execution during `turbine run`
- Cost-optimized for iterative development
- Sufficient for most implementation tasks

Configure in `~/.config/turbine/turbine.yaml`:

```yaml
backends:
  opencode:
    command: opencode
    models:
      fast:
        name: anthropic/claude-haiku-4.5
        variant: low
      slow: anthropic/claude-opus-4-5
  claude:
    command: claude
    models:
      fast: claude-3-5-sonnet-latest
      slow: claude-4-5-opus-latest
```

The `variant` field is a model modifier for reasoning effort (OpenCode only). Each model can have its own variant.

Override at runtime:

```bash
# Override model for all operations
turbine --model anthropic/claude-sonnet

# Override variant (OpenCode only)
turbine --variant high
```

### Task File

Current task definition is stored at:

- `./.turbine/task.yaml` (tracked in git)

Each task includes:

- `id` - Unique identifier
- `title` - Task title
- `status` - `todo`, `done`, or `failed`
- `description` - Detailed description
- `acceptance` - Acceptance criteria
- `verify` - Verification commands
- `commit_message` - Git commit message

Completed tasks are archived under:

- `./.turbine/archive/`

Progress is tracked at:

- `./.turbine/progress.md`

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
