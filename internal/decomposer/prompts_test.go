package decomposer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildExplorePrompt(t *testing.T) {
	prd := "Build a rocket."
	result := buildExplorePrompt(prd)

	assert.Contains(t, result, prd)
	assert.Contains(t, result, "codebase analyst")
	assert.Contains(t, result, "Do NOT create any files")
}

func TestBuildDecomposePrompt(t *testing.T) {
	prd := "Build a rocket."
	path := ".turbine/tasks.yaml"
	result := buildDecomposePrompt(prd, path)

	assert.Contains(t, result, prd)
	assert.Contains(t, result, path)
	assert.Contains(t, result, "task decomposer")
	assert.Contains(t, result, "Task Planning Methodology")
}

func TestBuildDecomposeFixPrompt(t *testing.T) {
	prd := "Build a rocket."
	failed := "invalid: yaml"
	errMsg := "missing fields"
	result := buildDecomposeFixPrompt(prd, failed, errMsg)

	assert.Contains(t, result, prd)
	assert.Contains(t, result, failed)
	assert.Contains(t, result, errMsg)
	assert.Contains(t, result, "Fix Invalid .turbine/tasks.yaml")
}
