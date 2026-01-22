# AGENTS.md — Respawn Engineering Rules

This file defines durable guidance for coding agents working on the **Respawn** codebase. Treat it as the project’s “house style” and safety policy.

## Mission

Build a minimal, reliable harness that:

- Reads `./.respawn/tasks.yaml` (tracked) and executes tasks autonomously.
- Stores run artifacts under `./.respawn/runs/` and resume state under `./.respawn/state/` (both gitignored).
- Uses local Git only (no branches, no push).
- Shells out to OpenCode / Claude Code and to verification commands.
- Optimizes for unattended overnight operation.

## Core Principles (DRY / KISS / YAGNI)

- Prefer the simplest correct implementation.
- Avoid abstractions until they pay rent (interfaces only where multiple implementations already exist).
- Favor stdlib over new dependencies.
- Small diffs: change only what the task requires.

## Baseline Tech

- Go 1.22+
- CLI: Cobra
- Tests: standard `testing` + `testify` where helpful
- YAML: `gopkg.in/yaml.v3`
- Logging: `log/slog` (structured, minimal)

## Repo Conventions

- Paths are relative to repo root.
- `.respawn/tasks.yaml` is the source of truth for tasks; update `status` only.
- `.respawn/runs/**` and `.respawn/state/**` are **always** gitignored.
- Default commands:
  - Runner: `respawn`
  - Decomposer: `respawn decompose --prd <path> [--yes]`

## Coding Style (Go)

### General

- `gofmt` clean; idiomatic Go naming; avoid cleverness.
- Keep packages small and single-purpose.
- Prefer explicit types and clear control flow over reflection/magic.

### Errors

- Wrap with context: `fmt.Errorf("do X: %w", err)`
- Define sentinel errors only when callers must branch on them.
- Prefer returning errors over panics.

### Context & Timeouts

- Every external process must be `exec.CommandContext(ctx, ...)` with a sensible timeout at the callsite.
- Propagate `context.Context` through runner/decomposer/backends.
- Ensure cancellation is honored and logged at INFO.

### Logging

- Use `slog` with consistent fields: `task_id`, `run_id`, `backend`, `session_id`, `attempt`, `cycle`, `cmd`.
- Quiet mode prints high-level progress; detailed logs go to `.respawn/runs/...`.
- Never log secrets or full environment dumps.

### CLI UX

- Output must be scannable in overnight logs:
  - task start/end, cycle/attempt changes, gate summary, artifact pointers.
- Avoid chatty logs by default; `--verbose` streams backend output.
- All interactive prompts must be skippable with `--yes`.

## Safety & Security Guardrails

Respawn is a tool that executes shell commands and invokes agents. Enforce safety by construction:

### File System Boundaries

- Never write outside repo root (except global config under `XDG_CONFIG_HOME` / `~/.config`).
- For any path from user/task input:
  - `filepath.Clean`
  - reject absolute paths (unless explicitly expected for global config)
  - ensure resolved path stays under repo root (prefix check on `filepath.Abs`)

### Shelling Out

- Prefer argv execution (`exec.CommandContext`) over `sh -c`.
- **Exception**: task verification commands are strings by design; run them as:
  - `/bin/sh -lc "<cmd>"`
- Always capture stdout/stderr to run artifacts (even in verbose mode).
- Never execute destructive repo-wide commands (e.g., `git clean -fdx`) unless the PRD explicitly requires it.

### Git Operations

- Must be strictly local:
  - no `git push`, no remote modifications
  - no automatic branches
- At normal start: if no resume state exists, require clean working tree and fail fast.
- If resume state exists: resume anyway (ignore dirtiness).
- Commits are created only after verification gates pass.
- Commit format:
  - Subject: from `commit_message` in tasks.yaml
  - Footer: exactly `Respawn: T-001`

### `.gitignore` Handling

- Validate that these are ignored:
  - `.respawn/runs/`
  - `.respawn/state/`
- If missing, prompt to add; `--yes` auto-adds.
- Update `.gitignore` idempotently:
  - do not reorder
  - append missing lines only
  - preserve newline at EOF

### Backend Invocation

- OpenCode:
  - new task → new session
  - retry-in-cycle → same session (`--continue`/`--session`)
  - reset cycle → new session
  - `--variant` applies only to OpenCode
- Claude Code:
  - run headless (`-p`) and fully autonomous (permission bypass per PRD)
- Backend binaries/args must be configurable in `respawn.yaml`.
- If backend isn’t found on PATH, fail with a clear message.

## Testing & Verification (Respawn Codebase)

Always add or update tests when behavior changes.

Minimum checks before considering a task done:

- `go test ./...`
- `golangci-lint run ./... --fix`
- Build succeeds: `go build ./...`

Always fix all errors or warnings reported by the linter before proceeding.

When changing:

- task DAG logic: add focused unit tests (deps, runnable selection, blocked behavior)
- `.gitignore` validation: add tests using temp repos and deterministic fixtures
- command execution: use fake runners or injectable exec layer; avoid flaky integration tests by default

## Determinism & Resume

- Persist only what is needed to resume safely (run id, active task id, cycle/attempt, backend + session id, last save point hash, artifact paths).
- Resume should be robust across restarts:
  - if backend session cannot be resumed, start a new session and continue the same task.
- Never mutate `.respawn/tasks.yaml` beyond `status`.

## What Not To Do

- Do not add new subcommands beyond `decompose` without explicit PRD change.
- Do not introduce background daemons, watchers, or concurrency.
- Do not add speculative features (“nice to have”) not required by tasks/PRD.
- Do not “improve” architecture during retries; keep changes surgical.

## When in Doubt

Choose the smallest change that:

- preserves safety boundaries,
- keeps behavior explicit,
- and is easy to reason about from logs the next morning.
