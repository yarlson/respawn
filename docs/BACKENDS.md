# AI Backend Integration

## Backend Invocation
- Binaries and arguments must be configurable in `respawn.yaml`.
- Fail with a clear message if backend is not on PATH.

## OpenCode
- New task → new session.
- Retry-in-cycle → same session (`--continue`/`--session`).
- Reset cycle → new session.
- `--variant` applies only to OpenCode.

## Claude Code
- Run headless (`-p`) and fully autonomous (permission bypass per PRD).

## CLI UX
- Output must be scannable: task transitions, gate summaries, artifact pointers.
- Avoid chatty logs; `--verbose` streams backend output.
- All interactive prompts must be skippable with `--yes`.
