package methodologies

import (
	"strings"
	"testing"
)

func TestMethodologyString(t *testing.T) {
	tests := []struct {
		methodology Methodology
		expected    string
	}{
		{MethodologyTDD, "tdd"},
		{MethodologyDebuggingLight, "debugging_light"},
		{MethodologyDebuggingFull, "debugging_full"},
		{MethodologyVerification, "verification"},
		{MethodologyPlanning, "planning"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.methodology.String()
			if result != tt.expected {
				t.Errorf("String() = %q, want %q", result, tt.expected)
			}
			if result == "" {
				t.Errorf("String() returned empty string for %v", tt.methodology)
			}
		})
	}
}

func TestMethodologyStringNonEmpty(t *testing.T) {
	methodologies := []Methodology{
		MethodologyTDD,
		MethodologyDebuggingLight,
		MethodologyDebuggingFull,
		MethodologyVerification,
		MethodologyPlanning,
	}

	for _, m := range methodologies {
		result := m.String()
		if result == "" {
			t.Errorf("String() returned empty string for methodology constant %v", m)
		}
	}
}

func TestMethodologyContent(t *testing.T) {
	tests := []struct {
		methodology Methodology
		name        string
	}{
		{MethodologyTDD, "TDD"},
		{MethodologyDebuggingLight, "DebuggingLight"},
		{MethodologyDebuggingFull, "DebuggingFull"},
		{MethodologyVerification, "Verification"},
		{MethodologyPlanning, "Planning"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.methodology.Content()
			if result == "" {
				t.Errorf("Content() returned empty string for %s", tt.name)
			}
		})
	}
}

func TestMethodologyContentInvalid(t *testing.T) {
	// Test that an invalid methodology returns empty string
	invalidMethodology := Methodology(999)
	result := invalidMethodology.Content()
	if result != "" {
		t.Errorf("Content() for invalid methodology should return empty string, got %q", result)
	}
}

func TestMethodologyStringInvalid(t *testing.T) {
	// Test that an invalid methodology returns "unknown"
	invalidMethodology := Methodology(999)
	result := invalidMethodology.String()
	if result != "unknown" {
		t.Errorf("String() for invalid methodology should return \"unknown\", got %q", result)
	}
}

func TestTDDMethodology(t *testing.T) {
	if TDDMethodology == "" {
		t.Error("TDDMethodology constant is empty")
	}
}

func TestTDDMethodologyContent(t *testing.T) {
	if !strings.Contains(TDDMethodology, "Iron Law") {
		t.Error("TDDMethodology does not contain 'Iron Law' core principle")
	}
}

func TestTDDMethodologyRedGreenRefactor(t *testing.T) {
	if !strings.Contains(TDDMethodology, "Red-Green-Refactor") {
		t.Error("TDDMethodology does not contain 'Red-Green-Refactor' cycle")
	}
}

func TestDebuggingLightMethodology(t *testing.T) {
	if DebuggingLightMethodology == "" {
		t.Error("DebuggingLightMethodology constant is empty")
	}
}

func TestDebuggingLightMethodologyIronLaw(t *testing.T) {
	if !strings.Contains(DebuggingLightMethodology, "Iron Law") {
		t.Error("DebuggingLightMethodology does not contain 'Iron Law' core principle")
	}
}

func TestDebuggingFullMethodology(t *testing.T) {
	if DebuggingFullMethodology == "" {
		t.Error("DebuggingFullMethodology constant is empty")
	}
}

func TestDebuggingFullMethodologyIronLaw(t *testing.T) {
	if !strings.Contains(DebuggingFullMethodology, "Iron Law") {
		t.Error("DebuggingFullMethodology does not contain 'Iron Law' core principle")
	}
}

func TestDebuggingFullMethodologyArchitecturalCheck(t *testing.T) {
	if !strings.Contains(DebuggingFullMethodology, "Architectural Check") {
		t.Error("DebuggingFullMethodology does not contain 'Architectural Check'")
	}
}

func TestVerificationMethodology(t *testing.T) {
	if VerificationMethodology == "" {
		t.Error("VerificationMethodology constant is empty")
	}
}

func TestVerificationMethodologyIronLaw(t *testing.T) {
	if !strings.Contains(VerificationMethodology, "Iron Law") {
		t.Error("VerificationMethodology does not contain 'Iron Law' core principle")
	}
}

func TestVerificationMethodologyGateFunction(t *testing.T) {
	if !strings.Contains(VerificationMethodology, "Gate Function") {
		t.Error("VerificationMethodology does not contain 'Gate Function'")
	}
}

func TestPlanningMethodology(t *testing.T) {
	if PlanningMethodology == "" {
		t.Error("PlanningMethodology constant is empty")
	}
}

func TestPlanningMethodologyCorePrinciple(t *testing.T) {
	if !strings.Contains(PlanningMethodology, "Core Principle") {
		t.Error("PlanningMethodology does not contain 'Core Principle'")
	}
}

func TestPlanningMethodologyBiteSizedTaskGranularity(t *testing.T) {
	if !strings.Contains(PlanningMethodology, "Bite-Sized Task Granularity") {
		t.Error("PlanningMethodology does not contain 'Bite-Sized Task Granularity'")
	}
}
