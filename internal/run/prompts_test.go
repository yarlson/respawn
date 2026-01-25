package run

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yarlson/turbine/internal/tasks"
)

func TestBuildTaskPrompt_ImplementationContainsTDD(t *testing.T) {
	ctx := promptContext{
		IsRetry:  false,
		Attempt:  1,
		Rotation: 1,
	}
	result := buildTaskPrompt(ctx, "Task: build something")

	assert.Contains(t, result, "coding agent")
	assert.Contains(t, result, "Test-Driven Development")
	assert.Contains(t, result, "Iron Law")
	assert.Contains(t, result, "Verification Before Completion")
	assert.Contains(t, result, "Task: build something")
}

func TestBuildTaskPrompt_RetryContainsDebuggingLight(t *testing.T) {
	ctx := promptContext{
		IsRetry:  true,
		Attempt:  2,
		Rotation: 1,
	}
	result := buildTaskPrompt(ctx, "Fix the failure")

	assert.Contains(t, result, "retry")
	assert.Contains(t, result, "Systematic Debugging")
	assert.Contains(t, result, "Root Cause Investigation")
	assert.NotContains(t, result, "Architectural Check")
}

func TestBuildTaskPrompt_RetryNewRotationContainsFullDebugging(t *testing.T) {
	ctx := promptContext{
		IsRetry:  true,
		Attempt:  1,
		Rotation: 2,
	}
	result := buildTaskPrompt(ctx, "Fresh rotation")

	assert.Contains(t, result, "Architectural Check")
	assert.Contains(t, result, "Test-Driven Development")
}

func TestImplementUserPrompt(t *testing.T) {
	task := tasks.Task{
		ID:          "T-001",
		Title:       "Test Task",
		Description: "Do something.",
		Acceptance:  []string{"It works."},
		Verify:      []string{"go test ./..."},
	}

	prompt := implementUserPrompt(task)

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

	prompt := retryUserPrompt(task, failureOutput)

	assert.Contains(t, prompt, "T-001")
	assert.Contains(t, prompt, "Failure Output")
	assert.Contains(t, prompt, failureOutput)
	assert.Contains(t, prompt, "Do something.")
	assert.Contains(t, prompt, "go test ./...")
}

func TestTrimFailureOutput(t *testing.T) {
	longOutput := ""
	for i := 0; i < 200; i++ {
		longOutput += "line\n"
	}
	trimmed := trimFailureOutput(longOutput)
	lines := strings.Split(strings.TrimSpace(trimmed), "\n")
	assert.LessOrEqual(t, len(lines), 100)

	hugeOutput := strings.Repeat("a", 10000)
	trimmed = trimFailureOutput(hugeOutput)
	assert.LessOrEqual(t, len(trimmed), 4096+3)
	assert.True(t, strings.HasPrefix(trimmed, "..."))
}

func TestDeterministicFormatting(t *testing.T) {
	task := tasks.Task{
		ID:          "T-001",
		Title:       "Test",
		Description: "Desc",
	}

	p1 := implementUserPrompt(task)
	p2 := implementUserPrompt(task)

	assert.Equal(t, p1, p2)
}
