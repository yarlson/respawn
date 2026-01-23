# Architecture & Core Logic

## Overview
Build a minimal, reliable harness that executes tasks autonomously from `./.turbine/tasks.yaml` and optimizes for unattended overnight operation.

## Core Principles
- **DRY / KISS / YAGNI**: Simplest correct implementation; avoid speculative abstractions.
- **Minimal Dependencies**: Favor Go stdlib.
- **Surgical Changes**: Small diffs; change only what is required.

## Repository Structure
- Paths are relative to repo root.
- `.turbine/tasks.yaml` is the source of truth for tasks.
- Run artifacts: `./.turbine/runs/` (gitignored).
- Resume state: `./.turbine/state/` (gitignored).

## Determinism & Resume
- Persist only essentials: run id, active task id, rotation/stroke, backend session id, artifact paths.
- Resume must be robust across restarts; if session resume fails, start new session but continue task.
- Never mutate `tasks.yaml` beyond the `status` field.
