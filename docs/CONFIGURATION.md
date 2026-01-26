# Configuration Guide

Turbine uses a YAML configuration file located at:

- `~/.config/turbine/turbine.yaml`
- or `${XDG_CONFIG_HOME}/turbine/turbine.yaml` (if `XDG_CONFIG_HOME` is set)

If the file is missing, Turbine uses built-in defaults.

## Configuration Structure

```yaml
defaults:
  backend: opencode # Default backend to use
  quiet: false # Suppress output
  retry:
    rotations: 3 # Number of new-session rotations
    strokes: 3 # Number of strokes per rotation

backends:
  claude:
    command: claude # Command to invoke the backend
    args: # Additional arguments
      - -p
      - --dangerously-skip-permissions
    models:
      fast: claude-3-5-sonnet-latest # Model for task execution
      slow: claude-4-5-opus-latest # Model for planning (AGENTS.md, task planning)
    variant: "" # Optional backend-specific variant

  opencode:
    command: opencode
    args: []
    models:
      fast: anthropic/claude-sonnet # Model for task execution
      slow: anthropic/claude-opus-4-5 # Model for planning
    variant: ""
```

## Default Configuration

If no config file exists, Turbine uses these defaults:

```yaml
defaults:
  backend: opencode
  quiet: false
  retry:
    rotations: 3
    strokes: 3

backends:
  opencode:
    command: opencode
    args: []
    models:
      fast: anthropic/claude-sonnet
      slow: anthropic/claude-opus-4-5

  claude:
    command: claude
    args:
      - -p
      - --dangerously-skip-permissions
    models:
      fast: claude-3-5-sonnet-latest
      slow: claude-4-5-opus-latest
```

## Model Strategy

Turbine uses different models for different operations:

### Slow Model

Used for one-time, high-stakes operations that benefit from higher quality:

- `turbine agents` - Generates AGENTS.md with methodology selection and project guidelines
- `turbine` (planning phase) - Plans the next task into `.turbine/task.yaml`

These operations happen once per project and produce artifacts used throughout development, so quality is prioritized over cost.

**Recommended models:**

- Claude: `claude-4-5-opus-latest`
- OpenCode: `anthropic/claude-opus-4-5`

### Fast Model

Used for repetitive implementation work during task execution:

- `turbine` (execution phase) - Executes the current task from `.turbine/task.yaml`
- Interactive fixes and retries

These operations happen frequently, so cost and speed are prioritized while maintaining sufficient quality for implementation.

**Recommended models:**

- Claude: `claude-3-5-sonnet-latest`
- OpenCode: `anthropic/claude-sonnet`

## Runtime Overrides

Override configuration at runtime using CLI flags:

```bash
# Use a specific backend
turbine --prd PRD.md --backend claude

# Use a specific model for the current operation
turbine agents --prd PRD.md --model claude-opus-4-5

# Use fast variant explicitly
turbine --prd PRD.md --variant fast

# Combine multiple overrides
turbine --backend claude --model claude-3-5-sonnet
```

## Environment Variables

| Variable          | Purpose                            | Example                |
| ----------------- | ---------------------------------- | ---------------------- |
| `XDG_CONFIG_HOME` | Override config directory location | `/home/user/.config`   |
| `NO_COLOR`        | Disable colored terminal output    | (any value to disable) |

## Examples

### Using Claude Backend

```yaml
defaults:
  backend: claude

backends:
  claude:
    command: claude
    args:
      - -p
      - --dangerously-skip-permissions
    models:
      fast: claude-3-5-sonnet-latest
      slow: claude-4-5-opus-latest
```

### Using OpenCode Backend

```yaml
defaults:
  backend: opencode

backends:
  opencode:
    command: opencode
    args: []
    models:
      fast: anthropic/claude-sonnet
      slow: anthropic/claude-opus-4-5
```

### Custom Retry Policy

```yaml
defaults:
  backend: claude
  retry:
    rotations: 5 # More rotations for critical projects
    strokes: 2 # Fewer strokes per rotation
```

### Quiet Mode

```yaml
defaults:
  backend: claude
  quiet: true # Suppress all output except errors
```

## Troubleshooting

### Configuration file not found

Turbine automatically falls back to defaults. To create a config file:

```bash
mkdir -p ~/.config/turbine
cat > ~/.config/turbine/turbine.yaml << 'EOF'
defaults:
  backend: claude

backends:
  claude:
    command: claude
    args: [-p, --dangerously-skip-permissions]
    models:
      fast: claude-3-5-sonnet-latest
      slow: claude-4-5-opus-latest
EOF
```

### Invalid YAML syntax

Check your YAML for proper indentation and syntax. Use an online YAML validator if needed.

### Backend command not found

Ensure the backend command (e.g., `claude` or `opencode`) is installed and available in `$PATH`:

```bash
# For Claude CLI
which claude

# For OpenCode CLI
which opencode
```

### Model not available

Some model names may not be valid for your backend. Check:

- Claude: Valid models include `claude-3-5-sonnet-latest`, `claude-4-5-opus-latest`
- OpenCode: Valid models include `anthropic/claude-sonnet`, `anthropic/claude-opus-4-5`
