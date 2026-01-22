# Respawn

Respawn is a Go-based CLI that autonomously drives coding agents through a resilient "Decompose → Execute → Verify → Commit" loop to implement a PRD end-to-end without human interaction.

## Usage

### Run tasks autonomously
```bash
respawn [flags]
```

### Decompose PRD into tasks
```bash
respawn decompose --prd <path> [flags]
```

## Configuration

### Config File
Global configuration is stored at:
- `~/.config/respawn/respawn.yaml`
- or `${XDG_CONFIG_HOME}/respawn/respawn.yaml`

### Tasks File
Task definitions are stored at:
- `./.respawn/tasks.yaml` (tracked in git)

### Run Artifacts
Execution logs and state are stored under:
- `./.respawn/runs/` (gitignored)
- `./.respawn/state/` (gitignored)

## Flags

- `--backend {opencode|claude}`: Select the coding agent backend.
- `--model <string>`: Specify the model name.
- `--variant <string>`: Specify the reasoning effort (OpenCode only).
- `--yes`: Auto-accept prompts (e.g., .gitignore updates).
- `--verbose`: Stream full backend output live.
- `--debug`: Include gate stdout/stderr inline in terminal output.
