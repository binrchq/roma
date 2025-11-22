package api

import (
	"net/http"

	"binrc.com/roma/core/model"
	"binrc.com/roma/core/operation"
	"binrc.com/roma/core/utils"
	"github.com/gin-gonic/gin"
)

type AuthController struct{}

func NewAuthController() *AuthController {
	return &AuthController{}
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token string      `json:"token"` // JWT token
	User  *model.User `json:"user"`  // 用户信息
}

// Login 用户登录
func (ac *AuthController) Login(c *gin.Context) {
	utilG := utils.Gin{C: c}
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utilG.Response(http.StatusBadRequest, utils.ERROR, "请输入用户名和密码")
		return
	}

	// 查找用户
	opUser := operation.NewUserOperation()
	user, err := opUser.GetUserByUsername(req.Username)
	if err != nil {
		utilG.Response(http.StatusUnauthorized, utils.ERROR, "用户名或密码错误")
		return
	}

	// 验证密码（目前密码是明文存储，直接比较）
	// TODO: 如果后续使用 bcrypt，需要在这里验证
	if user.Password != req.Password {
		utilG.Response(http.StatusUnauthorized, utils.ERROR, "用户名或密码错误")
		return
	}

	// 生成 JWT token
	token, err := utils.GenerateJWT(user.ID, user.Username)
	if err != nil {
		utilG.Response(http.StatusInternalServerError, utils.ERROR, "生成认证令牌失败")
		return
	}

	// 返回登录信息
	response := LoginResponse{
		Token: token,
		User:  user,
	}

	utilG.Response(http.StatusOK, utils.SUCCESS, response)
}

// Logout 用户登出
func (ac *AuthController) Logout(c *gin.Context) {
	utilG := utils.Gin{C: c}
	// 登出主要是客户端删除 token/API Key
	// 服务端可以选择使当前的 API Key 失效，但这里简化处理
	utilG.Response(http.StatusOK, utils.SUCCESS, "登出成功")
}

