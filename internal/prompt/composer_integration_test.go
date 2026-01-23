package prompt

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yarlson/turbine/internal/prompt/roles"
)

func TestComposerIntegration_ImplementationPromptContainsTDD(t *testing.T) {
	ctx := ExecutionContext{
		Phase:    PhaseImplement,
		Attempt:  1,
		Rotation: 1,
	}
	meths := SelectMethodologies(ctx)
	prompt := Compose(roles.RoleImplementer, meths, "Task: build something")

	assert.Contains(t, prompt, "coding agent")
	assert.Contains(t, prompt, "Test-Driven Development")
	assert.Contains(t, prompt, "Iron Law")
	assert.Contains(t, prompt, "Verification Before Completion")
	assert.Contains(t, prompt, "Task: build something")
}

func TestComposerIntegration_RetryPromptContainsDebugging(t *testing.T) {
	ctx := ExecutionContext{
		Phase:    PhaseRetry,
		Attempt:  2,
		Rotation: 1,
	}
	meths := SelectMethodologies(ctx)
	prompt := Compose(roles.RoleRetrier, meths, "Fix the failure")

	assert.Contains(t, prompt, "retry")
	assert.Contains(t, prompt, "Systematic Debugging")
	assert.Contains(t, prompt, "Root Cause Investigation")
	assert.NotContains(t, prompt, "Architectural Check") // Light debugging
}

func TestComposerIntegration_NewRotationPromptContainsFullDebugging(t *testing.T) {
	ctx := ExecutionContext{
		Phase:    PhaseRetry,
		Attempt:  1,
		Rotation: 2,
	}
	meths := SelectMethodologies(ctx)
	prompt := Compose(roles.RoleRetrier, meths, "Fresh rotation")

	assert.Contains(t, prompt, "Architectural Check") // Full debugging
	assert.Contains(t, prompt, "Test-Driven Development")
}

func TestComposerIntegration_DecomposePromptContainsPlanning(t *testing.T) {
	ctx := ExecutionContext{Phase: PhaseDecompose}
	meths := SelectMethodologies(ctx)
	prompt := Compose(roles.RoleDecomposer, meths, "PRD content here")

	assert.Contains(t, prompt, "task decomposer")
	assert.Contains(t, prompt, "Task Planning Methodology")
	assert.Contains(t, prompt, "Bite-Sized")
}

func TestComposerIntegration_ExplorePromptHasNoMethodologies(t *testing.T) {
	ctx := ExecutionContext{Phase: PhaseExplore}
	meths := SelectMethodologies(ctx)
	prompt := Compose(roles.RoleExplorer, meths, "Explore")

	assert.Contains(t, prompt, "codebase analyst")
	assert.NotContains(t, prompt, "Required Methodologies")
}

func TestComposerIntegration_AgentsPromptHasNoMethodologies(t *testing.T) {
	ctx := ExecutionContext{Phase: PhaseGenerateAgents}
	meths := SelectMethodologies(ctx)
	prompt := Compose(roles.RoleAgentsGenerator, meths, "Generate")

	assert.Contains(t, prompt, "AGENTS.md generator")
	assert.NotContains(t, prompt, "Required Methodologies")
}
