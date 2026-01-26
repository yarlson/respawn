# Architecture & Core Logic

## Overview

Build a minimal, reliable harness that plans and executes tasks autonomously from `./.turbine/task.yaml` with progress tracked in `./.turbine/progress.md`.

## Core Principles

- **DRY / KISS / YAGNI**: Simplest correct implementation; avoid speculative abstractions.
- **Minimal Dependencies**: Favor Go stdlib.
- **Surgical Changes**: Small diffs; change only what is required.

## Repository Structure

- Paths are relative to repo root.
- `.turbine/task.yaml` is the source of truth for the current task.
- `.turbine/archive/` stores completed task files.
- `.turbine/progress.md` captures narrative progress.
- Run artifacts: `./.turbine/runs/` (gitignored).
- Resume state: `./.turbine/state/` (gitignored).

## Determinism & Resume

- Persist only essentials: run id, active task id, rotation/stroke, backend session id, artifact paths.
- Resume must be robust across restarts; if session resume fails, start new session but continue task.
- Never mutate `task.yaml` beyond the `status` field during execution.
