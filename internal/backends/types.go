package backends

// SessionOptions configures a new backend session.
type SessionOptions struct {
	// Model is the primary model identifier (e.g., "claude-3-5-sonnet-20241022").
	Model string
	// Variant is an optional backend-specific variant (e.g., "opencode").
	Variant string
	// WorkingDir is the absolute path to the repository root.
	WorkingDir string
	// ArtifactsDir is the path where stdout/stderr and other run data should be captured.
	ArtifactsDir string
}

// SendOptions configures an individual message sent to the backend.
type SendOptions struct {
	// Model overrides the session's default model for this specific message.
	// If empty, uses the session's model.
	Model string
}

// Result captures the outcome of a Send operation.
type Result struct {
	// Output is the raw text response from the backend.
	Output string
	// Metadata contains backend-specific execution details (e.g., token usage, session IDs).
	Metadata map[string]string
}
