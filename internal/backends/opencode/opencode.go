package opencode

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

	"github.com/yarlson/respawn/internal/backends"
	"github.com/yarlson/respawn/internal/config"
)

type sessionInfo struct {
	opts              backends.SessionOptions
	externalSessionID string
	hasStarted        bool
}

// OpenCode implements the backends.Backend interface for OpenCode.
type OpenCode struct {
	cfg      config.Backend
	sessions map[string]*sessionInfo
	mu       sync.RWMutex
}

// New creates a new OpenCode backend with the given configuration.
func New(cfg config.Backend) *OpenCode {
	if cfg.Command == "" {
		cfg.Command = "opencode"
	}
	return &OpenCode{
		cfg:      cfg,
		sessions: make(map[string]*sessionInfo),
	}
}

// StartSession initializes a new session and returns a session ID.
func (b *OpenCode) StartSession(ctx context.Context, opts backends.SessionOptions) (string, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	sessionID := fmt.Sprintf("opencode-%d-%d", os.Getpid(), time.Now().UnixNano())
	b.sessions[sessionID] = &sessionInfo{
		opts: opts,
	}
	return sessionID, nil
}

func (b *OpenCode) buildCommandArgs(info *sessionInfo) ([]string, string) {
	cmdArgs := append([]string{}, b.cfg.Args...)

	// Ensure 'run' command is present
	hasRun := false
	for _, arg := range cmdArgs {
		if arg == "run" {
			hasRun = true
			break
		}
	}
	if !hasRun {
		cmdArgs = append([]string{"run"}, cmdArgs...)
	}

	// Always ensure --format json
	hasFormat := false
	for i := 0; i < len(cmdArgs); i++ {
		if cmdArgs[i] == "--format" {
			hasFormat = true
			break
		}
	}
	if !hasFormat {
		cmdArgs = append(cmdArgs, "--format", "json")
	}

	// Apply model/variant from session options (prioritized) or config
	model := info.opts.Model
	if model == "" {
		model = b.cfg.Model
	}
	if model != "" {
		cmdArgs = append(cmdArgs, "--model", model)
	}

	variant := info.opts.Variant
	if variant == "" {
		variant = b.cfg.Variant
	}
	if variant != "" {
		cmdArgs = append(cmdArgs, "--variant", variant)
	}

	// Session handling: New task starts without -c/--continue.
	// Retry in same cycle (indicated by hasStarted) re-uses session.
	var warning string
	if info.hasStarted {
		if info.externalSessionID != "" {
			cmdArgs = append(cmdArgs, "--session", info.externalSessionID)
		} else {
			cmdArgs = append(cmdArgs, "--continue")
			warning = "Warning: session ID unknown, falling back to --continue"
		}
	}

	return cmdArgs, warning
}

// Send transmits a prompt to OpenCode and returns the result.
func (b *OpenCode) Send(ctx context.Context, sessionID string, prompt string, opts backends.SendOptions) (*backends.Result, error) {
	b.mu.Lock()
	info, ok := b.sessions[sessionID]
	if !ok {
		b.mu.Unlock()
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}
	b.mu.Unlock()

	cmdArgs, warning := b.buildCommandArgs(info)
	if warning != "" {
		slog.Warn(warning, "session_id", sessionID)
	}

	cmd := exec.CommandContext(ctx, b.cfg.Command, cmdArgs...)
	cmd.Dir = info.opts.WorkingDir
	cmd.Stdin = strings.NewReader(prompt)

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err := cmd.Run()

	stdout := stdoutBuf.String()
	stderr := stderrBuf.String()

	// Parse NDJSON output from opencode
	output, extID := ParseOutput(stdout)
	if output == "" {
		output = stdout
	}

	// Capture to artifacts
	if info.opts.ArtifactsDir != "" {
		backendDir := filepath.Join(info.opts.ArtifactsDir, "backend")
		_ = os.MkdirAll(backendDir, 0755)

		timestamp := time.Now().Format("20060102-150405.000")
		_ = os.WriteFile(filepath.Join(backendDir, fmt.Sprintf("%s.stdout.txt", timestamp)), stdoutBuf.Bytes(), 0644)
		_ = os.WriteFile(filepath.Join(backendDir, fmt.Sprintf("%s.stderr.txt", timestamp)), stderrBuf.Bytes(), 0644)

		if warning != "" {
			_ = os.WriteFile(filepath.Join(backendDir, "warning.txt"), []byte(warning+"\n"), 0644)
		}
	}

	// Parse session ID from stdout for future reuse
	if extID != "" {
		info.externalSessionID = extID
	}
	info.hasStarted = true

	if err != nil {
		return nil, fmt.Errorf("opencode command failed: %w (stderr: %s)", err, stderr)
	}

	return &backends.Result{
		Output: output,
		Metadata: map[string]string{
			"session_id":          sessionID,
			"external_session_id": info.externalSessionID,
		},
	}, nil
}
