package roles

// ExplorerRole is defined in explorer.go
// DecomposerRole is defined in decomposer.go
// ImplementerRole is defined in implementer.go
// RetrierRole is defined in retrier.go
// AgentsGeneratorRole is defined in agents_generator.go

// Role identifies which agent role to use
type Role int

const (
	RoleExplorer Role = iota
	RoleDecomposer
	RoleImplementer
	RoleRetrier
	RoleAgentsGenerator
)

// String returns the role name for logging/debugging
func (r Role) String() string {
	switch r {
	case RoleExplorer:
		return "explorer"
	case RoleDecomposer:
		return "decomposer"
	case RoleImplementer:
		return "implementer"
	case RoleRetrier:
		return "retrier"
	case RoleAgentsGenerator:
		return "agents_generator"
	default:
		return "unknown"
	}
}

// Content returns the prompt content for this role
func (r Role) Content() string {
	switch r {
	case RoleExplorer:
		return ExplorerRole
	case RoleDecomposer:
		return DecomposerRole
	case RoleImplementer:
		return ImplementerRole
	case RoleRetrier:
		return RetrierRole
	case RoleAgentsGenerator:
		return AgentsGeneratorRole
	default:
		return ""
	}
}
