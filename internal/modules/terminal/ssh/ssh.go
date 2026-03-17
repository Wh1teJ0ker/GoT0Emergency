// Package ssh provides SSH-based terminal session implementation
// Handles remote terminal sessions via SSH with PTY support
package ssh

import (
	"GoT0Emergency/internal/pkg/log"
	"context"
	"encoding/base64"
	"io"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"golang.org/x/crypto/ssh"
)

// SSHTerminal represents an SSH terminal session
type SSHTerminal struct {
	session *ssh.Session    // SSH session
	stdin   io.WriteCloser  // Standard input pipe
	stdout  io.Reader       // Standard output pipe
	ctx     context.Context // Context for event emission
	id      string          // Terminal session identifier
}

// NewSSHTerminal creates a new SSH terminal session
// ctx: application context for event emission
// client: SSH client connection
// id: unique terminal identifier
// rows, cols: initial terminal dimensions
// Returns: SSHTerminal instance and error if session creation fails
func NewSSHTerminal(ctx context.Context, client *ssh.Client, id string, rows, cols int) (*SSHTerminal, error) {
	session, err := client.NewSession()
	if err != nil {
		return nil, log.Errorf("failed to create session: %w", err)
	}

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	if err := session.RequestPty("xterm-256color", rows, cols, modes); err != nil {
		session.Close()
		return nil, log.Errorf("request for pty failed: %w", err)
	}

	stdin, err := session.StdinPipe()
	if err != nil {
		session.Close()
		return nil, err
	}

	stdout, err := session.StdoutPipe()
	if err != nil {
		session.Close()
		return nil, err
	}

	// Merge stderr into stdout for simple terminal
	session.Stderr = session.Stdout

	if err := session.Shell(); err != nil {
		session.Close()
		return nil, log.Errorf("failed to start shell: %w", err)
	}

	t := &SSHTerminal{
		session: session,
		stdin:   stdin,
		stdout:  stdout,
		ctx:     ctx,
		id:      id,
	}

	go t.readLoop()

	return t, nil
}

// readLoop continuously reads from stdout and emits data to frontend
func (t *SSHTerminal) readLoop() {
	buf := make([]byte, 1024)
	for {
		n, err := t.stdout.Read(buf)
		if err != nil {
			if err == io.EOF {
				runtime.EventsEmit(t.ctx, "terminal:closed:"+t.id)
				return
			}
			return
		}

		if n > 0 {
			data := base64.StdEncoding.EncodeToString(buf[:n])
			runtime.EventsEmit(t.ctx, "terminal:data:"+t.id, data)
		}
	}
}

// Write sends data to the terminal (user input)
// data: bytes to write
// Returns: number of bytes written and error if write fails
func (t *SSHTerminal) Write(data []byte) (int, error) {
	return t.stdin.Write(data)
}

// Resize changes the terminal dimensions
// rows, cols: new terminal size
// Returns: error if resize fails
func (t *SSHTerminal) Resize(rows, cols int) error {
	return t.session.WindowChange(rows, cols)
}

// Close closes the terminal session
func (t *SSHTerminal) Close() error {
	return t.session.Close()
}

// Wait waits for the terminal process to exit
func (t *SSHTerminal) Wait() error {
	return t.session.Wait()
}
