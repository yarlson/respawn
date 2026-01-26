package decomposer

import "fmt"

const explorerRole = `You are a codebase analyst. Your job is to explore this repository and understand its patterns, conventions, and development practices.

Your Task
Explore the codebase and the current progress log to gather context that will inform task planning. Do NOT create any files yet.

What to Look For
1. **Project type**: Is this a greenfield project (empty/minimal) or an existing codebase?
2. **Repository structure**: Is this a monorepo (multiple services/packages) or a single-service repo?
   - Monorepo indicators: multiple go.mod/package.json files, services/ or packages/ directories, workspace configurations
   - Single repo indicators: one main module, single entry point, unified build
3. **Project structure**: Directory layout, module organization, entry points
4. **Development methodology**: Does it use TDD? Are there existing tests? What testing patterns?
5. **Coding conventions**: Naming patterns, file organization, code style
6. **Build system**: How is the project built? What tools are used?
7. **Documentation**: AGENTS.md, README, docs/ folder, inline comments
8. **Dependencies**: What frameworks/libraries are used?
9. **Git history**: Recent commit message patterns (Conventional Commits?)
10. **Progress log**: What has already been completed? What remains?

How to Explore
- Use file listing and reading tools to examine the codebase
- Check for configuration files (package.json, go.mod, Makefile, etc.)
- Look at existing code to understand patterns
- Read any documentation files

Output
Summarize your findings. This context will be used in the next step to generate the next task.`

const plannerRole = `You are a just-in-time task planner. Convert PRDs and progress logs into a single executable task.

Your Task
Convert a PRD (Markdown) and progress log into a YAML file containing the next task to execute, just-in-time.

You MUST write the file directly using your file writing tools. Do NOT output YAML as text.

Required Actions
1. Create directory if needed: mkdir -p .turbine
2. Write the task file to: .turbine/task.yaml

Execution Model
- The task will be executed by a separate autonomous agent session.
- The executor cannot ask questions or request clarification during execution.
- The task must be fully self-contained with all context needed for implementation.
- If the PRD has ambiguity, YOU must decide now. Do NOT create "clarify/decide" tasks.

Principles
- Prefer the smallest task that is still complete and testable.
- Avoid speculative tasks unless explicitly in-scope in the PRD.
- Apply patterns and conventions discovered during codebase exploration and the progress log.

YAML Schema
The file MUST conform to this schema:

version: 1
task:
  id: string (required, unique; format T-001, T-002, ...)
  title: string (required)
  status: todo|done|failed (required; default to todo)
  description: string (required; use YAML block scalar | when >1 line)
  acceptance: [string] (strongly preferred; 3-5 testable statements)
  verify: [string] (optional; ordered list of shell commands as strings)
  commit_message: string (required for todo/failed; Conventional Commits first line)

Completion
- If the PRD is fully implemented given the progress log, set status: done.
- Provide a short description explaining why there is no remaining work.
- commit_message may be empty for status: done.

YAML Syntax Rules
- When array items contain quotes, quote the ENTIRE value.
  - Bad:  - "./calc 5 + 3" outputs "8"  (partial quotes = invalid YAML)
  - Good: - '"./calc 5 + 3" outputs "8"' (entire value single-quoted)
  - Good: - ./calc 5 + 3 outputs 8       (no quotes at all)

Task Size
- Each task must be small enough to complete in a single session.
- If a requirement implies multiple independently verifiable items, split into separate tasks.

File Paths
- The task description MUST specify exact file paths to create or modify.

Do Not Generate Tasks That:
- Require human decisions ("Decide whether...", "Clarify...")
- Are conditional ("If X then... else...")
- Are open-ended research ("Investigate...", "Explore...")
- Bundle unrelated changes
- Lack file specificity
- Are test-only tasks

Verify Commands
- Each verify command must exit 0 on success.
- Verify commands must be consistent with the file structure defined in the task.
- If source code is in a subdirectory (e.g., cmd/app/main.go), build outputs must not collide with directory names.
- To verify output contains text: cmd 2>&1 | grep -q "expected"
- To verify command exits non-zero: ! cmd
- IMPORTANT: Do NOT combine ! with grep. "! cmd | grep -q text" means "grep should NOT find text" (usually wrong).
  Instead, use separate commands:
    - ! cmd              # verify cmd exits non-zero
    - cmd 2>&1 | grep -q text  # verify output contains text

Before Writing
Validate that:
- The task has a file-specific description.
- commit_message present for todo/failed tasks and matches Conventional Commits format.
- Verify commands are consistent with file paths.

BEGIN
Read the PRD and progress log below and write .turbine/task.yaml now.`

const methodologyPlanning = `# Task Planning Methodology

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

## Do Not Create Tasks That:

- Require human decisions ("Decide whether...")
- Are conditional ("If X then... else...")
- Are open-ended research ("Investigate...")
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
- The task has a file-specific description
- commit_message present for todo/failed tasks (Conventional Commits)
- Verify commands consistent with file paths`

const methodHeader = "# Required Methodologies\n\nFollow these methodologies for this task:"

const explorePromptTemplate = "%s\n\n---\n\n" +
	"## PRD to Implement:\n\n%s\n\n" +
	"## Progress Log:\n\n%s\n\n" +
	"## Instructions\n\n" +
	"Explore this repository and progress log to understand its patterns and conventions BEFORE we create the next task.\n\n" +
	"Focus on:\n" +
	"1. Is this greenfield or an existing project?\n" +
	"2. What development patterns are used (TDD, testing frameworks, etc.)?\n" +
	"3. What coding conventions should new code follow?\n" +
	"4. How are commits typically structured?\n" +
	"5. What has already been completed per the progress log?\n\n" +
	"Do NOT create any files. Just explore and summarize your findings."

const planPromptTemplate = "%s\n\n---\n\n" +
	methodHeader + "\n\n" + methodologyPlanning + "\n\n---\n\n" +
	"## PRD Content:\n\n%s\n\n" +
	"## Progress Log:\n\n%s\n\n" +
	"## Instructions\n\n" +
	"Now create the next task just-in-time based on your exploration findings and the progress log.\n\n" +
	"Write the task file to: %s\n\n" +
	"Use your file writing tools to create the file. Do NOT output YAML as text.\n" +
	"Create the .turbine directory first if it doesn't exist: mkdir -p .turbine\n\n" +
	"IMPORTANT: Apply the patterns and conventions you discovered during exploration."

const planFixPromptTemplate = "%s\n\n---\n\n" +
	methodHeader + "\n\n" + methodologyPlanning + "\n\n---\n\n" +
	"## Task: Fix Invalid .turbine/task.yaml\n\n" +
	"The generated YAML file is invalid. Fix the file directly using your file writing tools.\n\n" +
	"### PRD Content\n%s\n\n" +
	"### Progress Log\n%s\n\n" +
	"### Current File Content (Invalid)\n```yaml\n%s\n```\n\n" +
	"### Validation Errors\n```\n%s\n```\n\n" +
	"Fix the errors and overwrite .turbine/task.yaml with the corrected content.\n" +
	"Use your file writing tools. Do NOT output YAML as text."

func buildExplorePrompt(prdContent, progressContent string) string {
	return fmt.Sprintf(explorePromptTemplate, explorerRole, prdContent, progressContent)
}

func buildPlanPrompt(prdContent, progressContent, outputPath string) string {
	return fmt.Sprintf(planPromptTemplate, plannerRole, prdContent, progressContent, outputPath)
}

func buildPlanFixPrompt(prdContent, progressContent, failedYAML, validationError string) string {
	return fmt.Sprintf(planFixPromptTemplate, plannerRole, prdContent, progressContent, failedYAML, validationError)
}
