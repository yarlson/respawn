package claude

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/yarlson/turbine/internal/backends"
	"github.com/yarlson/turbine/internal/config"
)

type sessionInfo struct {
	opts       backends.SessionOptions
	hasStarted bool
}

type Backend struct {
	cfg      config.Backend
	sessions map[string]*sessionInfo
	mu       sync.RWMutex
}

func New(cfg config.Backend) *Backend {
	if cfg.Command == "" {
		cfg.Command = "claude"
	}
	return &Backend{
		cfg:      cfg,
		sessions: make(map[string]*sessionInfo),
	}
}

// StartSession initializes a new session and returns a session ID.
func (b *Backend) StartSession(ctx context.Context, opts backends.SessionOptions) (string, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Generate a unique session ID for this run.
	sessionID := fmt.Sprintf("claude-%d-%d", os.Getpid(), time.Now().UnixNano())

	if opts.Variant != "" {
		msg := fmt.Sprintf("Warning: variant '%s' is ignored for claude backend", opts.Variant)
		slog.Warn(msg)
		if opts.ArtifactsDir != "" {
			warnPath := filepath.Join(opts.ArtifactsDir, "backend", "warning.txt")
			_ = os.MkdirAll(filepath.Dir(warnPath), 0755)
			_ = os.WriteFile(warnPath, []byte(msg+"\n"), 0644)
		}
	}

	b.sessions[sessionID] = &sessionInfo{
		opts: opts,
	}
	return sessionID, nil
}

func (b *Backend) Send(ctx context.Context, sessionID string, prompt string, opts backends.SendOptions) (*backends.Result, error) {
	b.mu.Lock()
	info, ok := b.sessions[sessionID]
	if !ok {
		b.mu.Unlock()
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}
	b.mu.Unlock()

	caps := DetectCapabilities(ctx, b.cfg.Command)

	cmdArgs := append([]string{}, b.cfg.Args...)

	// Apply model from session options
	if info.opts.Model != "" {
		cmdArgs = append(cmdArgs, "--model", info.opts.Model)
	}

	useContinue := false
	if info.hasStarted {
		if caps.HasContinue {
			cmdArgs = append(cmdArgs, "--continue")
			useContinue = true
		} else if caps.HasResume {
			cmdArgs = append(cmdArgs, "--resume")
			useContinue = true
		} else {
			slog.Info("Claude session reuse requested but not supported by CLI; re-running without continue/resume flags", "session_id", sessionID)
		}
	}

	cmd := exec.CommandContext(ctx, b.cfg.Command, cmdArgs...)
	cmd.Dir = info.opts.WorkingDir
	cmd.Stdin = strings.NewReader(prompt)

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err := cmd.Run()

	// Mark as started regardless of error, as the CLI might have created state.
	info.hasStarted = true

	stdout := stdoutBuf.String()
	stderr := stderrBuf.String()

	// Capture to artifacts
	if info.opts.ArtifactsDir != "" {
		backendDir := filepath.Join(info.opts.ArtifactsDir, "backend")
		_ = os.MkdirAll(backendDir, 0755)

		timestamp := time.Now().Format("20060102-150405.000")
		_ = os.WriteFile(filepath.Join(backendDir, fmt.Sprintf("%s.stdout.txt", timestamp)), stdoutBuf.Bytes(), 0644)
		_ = os.WriteFile(filepath.Join(backendDir, fmt.Sprintf("%s.stderr.txt", timestamp)), stderrBuf.Bytes(), 0644)

		if info.hasStarted && !useContinue && (caps.HasContinue || caps.HasResume) {
			note := "Note: Reuse was requested but this was the first message in this process session.\n"
			_ = os.WriteFile(filepath.Join(backendDir, "session_note.txt"), []byte(note), 0644)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("claude command failed: %w (stderr: %s)", err, stderr)
	}

	return &backends.Result{
		Output: stdout,
		Metadata: map[string]string{
			"session_id": sessionID,
			"command":    b.cfg.Command,
		},
	}, nil
}
