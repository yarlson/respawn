# Safety & Security Guardrails

## File System Boundaries

- Never write outside repo root (except global config under `XDG_CONFIG_HOME` / `~/.config`).
- For any path from user/task input:
  - Use `filepath.Clean`.
  - Reject absolute paths (unless for global config).
  - Ensure resolved path stays under repo root (prefix check on `filepath.Abs`).

## Shelling Out

- Prefer argv execution (`exec.CommandContext`) over `sh -c`.
- **Exception**: Task verification commands run as `/bin/sh -lc "<cmd>"`.
- Capture stdout/stderr to run artifacts.
- Never execute destructive commands like `git clean -fdx` unless explicitly required.

## Git Operations

- Strictly local: no `git push`, no remote modifications, no automatic branches.
- Require clean working tree on start if no resume state exists.
- Commits created only after verification gates pass.
- Format:
- Subject: from `commit_message` in task.yaml
  - Footer: exactly `Turbine: T-001`

## .gitignore Handling

- Ensure `.turbine/runs/` and `.turbine/state/` are ignored.
- Update `.gitignore` idempotently: append missing lines only, preserve newline at EOF.
