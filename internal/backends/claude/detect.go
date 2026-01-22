package claude

import (
	"context"
	"os/exec"
	"strings"
	"sync"
)

var (
	capabilities     *Capabilities
	capabilitiesOnce sync.Once
)

// Capabilities represents the detected flags of the claude CLI.
type Capabilities struct {
	HasContinue bool
	HasResume   bool
}

// DetectCapabilities runs 'claude --help' once and parses its output.
func DetectCapabilities(ctx context.Context, command string) *Capabilities {
	capabilitiesOnce.Do(func() {
		capabilities = &Capabilities{}
		if command == "" {
			command = "claude"
		}

		cmd := exec.CommandContext(ctx, command, "--help")
		output, err := cmd.CombinedOutput()
		if err != nil {
			// Fallback to -h if --help fails
			cmd = exec.CommandContext(ctx, command, "-h")
			output, err = cmd.CombinedOutput()
			if err != nil {
				return
			}
		}

		outStr := string(output)
		capabilities.HasContinue = strings.Contains(outStr, "--continue")
		capabilities.HasResume = strings.Contains(outStr, "--resume")
	})

	return capabilities
}

// ParseHelpOutput is a helper for testing.
func ParseHelpOutput(output string) *Capabilities {
	return &Capabilities{
		HasContinue: strings.Contains(output, "--continue"),
		HasResume:   strings.Contains(output, "--resume"),
	}
}
