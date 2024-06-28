package services

import (
	"bitrec.ai/roma/core/jump"
	"bitrec.ai/roma/core/sshd"
	"github.com/brckubo/ssh"
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
	// sshd.ExecuteSCP(args, sess)
}
