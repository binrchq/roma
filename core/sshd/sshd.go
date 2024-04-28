package sshd

import (
	"errors"
	"fmt"
	"net"
	"strings"

	"bitrec.ai/roma/core/operation"
	"bitrec.ai/roma/core/utils/logger"
	"github.com/brckubo/ssh"
	"github.com/fatih/color"
	gossh "golang.org/x/crypto/ssh"
)

// NewTerminal NewTerminal
func NewTerminal(sess *ssh.Session, ip string, port int, sshUser string, key string, resType string) error {
	upstreamClient, err := NewSSHClient(ip, port, sshUser, key, resType)
	if err != nil {
		return nil
	}

	upstreamSess, err := upstreamClient.NewSession()
	if err != nil {
		return nil
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
	client, err := gossh.Dial("tcp", addr, configs)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}
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

// ErrorInfo ErrorInfo
func ErrorInfo(err error, sess *ssh.Session) {
	read := color.New(color.FgRed)
	read.Fprint(*sess, fmt.Sprintf("%s\n", err))
}
