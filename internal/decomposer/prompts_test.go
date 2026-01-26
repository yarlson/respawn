package decomposer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildExplorePrompt(t *testing.T) {
	prd := "Build a rocket."
	progress := "- 2026-01-26T00:00:00Z Initialized"
	result := buildExplorePrompt(prd, progress)

	assert.Contains(t, result, prd)
	assert.Contains(t, result, progress)
	assert.Contains(t, result, "codebase analyst")
	assert.Contains(t, result, "Do NOT create any files")
}

func TestBuildPlanPrompt(t *testing.T) {
	prd := "Build a rocket."
	progress := "- 2026-01-26T00:00:00Z Initialized"
	path := ".turbine/task.yaml"
	result := buildPlanPrompt(prd, progress, path)

	assert.Contains(t, result, prd)
	assert.Contains(t, result, progress)
	assert.Contains(t, result, path)
	assert.Contains(t, result, "task planner")
	assert.Contains(t, result, "Task Planning Methodology")
}

func TestBuildPlanFixPrompt(t *testing.T) {
	prd := "Build a rocket."
	progress := "- 2026-01-26T00:00:00Z Initialized"
	failed := "invalid: yaml"
	errMsg := "missing fields"
	result := buildPlanFixPrompt(prd, progress, failed, errMsg)

	assert.Contains(t, result, prd)
	assert.Contains(t, result, progress)
	assert.Contains(t, result, failed)
	assert.Contains(t, result, errMsg)
	assert.Contains(t, result, "Fix Invalid .turbine/task.yaml")
}
