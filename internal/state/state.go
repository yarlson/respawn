package state

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

const (
	stateRelPath = ".turbine/state/run.json"
)

// Load reads the run state from .turbine/state/run.json.
func Load(repoRoot string) (*RunState, bool, error) {
	path := filepath.Join(repoRoot, stateRelPath)
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("read state file: %w", err)
	}

	var state RunState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, true, fmt.Errorf("unmarshal state: %w", err)
	}

	return &state, true, nil
}

// Save writes the run state to .turbine/state/run.json.
func Save(repoRoot string, state *RunState) error {
	path := filepath.Join(repoRoot, stateRelPath)
	dir := filepath.Dir(path)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create state directory: %w", err)
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal state: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write state file: %w", err)
	}

	return nil
}

// Clear removes the run state file.
func Clear(repoRoot string) error {
	path := filepath.Join(repoRoot, stateRelPath)
	if err := os.Remove(path); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("remove state file: %w", err)
	}
	return nil
}
