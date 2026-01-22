package claude

import (
	"context"
	"os"
	"path/filepath"
	"respawn/internal/backends"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseHelpOutput(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		expected Capabilities
	}{
		{
			name:   "supports both",
			output: "Usage: claude [options]\n\nOptions:\n  --continue  Continue last session\n  --resume    Resume session",
			expected: Capabilities{
				HasContinue: true,
				HasResume:   true,
			},
		},
		{
			name:   "supports only continue",
			output: "Usage: claude [options]\n\nOptions:\n  --continue  Continue last session",
			expected: Capabilities{
				HasContinue: true,
				HasResume:   false,
			},
		},
		{
			name:   "supports only resume",
			output: "Usage: claude [options]\n\nOptions:\n  --resume    Resume session",
			expected: Capabilities{
				HasContinue: false,
				HasResume:   true,
			},
		},
		{
			name:   "supports none",
			output: "Usage: claude [options]\n\nOptions:\n  -p, --prompt  Send prompt",
			expected: Capabilities{
				HasContinue: false,
				HasResume:   false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			caps := ParseHelpOutput(tt.output)
			assert.Equal(t, tt.expected.HasContinue, caps.HasContinue)
			assert.Equal(t, tt.expected.HasResume, caps.HasResume)
		})
	}
}

func TestClaudeBackend_StartSession_VariantWarning(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "claude-test-*")
	assert.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	b := New("claude", []string{"-p"})
	opts := backends.SessionOptions{
		Variant:      "some-variant",
		ArtifactsDir: tmpDir,
	}

	sessionID, err := b.StartSession(context.Background(), opts)
	assert.NoError(t, err)
	assert.NotEmpty(t, sessionID)

	// Check if warning file was created
	warnPath := filepath.Join(tmpDir, "backend", "warning.txt")
	assert.FileExists(t, warnPath)
	content, err := os.ReadFile(warnPath)
	assert.NoError(t, err)
	assert.Contains(t, string(content), "variant 'some-variant' is ignored")
}

func TestClaudeBackend_Send_SessionNotFound(t *testing.T) {
	b := New("claude", []string{"-p"})
	_, err := b.Send(context.Background(), "non-existent", "hello", backends.SendOptions{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "session not found")
}
