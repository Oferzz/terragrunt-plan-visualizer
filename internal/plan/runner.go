package plan

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// RunnerConfig configures how terragrunt/terraform commands are executed.
type RunnerConfig struct {
	WorkingDir   string
	PlanTimeout  time.Duration
	ApplyTimeout time.Duration
	ExtraArgs    []string
}

// RunPlan executes terragrunt plan and returns the raw JSON plan output.
func RunPlan(ctx context.Context, cfg RunnerConfig) ([]byte, string, error) {
	planFile := filepath.Join(cfg.WorkingDir, "tfplan")

	// Run terragrunt plan
	args := append([]string{"plan", "-out=" + planFile}, cfg.ExtraArgs...)
	planCtx, cancel := context.WithTimeout(ctx, cfg.PlanTimeout)
	defer cancel()

	var stderr bytes.Buffer
	cmd := exec.CommandContext(planCtx, "terragrunt", args...)
	cmd.Dir = cfg.WorkingDir
	cmd.Stderr = &stderr
	cmd.Stdout = os.Stderr // Show plan output to user via stderr

	if err := cmd.Run(); err != nil {
		return nil, "", classifyError(err, stderr.String(), planCtx)
	}

	// Run terragrunt show -json to get structured output
	showCtx, showCancel := context.WithTimeout(ctx, 2*time.Minute)
	defer showCancel()

	var showOut bytes.Buffer
	var showErr bytes.Buffer
	showCmd := exec.CommandContext(showCtx, "terragrunt", "show", "-json", planFile)
	showCmd.Dir = cfg.WorkingDir
	showCmd.Stdout = &showOut
	showCmd.Stderr = &showErr

	if err := showCmd.Run(); err != nil {
		return nil, "", fmt.Errorf("failed to show plan JSON: %w\n%s", err, showErr.String())
	}

	return showOut.Bytes(), planFile, nil
}

// RunApply executes terragrunt apply with the given plan file.
// The outputFn callback is called with each line of output for streaming.
func RunApply(ctx context.Context, cfg RunnerConfig, planFile string, outputFn func(line string)) error {
	applyCtx, cancel := context.WithTimeout(ctx, cfg.ApplyTimeout)
	defer cancel()

	args := []string{"apply", planFile}
	cmd := exec.CommandContext(applyCtx, "terragrunt", args...)
	cmd.Dir = cfg.WorkingDir

	// Capture combined output for streaming
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return classifyError(err, "", applyCtx)
	}

	buf := make([]byte, 4096)
	for {
		n, readErr := stdout.Read(buf)
		if n > 0 {
			lines := strings.Split(string(buf[:n]), "\n")
			for _, line := range lines {
				if line != "" && outputFn != nil {
					outputFn(line)
				}
			}
		}
		if readErr != nil {
			break
		}
	}

	if err := cmd.Wait(); err != nil {
		return classifyError(err, stderr.String(), applyCtx)
	}

	return nil
}

// RunUnlock force-unlocks a terraform state.
func RunUnlock(ctx context.Context, workingDir string, lockID string) error {
	unlockCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var stderr bytes.Buffer
	cmd := exec.CommandContext(unlockCtx, "terragrunt", "force-unlock", "-force", lockID)
	cmd.Dir = workingDir
	cmd.Stderr = &stderr
	cmd.Stdout = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to unlock state: %w\n%s", err, stderr.String())
	}

	return nil
}

var lockIDRegex = regexp.MustCompile(`Lock Info:\s*ID:\s*(\S+)`)
var lockWhoRegex = regexp.MustCompile(`Who:\s*(.+)`)
var lockOperationRegex = regexp.MustCompile(`Operation:\s*(.+)`)
var lockCreatedRegex = regexp.MustCompile(`Created:\s*(.+)`)

func classifyError(err error, stderr string, ctx context.Context) error {
	if ctx.Err() == context.DeadlineExceeded {
		return &CLIError{Type: "timeout", Message: "operation timed out", Details: stderr}
	}

	combined := strings.ToLower(stderr)

	// State lock detection
	if strings.Contains(combined, "lock") && (strings.Contains(combined, "acquired") || strings.Contains(combined, "locked")) {
		lockInfo := parseLockInfo(stderr)
		return &StateLockError{
			CLIError: CLIError{Type: "lock", Message: "state is locked", Details: stderr},
			Lock:     lockInfo,
		}
	}

	// Auth errors
	if strings.Contains(combined, "no valid credential") ||
		strings.Contains(combined, "expired token") ||
		strings.Contains(combined, "security token") ||
		strings.Contains(combined, "not authorized") ||
		strings.Contains(combined, "accessdenied") {
		return &CLIError{
			Type:    "auth",
			Message: "authentication/authorization error",
			Details: stderr,
			Hint:    "Check your cloud provider credentials. For AWS, try: aws sso login",
		}
	}

	// Network errors
	if strings.Contains(combined, "dial tcp") ||
		strings.Contains(combined, "connection refused") ||
		strings.Contains(combined, "no such host") {
		return &CLIError{
			Type:    "network",
			Message: "network connectivity error",
			Details: stderr,
			Hint:    "Check your network connection and VPN status",
		}
	}

	return &CLIError{
		Type:    "terraform",
		Message: err.Error(),
		Details: stderr,
	}
}

func parseLockInfo(stderr string) LockInfo {
	info := LockInfo{}
	if m := lockIDRegex.FindStringSubmatch(stderr); len(m) > 1 {
		info.ID = m[1]
	}
	if m := lockWhoRegex.FindStringSubmatch(stderr); len(m) > 1 {
		info.Who = strings.TrimSpace(m[1])
	}
	if m := lockOperationRegex.FindStringSubmatch(stderr); len(m) > 1 {
		info.Operation = strings.TrimSpace(m[1])
	}
	if m := lockCreatedRegex.FindStringSubmatch(stderr); len(m) > 1 {
		t, err := time.Parse(time.RFC3339, strings.TrimSpace(m[1]))
		if err == nil {
			info.Created = t
		}
	}
	return info
}

// StateLockError wraps a CLI error with lock information.
type StateLockError struct {
	CLIError
	Lock LockInfo
}

func (e *StateLockError) Error() string {
	return fmt.Sprintf("state locked by %s (lock ID: %s)", e.Lock.Who, e.Lock.ID)
}

// CLIErrorFromErr extracts a CLIError from an error, or creates one.
func CLIErrorFromErr(err error) *CLIError {
	if cliErr, ok := err.(*CLIError); ok {
		return cliErr
	}
	if lockErr, ok := err.(*StateLockError); ok {
		return &lockErr.CLIError
	}
	return &CLIError{
		Type:    "unknown",
		Message: err.Error(),
	}
}
