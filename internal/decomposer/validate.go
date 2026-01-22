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

	return ""
}
