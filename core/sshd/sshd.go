package sshd

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"binrc.com/roma/core/operation"
	"binrc.com/roma/core/utils/logger"
	"github.com/fatih/color"
	"github.com/loganchef/ssh"
	gossh "golang.org/x/crypto/ssh"
)

// NewTerminal NewTerminal
func NewTerminal(sess *ssh.Session, ip string, port int, sshUser string, key string, resType string) error {
	upstreamClient, err := NewSSHClient(ip, port, sshUser, key, resType)
	if err != nil {
		return err
	}

	upstreamSess, err := upstreamClient.NewSession()
	if err != nil {
		return err
	}
	defer upstreamSess.Close()

	upstreamSess.Stdout = *sess
	upstreamSess.Stdin = *sess
	upstreamSess.Stderr = *sess

	pty, winCh, _ := (*sess).Pty()

	if err := upstreamSess.RequestPty(pty.Term, pty.Window.Height, pty.Window.Width, gossh.TerminalModes{}); err != nil {
		return err
	}

	if err := upstreamSess.Shell(); err != nil {
		return err
	}

	go func() {
		for win := range winCh {
			upstreamSess.WindowChange(win.Height, win.Width)
		}
	}()

	if err := upstreamSess.Wait(); err != nil {
		return err
	}

	return nil
}

// NewSSHClient NewSSHClient
func NewSSHClient(ip string, port int, sshUser string, key string, resType string) (*gossh.Client, error) {
	if key == "" {
		op := operation.NewPassportOperation()
		keys, err := op.GetPassportByType(resType)
		if err != nil {
			logger.Logger.Error(err)
			return nil, err
		}
		key = keys[0].Passport
		sshUser = keys[0].ServiceUser
	}
	signer, err := gossh.ParsePrivateKey([]byte(key))
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}

	configs := &gossh.ClientConfig{
		User: "root",
		Auth: []gossh.AuthMethod{
			gossh.PublicKeys(signer),
		},
		HostKeyCallback: gossh.HostKeyCallback(func(hostname string, remote net.Addr, key gossh.PublicKey) error { return nil }),
	}

	addr := fmt.Sprintf("%s:%d", ip, port)

	// 使用带超时的 TCP 连接（10秒超时）
	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}

	// 在 TCP 连接上建立 SSH 连接（30秒握手超时）
	sshConn, chans, reqs, err := gossh.NewClientConn(conn, addr, configs)
	if err != nil {
		conn.Close()
		logger.Logger.Error(err)
		return nil, err
	}

	client := gossh.NewClient(sshConn, chans, reqs)
	return client, nil
}

// ParseRawCommand ParseRawCommand
func ParseRawCommand(command string) (string, []string, error) {
	parts := strings.Split(command, " ")

	if len(parts) < 1 {
		return "", nil, errors.New("No command in payload: " + command)
	}

	if len(parts) < 2 {
		return parts[0], []string{}, nil
	}

	return parts[0], parts[1:], nil
}

// ParseRemainingCommand removes the used portion from rawCmd
func ParseRemainingCommand(rawCmd string) (string, []string, error) {
	// List of known SSH parameters to be removed
	sshParams := map[string]bool{
		"-p": true,
		"-i": true,
		"-l": true,
		"-o": true,
		"-P": true,
	}

	parts := strings.Fields(rawCmd)
	var remainingParts []string
	skipNext := false

	for i, part := range parts {
		if skipNext {
			skipNext = false
			continue
		}

		// Check for SSH parameters
		if _, isSSHParam := sshParams[part]; isSSHParam {
			skipNext = true // Skip the next part which is the value of the SSH parameter
			continue
		}

		// If not an SSH parameter, add to remaining parts
		remainingParts = parts[i:]
		break
	}

	if len(remainingParts) == 0 {
		return "", nil, nil
	}

	remainingCmd := remainingParts[0]
	remainingArgs := remainingParts[1:]
	return remainingCmd, remainingArgs, nil
}

// ErrorInfo ErrorInfo
func ErrorInfo(err error, sess *ssh.Session) {
	read := color.New(color.FgRed)
	read.Fprint(*sess, fmt.Sprintf("%s\n", err))
}
