# Replay Migration Plan

This plan migrates Turbine from handwritten agent calls to replay workflows while
preserving current behavior (including session reuse via --continue).

## Phase 0 — Replay Readiness
- Format modified replay files (at minimum `replay/executor.go`).
- Run `go test ./...` in `/Users/yaroslavk/home/replay`.
- Add an executor test that asserts `RunParams.Continue` is set on step 2 in a
  session (update `replay/executor_test.go`).
- Optional: add `Continue bool` to `replay.Step` and honor it in the executor to
  explicitly request `--continue` on single-step workflows when retrying.

## Phase 1 — Provider Wiring in Turbine
- Add `github.com/yarlson/replay` to `turbine/go.mod`.
- Create `internal/replayprovider` to map Turbine config to replay providers:
  - `config.Backend.Command` -> `provider.Config.Executable`
  - `config.Backend.Args` -> `provider.Config.Args`
  - `config.Backend.Env` (if added) -> `provider.Config.Env`
- Replace `internal/backends` usage in `cmd/turbine/root.go` with replay provider
  creation.

## Phase 2 — Workflow Core + Store Adapter
- Implement `store.Store` adapter under Turbine (e.g., `internal/replaystore`) to
  persist:
  - Workflow state -> `.turbine/state/replay.json`
  - Step summaries -> `.turbine/state/steps.json`
  - Events -> `.turbine/runs/<runID>/events.log`
  - Raw payloads -> `.turbine/runs/<runID>/raw/<step>.ndjson`
- Update runner initialization to create `replay.Executor` with
  `replay.WithStore(...)`.

## Phase 3 — Decomposer via Workflow
- Replace direct backend calls in `internal/decomposer/decomposer.go` with a
  replay workflow:
  - Single session, two steps: explore (fast model) -> decompose (slow model).
  - `Session.SystemPrompt = prompt.Compose(roles.RoleExplorer/RoleDecomposer, ...)`.
  - `Step.Prompt = prompt.ExploreUserPrompt` and `prompt.DecomposeUserPrompt`.
- Capture Step 1 text into workflow context for Step 2 template rendering.
- Preserve validation retries by rerunning only the decompose step with
  `prompt.DecomposeFixPrompt`.

## Phase 4 — Agents Generation via Workflow
- Replace `internal/agents/agents.go` with a one-step workflow:
  - `Session.SystemPrompt = prompt.Compose(roles.RoleAgentsGenerator, ...)`.
  - `Step.Prompt = prompt.AgentsUserPrompt(...)`.
- Move `validateOutput` into `Step.PostHook` so missing outputs fail the step.

## Phase 5 — Task Execution via Workflow
- In `internal/run/execute_task.go`, build a workflow per attempt:
  - `Workflow.WorkingDir = repoRoot`.
  - `Workflow.Model = model`, `Workflow.Variant = variant`.
  - `Session.SystemPrompt = prompt.Compose(role, meths, "")`.
  - `Step.Prompt = prompt.ImplementUserPrompt(...)` or
    `prompt.RetryUserPrompt(...)`.
- Run verification in `Step.PostHook` using `RunVerification`; return error on
  failure to trigger the next stroke in Turbine's retry loop.
- Capture `EventKindText` into `lastFailureOutput` for retries (event hook or
  collector).
- If Step-level continue exists, set it when `stroke > 1` to reuse sessions.

## Phase 6 — CLI + UX Alignment
- Keep `turbine run` UX stable by mapping replay events to existing UI output in
  `internal/ui` (start/end markers, failure summaries).
- Update `docs/BACKENDS.md` and `docs/ARCHITECTURE.md` to document replay as the
  execution engine.

## Phase 7 — Cleanup + Tests
- Update unit tests around decomposer, agents, and runner to validate workflow
  usage and state persistence.
- Remove or quarantine `internal/backends` once replay is fully wired.
