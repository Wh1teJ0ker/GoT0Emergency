package executor

type Executor interface {
	Exec(cmd string) (string, error)
}
