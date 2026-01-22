````markdown
# Respawn — Extremely Detailed PRD

## 0. Document Control

- **Project:** Respawn
- **Type:** Greenfield CLI (Go)
- **Primary users:** Developers (solo → small team → big team), running locally (not CI).
- **Default mode:** Fully autonomous overnight execution.
- **Key constraint:** After start, Respawn assumes all repo changes are made by Respawn; it must not require human interaction except where explicitly allowed via prompts that can be bypassed with `--yes`.

---

## 1. Overview

Respawn is a Go-based CLI that autonomously drives coding agents (OpenCode and Claude Code) through a resilient “Decompose → Execute → Verify → Commit” loop to implement a PRD end-to-end without human interaction.

Respawn operates from a task graph stored at:

- `./.respawn/tasks.yaml` (tracked; source of truth)

It generates and stores all run artifacts under:

- `./.respawn/runs/` (gitignored; keep everything)
- `./.respawn/state/` (gitignored; persistent resume state)

Respawn uses Git save points: after a task verifies successfully, Respawn commits changes using the task’s Conventional Commit first line plus a mandatory footer:

- `Respawn: T-001`

Respawn is strictly local:
- no branch creation
- no push
- no PR automation
- runs on the current branch

---

## 2. Goals

### G1 — Fully autonomous execution
- Respawn can run unattended overnight and progress through the task DAG.
- No interactive mid-run questions for the user.
- When prompts are needed at startup (e.g., `.gitignore` missing ignores), they must be bypassable with `--yes`.

### G2 — Resumable after interruptions
- If interrupted mid-task (SIGINT, crash, reboot), the next run resumes from persistent state and continues the in-progress task.
- Resume must bypass the “clean working tree” requirement.

### G3 — Correctness gating
- A task is marked `done` only if all task verification commands pass.
- Failed tasks do not stop the run; Respawn continues with any other runnable tasks.

### G4 — Minimal CLI surface (KISS/YAGNI)
- Exactly two commands:
  - `respawn` (runner)
  - `respawn decompose` (PRD → tasks.yaml)
- No “run task by ID” mode.
- No additional subcommands in v1.

### G5 — Backend flexibility
- Shell out to:
  - OpenCode (`opencode run`)
  - Claude Code (`claude`)
- Backend selection precedence:
  1) CLI flags
  2) global config file
  3) default backend (`opencode`)

---

## 3. Non-Goals

- CI integration (explicitly out of scope).
- Windows support (v1).
- Parallel task execution.
- Automated branch creation / merging / rebasing.
- Remote orchestration or push-to-remote.
- Managing “durable guidance” files (e.g., generating `AGENTS.md`, `CLAUDE.md`, `.claude/rules/`). Respawn only warns when absent.
- Storing attempt counters or metadata in `tasks.yaml` (status-only updates).

---

## 4. Primary User Journeys

### J1 — Decompose PRD into tasks (one-time or occasional)
1. Developer writes/has `PRD.md`.
2. Runs:
   - `respawn decompose --prd PRD.md`
3. Respawn:
   - validates `.gitignore` ignores, prompts to add if missing (or auto-add with `--yes`)
   - if `./.respawn/tasks.yaml` exists, prompts to overwrite (or auto-overwrite with `--yes`)
   - shells out to backend to produce `./.respawn/tasks.yaml` with a DAG, verify commands, and commit messages.

### J2 — Run tasks autonomously overnight
1. Developer ensures repo is clean (unless resuming).
2. Runs:
   - `respawn`
3. Respawn:
   - validates `.gitignore` ignores, prompts to add if missing (or auto-add with `--yes`)
   - enforces clean working tree at start (unless resume state exists)
   - executes tasks in DAG order with retry/reset policy
   - commits save points
   - prints a scannable progress log.

### J3 — Resume after interruption
1. Previous run interrupted mid-task.
2. Developer runs:
   - `respawn`
3. Respawn:
   - detects in-progress run state in `./.respawn/state/`
   - resumes immediately, bypassing clean-tree requirement.

---

## 5. UX & Output Requirements (Quiet Mode Default)

### Default output: “quiet but informative”
Respawn must print what is happening (task, progress), optimized for next-morning scanning.

#### Required console events
- Startup summary:
  - repo root path
  - backend/model/variant
  - tasks counts: total / done / runnable / blocked / failed
- Per task:
  - Task start: `TASK T-007 <title>`
  - Session: backend + session id (or best available identifier)
  - Progress transitions: `cycle 1/3 attempt 2/3`
  - Verification summary (each command pass/fail + duration)
  - Failure pointer: path to `.respawn/runs/<run-id>/...`
  - Reset announcement (when moving to next cycle)
  - Success: commit hash
  - Failure after full policy: mark failed + continue
- End summary:
  - done / failed / blocked counts
  - exit code semantics

### Verbosity flags
- `--verbose`: stream full backend output live (still saved to run artifacts).
- `--debug`: include gate stdout/stderr inline in terminal output (still saved to run artifacts).

---

## 6. Tech Stack & Dependencies

- Go 1.22+
- Cobra (CLI)
- Testify (tests)
- `gopkg.in/yaml.v3` (YAML)
- Stdlib:
  - `os/exec`, `context`, `encoding/json`, `log/slog`, `path/filepath`

External requirements:
- `git` on PATH
- `opencode` and/or `claude` on PATH (or configured with explicit command paths)

---

## 7. Filesystem Layout

### Repository-scoped (tracked/ignored)

- `./.respawn/tasks.yaml` (tracked)
- `./.respawn/runs/**` (ignored, keep everything)
- `./.respawn/state/**` (ignored, keep everything)

### Global config (global-only; no per-project config)
Config resolution (non-Windows):
1. If `XDG_CONFIG_HOME` is set:
   - `${XDG_CONFIG_HOME}/respawn/respawn.yaml`
2. Else:
   - `~/.config/respawn/respawn.yaml`

---

## 8. CLI Specification

### 8.1 Runner

**Command:**
- `respawn`

**Flags (v1):**
- `--backend {opencode|claude}`
- `--model <string>` (provider/model format for OpenCode; literal model name for Claude)
- `--variant <string>` (OpenCode only; ignored with a soft warning for Claude)
- `--yes` (auto-accept prompts: `.gitignore` updates, etc.)
- `--verbose`
- `--debug`

### 8.2 Decomposer

**Command:**
- `respawn decompose --prd <path> [--yes]`

**Flags:**
- `--prd <path>` (required)
- same backend selection flags as runner:
  - `--backend`, `--model`, `--variant` (variant applies only to OpenCode)
- `--yes`:
  - overwrite existing `./.respawn/tasks.yaml` without prompting
  - auto-apply `.gitignore` updates without prompting
- `--verbose`, `--debug`

---

## 9. `.gitignore` Validation and Mutation

### 9.1 Required ignores
Respawn requires the repository `.gitignore` to ignore:
- `.respawn/runs/`
- `.respawn/state/`

### 9.2 Validation behavior
At the start of both:
- `respawn`
- `respawn decompose`

Respawn must validate ignores by using Git’s ignore semantics (not homemade globbing). Recommended approach:
- `git check-ignore -q .respawn/runs/`
- `git check-ignore -q .respawn/state/`

### 9.3 Prompt to add missing ignores
If ignores are missing:
- prompt: “`.gitignore` is missing required Respawn ignores. Add them? [y/N]”
- if `--yes`: apply without prompting

### 9.4 Mutation rules (idempotent)
When adding lines:
- append missing lines only
- do not reorder `.gitignore`
- preserve newline at EOF

---

## 10. Git Working Tree Policy

### 10.1 Normal start
If there is no resume state indicating an in-progress run:
- require clean working tree at start
- if dirty: error out (non-zero)

### 10.2 Resume start
If an in-progress run is detected:
- resume anyway (bypass dirty check)

### 10.3 Branching / remotes
- do not create branches
- do not push
- do not touch remotes

---

## 11. Task Graph Model (`./.respawn/tasks.yaml`)

### 11.1 YAML schema (v1)

```yaml
version: 1
tasks:
  - id: T-001
    title: "Short title"
    status: todo|done|failed
    deps: [T-000]               # optional
    description: |              # required for leaf tasks
      Exact instructions with explicit file paths to create/modify.
    acceptance:                 # recommended (3–5 testable lines)
      - "..."
    verify:                     # ordered list of shell command strings
      - "go test ./..."
    commit_message: "feat(cli): add runner skeleton"
````

### 11.2 Task ID format

* `T-001`, `T-002`, … (3-digit, zero-padded)

### 11.3 Status semantics

* `todo`: pending
* `done`: verified and committed
* `failed`: exhausted full retry policy; Respawn moves on

**Important:** Respawn persists **status only** in `tasks.yaml` (no counters, timestamps, etc.).

### 11.4 Dependency semantics

* A task is runnable iff:

  * `status == todo`
  * all `deps` tasks exist and have `status == done`

If a dependency is `failed`, dependent tasks are implicitly blocked forever (remain `todo` but non-runnable).

### 11.5 Task selection tie-breaker

* stable file order in `tasks.yaml` (KISS)

---

## 12. Runner State Machine

### 12.1 High-level loop

For each runnable task:

1. Start task session (new backend session).
2. Execute agent implementation step.
3. Run verification commands.
4. If verification passes:

  * mark task `done`
  * commit save point
  * move to next runnable task
5. If verification fails:

  * retry within cycle up to N attempts (same session)
6. If attempts exhausted:

  * reset repo to last save point
  * start new session
  * continue next cycle
7. If cycles exhausted:

  * mark task `failed`
  * continue to next runnable task

### 12.2 Retry policy

* Two configurable integers:

  * `attempts` (inner attempts per cycle)
  * `cycles` (outer resets)
* Defaults: 3×3
* Same-session rule:

  * retries in same cycle reuse the same backend session
* Reset rule:

  * between cycles Respawn resets changes and starts a new backend session

### 12.3 Reset mechanics (“reset to last save point”)

Reset must return the repo to a known consistent state:

* recommended:

  * `git reset --hard <last_save_point_hash>`
  * ensure index + working tree match save point
* untracked files: do **not** delete (avoid destructive commands); rely on agents not creating junk, and on `.gitignore` for Respawn artifacts.

---

## 13. Verification Execution

### 13.1 Command representation

`verify` in tasks.yaml is a list of shell command strings:

```yaml
verify:
  - "go test ./..."
  - "golangci-lint run"
```

### 13.2 Execution mechanism

* Execute each verify command via:

  * `/bin/sh -lc "<cmd>"`
* Capture stdout/stderr and durations.
* On first failure: stop remaining gates and treat attempt as failed.

### 13.3 Artifacts

Write each gate output to:

* `.respawn/runs/<run-id>/verify/NN.log`

---

## 14. Save Point Commits

### 14.1 Conventional Commit subject line

Per-task field:

* `commit_message: "feat(cli): add runner skeleton"`

Respawn uses it verbatim as the commit subject.

### 14.2 Commit footer

Respawn appends exactly one footer line:

* `Respawn: T-001`

### 14.3 Commit content

Each save point commit must include:

* code changes from task
* `./.respawn/tasks.yaml` updated to `status: done` for this task

### 14.4 Failure commits

No commits on failure.

* If task fails after 3×3, Respawn only updates `status: failed` in `tasks.yaml`.

  * That update is not committed unless it is part of a later successful task’s commit (acceptable) OR you can explicitly choose to commit status-only changes later (not required in v1).
  * Primary constraint: do not break autonomy; do not require manual cleanup.

---

## 15. Resume Behavior

### 15.1 Resume trigger

If `.respawn/state/run.json` exists and indicates a task in progress:

* Respawn resumes the same task
* bypass clean-tree check

### 15.2 Resume strategy

* Attempt to continue backend session if possible (session id known).
* If backend session cannot be resumed:

  * start a new session and re-send the task prompt (same task, same cycle/attempt counters as persisted or best-effort)

### 15.3 Persistence rules

Persistent state lives under:

* `.respawn/state/` (gitignored; keep everything)

State must be sufficient to:

* know active task id
* know current cycle/attempt counts
* know backend + session id (if any)
* know last save point commit hash
* know current run-id artifact directory

---

## 16. Backends

### 16.1 Backend selection precedence

1. CLI flags
2. config
3. default: OpenCode

### 16.2 OpenCode backend (`opencode run`)

Respawn shells out to `opencode run` and uses:

* `--model <provider/model>`
* `--variant <value>` (provider-specific reasoning effort; e.g., high/low/max/minimal)
* session control:

  * within cycle retries: same session (`--continue` or `--session`)
  * between cycles: new session

### 16.3 Claude Code backend (`claude`)

Respawn shells out to Claude Code in fully autonomous mode:

* configured args include permission bypass
* prompt is delivered via stdin
* session continuation is best-effort (depending on available CLI capabilities)

### 16.4 Variant semantics

* `--variant` is only meaningful for OpenCode.
* If `--variant` is set while backend=claude:

  * emit a soft warning and ignore it.

### 16.5 Command configurability

Respawn must allow specifying per-backend command + base args in `respawn.yaml`:

* custom binary path
* wrappers
* fixed args

Example:

```yaml
backends:
  opencode:
    command: "opencode"
    args: ["run", "--format", "json"]
    model: "provider/model"
    variant: "high"
  claude:
    command: "claude"
    args: ["-p", "--dangerously-skip-permissions"]
    model: "claude-model-name"
```

---

## 17. Run Artifacts (`.respawn/runs/`)

### 17.1 Retention

Keep everything. Do not rotate/delete in v1.

### 17.2 Layout (normative)

```
.respawn/runs/<run-id>/
  meta.json
  prompts/
    system.txt
    user.txt
  backend/
    stdout.log
    stderr.log
    session.json
  verify/
    01.log
    02.log
  git/
    status_before.txt
    diff_after_attempt.patch
```

### 17.3 Run ID

* unique and sortable (timestamp-based recommended)
* example: `20260122-103012Z-<short-rand>`

---

## 18. Prompting (Adopted from Ralph)

Prompts are versioned harness contracts. Respawn uses three prompt families:

* Decompose (PRD → tasks.yaml)
* Implement (normal iteration)
* Retry (fix-only after verification failure)

### 18.1 Prompt delivery

* Provide prompts via stdin.
* Prefer a two-part prompt:

  * system prompt (harness rules)
  * user prompt (task or PRD content)
* If backend lacks a strict “system prompt” channel, concatenate:

  * `SYSTEM:\n...\n\nUSER:\n...`

---

### 18.2 Decomposer System Prompt (canonical)

```text
You are Task Decomposer, a PRD→Execution Plan agent.

GOAL
Convert an input PRD (Markdown) into a single YAML file at .respawn/tasks.yaml containing a dependency-aware task DAG that is directly executable by autonomous coding sessions.

EXECUTION MODEL (CRITICAL)
- Each task will be executed by a separate autonomous agent session.
- The executor cannot ask questions or request clarification during execution.
- Tasks must be fully self-contained with all context needed for implementation.
- If the PRD has ambiguity, YOU must decide now. Do NOT create “clarify/decide” tasks.

CORE PRINCIPLES (DRY / KISS / YAGNI)
- Prefer the smallest task graph that is still complete and testable.
- Avoid speculative tasks unless explicitly in-scope in the PRD.
- Use deps to enforce correct execution order.

INPUT
- The user provides PRD.md content. Treat it as the source of truth.

OUTPUT (HARD REQUIREMENTS)
- Output YAML only. No prose, no markdown fences, no commentary.
- The YAML MUST conform to this schema:

version: 1
tasks:
  - id: string (required, unique; format T-001, T-002, ...)
    title: string (required)
    status: todo|done|failed (required; default to todo for all generated tasks)
    deps: [string] (optional; each must reference an existing task id)
    description: string (required for leaf tasks; use YAML block scalar | when >1 line)
    acceptance: [string] (strongly preferred; 3–5 testable statements)
    verify: [string] (optional; ordered list of shell commands as strings)
    commit_message: string (required; Conventional Commits first line)

TASK ATOMICITY (CRITICAL)
- Each task must be small enough to complete in a single session.
- If a requirement implies multiple independently verifiable items, split into separate tasks.

FILE EXPLICITNESS (CRITICAL)
- Every task description MUST specify exact file paths to create or modify.

DEPENDENCY ORDERING (STRICT)
- Scaffolding tasks must come before feature tasks.
- Prefer explicit deps over relying on creation order.

FORBIDDEN PATTERNS
Never generate tasks that:
- Require human decisions (“Decide whether…”, “Clarify…”)
- Are conditional (“If X then… else…”)
- Are open-ended research (“Investigate…”, “Explore…”)
- Bundle unrelated changes
- Lack file specificity
- Are test-only tasks

QUALITY GATE BEFORE FINAL OUTPUT
- All deps reference existing task IDs; no cycles.
- Every task has file-specific description.
- commit_message present for every task and matches Conventional Commits format.

BEGIN
Convert the provided PRD into .respawn/tasks.yaml now.
```

---

### 18.3 Decomposer Fix Prompt Template (validation retry)

```text
The following tasks YAML was generated from this PRD but has validation errors.
Please fix the YAML and output ONLY the corrected YAML (no explanations).

## Original PRD:
{PRD_CONTENT}

## Failed YAML:
{YAML_CONTENT}

## Validation Errors:
{ERRORS}

Output the corrected YAML only:
```

Retry policy for decomposer validation fixes:

* max retries: 2

---

### 18.4 Implementation System Prompt (canonical)

```text
You are a coding agent working within the Respawn harness.

## Your Role
You implement exactly one task at a time. The harness manages task selection, verification, and commits.

## Rules
1. Implement ONLY the task described below. Do not work on other tasks.
2. Prefer minimal, surgical changes. Avoid over-engineering.
3. Follow existing codebase patterns and conventions.
4. Run the verification commands relevant to this task and fix any failures.
5. Do NOT commit changes — Respawn will commit after verification passes.
6. If you add new behavior, add or update tests in the same task.

## Completion
When done:
- The task acceptance criteria are satisfied
- Verification commands pass
- Stop and let the harness verify and commit
```

---

### 18.5 Retry System Prompt (fix-only)

```text
You are a coding agent working within the Respawn harness. This is a RETRY after verification failure.

## Your Role
Fix the verification failure for this task. This is a retry — focus only on fixing the failure.

## Rules
1. Analyze the verification failure output carefully.
2. Fix ONLY what is necessary to make verification pass. Do not add new features.
3. Make minimal, surgical changes. Do not refactor unrelated code.
4. Run verification commands relevant to the fix.
5. Do NOT commit changes — Respawn will commit after verification passes.
```

---

## 19. Decomposer Mode Details

### 19.1 Invocation

* `respawn decompose --prd <path> [--yes]`

### 19.2 Overwrite behavior

* If `./.respawn/tasks.yaml` exists:

  * prompt to overwrite
  * `--yes`: overwrite without prompting

### 19.3 `.gitignore` handling

* Validate required ignores
* If missing:

  * prompt to add
  * `--yes`: add without prompting

### 19.4 Output requirements

Decomposer must generate:

* minimal DAG
* file-specific descriptions
* acceptance criteria
* verify commands when obvious
* `commit_message` per task

---

## 20. Security & Safety Requirements

### 20.1 Filesystem boundary

* Respawn must not write outside:

  * repo root (for `.respawn/**`)
  * config dir (`XDG_CONFIG_HOME` or `~/.config`)
* Reject unsafe paths for `--prd` if desired (best effort):

  * allow reading PRD anywhere, but all writes must remain within repo root.

### 20.2 Shell execution safety

* For verify commands, `/bin/sh -lc` is allowed by design.
* Capture outputs; do not echo secrets.
* Do not perform destructive cleanup commands automatically.

---

## 21. Exit Code Semantics

* Exit `0` if:

  * run completes and there are **no** `failed` tasks
* Exit non-zero if:

  * any task is `failed`
  * or a fatal precondition occurs (missing tasks.yaml in runner mode, missing backend binaries, etc.)

---

## 22. Acceptance Criteria (System-Level)

### Runner

* `respawn` errors if `./.respawn/tasks.yaml` is missing.
* On normal start (no resume):

  * errors if working tree is dirty.
* On resume start:

  * resumes anyway even if dirty.
* Executes tasks in DAG order.
* Applies retry policy 3×3 by default and is configurable.
* Marks task `done` only after all verify commands succeed.
* Commits with:

  * subject from `commit_message`
  * footer `Respawn: T-001`
* Marks task `failed` after exhausting full policy and continues other runnable tasks.
* Writes run artifacts to `.respawn/runs/` and state to `.respawn/state/`.

### Decomposer

* `respawn decompose --prd <path>` generates `./.respawn/tasks.yaml`.
* If tasks.yaml exists:

  * prompts to overwrite, `--yes` overwrites.
* Validates `.gitignore` and prompts to add required ignores (`--yes` applies).
* Validates generated YAML; retries fix up to 2 times on validation errors.

---

## 23. Decisions Made (Complete List)

* CLI is Go + Cobra; state handled in files.
* Two commands only: `respawn` and `respawn decompose`.
* No “run task by ID” mode; fully automatic DAG scheduling.
* Tasks file path is `./.respawn/tasks.yaml` and **must not be overwritten** except in decomposer mode with prompt/`--yes`.
* DAG is required; selection tie-breaker is tasks.yaml order.
* Status-only updates in tasks.yaml.
* Retry policy is 3×3 by default; both numbers configurable.
* Retry in same cycle uses the same backend session.
* Reset between cycles starts a new session (close to a new task).
* If a task fails after full policy: continue with other runnable tasks (B behavior).
* Commit messages:

  * provided per task (`commit_message` string)
  * footer line: `Respawn: T-001`
* `.respawn/` artifacts stored under subdirs and kept:

  * `.respawn/runs/` and `.respawn/state/` (gitignored)
* `.gitignore` must contain ignores for runs/state; Respawn validates and prompts to add (both decompose and run).
* Clean-tree check enforced only at initial run start; after start, don’t enforce between tasks.
* On resume, bypass clean-tree requirement and resume anyway.
* Run on current branch; no branch automation.
* Strictly local; no push.
* `--variant` is OpenCode-only (model variant high/low/etc).
* Backend commands configurable via global config.
* Default output is quiet (progress-oriented); `--verbose` streams backend output.

---

## 24. Appendix: Example tasks.yaml

```yaml
version: 1
tasks:
  - id: T-001
    title: "Initialize Respawn runner skeleton"
    status: todo
    deps: []
    description: |
      Create the CLI entrypoint and runner orchestration skeleton.

      Files:
      - Create cmd/respawn/main.go
      - Create internal/run/runner.go
      - Create internal/tasks/loader.go

      Requirements:
      - Load global config
      - Validate repo and .gitignore ignores
      - Load .respawn/tasks.yaml and print a summary
      - Enforce clean working tree at start (unless resume state exists)
    acceptance:
      - "cmd/respawn/main.go builds and runs `respawn`"
      - "Runner errors with a clear message if .respawn/tasks.yaml is missing"
      - "Runner enforces clean working tree only when not resuming"
      - "go test ./... passes"
    verify:
      - "go test ./..."
      - "go build ./..."
    commit_message: "feat(cli): add runner skeleton"
```
