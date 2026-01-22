package prompt

import (
	"fmt"
	"strings"

	"respawn/internal/tasks"
)

// DecomposeUserPrompt builds the user prompt for the decomposition task.
func DecomposeUserPrompt(prdContent, outputPath string) string {
	return fmt.Sprintf("## PRD Content:\n\n%s\n\n## Output Path:\n%s", prdContent, outputPath)
}

// DecomposeFixPrompt builds the user prompt for fixing invalid decomposition YAML.
func DecomposeFixPrompt(prdContent, failedYAML, validationError string) string {
	var b strings.Builder

	b.WriteString("## Task: Fix Invalid .respawn/tasks.yaml\n\n")
	b.WriteString("The generated YAML is invalid. Please fix it based on the PRD and validation errors.\n\n")

	b.WriteString("### PRD Content\n")
	b.WriteString(prdContent)
	b.WriteString("\n\n")

	b.WriteString("### Failed YAML\n")
	b.WriteString("```yaml\n")
	b.WriteString(failedYAML)
	b.WriteString("\n```\n\n")

	b.WriteString("### Validation Errors\n")
	b.WriteString("```\n")
	b.WriteString(validationError)
	b.WriteString("\n```\n\n")

	b.WriteString("Return ONLY the corrected YAML. No prose, no markdown fences.")

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
