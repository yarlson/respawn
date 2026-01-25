package decomposer

import (
	"fmt"
	"strings"
)

const promptBlockSeparator = "\n\n---\n\n"

const explorerRole = `You are a codebase analyst. Your job is to explore this repository and understand its patterns, conventions, and development practices.

Your Task
Explore the codebase to gather context that will inform task planning. Do NOT create any files yet.

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

How to Explore
- Use file listing and reading tools to examine the codebase
- Check for configuration files (package.json, go.mod, Makefile, etc.)
- Look at existing code to understand patterns
- Read any documentation files

Output
Summarize your findings. This context will be used in the next step to generate appropriate tasks.`

const decomposerRole = `You are a task decomposer. Convert PRDs into executable task plans.

Your Task
Convert a PRD (Markdown) into a YAML file containing a dependency-aware task DAG.

You MUST write the file directly using your file writing tools. Do NOT output YAML as text.

Required Actions
1. Create directory if needed: mkdir -p .turbine
2. Write the tasks file to: .turbine/tasks.yaml

Execution Model
- Each task will be executed by a separate autonomous agent session.
- The executor cannot ask questions or request clarification during execution.
- Tasks must be fully self-contained with all context needed for implementation.
- If the PRD has ambiguity, YOU must decide now. Do NOT create "clarify/decide" tasks.

Principles
- Prefer the smallest task graph that is still complete and testable.
- Avoid speculative tasks unless explicitly in-scope in the PRD.
- Use deps to enforce correct execution order.
- Apply patterns and conventions discovered during codebase exploration.

YAML Schema
The file MUST conform to this schema:

version: 1
tasks:
  - id: string (required, unique; format T-001, T-002, ...)
    title: string (required)
    status: todo|done|failed (required; default to todo for all generated tasks)
    deps: [string] (optional; each must reference an existing task id)
    description: string (required for leaf tasks; use YAML block scalar | when >1 line)
    acceptance: [string] (strongly preferred; 3-5 testable statements)
    verify: [string] (optional; ordered list of shell commands as strings)
    commit_message: string (required; Conventional Commits first line)

YAML Syntax Rules
- When array items contain quotes, quote the ENTIRE value.
  - Bad:  - "./calc 5 + 3" outputs "8"  (partial quotes = invalid YAML)
  - Good: - '"./calc 5 + 3" outputs "8"' (entire value single-quoted)
  - Good: - ./calc 5 + 3 outputs 8       (no quotes at all)

Task Size
- Each task must be small enough to complete in a single session.
- If a requirement implies multiple independently verifiable items, split into separate tasks.

File Paths
- Every task description MUST specify exact file paths to create or modify.

Dependencies
- Scaffolding tasks must come before feature tasks.
- Prefer explicit deps over relying on creation order.

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
- All deps reference existing task IDs; no cycles.
- Every task has file-specific description.
- commit_message present for every task and matches Conventional Commits format.
- Verify commands are consistent with file paths.

BEGIN
Read the PRD below and write .turbine/tasks.yaml now.`

const methodologyPlanning = "# Task Planning Methodology\n\n" +
	"## Core Principle\n\n" +
	"Write tasks assuming the executor has zero context for the codebase. Document everything needed: which files to touch, exact code, how to test. Give bite-sized tasks. DRY. YAGNI. TDD. Frequent commits.\n\n" +
	"## Bite-Sized Task Granularity\n\n" +
	"Each task step is one action (2-5 minutes):\n" +
	"- \"Write the failing test\" - one step\n" +
	"- \"Run it to make sure it fails\" - one step\n" +
	"- \"Implement the minimal code to make the test pass\" - one step\n" +
	"- \"Run the tests and make sure they pass\" - one step\n" +
	"- \"Commit\" - one step\n\n" +
	"## Task Requirements\n\n" +
	"Every task MUST include:\n\n" +
	"1. **Exact file paths**\n" +
	"   - Create: `exact/path/to/file.go`\n" +
	"   - Modify: `exact/path/to/existing.go`\n" +
	"   - Test: `exact/path/to/file_test.go`\n\n" +
	"2. **Complete code** (not \"add validation\")\n" +
	"   - Show the actual code to write\n" +
	"   - No ambiguity\n\n" +
	"3. **Exact commands with expected output**\n" +
	"   - Run: `go test ./path/...`\n" +
	"   - Expected: PASS or specific failure message\n\n" +
	"4. **Clear acceptance criteria**\n" +
	"   - 3-5 testable statements\n" +
	"   - Each can be verified independently\n\n" +
	"## Task Sizing Rules\n\n" +
	"- Each task completable in a single session\n" +
	"- If requirement implies multiple verifiable items, split into separate tasks\n" +
	"- Scaffolding tasks before feature tasks\n" +
	"- Explicit deps over relying on creation order\n\n" +
	"## Do Not Create Tasks That:\n\n" +
	"- Require human decisions (\"Decide whether...\")\n" +
	"- Are conditional (\"If X then... else...\")\n" +
	"- Are open-ended research (\"Investigate...\")\n" +
	"- Bundle unrelated changes\n" +
	"- Lack file specificity\n" +
	"- Are test-only (tests should be part of feature tasks)\n\n" +
	"## Verification Commands\n\n" +
	"- Each verify command must exit 0 on success\n" +
	"- Verify commands must match file structure in task\n" +
	"- To verify output contains text: `cmd 2>&1 | grep -q \"expected\"`\n" +
	"- To verify command exits non-zero: `! cmd`\n" +
	"- Do NOT combine `!` with grep\n\n" +
	"## Before Finalizing\n\n" +
	"Validate that:\n" +
	"- All deps reference existing task IDs; no cycles\n" +
	"- Every task has file-specific description\n" +
	"- commit_message present for every task (Conventional Commits)\n" +
	"- Verify commands consistent with file paths"

func buildExplorePrompt(prdContent string) string {
	return buildPrompt(explorerRole, nil, exploreUserPrompt(prdContent))
}

func buildDecomposePrompt(prdContent, outputPath string) string {
	return buildPrompt(decomposerRole, []string{methodologyPlanning}, decomposeUserPrompt(prdContent, outputPath))
}

func buildDecomposeFixPrompt(prdContent, failedYAML, validationError string) string {
	return buildPrompt(decomposerRole, []string{methodologyPlanning}, decomposeFixPrompt(prdContent, failedYAML, validationError))
}

func buildPrompt(role string, methods []string, userPrompt string) string {
	blocks := []string{role}
	if methodsBlock := formatMethodologies(methods); methodsBlock != "" {
		blocks = append(blocks, methodsBlock)
	}
	blocks = append(blocks, userPrompt)
	return joinPromptBlocks(blocks...)
}

func joinPromptBlocks(blocks ...string) string {
	trimmed := make([]string, 0, len(blocks))
	for _, block := range blocks {
		value := strings.TrimSpace(block)
		if value == "" {
			continue
		}
		trimmed = append(trimmed, value)
	}
	return strings.Join(trimmed, promptBlockSeparator)
}

func formatMethodologies(methods []string) string {
	items := make([]string, 0, len(methods))
	for _, method := range methods {
		value := strings.TrimSpace(method)
		if value == "" {
			continue
		}
		items = append(items, value)
	}
	if len(items) == 0 {
		return ""
	}

	intro := "# Required Methodologies\n\nFollow these methodologies for this task:"
	return joinPromptBlocks(intro, strings.Join(items, promptBlockSeparator))
}

func exploreUserPrompt(prdContent string) string {
	return fmt.Sprintf("## PRD to Implement:\n\n%s\n\n## Instructions\n\nExplore this repository to understand its patterns and conventions BEFORE we create tasks.\n\nFocus on:\n1. Is this greenfield or an existing project?\n2. What development patterns are used (TDD, testing frameworks, etc.)?\n3. What coding conventions should new code follow?\n4. How are commits typically structured?\n\nDo NOT create any files. Just explore and summarize your findings.", prdContent)
}

func decomposeUserPrompt(prdContent, outputPath string) string {
	return fmt.Sprintf("## PRD Content:\n\n%s\n\n## Instructions\n\nNow create the tasks file based on your exploration findings.\n\nWrite the tasks file to: %s\n\nUse your file writing tools to create the file. Do NOT output YAML as text.\nCreate the .turbine directory first if it doesn't exist: mkdir -p .turbine\n\nIMPORTANT: Apply the patterns and conventions you discovered during exploration.", prdContent, outputPath)
}

func decomposeFixPrompt(prdContent, failedYAML, validationError string) string {
	return fmt.Sprintf("## Task: Fix Invalid .turbine/tasks.yaml\n\n"+
		"The generated YAML file is invalid. Fix the file directly using your file writing tools.\n\n"+
		"### PRD Content\n%s\n\n"+
		"### Current File Content (Invalid)\n```yaml\n%s\n```\n\n"+
		"### Validation Errors\n```\n%s\n```\n\n"+
		"Fix the errors and overwrite .turbine/tasks.yaml with the corrected content.\n"+
		"Use your file writing tools. Do NOT output YAML as text.", prdContent, failedYAML, validationError)
}
