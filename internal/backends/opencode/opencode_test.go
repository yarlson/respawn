package opencode

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yarlson/turbine/internal/backends"
	"github.com/yarlson/turbine/internal/config"
)

func TestBuildCommandArgs(t *testing.T) {
	tests := []struct {
		name       string
		cfg        config.Backend
		opts       backends.SessionOptions
		hasStarted bool
		extID      string
		expected   []string
	}{
		{
			name: "default run",
			cfg: config.Backend{
				Command: "opencode",
			},
			opts:     backends.SessionOptions{},
			expected: []string{"run", "--format", "json"},
		},
		{
			name: "with model and variant from config",
			cfg: config.Backend{
				Command: "opencode",
				Models: config.Models{
					Slow: "m1",
					Fast: "m1-fast",
				},
				Variant: "v1",
			},
			opts:     backends.SessionOptions{},
			expected: []string{"run", "--format", "json", "--model", "m1", "--variant", "v1"},
		},
		{
			name: "override model and variant from opts",
			cfg: config.Backend{
				Command: "opencode",
				Models: config.Models{
					Slow: "m1",
					Fast: "m1-fast",
				},
				Variant: "v1",
			},
			opts: backends.SessionOptions{
				Model:   "m2",
				Variant: "v2",
			},
			expected: []string{"run", "--format", "json", "--model", "m2", "--variant", "v2"},
		},
		{
			name: "retry with known session id",
			cfg: config.Backend{
				Command: "opencode",
			},
			opts:       backends.SessionOptions{},
			hasStarted: true,
			extID:      "sid123",
			expected:   []string{"run", "--format", "json", "--session", "sid123"},
		},
		{
			name: "retry without known session id",
			cfg: config.Backend{
				Command: "opencode",
			},
			opts:       backends.SessionOptions{},
			hasStarted: true,
			expected:   []string{"run", "--format", "json", "--continue"},
		},
		{
			name: "custom args",
			cfg: config.Backend{
				Command: "opencode",
				Args:    []string{"--verbose"},
			},
			opts:     backends.SessionOptions{},
			expected: []string{"run", "--verbose", "--format", "json"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := New(tt.cfg)
			info := &sessionInfo{
				opts:              tt.opts,
				hasStarted:        tt.hasStarted,
				externalSessionID: tt.extID,
			}
			args, _ := b.buildCommandArgs(info, "") // empty model override
			assert.Equal(t, tt.expected, args)
		})
	}
}

func TestParseOutput(t *testing.T) {
	tests := []struct {
		name      string
		output    string
		sessionID string
	}{
		{
			name:      "valid json",
			output:    `{"sessionID": "sid123", "output": "hello"}`,
			sessionID: "sid123",
		},
		{
			name:      "json with prefix",
			output:    "Welcome!\n" + `{"sessionID": "sid456"}`,
			sessionID: "sid456",
		},
		{
			name:      "no json",
			output:    "just some text",
			sessionID: "",
		},
		{
			name:      "empty session id",
			output:    `{"output": "no session"}`,
			sessionID: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, sid := ParseOutput(tt.output)
			assert.Equal(t, tt.sessionID, sid)
		})
	}
}
