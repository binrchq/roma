package services

import (
	"binrc.com/roma/core/jump"
	"binrc.com/roma/core/sshd"
	"github.com/loganchef/ssh"
)

func SessionHandler(sess *ssh.Session) {
	defer func() {
		(*sess).Close()
	}()

	rawCmd := (*sess).RawCommand()
	cmd, args, err := sshd.ParseRawCommand(rawCmd)
	if err != nil {
		sshd.ErrorInfo(err, sess)
		return
	}
	switch cmd {
	case "scp":
		scpHandler(args, sess) //检测SCP命令执行逻辑
	default:
		remainingCmd, remainingArgs, err := sshd.ParseRemainingCommand(rawCmd)
		if err != nil {
			sshd.ErrorInfo(err, sess)
			return
		}
		sshHandler(remainingCmd, remainingArgs, sess)
	}
}

func sshHandler(remainingCmd string, remainingArgs []string, sess *ssh.Session) {
	jps := jump.Service{}
	jps.Run(remainingCmd, remainingArgs, sess)
}

func scpHandler(args []string, sess *ssh.Session) {
	// SCP 非交互式执行，直接传输文件并退出
	err := sshd.ExecuteSCP(args, sess)
	if err != nil {
		sshd.ErrorInfo(err, sess)
	}
	(*sess).Close()
}
