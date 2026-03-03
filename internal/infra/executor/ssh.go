package executor

import (
	"bytes"
	"fmt"
	"time"

	"golang.org/x/crypto/ssh"
)

type SSHExecutor struct {
	client *ssh.Client
}

func NewSSHExecutor(client *ssh.Client) *SSHExecutor {
	return &SSHExecutor{client: client}
}

func (e *SSHExecutor) Exec(cmdStr string) (string, error) {
	if e.client == nil {
		return "", fmt.Errorf("ssh client is nil")
	}

	session, err := e.client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	err = session.Run(cmdStr)
	if err != nil {
		return "", fmt.Errorf("ssh exec failed: %w, stderr: %s", err, stderr.String())
	}

	return stdout.String(), nil
}

// Helper to create client (usually handled in session manager, but useful here for simple setup)
func ConnectSSH(ip string, port int, user string, authMethod ssh.AuthMethod) (*ssh.Client, error) {
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			authMethod,
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // MVP: Ignore host key verification
		Timeout:         5 * time.Second,
	}

	addr := fmt.Sprintf("%s:%d", ip, port)
	return ssh.Dial("tcp", addr, config)
}
