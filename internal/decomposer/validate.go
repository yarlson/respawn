package decomposer

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/yarlson/respawn/internal/tasks"

	"gopkg.in/yaml.v3"
)

// validateTasksYAML parses and validates tasks YAML.
func validateTasksYAML(yamlContent string) (*tasks.TaskList, error) {
	var taskList tasks.TaskList
	if err := yaml.Unmarshal([]byte(yamlContent), &taskList); err != nil {
		return nil, fmt.Errorf("unmarshal tasks yaml: %w", err)
	}

	if err := taskList.Validate(); err != nil {
		return nil, fmt.Errorf("validate tasks: %w", err)
	}

	return &taskList, nil
}

// extractYAML attempts to find YAML content in backend output.
func extractYAML(output string) string {
	// Try to find markdown code block first
	re := regexp.MustCompile("(?s)```(?:yaml)?\n(.*?)\n```")
	matches := re.FindStringSubmatch(output)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	// If no code block, look for something that looks like YAML (version: 1 and tasks:)
	if strings.Contains(output, "version:") && strings.Contains(output, "tasks:") {
		return strings.TrimSpace(output)
	}

	// Also accept YAML that starts with tasks: (missing version) - we'll get validation error
	// but at least we'll extract it for retry/fix
	if strings.Contains(output, "tasks:") {
		// Find where tasks: starts and extract from there
		idx := strings.Index(output, "tasks:")
		if idx >= 0 {
			// Check if there's version: before tasks:
			beforeTasks := output[:idx]
			if versionIdx := strings.LastIndex(beforeTasks, "version:"); versionIdx >= 0 {
				return strings.TrimSpace(output[versionIdx:])
			}
			// No version found, prepend it
			extracted := strings.TrimSpace(output[idx:])
			return "version: 1\n" + extracted
		}
	}

	return ""
}
