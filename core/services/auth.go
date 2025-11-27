package services

import (
	"log"
	"strings"

	"binrc.com/roma/core/operation"
	"github.com/loganchef/ssh"
)

// PublicKeyAuth 公钥认证函数（内部使用，不包含安全检查）
// 输入: ctx - SSH上下文；key - 客户端提供的公钥
// 输出: bool - 是否认证成功
// 必要性: 这是公钥认证的核心逻辑，由SecurePublicKeyAuth包装后使用
func PublicKeyAuth(ctx ssh.Context, key ssh.PublicKey) bool {
	username := ctx.User()
	op := operation.NewUserOperation()
	user, err := op.GetUserByUsername(username)
	if err != nil {
		log.Printf("PublicKeyAuth: 用户 %s 不存在: %v", username, err)
		return false
	}

	// 如果用户没有设置公钥，拒绝认证
	if user.PublicKey == "" {
		log.Printf("PublicKeyAuth: 用户 %s 没有设置公钥", username)
		return false
	}

	// 清理公钥字符串（去除前后空白和换行符）
	pub := strings.TrimSpace(user.PublicKey)
	if pub == "" {
		log.Printf("PublicKeyAuth: 用户 %s 的公钥为空", username)
		return false
	}

	// 解析数据库中存储的公钥
	allowed, _, _, _, err := ssh.ParseAuthorizedKey([]byte(pub))
	if err != nil {
		log.Printf("PublicKeyAuth: 用户 %s 的公钥解析失败: %v, 公钥内容: %s", username, err, pub[:min(len(pub), 50)])
		return false
	}

	// 比较客户端提供的公钥和数据库中存储的公钥
	isEqual := ssh.KeysEqual(key, allowed)
	if !isEqual {
		log.Printf("PublicKeyAuth: 用户 %s 的公钥不匹配", username)
	}
	return isEqual
}

// min 返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
