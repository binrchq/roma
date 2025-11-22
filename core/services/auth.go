package services

import (
	"log"

	"binrc.com/roma/core/operation"
	"github.com/loganchef/ssh"
)

//	func PasswordAuth(ctx ssh.Context, pass string) bool {
//		configs.Conf.ReadFrom(*configs.ConfPath)
//		var success bool
//		if (len(*configs.Conf.Users)) < 1 {
//			success = (pass == "newuser")
//		} else {
//			success = jump.VarifyUser(ctx, pass)
//		}
//		if !success {
//			time.Sleep(time.Second * 3)
//		}
//		return success
//	}

func PublicKeyAuth(ctx ssh.Context, key ssh.PublicKey) bool {
	var pub string

	// configs.Conf.ReadFrom(*configs.ConfPath)
	op := operation.NewUserOperation()
	user, err := op.GetUserByUsername(ctx.User())
	if err != nil {
		return false
	}
	pub = user.PublicKey
	allowed, _, _, _, err := ssh.ParseAuthorizedKey([]byte(pub))
	if err != nil {
		log.Println("Error parsing authorized key:", err)
		// 处理解析错误
	}
	return ssh.KeysEqual(key, allowed)
}
