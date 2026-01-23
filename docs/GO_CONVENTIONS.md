# Go Coding Conventions

## General
- Idiomatic Go naming; prefer explicit types and clear control flow.
- Keep packages small and single-purpose.
- Run `go test ./...` and `golangci-lint run ./... --fix` before completing tasks.
- Ensure `go build ./...` succeeds.

## Errors
- Wrap with context: `fmt.Errorf("do X: %w", err)`
- Define sentinel errors only when callers must branch on them.
- Prefer returning errors over panics.

## Context & Timeouts
- Use `exec.CommandContext(ctx, ...)` for external processes with sensible timeouts.
- Propagate `context.Context` through runner/decomposer/backends.
- Ensure cancellation is honored and logged at INFO.

## Logging
- Use `log/slog` with consistent fields: `task_id`, `run_id`, `backend`, `session_id`, `attempt`, `cycle`, `cmd`.
- Quiet mode for high-level progress; detailed logs to `.turbine/runs/...`.
- Never log secrets or environment dumps.
