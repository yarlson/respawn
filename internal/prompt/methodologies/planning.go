package methodologies

// PlanningMethodology contains task planning guidance for decomposition.
const PlanningMethodology = `# Task Planning Methodology

## Core Principle

Write tasks assuming the executor has zero context for the codebase. Document everything needed: which files to touch, exact code, how to test. Give bite-sized tasks. DRY. YAGNI. TDD. Frequent commits.

## Bite-Sized Task Granularity

Each task step is one action (2-5 minutes):
- "Write the failing test" - one step
- "Run it to make sure it fails" - one step  
- "Implement the minimal code to make the test pass" - one step
- "Run the tests and make sure they pass" - one step
- "Commit" - one step

## Task Requirements

Every task MUST include:

1. **Exact file paths**
   - Create: ` + "`exact/path/to/file.go`" + `
   - Modify: ` + "`exact/path/to/existing.go`" + `
   - Test: ` + "`exact/path/to/file_test.go`" + `

2. **Complete code** (not "add validation")
   - Show the actual code to write
   - No ambiguity

3. **Exact commands with expected output**
   - Run: ` + "`go test ./path/...`" + `
   - Expected: PASS or specific failure message

4. **Clear acceptance criteria**
   - 3-5 testable statements
   - Each can be verified independently

## Task Sizing Rules

- Each task completable in a single session
- If requirement implies multiple verifiable items, split into separate tasks
- Scaffolding tasks before feature tasks
- Explicit deps over relying on creation order

## Do Not Create Tasks That:

- Require human decisions ("Decide whether…")
- Are conditional ("If X then… else…")
- Are open-ended research ("Investigate…")
- Bundle unrelated changes
- Lack file specificity
- Are test-only (tests should be part of feature tasks)

## Verification Commands

- Each verify command must exit 0 on success
- Verify commands must match file structure in task
- To verify output contains text: ` + "`cmd 2>&1 | grep -q \"expected\"`" + `
- To verify command exits non-zero: ` + "`! cmd`" + `
- Do NOT combine ` + "`!`" + ` with grep

## Before Finalizing

Validate that:
- All deps reference existing task IDs; no cycles
- Every task has file-specific description
- commit_message present for every task (Conventional Commits)
- Verify commands consistent with file paths`
