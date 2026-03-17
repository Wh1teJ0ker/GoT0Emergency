// Package executor provides command execution interface and implementations
// Supports local execution and SSH remote execution modes
package executor

import (
	"bytes"
	"os/exec"
	"runtime"

	"GoT0Emergency/internal/pkg/log"
)

// LocalExecutor executes commands on the local host
type LocalExecutor struct{}

// NewLocalExecutor creates a new local executor instance
func NewLocalExecutor() *LocalExecutor {
	return &LocalExecutor{}
}

// Exec executes a command on the local host
// cmdStr: command string to execute
// Returns: command stdout and error if execution fails
// Uses cmd.exe on Windows and sh on Unix-like systems
func (e *LocalExecutor) Exec(cmdStr string) (string, error) {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", cmdStr)
	} else {
		cmd = exec.Command("sh", "-c", cmdStr)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", log.Errorf("local exec failed: %w, stderr: %s", err, stderr.String())
	}

	return stdout.String(), nil
}
