package local

import (
	"context"
	"encoding/base64"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strconv"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"GoT0Emergency/internal/pkg/log"
)

type LocalTerminal struct {
	// ptmx 在不同平台上含义不同：
	// - Unix: pty 主设备文件（同时用于读写）
	// - Windows: stdout 管道（只读）
	ptmx      *os.File
	stdin     io.WriteCloser // 单独的 stdin，Windows 必需，Unix 可选
	ptyCloser io.Closer      // Unix 需要单独关闭 pty，Windows 为 nil
	cmd       *exec.Cmd      // Unix/Windows Pipe 模式下使用
	proc      *os.Process    // Windows ConPTY 模式下使用
	ctx       context.Context
	id        string
}

// NewLocalTerminal 创建一个新的本地终端实例
func NewLocalTerminal(ctx context.Context, id string) (*LocalTerminal, error) {
	var ptmx *os.File
	var stdin io.WriteCloser
	var ptyCloser io.Closer
	var cmd *exec.Cmd
	var proc *os.Process
	var err error

	// 使用运行时检查来选择启动逻辑
	if runtime.GOOS == "windows" {
		ptmx, stdin, ptyCloser, cmd, proc, err = startWindowsShell()
	} else {
		ptmx, stdin, ptyCloser, cmd, proc, err = startUnixShell()
	}

	if err != nil {
		log.Error("Failed to start shell: " + err.Error())
		return nil, err
	}

	t := &LocalTerminal{
		ptmx:      ptmx,
		stdin:     stdin,
		ptyCloser: ptyCloser,
		cmd:       cmd,
		proc:      proc,
		ctx:       ctx,
		id:        id,
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
				log.Info("Terminal closed (EOF): " + t.id)
				wruntime.EventsEmit(t.ctx, "terminal:closed:"+t.id)
				return
			}
			log.Error("Terminal read error: " + err.Error())
			return
		}

		if n > 0 {
			data := base64.StdEncoding.EncodeToString(buf[:n])
			wruntime.EventsEmit(t.ctx, "terminal:data:"+t.id, data)
		}
	}
}

func (t *LocalTerminal) Write(data []byte) (int, error) {
	if t.stdin != nil {
		return t.stdin.Write(data)
	}
	return t.ptmx.Write(data)
}

func (t *LocalTerminal) Resize(rows, cols int) error {
	// 运行时判断
	if runtime.GOOS == "windows" {
		return resizeWindowsTerminal(t.ptmx, rows, cols)
	}
	return resizeUnixTerminal(t.ptmx, rows, cols)
}

func (t *LocalTerminal) Close() error {
	log.Info("Closing terminal: " + t.id)

	if t.stdin != nil {
		t.stdin.Close()
	}

	if t.ptyCloser != nil {
		t.ptyCloser.Close()
	}

	return nil
}

func (t *LocalTerminal) Wait() error {
	// 优先使用 Process 等待 (ConPTY)
	if t.proc != nil {
		_, err := t.proc.Wait()
		if t.ptmx != nil {
			t.ptmx.Close()
		}
		if err != nil {
			log.Error("Terminal process exited with error: " + err.Error())
		}
		return err
	}

	// 其次使用 Cmd 等待 (Pipe/Unix)
	if t.cmd != nil {
		err := t.cmd.Wait()
		if t.ptmx != nil {
			t.ptmx.Close()
		}
		if err != nil {
			log.Error("Terminal process exited with error: " + err.Error())
		}
		return err
	}
	
	return nil
}

// 辅助函数：Int 转 String，替代 fmt.Sprintf
func itoa(i int) string {
	return strconv.Itoa(i)
}
