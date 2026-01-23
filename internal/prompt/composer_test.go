package prompt

import (
	"strings"
	"testing"

	"github.com/yarlson/turbine/internal/prompt/methodologies"
	"github.com/yarlson/turbine/internal/prompt/roles"
)

// TestSelectMethodologiesPhaseExplore verifies PhaseExplore returns empty slice
func TestSelectMethodologiesPhaseExplore(t *testing.T) {
	ctx := ExecutionContext{Phase: PhaseExplore}
	result := SelectMethodologies(ctx)
	if len(result) > 0 {
		t.Errorf("Expected empty slice for PhaseExplore, got %v", result)
	}
}

// TestSelectMethodologiesPhaseDecompose verifies PhaseDecompose returns Planning
func TestSelectMethodologiesPhaseDecompose(t *testing.T) {
	ctx := ExecutionContext{Phase: PhaseDecompose}
	result := SelectMethodologies(ctx)
	if len(result) != 1 {
		t.Errorf("Expected 1 methodology for PhaseDecompose, got %d", len(result))
	}
	if len(result) > 0 && result[0] != methodologies.MethodologyPlanning {
		t.Errorf("Expected MethodologyPlanning, got %v", result[0])
	}
}

// TestSelectMethodologiesPhaseGenerateAgents verifies PhaseGenerateAgents returns empty slice
func TestSelectMethodologiesPhaseGenerateAgents(t *testing.T) {
	ctx := ExecutionContext{Phase: PhaseGenerateAgents}
	result := SelectMethodologies(ctx)
	if len(result) > 0 {
		t.Errorf("Expected empty slice for PhaseGenerateAgents, got %v", result)
	}
}

// TestSelectMethodologiesPhaseImplement verifies PhaseImplement returns TDD+Verification
func TestSelectMethodologiesPhaseImplement(t *testing.T) {
	ctx := ExecutionContext{Phase: PhaseImplement}
	result := SelectMethodologies(ctx)
	if len(result) != 2 {
		t.Errorf("Expected 2 methodologies for PhaseImplement, got %d", len(result))
	}
	if len(result) >= 2 {
		if result[0] != methodologies.MethodologyTDD {
			t.Errorf("Expected first methodology to be MethodologyTDD, got %v", result[0])
		}
		if result[1] != methodologies.MethodologyVerification {
			t.Errorf("Expected second methodology to be MethodologyVerification, got %v", result[1])
		}
	}
}

// TestSelectMethodologiesPhaseRetryNewRotation verifies PhaseRetry with new rotation (Rotation>1, Attempt=1)
func TestSelectMethodologiesPhaseRetryNewRotation(t *testing.T) {
	ctx := ExecutionContext{
		Phase:    PhaseRetry,
		Rotation: 2,
		Attempt:  1,
	}
	result := SelectMethodologies(ctx)
	if len(result) != 3 {
		t.Errorf("Expected 3 methodologies for PhaseRetry new rotation, got %d", len(result))
	}
	if len(result) >= 3 {
		if result[0] != methodologies.MethodologyDebuggingFull {
			t.Errorf("Expected first methodology to be MethodologyDebuggingFull, got %v", result[0])
		}
		if result[1] != methodologies.MethodologyTDD {
			t.Errorf("Expected second methodology to be MethodologyTDD, got %v", result[1])
		}
		if result[2] != methodologies.MethodologyVerification {
			t.Errorf("Expected third methodology to be MethodologyVerification, got %v", result[2])
		}
	}
}

// TestSelectMethodologiesPhaseRetrySameRotation verifies PhaseRetry within same rotation
func TestSelectMethodologiesPhaseRetrySameRotation(t *testing.T) {
	ctx := ExecutionContext{
		Phase:    PhaseRetry,
		Rotation: 1,
		Attempt:  2,
	}
	result := SelectMethodologies(ctx)
	if len(result) != 2 {
		t.Errorf("Expected 2 methodologies for PhaseRetry same rotation, got %d", len(result))
	}
	if len(result) >= 2 {
		if result[0] != methodologies.MethodologyDebuggingLight {
			t.Errorf("Expected first methodology to be MethodologyDebuggingLight, got %v", result[0])
		}
		if result[1] != methodologies.MethodologyVerification {
			t.Errorf("Expected second methodology to be MethodologyVerification, got %v", result[1])
		}
	}
}

// TestComposeWithMethodologies verifies Compose correctly assembles role, methodologies, and task
func TestComposeWithMethodologies(t *testing.T) {
	// Create a slice with two methodologies that have actual content
	meths := []methodologies.Methodology{
		methodologies.MethodologyPlanning,
		methodologies.MethodologyVerification,
	}
	taskContent := "Review this code"
	role := roles.RoleExplorer

	result := Compose(role, meths, taskContent)

	// Verify structure
	if !strings.Contains(result, "# Required Methodologies") {
		t.Error("Missing methodology header")
	}
	if !strings.Contains(result, "Review this code") {
		t.Error("Missing task content")
	}

	// Verify separators
	if !strings.Contains(result, "---") {
		t.Error("Missing separator")
	}
}

// TestComposeWithoutMethodologies verifies Compose handles empty methodology list
func TestComposeWithoutMethodologies(t *testing.T) {
	var meths []methodologies.Methodology
	taskContent := "Write this feature"
	role := roles.RoleExplorer

	result := Compose(role, meths, taskContent)

	// Verify structure
	if strings.Contains(result, "# Required Methodologies") {
		t.Error("Should not have methodology header when no methodologies")
	}
	if !strings.Contains(result, "Write this feature") {
		t.Error("Missing task content")
	}
}

// TestComposeWithoutTaskContent verifies Compose handles empty task content
func TestComposeWithoutTaskContent(t *testing.T) {
	meths := []methodologies.Methodology{
		methodologies.MethodologyPlanning,
	}
	taskContent := ""
	role := roles.RoleExplorer

	result := Compose(role, meths, taskContent)

	// Verify structure
	if !strings.Contains(result, "# Required Methodologies") {
		t.Error("Missing methodology header")
	}

	// Should not have extra separator at the end
	if strings.HasSuffix(strings.TrimSpace(result), "---") {
		t.Error("Should not have trailing separator when no task content")
	}
}

// TestPhaseString verifies Phase.String() returns correct names
func TestPhaseString(t *testing.T) {
	tests := []struct {
		phase    Phase
		expected string
	}{
		{PhaseExplore, "explore"},
		{PhaseDecompose, "decompose"},
		{PhaseImplement, "implement"},
		{PhaseRetry, "retry"},
		{PhaseGenerateAgents, "generate_agents"},
		{Phase(999), "unknown"},
	}

	for _, test := range tests {
		if test.phase.String() != test.expected {
			t.Errorf("Phase %v: expected %q, got %q", test.phase, test.expected, test.phase.String())
		}
	}
}
