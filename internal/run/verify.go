package run

import (
	"context"
	"fmt"
	"os/exec"
	"time"
)

// VerifyResult captures the outcome of a single verification command.
type VerifyResult struct {
	Command  string
	Duration time.Duration
	LogPath  string
}

// VerifyError is returned when a verification command fails.
type VerifyError struct {
	Command  string
	ExitCode int
	LogPath  string
	Err      error
}

func (e *VerifyError) Error() string {
	return fmt.Sprintf("verification failed: %q exited %d (see %s): %v", e.Command, e.ExitCode, e.LogPath, e.Err)
}

func (e *VerifyError) Unwrap() error {
	return e.Err
}

// RunVerification executes a set of verification commands sequentially.
// It stops at the first failure.
func RunVerification(ctx context.Context, artifacts *Artifacts, commands []string) ([]VerifyResult, error) {
	results := make([]VerifyResult, 0, len(commands))

	for i, cmd := range commands {
		start := time.Now()

		// Execute as: /bin/sh -lc "<cmd>"
		execCmd := exec.CommandContext(ctx, "/bin/sh", "-lc", cmd)

		// Capture combined output
		output, err := execCmd.CombinedOutput()
		duration := time.Since(start)

		// Log filename: NN.log (1-based stable numbering)
		logFilename := fmt.Sprintf("%02d.log", i+1)
		logPath, logErr := artifacts.WriteFile(SubDirVerify, logFilename, string(output))
		if logErr != nil {
			return results, fmt.Errorf("write verify log: %w", logErr)
		}

		res := VerifyResult{
			Command:  cmd,
			Duration: duration,
			LogPath:  logPath,
		}
		results = append(results, res)

		if err != nil {
			exitCode := -1
			if exitErr, ok := err.(*exec.ExitError); ok {
				exitCode = exitErr.ExitCode()
			}
			return results, &VerifyError{
				Command:  cmd,
				ExitCode: exitCode,
				LogPath:  logPath,
				Err:      err,
			}
		}
	}

	return results, nil
}
