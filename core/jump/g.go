package jump

import (
	"binrc.com/roma/core/tui"
	"github.com/loganchef/ssh"
)

func init() {

}

// Service Service
type Service struct {
	sess       *ssh.Session
	terminalUI *tui.TUI
}

func (jps *Service) setSession(sess *ssh.Session) {
	jps.sess = sess
}

// relogin, err := Configurate(sess)
//
//	if err != nil {
//		logger.Logger.Error("%s\n", err)
//		return
//	}
//
//	if relogin {
//		sshd.Info("Please login again with your new acount. \n", sess)
//		sshConn := (*sess).Context().Value(ssh.ContextKeyConn).(gossh.Conn)
//		sshConn.Close()
//		return
//	}
//
// Run jump
func (jps *Service) Run(remainingCmd string, remainingArgs []string, sess *ssh.Session) {
	defer func() {
		(*sess).Exit(0)
	}()

	//clear
	// 发送清空终端的命令
	// req := &gossh.Request{
	// 	Type:      "exec",
	// 	WantReply: true,
	// 	Payload:   []byte("clear"),
	// }

	// if err != nil {
	// 	log.Panicln("Failed to clear terminal" + err.Error())
	// }
	// // 检查请求的回复
	// if !aa {
	// 	log.Panicln("Failed to clear terminal" + err.Error())
	// }

	jps.setSession(sess)
	jps.terminalUI = &tui.TUI{}
	jps.terminalUI.SetSession(jps.sess)
	jps.terminalUI.ShowMainMenu(remainingCmd, remainingArgs)
}

// VarifyUser VarifyUser
// func VarifyUser(ctx ssh.Context, pass string) bool {
// 	username := ctx.User()
// 	logger.Logger.Debugf("VarifyUser username: %s\n", username)
// 	for _, user := range *config.Conf.Users {
// 		// Todo Password hash
// 		if user.Username == username && user.HashPasswd == pass {
// 			return true
// 		}
// 	}
// 	return false
// }

// Configurate Configurate
// func Configurate(sess *ssh.Session) (bool, error) {
// 	if *config.ConfPath == "" {
// 		return false, errors.New("Please specify a config file. ")
// 	}
// 	logger.Logger.Info("Read config file", *config.ConfPath)
// 	if !utils.FileExited(*config.ConfPath) {
// 		_, _, err := pui.CreateUser(false, true, sess)
// 		if err != nil {
// 			sshd.ErrorInfo(err, sess)
// 			return false, err
// 		}
// 		config.Conf.SaveTo(*config.ConfPath)
// 		return true, nil
// 	} else {
// 		config.Conf.ReadFrom(*config.ConfPath)
// 		if config.Conf.Users == nil || len(*config.Conf.Users) < 1 {
// 			_, _, err := pui.CreateUser(false, true, sess)
// 			if err != nil {
// 				sshd.ErrorInfo(err, sess)
// 				return false, err
// 			}
// 			config.Conf.SaveTo(*config.ConfPath)
// 			return true, nil
// 		}
// 	}
// 	return false, nil
// }
