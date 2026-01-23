package agents

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateOutput(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(dir string) error
		expectError bool
		errorMsg    string
	}{
		{
			name: "all files present",
			setup: func(dir string) error {
				// Create AGENTS.md
				if err := os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte("# Test"), 0644); err != nil {
					return err
				}
				// Create CLAUDE.md symlink
				return os.Symlink("AGENTS.md", filepath.Join(dir, "CLAUDE.md"))
			},
			expectError: false,
		},
		{
			name: "missing AGENTS.md",
			setup: func(dir string) error {
				// Only create CLAUDE.md symlink (will be broken)
				return os.Symlink("AGENTS.md", filepath.Join(dir, "CLAUDE.md"))
			},
			expectError: true,
			errorMsg:    "AGENTS.md was not created",
		},
		{
			name: "missing CLAUDE.md symlink",
			setup: func(dir string) error {
				// Only create AGENTS.md
				return os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte("# Test"), 0644)
			},
			expectError: true,
			errorMsg:    "CLAUDE.md symlink was not created",
		},
		{
			name: "CLAUDE.md is regular file not symlink",
			setup: func(dir string) error {
				// Create AGENTS.md
				if err := os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte("# Test"), 0644); err != nil {
					return err
				}
				// Create CLAUDE.md as regular file
				return os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("# Test"), 0644)
			},
			expectError: true,
			errorMsg:    "CLAUDE.md exists but is not a symlink",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			if err := tt.setup(tmpDir); err != nil {
				t.Fatalf("setup failed: %v", err)
			}

			g := &Generator{repoRoot: tmpDir}
			err := g.validateOutput()

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errorMsg)
				} else if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("expected error %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}
