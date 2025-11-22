package api

import (
	"encoding/base64"
	"net/http"

	"binrc.com/roma/core/model"
	"binrc.com/roma/core/operation"
	"binrc.com/roma/core/sshd"
	"binrc.com/roma/core/utils"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/ssh"
)

type SSHKeyController struct{}

func NewSSHKeyController() *SSHKeyController {
	return &SSHKeyController{}
}

type SSHKeyResponse struct {
	PublicKey  string `json:"public_key"`  // SSH 公钥
	PrivateKey string `json:"private_key"` // SSH 私钥（仅创建时返回）
}

// GetMySSHKey 获取当前用户的 SSH 公钥
func (sc *SSHKeyController) GetMySSHKey(c *gin.Context) {
	utilG := utils.Gin{C: c}

	// 从上下文获取用户
	user, exists := c.Get("user")
	if !exists {
		utilG.Response(http.StatusUnauthorized, utils.ERROR, "未认证")
		return
	}

	currentUser := user.(*model.User)

	// 如果用户没有公钥，返回空
	if currentUser.PublicKey == "" {
		utilG.Response(http.StatusOK, utils.SUCCESS, SSHKeyResponse{
			PublicKey:  "",
			PrivateKey: "",
		})
		return
	}

	// 只返回公钥的头尾部分（安全考虑）
	maskedPublicKey := maskKey(currentUser.PublicKey)

	utilG.Response(http.StatusOK, utils.SUCCESS, SSHKeyResponse{
		PublicKey:  maskedPublicKey,
		PrivateKey: "", // 不返回私钥
	})
}

type UploadSSHKeyRequest struct {
	PublicKey  string `json:"public_key" binding:"required"`  // SSH 公钥
	PrivateKey string `json:"private_key" binding:"required"` // SSH 私钥（Base64 编码）
}

// UploadSSHKey 上传用户的 SSH 公私钥
func (sc *SSHKeyController) UploadSSHKey(c *gin.Context) {
	utilG := utils.Gin{C: c}

	// 从上下文获取用户
	user, exists := c.Get("user")
	if !exists {
		utilG.Response(http.StatusUnauthorized, utils.ERROR, "未认证")
		return
	}

	currentUser := user.(*model.User)

	var req UploadSSHKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utilG.Response(http.StatusBadRequest, utils.ERROR, "请提供公钥和私钥")
		return
	}

	// 验证公钥格式
	if len(req.PublicKey) < 20 {
		utilG.Response(http.StatusBadRequest, utils.ERROR, "公钥格式无效")
		return
	}

	// 解析公钥
	publicKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(req.PublicKey))
	if err != nil {
		utilG.Response(http.StatusBadRequest, utils.ERROR, "公钥格式无效")
		return
	}

	// 验证私钥格式（尝试解码 Base64）
	privateKeyBytes, err := base64.StdEncoding.DecodeString(req.PrivateKey)
	if err != nil {
		utilG.Response(http.StatusBadRequest, utils.ERROR, "私钥格式无效（需要 Base64 编码）")
		return
	}

	// 解析私钥
	signer, err := ssh.ParsePrivateKey(privateKeyBytes)
	if err != nil {
		utilG.Response(http.StatusBadRequest, utils.ERROR, "私钥格式无效")
		return
	}

	// 验证公私钥是否匹配（比较公钥的 Marshal 结果）
	publicKeyBytes := ssh.MarshalAuthorizedKey(publicKey)
	signerPublicKeyBytes := ssh.MarshalAuthorizedKey(signer.PublicKey())
	if string(publicKeyBytes) != string(signerPublicKeyBytes) {
		utilG.Response(http.StatusBadRequest, utils.ERROR, "公钥和私钥不匹配")
		return
	}

	// 更新用户的公钥（私钥不存储在数据库中，由用户自己保管）
	currentUser.PublicKey = req.PublicKey

	opUser := operation.NewUserOperation()
	_, err = opUser.UpdateUser(currentUser)
	if err != nil {
		utilG.Response(http.StatusInternalServerError, utils.ERROR, "更新 SSH 密钥失败")
		return
	}

	utilG.Response(http.StatusOK, utils.SUCCESS, "SSH 密钥上传成功")
}

// GenerateSSHKey 重新生成用户的 SSH 公私钥
func (sc *SSHKeyController) GenerateSSHKey(c *gin.Context) {
	utilG := utils.Gin{C: c}

	// 从上下文获取用户
	user, exists := c.Get("user")
	if !exists {
		utilG.Response(http.StatusUnauthorized, utils.ERROR, "未认证")
		return
	}

	currentUser := user.(*model.User)

	// 生成新的 SSH 密钥对
	privateKeyBase64, publicKeyBase64, err := sshd.GenKey()
	if err != nil {
		utilG.Response(http.StatusInternalServerError, utils.ERROR, "生成 SSH 密钥失败")
		return
	}

	// 解码公钥（从 Base64 转为字符串）
	publicKeyStr := string(publicKeyBase64)

	// 更新用户的公钥
	currentUser.PublicKey = publicKeyStr

	opUser := operation.NewUserOperation()
	_, err = opUser.UpdateUser(currentUser)
	if err != nil {
		utilG.Response(http.StatusInternalServerError, utils.ERROR, "保存 SSH 密钥失败")
		return
	}

	// 返回公私钥（仅创建时返回私钥）
	utilG.Response(http.StatusOK, utils.SUCCESS, SSHKeyResponse{
		PublicKey:  publicKeyStr,
		PrivateKey: string(privateKeyBase64), // 返回 Base64 编码的私钥
	})
}
