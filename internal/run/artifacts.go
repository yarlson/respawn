package run

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	// RunsDir is the base directory for all runs, relative to repo root.
	RunsDir = ".turbine/runs"

	// Subdirectories within a run directory.
	SubDirPrompts = "prompts"
	SubDirBackend = "backend"
	SubDirVerify  = "verify"
	SubDirGit     = "git"
)

// Artifacts manages the directory layout and file persistence for a single run.
type Artifacts struct {
	root string
}

// NewArtifacts initializes the directory layout for a runID under repoRoot.
// It creates all required subdirectories.
func NewArtifacts(repoRoot, runID string) (*Artifacts, error) {
	if repoRoot == "" {
		return nil, fmt.Errorf("repo root cannot be empty")
	}
	if runID == "" {
		return nil, fmt.Errorf("run ID cannot be empty")
	}

	runRoot := filepath.Join(repoRoot, RunsDir, runID)

	subDirs := []string{
		runRoot,
		filepath.Join(runRoot, SubDirPrompts),
		filepath.Join(runRoot, SubDirBackend),
		filepath.Join(runRoot, SubDirVerify),
		filepath.Join(runRoot, SubDirGit),
	}

	for _, dir := range subDirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("create artifact directory: %w", err)
		}
	}

	return &Artifacts{root: runRoot}, nil
}

// Root returns the absolute path to the run's artifact directory.
func (a *Artifacts) Root() string {
	return a.root
}

// WriteFile writes a small text file into the specified subdirectory of the run.
func (a *Artifacts) WriteFile(subDir, filename string, content string) (string, error) {
	destDir := filepath.Join(a.root, subDir)
	if _, err := os.Stat(destDir); os.IsNotExist(err) {
		return "", fmt.Errorf("subdirectory %s does not exist", subDir)
	}

	path := filepath.Join(destDir, filename)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("write artifact file: %w", err)
	}

	return path, nil
}
