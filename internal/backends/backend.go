package backends

import "context"

// Backend defines the interface for interacting with LLM agents/backends.
// It is used by the runner to execute tasks and the decomposer to break down PRDs.
type Backend interface {
	// StartSession initializes a new session. It returns a sessionID
	// that should be used in subsequent Send calls to maintain continuity.
	StartSession(ctx context.Context, opts SessionOptions) (sessionID string, err error)

	// Send transmits a prompt to the backend within the context of a session.
	// It returns a Result containing the backend's response and metadata.
	Send(ctx context.Context, sessionID string, prompt string, opts SendOptions) (*Result, error)
}
