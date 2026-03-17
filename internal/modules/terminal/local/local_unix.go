//go:build linux || darwin
// +build linux darwin

// Package local provides local terminal session implementation
// Supports Windows ConPTY and Unix PTY for native terminal emulation
package local

import (
	"errors"
	"io"
	"os"
	"os/exec"

	"github.com/creack/pty"

	"GoT0Emergency/internal/pkg/log"
)

// startUnixShell starts a shell on Unix systems using PTY
// Returns: ptyx (stdout/stdin), closer, cmd, process, error
func startUnixShell() (*os.File, *os.File, io.Closer, *exec.Cmd, *os.Process, error) {
	// Get default shell
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/bash"
	}

	cmd := exec.Command(shell)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "TERM=xterm-256color")

	// Start with pty
	ptmx, err := pty.Start(cmd)
	if err != nil {
		log.Error("Failed to start pty: " + err.Error())
		return nil, nil, nil, nil, nil, errors.New("failed to start pty: " + err.Error())
	}

	// On Unix, ptmx is both stdin and stdout
	// ptmx returned by pty.Start is the master device file
	// We need to close it, so return ptmx as Closer
	// proc is nil because we use cmd for Wait
	return ptmx, ptmx, ptmx, cmd, nil, nil
}

// startWindowsShell is a stub for Unix build
func startWindowsShell() (*os.File, *os.File, io.Closer, *exec.Cmd, *os.Process, error) {
	return nil, nil, nil, nil, nil, errors.New("windows shell not supported on unix")
}

// resizeUnixTerminal resizes a Unix terminal
// ptmx: pty master file descriptor
// rows, cols: new terminal dimensions
func resizeUnixTerminal(ptmx *os.File, rows, cols int) error {
	return pty.Setsize(ptmx, &pty.Winsize{
		Rows: uint16(rows),
		Cols: uint16(cols),
		X:    0,
		Y:    0,
	})
}

// resizeWindowsTerminal is a stub for Unix build
func resizeWindowsTerminal(_ *os.File, _, _ int) error {
	return nil
}
