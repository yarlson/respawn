package methodologies

// Methodology identifies which methodology to inject
type Methodology int

const (
	MethodologyTDD Methodology = iota
	MethodologyDebuggingLight
	MethodologyDebuggingFull
	MethodologyVerification
	MethodologyPlanning
)

// String returns the methodology name for logging/debugging
func (m Methodology) String() string {
	switch m {
	case MethodologyTDD:
		return "tdd"
	case MethodologyDebuggingLight:
		return "debugging_light"
	case MethodologyDebuggingFull:
		return "debugging_full"
	case MethodologyVerification:
		return "verification"
	case MethodologyPlanning:
		return "planning"
	default:
		return "unknown"
	}
}

// Content returns the prompt content for this methodology
func (m Methodology) Content() string {
	switch m {
	case MethodologyTDD:
		return TDDMethodology
	case MethodologyDebuggingLight:
		return DebuggingLightMethodology
	case MethodologyDebuggingFull:
		return DebuggingFullMethodology
	case MethodologyVerification:
		return VerificationMethodology
	case MethodologyPlanning:
		return PlanningMethodology
	default:
		return ""
	}
}
