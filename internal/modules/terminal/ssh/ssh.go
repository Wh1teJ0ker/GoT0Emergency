package ssh

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"

	"golang.org/x/crypto/ssh"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type SSHTerminal struct {
	session *ssh.Session
	stdin   io.WriteCloser
	stdout  io.Reader
	ctx     context.Context
	id      string
}

func NewSSHTerminal(ctx context.Context, client *ssh.Client, id string, rows, cols int) (*SSHTerminal, error) {
	session, err := client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	if err := session.RequestPty("xterm-256color", rows, cols, modes); err != nil {
		session.Close()
		return nil, fmt.Errorf("request for pty failed: %w", err)
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
		return nil, fmt.Errorf("failed to start shell: %w", err)
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

func (t *SSHTerminal) Write(data []byte) (int, error) {
	return t.stdin.Write(data)
}

func (t *SSHTerminal) Resize(rows, cols int) error {
	return t.session.WindowChange(rows, cols)
}

func (t *SSHTerminal) Close() error {
	return t.session.Close()
}

func (t *SSHTerminal) Wait() error {
	return t.session.Wait()
}
