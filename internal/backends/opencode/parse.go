package opencode

import (
	"encoding/json"
	"strings"
)

// OpencodeOutput represents the expected JSON structure from opencode run --format json.
type OpencodeOutput struct {
	SessionID string `json:"session_id"`
	Output    string `json:"output"`
}

// parseSessionID extracts the session ID from the raw output string.
// It looks for a JSON object in the output and attempts to unmarshal it.
func parseSessionID(output string) string {
	// Best-effort JSON extraction.
	// opencode might output other text before or after the JSON.
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "{") {
			continue
		}

		var result OpencodeOutput
		if err := json.Unmarshal([]byte(line), &result); err == nil {
			if result.SessionID != "" {
				return result.SessionID
			}
		}
	}
	return ""
}
