package prompt

import (
	"strings"
	"testing"

	"respawn/internal/tasks"

	"github.com/stretchr/testify/assert"
)

func TestSystemPrompts(t *testing.T) {
	assert.Contains(t, DecomposerSystemPrompt, "You are Task Decomposer")
	assert.Contains(t, DecomposerSystemPrompt, ".respawn/tasks.yaml")
	assert.Contains(t, ImplementSystemPrompt, "You are a coding agent working within the Respawn harness.")
	assert.Contains(t, RetrySystemPrompt, "RETRY after verification failure.")
}

func TestDecomposeUserPrompt(t *testing.T) {
	prd := "Build a rocket."
	path := ".respawn/tasks.yaml"
	prompt := DecomposeUserPrompt(prd, path)

	assert.Contains(t, prompt, prd)
	assert.Contains(t, prompt, path)
	assert.Contains(t, prompt, "## PRD Content:")
}

func TestImplementUserPrompt(t *testing.T) {
	task := tasks.Task{
		ID:          "T-001",
		Title:       "Test Task",
		Description: "Do something.",
		Acceptance:  []string{"It works."},
		Verify:      []string{"go test ./..."},
	}

	prompt := ImplementUserPrompt(task)

	assert.Contains(t, prompt, "T-001")
	assert.Contains(t, prompt, "Test Task")
	assert.Contains(t, prompt, "Do something.")
	assert.Contains(t, prompt, "It works.")
	assert.Contains(t, prompt, "go test ./...")
	assert.Contains(t, prompt, "### Description")
	assert.Contains(t, prompt, "### Acceptance Criteria")
	assert.Contains(t, prompt, "### Verification Commands")
}

func TestRetryUserPrompt(t *testing.T) {
	task := tasks.Task{
		ID:          "T-001",
		Title:       "Test Task",
		Description: "Do something.",
		Verify:      []string{"go test ./..."},
	}
	failureOutput := "Error: something went wrong\nLine 2\nLine 3"

	prompt := RetryUserPrompt(task, failureOutput)

	assert.Contains(t, prompt, "T-001")
	assert.Contains(t, prompt, "Failure Output")
	assert.Contains(t, prompt, failureOutput)
	assert.Contains(t, prompt, "Do something.")
	assert.Contains(t, prompt, "go test ./...")
}

func TestTrimFailureOutput(t *testing.T) {
	// Test line trimming
	longOutput := ""
	for i := 0; i < 200; i++ {
		longOutput += "line\n"
	}
	trimmed := trimFailureOutput(longOutput)
	lines := strings.Split(strings.TrimSpace(trimmed), "\n")
	assert.LessOrEqual(t, len(lines), 100)

	// Test char trimming
	hugeOutput := strings.Repeat("a", 10000)
	trimmed = trimFailureOutput(hugeOutput)
	assert.LessOrEqual(t, len(trimmed), 4096+3) // +3 for "..."
	assert.True(t, strings.HasPrefix(trimmed, "..."))
}

func TestDeterministicFormatting(t *testing.T) {
	task := tasks.Task{
		ID:          "T-001",
		Title:       "Test",
		Description: "Desc",
	}

	p1 := ImplementUserPrompt(task)
	p2 := ImplementUserPrompt(task)

	assert.Equal(t, p1, p2)
}
