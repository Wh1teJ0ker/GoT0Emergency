package local

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/creack/pty"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type LocalTerminal struct {
	ptmx   *os.File
	cmd    *exec.Cmd
	ctx    context.Context
	id     string // Terminal ID for event emission
}

func NewLocalTerminal(ctx context.Context, id string) (*LocalTerminal, error) {
	// Default shell
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/bash"
	}

	c := exec.Command(shell)
	// Set environment variables if needed
	c.Env = os.Environ()
	c.Env = append(c.Env, "TERM=xterm-256color")

	ptmx, err := pty.Start(c)
	if err != nil {
		return nil, fmt.Errorf("failed to start pty: %w", err)
	}

	t := &LocalTerminal{
		ptmx: ptmx,
		cmd:  c,
		ctx:  ctx,
		id:   id,
	}

	// Start reading routine
	go t.readLoop()

	return t, nil
}

func (t *LocalTerminal) readLoop() {
	buf := make([]byte, 1024)
	for {
		n, err := t.ptmx.Read(buf)
		if err != nil {
			if err == io.EOF {
				runtime.EventsEmit(t.ctx, "terminal:closed:"+t.id)
				return
			}
			// Handle other errors?
			return
		}

		if n > 0 {
			// Encode to base64 to avoid encoding issues in JSON/Events
			data := base64.StdEncoding.EncodeToString(buf[:n])
			runtime.EventsEmit(t.ctx, "terminal:data:"+t.id, data)
		}
	}
}

func (t *LocalTerminal) Write(data []byte) (int, error) {
	return t.ptmx.Write(data)
}

func (t *LocalTerminal) Resize(rows, cols int) error {
	return pty.Setsize(t.ptmx, &pty.Winsize{
		Rows: uint16(rows),
		Cols: uint16(cols),
		X:    0,
		Y:    0,
	})
}

func (t *LocalTerminal) Close() error {
	// Kill the process
	if t.cmd.Process != nil {
		t.cmd.Process.Kill()
	}
	return t.ptmx.Close()
}

func (t *LocalTerminal) Wait() error {
	return t.cmd.Wait()
}
