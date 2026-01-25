package run

import (
	"fmt"
	"strings"

	"github.com/yarlson/turbine/internal/tasks"
)

const promptBlockSeparator = "\n\n---\n\n"

const implementerRole = `You are a coding agent working within the Turbine harness.

Your Role
You implement exactly one task at a time. The harness manages task selection, verification, and commits.

Rules
1. Implement ONLY the task described below. Do not work on other tasks.
2. Prefer minimal, surgical changes. Avoid over-engineering.
3. Follow existing codebase patterns and conventions.
4. Run the verification commands relevant to this task and fix any failures.
5. Do NOT commit changes — Turbine will commit after verification passes.

Completion
When done:
- The task acceptance criteria are satisfied
- Verification commands pass
- Stop and let the harness verify and commit`

const retrierRole = `You are a coding agent working within the Turbine harness. This is a retry after verification failure.

Your Role
Fix the verification failure for this task. Focus only on making verification pass.

Rules
1. Analyze the verification failure output carefully.
2. Fix ONLY what is necessary to make verification pass.
3. Make minimal, surgical changes. Do not refactor unrelated code.
4. Run verification commands relevant to the fix.
5. Do NOT commit changes — Turbine will commit after verification passes.`

const methodologyTDD = `# Test-Driven Development

## The Iron Law

NO PRODUCTION CODE WITHOUT A FAILING TEST FIRST

Write code before the test? Delete it. Start over. No exceptions.

## Red-Green-Refactor Cycle

### RED - Write Failing Test

Write one minimal test showing what should happen.

Requirements:
- One behavior per test
- Clear descriptive name
- Real code (no mocks unless unavoidable)

### Verify RED - Watch It Fail

MANDATORY. Never skip.

Run the test. Confirm:
- Test fails (not errors)
- Failure message is expected
- Fails because feature missing (not typos)

Test passes immediately? You're testing existing behavior. Fix test.

### GREEN - Minimal Code

Write simplest code to pass the test.

Don't add features, refactor other code, or "improve" beyond the test.

### Verify GREEN - Watch It Pass

MANDATORY.

Run the test. Confirm:
- Test passes
- Other tests still pass
- Output clean (no errors, warnings)

Test fails? Fix code, not test.

### REFACTOR - Clean Up

After green only:
- Remove duplication
- Improve names
- Extract helpers

Keep tests green. Don't add behavior.

## Common Rationalizations - All Mean DELETE AND START OVER

| Excuse | Reality |
|--------|---------|
| "Too simple to test" | Simple code breaks. Test takes 30 seconds. |
| "I'll test after" | Tests passing immediately prove nothing. |
| "Already manually tested" | Ad-hoc != systematic. No record, can't re-run. |
| "Keep as reference" | You'll adapt it. That's testing after. Delete means delete. |
| "Need to explore first" | Fine. Throw away exploration, start with TDD. |

## Red Flags - STOP and Start Over

- Code written before test
- Test passes immediately
- Can't explain why test failed
- Rationalizing "just this once"
- "Tests after achieve the same purpose"

## Verification Checklist

Before claiming task complete:
- Every new function/method has a test
- Watched each test fail before implementing
- Wrote minimal code to pass each test
- All tests pass
- Output clean (no errors, warnings)

Can't check all boxes? You skipped TDD. Start over.`

const methodologyVerification = `# Verification Before Completion

## The Iron Law

NO COMPLETION CLAIMS WITHOUT FRESH VERIFICATION EVIDENCE

If you haven't run the verification command, you cannot claim it passes.

## The Gate Function

BEFORE claiming any status:

1. IDENTIFY: What command proves this claim?
2. RUN: Execute the FULL command (fresh, complete)
3. READ: Full output, check exit code, count failures
4. VERIFY: Does output confirm the claim?
   - If NO: State actual status with evidence
   - If YES: State claim WITH evidence
5. ONLY THEN: Make the claim

Skip any step = not verified

## Common Failures

| Claim | Requires | Not Sufficient |
|-------|----------|----------------|
| Tests pass | Test command output: 0 failures | Previous run, "should pass" |
| Build succeeds | Build command: exit 0 | Linter passing |
| Bug fixed | Test original symptom: passes | Code changed, assumed fixed |

## Red Flags - STOP

- Using "should", "probably", "seems to"
- About to claim done without running verification
- Relying on partial verification
- Thinking "just this once"

## Rationalization Prevention

| Excuse | Reality |
|--------|---------|
| "Should work now" | RUN the verification |
| "I'm confident" | Confidence != evidence |
| "Just this once" | No exceptions |
| "Partial check is enough" | Partial proves nothing |

## Key Patterns

Tests:
OK [Run test command] [See: 34/34 pass] "All tests pass"
NO "Should pass now" / "Looks correct"

Build:
OK [Run build] [See: exit 0] "Build passes"
NO "Linter passed" (linter doesn't check compilation)

## The Bottom Line

Run the command. Read the output. THEN claim the result.

This is non-negotiable.`

const methodologyDebuggingLight = `# Systematic Debugging (Focused)

## The Iron Law

NO FIXES WITHOUT ROOT CAUSE INVESTIGATION FIRST

If you haven't investigated, you cannot propose fixes.

## Phase 1: Root Cause Investigation

BEFORE attempting ANY fix:

1. **Read Error Messages Carefully**
   - Don't skip past errors or warnings
   - Read stack traces completely
   - Note line numbers, file paths, error codes

2. **Reproduce Consistently**
   - Can you trigger it reliably?
   - What are the exact steps?

3. **Check Recent Changes**
   - What changed that could cause this?
   - Git diff, recent commits

4. **Trace Data Flow**
   - Where does bad value originate?
   - What called this with bad value?
   - Keep tracing up until you find the source
   - Fix at source, not at symptom

## Phase 2: Hypothesis and Testing

1. **Form Single Hypothesis**
   - State clearly: "I think X is the root cause because Y"
   - Be specific, not vague

2. **Test Minimally**
   - Make the SMALLEST possible change to test hypothesis
   - One variable at a time
   - Don't fix multiple things at once

3. **Verify Before Continuing**
   - Did it work? Yes -> implement fix
   - Didn't work? Form NEW hypothesis
   - DON'T add more fixes on top

## Red Flags - STOP and Return to Phase 1

- "Quick fix for now, investigate later"
- "Just try changing X and see if it works"
- "Add multiple changes, run tests"
- "It's probably X, let me fix that"
- Proposing solutions before tracing data flow

## Quick Reference

| Phase | Key Activities | Success Criteria |
| **1. Root Cause** | Read errors, reproduce, check changes | Understand WHAT and WHY |
| **2. Hypothesis** | Form theory, test minimally | Confirmed or new hypothesis |`

const methodologyDebuggingFull = `# Systematic Debugging (Full Process)

## The Iron Law

NO FIXES WITHOUT ROOT CAUSE INVESTIGATION FIRST

If you haven't completed Phase 1, you cannot propose fixes.

## Phase 1: Root Cause Investigation

BEFORE attempting ANY fix:

1. **Read Error Messages Carefully**
   - Don't skip past errors or warnings
   - Read stack traces completely
   - Note line numbers, file paths, error codes

2. **Reproduce Consistently**
   - Can you trigger it reliably?
   - What are the exact steps?
   - If not reproducible -> gather more data, don't guess

3. **Check Recent Changes**
   - What changed that could cause this?
   - Git diff, recent commits
   - Environmental differences

4. **Gather Evidence in Multi-Component Systems**

   WHEN system has multiple components:
   - Log what data enters each component
   - Log what data exits each component
   - Verify environment/config propagation
   - Run once to gather evidence showing WHERE it breaks
   - THEN analyze evidence to identify failing component

5. **Trace Data Flow**
   - Where does bad value originate?
   - What called this with bad value?
   - Keep tracing up until you find the source
   - Fix at source, not at symptom

## Phase 2: Pattern Analysis

1. **Find Working Examples**
   - Locate similar working code in same codebase
   - What works that's similar to what's broken?

2. **Compare Against References**
   - If implementing pattern, read reference implementation COMPLETELY
   - Don't skim - read every line

3. **Identify Differences**
   - What's different between working and broken?
   - List every difference, however small

## Phase 3: Hypothesis and Testing

1. **Form Single Hypothesis**
   - State clearly: "I think X is the root cause because Y"
   - Be specific, not vague

2. **Test Minimally**
   - Make the SMALLEST possible change to test hypothesis
   - One variable at a time
   - Don't fix multiple things at once

3. **Verify Before Continuing**
   - Did it work? Yes -> Phase 4
   - Didn't work? Form NEW hypothesis
   - DON'T add more fixes on top

## Phase 4: Implementation

1. **Create Failing Test Case**
   - Simplest possible reproduction
   - MUST have before fixing

2. **Implement Single Fix**
   - Address the root cause identified
   - ONE change at a time
   - No "while I'm here" improvements

3. **Verify Fix**
   - Test passes now?
   - No other tests broken?

## CRITICAL: Architectural Check

**If this is rotation 2 or 3, previous approaches failed completely.**

STOP and question fundamentals:
- Is this pattern fundamentally sound?
- Are we fighting the wrong architecture?
- Should we try a completely different approach?

Pattern indicating architectural problem:
- Each fix reveals new problem in different place
- Fixes require "massive refactoring" to implement
- Each fix creates new symptoms elsewhere

**Consider: Is the current approach correct, or should we rethink it entirely?**

## Red Flags - STOP and Return to Phase 1

- "Quick fix for now, investigate later"
- "Just try changing X and see if it works"
- "It's probably X, let me fix that"
- Proposing solutions before tracing data flow
- "One more fix attempt" after multiple failures

## Quick Reference

| Phase | Key Activities | Success Criteria |
| **1. Root Cause** | Read errors, reproduce, check changes, gather evidence | Understand WHAT and WHY |
| **2. Pattern** | Find working examples, compare | Identify differences |
| **3. Hypothesis** | Form theory, test minimally | Confirmed or new hypothesis |
| **4. Implementation** | Create test, fix, verify | Bug resolved, tests pass |`

type promptContext struct {
	IsRetry  bool
	Attempt  int
	Rotation int
}

func buildTaskPrompt(ctx promptContext, userPrompt string) string {
	role := implementerRole
	if ctx.IsRetry {
		role = retrierRole
	}

	blocks := []string{role}
	if methods := formatMethodologies(selectMethodologies(ctx)); methods != "" {
		blocks = append(blocks, methods)
	}
	blocks = append(blocks, userPrompt)

	return joinPromptBlocks(blocks...)
}

func selectMethodologies(ctx promptContext) []string {
	if !ctx.IsRetry {
		return []string{methodologyTDD, methodologyVerification}
	}
	if ctx.Rotation > 1 && ctx.Attempt == 1 {
		return []string{methodologyDebuggingFull, methodologyTDD, methodologyVerification}
	}
	return []string{methodologyDebuggingLight, methodologyVerification}
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

func implementUserPrompt(task tasks.Task) string {
	blocks := []string{
		fmt.Sprintf("## Task: %s (%s)\n\n### Description\n%s", task.Title, task.ID, task.Description),
	}

	if len(task.Acceptance) > 0 {
		blocks = append(blocks, formatSection("Acceptance Criteria", formatBulletList(task.Acceptance)))
	}

	if len(task.Verify) > 0 {
		blocks = append(blocks, formatSection("Verification Commands", formatCommandList(task.Verify)))
	}

	return strings.Join(blocks, "\n\n")
}

func retryUserPrompt(task tasks.Task, failureOutput string) string {
	blocks := []string{
		fmt.Sprintf("## Retry Task: %s (%s)\n\n### Failure Output\n```\n%s\n```", task.Title, task.ID, trimFailureOutput(failureOutput)),
		formatSection("Original Task Description", task.Description),
	}

	if len(task.Verify) > 0 {
		blocks = append(blocks, formatSection("Verification Commands", formatCommandList(task.Verify)))
	}

	return strings.Join(blocks, "\n\n")
}

func trimFailureOutput(out string) string {
	const maxLines = 100
	const maxChars = 4096

	lines := strings.Split(out, "\n")
	if len(lines) > maxLines {
		out = strings.Join(lines[len(lines)-maxLines:], "\n")
	}

	if len(out) > maxChars {
		out = "..." + out[len(out)-maxChars:]
	}

	return out
}

func formatSection(title, body string) string {
	value := strings.TrimSpace(body)
	if value == "" {
		return ""
	}
	return fmt.Sprintf("### %s\n%s", title, value)
}

func formatBulletList(items []string) string {
	lines := make([]string, 0, len(items))
	for _, item := range items {
		value := strings.TrimSpace(item)
		if value == "" {
			continue
		}
		lines = append(lines, fmt.Sprintf("- %s", value))
	}
	return strings.Join(lines, "\n")
}

func formatCommandList(items []string) string {
	lines := make([]string, 0, len(items))
	for _, item := range items {
		value := strings.TrimSpace(item)
		if value == "" {
			continue
		}
		lines = append(lines, fmt.Sprintf("- `%s`", value))
	}
	return strings.Join(lines, "\n")
}
