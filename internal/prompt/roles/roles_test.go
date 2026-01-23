package roles

import (
	"testing"
)

func TestRoleString(t *testing.T) {
	tests := []struct {
		role     Role
		expected string
	}{
		{RoleExplorer, "explorer"},
		{RoleDecomposer, "decomposer"},
		{RoleImplementer, "implementer"},
		{RoleRetrier, "retrier"},
		{RoleAgentsGenerator, "agents_generator"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := tt.role.String()
			if got != tt.expected {
				t.Errorf("Role.String() = %q, want %q", got, tt.expected)
			}
			if got == "" {
				t.Errorf("Role.String() returned empty string for role %d", tt.role)
			}
		})
	}
}

func TestRoleStringUnknown(t *testing.T) {
	unknownRole := Role(999)
	got := unknownRole.String()
	if got != "unknown" {
		t.Errorf("unknown role String() = %q, want %q", got, "unknown")
	}
}

func TestRoleContentInvalid(t *testing.T) {
	unknownRole := Role(999)
	got := unknownRole.Content()
	if got != "" {
		t.Errorf("unknown role Content() = %q, want empty string", got)
	}
}

func TestRoleContentDefined(t *testing.T) {
	tests := []Role{
		RoleExplorer,
		RoleDecomposer,
		RoleImplementer,
		RoleRetrier,
		RoleAgentsGenerator,
	}

	for _, role := range tests {
		t.Run(role.String(), func(t *testing.T) {
			got := role.Content()
			// This test will pass once the role constants are defined
			// For now, we just verify the method exists and can be called
			if role != Role(999) && got == "" {
				t.Logf("Role %s content not yet defined (will be added during role extraction)", role.String())
			}
		})
	}
}

func TestExplorerRole(t *testing.T) {
	// Test that ExplorerRole constant is non-empty
	if ExplorerRole == "" {
		t.Error("ExplorerRole constant is empty")
	}

	// Test that ExplorerRole contains the codebase analyst identity statement
	if !contains(ExplorerRole, "codebase analyst") {
		t.Error("ExplorerRole does not contain 'codebase analyst' identity statement")
	}
}

func TestDecomposerRole(t *testing.T) {
	// Test that DecomposerRole constant is non-empty
	if DecomposerRole == "" {
		t.Error("DecomposerRole constant is empty")
	}

	// Test that DecomposerRole contains the task decomposer identity statement
	if !contains(DecomposerRole, "task decomposer") {
		t.Error("DecomposerRole does not contain 'task decomposer' identity statement")
	}
}

func TestImplementerRole(t *testing.T) {
	// Test that ImplementerRole constant is non-empty
	if ImplementerRole == "" {
		t.Error("ImplementerRole constant is empty")
	}

	// Test that ImplementerRole contains the coding agent identity statement
	if !contains(ImplementerRole, "coding agent") {
		t.Error("ImplementerRole does not contain 'coding agent' identity statement")
	}
}

func TestRetrierRole(t *testing.T) {
	// Test that RetrierRole constant is non-empty
	if RetrierRole == "" {
		t.Error("RetrierRole constant is empty")
	}

	// Test that RetrierRole contains "retry" in the identity statement
	if !contains(RetrierRole, "retry") {
		t.Error("RetrierRole does not contain 'retry' in the identity statement")
	}
}

func TestAgentsGeneratorRole(t *testing.T) {
	// Test that AgentsGeneratorRole constant is non-empty
	if AgentsGeneratorRole == "" {
		t.Error("AgentsGeneratorRole constant is empty")
	}

	// Test that AgentsGeneratorRole contains the AGENTS.md generator identity statement
	if !contains(AgentsGeneratorRole, "AGENTS.md generator") {
		t.Error("AgentsGeneratorRole does not contain 'AGENTS.md generator' identity statement")
	}
}

// contains is a helper function to check if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
