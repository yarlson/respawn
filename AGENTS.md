# AGENTS.md — Turbine Engineering Rules

Turbine is a minimal Go-based harness for planning and executing the next task from `.turbine/task.yaml` and `.turbine/progress.md`.

## Core Stack

- **Go 1.22+** (Cobra, slog, gopkg.in/yaml.v3)
- **Local Git** (No push, no branches)

## Primary Commands

- **Runner**: `turbine --prd <path> [--yes]`
- **AGENTS.md Generator**: `turbine agents --prd <path> [--yes]`

## Guidelines (Progressive Disclosure)

Refer to these specialized guides for detailed rules:

- [Go Conventions & Testing](docs/GO_CONVENTIONS.md)
- [Safety & Security Guardrails](docs/SAFETY.md)
- [Architecture & State Management](docs/ARCHITECTURE.md)
- [AI Backend Integration](docs/BACKENDS.md)

---

_Choose the smallest change that preserves safety, keeps behavior explicit, and is easy to reason about from logs._
