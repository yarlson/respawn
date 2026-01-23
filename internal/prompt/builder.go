package prompt

import (
	"fmt"
	"strings"

	"github.com/yarlson/turbine/internal/tasks"
)

// ExploreUserPrompt builds the user prompt for Phase 1: codebase exploration.
func ExploreUserPrompt(prdContent string) string {
	return fmt.Sprintf(`## PRD to Implement:

%s

## Instructions

Explore this repository to understand its patterns and conventions BEFORE we create tasks.

Focus on:
1. Is this greenfield or an existing project?
2. What development patterns are used (TDD, testing frameworks, etc.)?
3. What coding conventions should new code follow?
4. How are commits typically structured?

Do NOT create any files. Just explore and summarize your findings.`, prdContent)
}

// DecomposeUserPrompt builds the user prompt for Phase 2: task generation.
func DecomposeUserPrompt(prdContent, outputPath string) string {
	return fmt.Sprintf(`## PRD Content:

%s

## Instructions

Now create the tasks file based on your exploration findings.

Write the tasks file to: %s

Use your file writing tools to create the file. Do NOT output YAML as text.
Create the .turbine directory first if it doesn't exist: mkdir -p .turbine

IMPORTANT: Apply the patterns and conventions you discovered during exploration.`, prdContent, outputPath)
}

// DecomposeFixPrompt builds the user prompt for fixing invalid decomposition YAML.
func DecomposeFixPrompt(prdContent, failedYAML, validationError string) string {
	var b strings.Builder

	b.WriteString("## Task: Fix Invalid .turbine/tasks.yaml\n\n")
	b.WriteString("The generated YAML file is invalid. Fix the file directly using your file writing tools.\n\n")

	b.WriteString("### PRD Content\n")
	b.WriteString(prdContent)
	b.WriteString("\n\n")

	b.WriteString("### Current File Content (Invalid)\n")
	b.WriteString("```yaml\n")
	b.WriteString(failedYAML)
	b.WriteString("\n```\n\n")

	b.WriteString("### Validation Errors\n")
	b.WriteString("```\n")
	b.WriteString(validationError)
	b.WriteString("\n```\n\n")

	b.WriteString("Fix the errors and overwrite .turbine/tasks.yaml with the corrected content.\n")
	b.WriteString("Use your file writing tools. Do NOT output YAML as text.")

	return b.String()
}

// ImplementUserPrompt builds the user prompt for an implementation task.
func ImplementUserPrompt(task tasks.Task) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("## Task: %s (%s)\n\n", task.Title, task.ID))
	b.WriteString(fmt.Sprintf("### Description\n%s\n\n", task.Description))

	if len(task.Acceptance) > 0 {
		b.WriteString("### Acceptance Criteria\n")
		for _, ac := range task.Acceptance {
			b.WriteString(fmt.Sprintf("- %s\n", ac))
		}
		b.WriteString("\n")
	}

	if len(task.Verify) > 0 {
		b.WriteString("### Verification Commands\n")
		for _, v := range task.Verify {
			b.WriteString(fmt.Sprintf("- `%s`\n", v))
		}
		b.WriteString("\n")
	}

	return b.String()
}

// RetryUserPrompt builds the user prompt for a retry after failure.
func RetryUserPrompt(task tasks.Task, failureOutput string) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("## Retry Task: %s (%s)\n\n", task.Title, task.ID))
	b.WriteString("### Failure Output\n")
	b.WriteString("```\n")
	b.WriteString(trimFailureOutput(failureOutput))
	b.WriteString("\n```\n\n")

	b.WriteString("### Original Task Description\n")
	b.WriteString(task.Description)
	b.WriteString("\n\n")

	if len(task.Verify) > 0 {
		b.WriteString("### Verification Commands\n")
		for _, v := range task.Verify {
			b.WriteString(fmt.Sprintf("- `%s`\n", v))
		}
		b.WriteString("\n")
	}

	return b.String()
}

// trimFailureOutput limits the failure output to a reasonable size (e.g., last 100 lines or 4KB).
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
