package agents

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildAgentsPrompt(t *testing.T) {
	prd := "Build a CLI tool for managing tasks."
	result := buildAgentsPrompt(prd)

	assert.Contains(t, result, prd)
	assert.Contains(t, result, "AGENTS.md generator")
	assert.Contains(t, result, "docs/TESTING.md")
	assert.Contains(t, result, "TDD")
	assert.Contains(t, result, "CLAUDE.md")
}
