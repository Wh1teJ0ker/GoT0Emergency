// Package executor 提供命令执行接口和实现
// 支持本地执行和 SSH 远程执行两种模式
package executor

// Executor 命令执行接口
// 用于统一本地和远程命令执行的调用方式
type Executor interface {
	// Exec 执行指定的命令并返回输出
	// cmd: 要执行的 shell 命令
	// 返回：命令输出和错误信息
	Exec(cmd string) (string, error)
}
