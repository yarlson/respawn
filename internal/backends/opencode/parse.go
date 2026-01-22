package opencode

import (
	"encoding/json"
	"strings"
)

// OpencodeOutput represents the expected JSON structure from opencode run --format json.
// OpenCode outputs NDJSON with different message types.
type OpencodeOutput struct {
	Type      string `json:"type"`
	SessionID string `json:"sessionID"`
	Part      struct {
		Type      string `json:"type"`
		Text      string `json:"text"`
		SessionID string `json:"sessionID"`
	} `json:"part"`
}

// ParseOutput extracts text content and session ID from opencode NDJSON output.
func ParseOutput(output string) (text string, sessionID string) {
	var textParts []string

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "{") {
			continue
		}

		var result OpencodeOutput
		if err := json.Unmarshal([]byte(line), &result); err != nil {
			continue
		}

		// Capture session ID from any message
		if result.SessionID != "" && sessionID == "" {
			sessionID = result.SessionID
		}

		// Capture text content from "text" type messages
		if result.Type == "text" && result.Part.Text != "" {
			textParts = append(textParts, result.Part.Text)
		}
	}

	text = strings.Join(textParts, "\n")
	return text, sessionID
}

// parseSessionID extracts the session ID from the raw output string.
// It looks for a JSON object in the output and attempts to unmarshal it.
func parseSessionID(output string) string {
	_, sessionID := ParseOutput(output)
	return sessionID
}
