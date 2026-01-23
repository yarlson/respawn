package prompt

import (
	"strings"

	"github.com/yarlson/turbine/internal/prompt/methodologies"
	"github.com/yarlson/turbine/internal/prompt/roles"
)

// Phase represents the current execution phase
type Phase int

const (
	PhaseExplore Phase = iota
	PhaseDecompose
	PhaseImplement
	PhaseRetry
	PhaseGenerateAgents
)

// String returns the phase name for logging
func (p Phase) String() string {
	switch p {
	case PhaseExplore:
		return "explore"
	case PhaseDecompose:
		return "decompose"
	case PhaseImplement:
		return "implement"
	case PhaseRetry:
		return "retry"
	case PhaseGenerateAgents:
		return "generate_agents"
	default:
		return "unknown"
	}
}

// ExecutionContext provides context for methodology selection
type ExecutionContext struct {
	Phase    Phase
	Attempt  int // 1, 2, 3 within rotation (called "stroke" in runner)
	Rotation int // 1, 2, 3
}

// SelectMethodologies returns the methodologies to inject based on execution context.
// Selection rules from PRD:
// - Phase=Explore: none
// - Phase=Decompose: Planning
// - Phase=GenerateAgents: none (self-contained role)
// - Phase=Implement: TDD, Verification
// - Phase=Retry, Rotation>1, Attempt=1: DebuggingFull, TDD, Verification
// - Phase=Retry (other): DebuggingLight, Verification
func SelectMethodologies(ctx ExecutionContext) []methodologies.Methodology {
	switch ctx.Phase {
	case PhaseExplore:
		return nil
	case PhaseDecompose:
		return []methodologies.Methodology{methodologies.MethodologyPlanning}
	case PhaseGenerateAgents:
		return nil
	case PhaseImplement:
		return []methodologies.Methodology{
			methodologies.MethodologyTDD,
			methodologies.MethodologyVerification,
		}
	case PhaseRetry:
		if ctx.Rotation > 1 && ctx.Attempt == 1 {
			// Fresh rotation after complete failure - full debugging + architectural review
			return []methodologies.Methodology{
				methodologies.MethodologyDebuggingFull,
				methodologies.MethodologyTDD,
				methodologies.MethodologyVerification,
			}
		}
		// Retry within same rotation - focused debugging
		return []methodologies.Methodology{
			methodologies.MethodologyDebuggingLight,
			methodologies.MethodologyVerification,
		}
	default:
		return nil
	}
}

// Compose assembles a final prompt from role, methodologies, and task content.
// Format:
//
//	[Role Content]
//	---
//	# Required Methodologies
//	[Methodology 1 Content]
//	---
//	[Methodology 2 Content]
//	---
//	[User/Task Content]
func Compose(role roles.Role, meths []methodologies.Methodology, taskContent string) string {
	var b strings.Builder

	// Role content
	b.WriteString(role.Content())

	// Methodologies (if any)
	if len(meths) > 0 {
		b.WriteString("\n\n---\n\n")
		b.WriteString("# Required Methodologies\n\n")
		b.WriteString("Follow these methodologies for this task:\n\n")

		for i, m := range meths {
			b.WriteString(m.Content())
			if i < len(meths)-1 {
				b.WriteString("\n\n---\n\n")
			}
		}
	}

	// Task content
	if taskContent != "" {
		b.WriteString("\n\n---\n\n")
		b.WriteString(taskContent)
	}

	return b.String()
}
