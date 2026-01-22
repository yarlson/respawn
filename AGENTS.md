# AGENTS.md — Respawn Engineering Rules

Respawn is a minimal Go-based harness for executing autonomous tasks from `.respawn/tasks.yaml`.

## Core Stack

- **Go 1.22+** (Cobra, slog, gopkg.in/yaml.v3)
- **Local Git** (No push, no branches)

## Primary Commands

- **Runner**: `respawn`
- **Decomposer**: `respawn decompose --prd <path> [--yes]`

## Guidelines (Progressive Disclosure)

Refer to these specialized guides for detailed rules:

- [Go Conventions & Testing](docs/GO_CONVENTIONS.md)
- [Safety & Security Guardrails](docs/SAFETY.md)
- [Architecture & State Management](docs/ARCHITECTURE.md)
- [AI Backend Integration](docs/BACKENDS.md)

---

_Choose the smallest change that preserves safety, keeps behavior explicit, and is easy to reason about from logs._
