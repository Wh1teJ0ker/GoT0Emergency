//go:build linux || darwin
// +build linux darwin

package local

import (
	"errors"
	"io"
	"os"
	"os/exec"

	"github.com/creack/pty"

	"GoT0Emergency/internal/pkg/log"
)

// startUnixShell 在 Unix 系统上启动 shell
// 返回: pty主设备, stdin(=ptmx), pty关闭器, cmd, proc, 错误
func startUnixShell() (*os.File, *os.File, io.Closer, *exec.Cmd, *os.Process, error) {
	// 获取默认 shell
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/bash"
	}

	cmd := exec.Command(shell)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "TERM=xterm-256color")

	// 使用 pty 启动
	ptmx, err := pty.Start(cmd)
	if err != nil {
		log.Error("Failed to start pty: " + err.Error())
		return nil, nil, nil, nil, nil, errors.New("failed to start pty: " + err.Error())
	}

	// 在 Unix 上，ptmx 同时是 stdin 和 stdout
	// pty.Start 返回的 ptmx 就是主设备文件
	// 我们需要关闭它，所以返回 ptmx 作为 Closer
	// proc 为 nil，因为我们使用 cmd 等待
	return ptmx, ptmx, ptmx, cmd, nil, nil
}

// startWindowsShell is a stub for Unix build
func startWindowsShell() (*os.File, *os.File, io.Closer, *exec.Cmd, *os.Process, error) {
	return nil, nil, nil, nil, nil, errors.New("windows shell not supported on unix")
}

// resizeUnixTerminal 调整 Unix 终端大小
func resizeUnixTerminal(ptmx *os.File, rows, cols int) error {
	return pty.Setsize(ptmx, &pty.Winsize{
		Rows: uint16(rows),
		Cols: uint16(cols),
		X:    0,
		Y:    0,
	})
}

// Stub for Windows resize on Unix build
func resizeWindowsTerminal(_ *os.File, _, _ int) error {
	return nil
}
