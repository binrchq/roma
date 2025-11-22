package middleware

import (
	"binrc.com/roma/core/operation"
	"binrc.com/roma/core/utils"
	"github.com/gin-gonic/gin"
)

// JWTAuth JWT 认证中间件
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		utilG := utils.Gin{C: c}

		// 从 Header 获取 token
		token := c.GetHeader("Authorization")
		if token == "" {
			utilG.Response(401, utils.ERROR, "未提供认证令牌")
			c.Abort()
			return
		}

		// 移除 "Bearer " 前缀
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}

		// 解析 JWT token
		claims, err := utils.ParseJWT(token)
		if err != nil {
			utilG.Response(401, utils.ERROR, "无效的认证令牌")
			c.Abort()
			return
		}

		// 根据 user_id 获取用户信息
		opUser := operation.NewUserOperation()
		user, err := opUser.GetUserByID(claims.UserID)
		if err != nil {
			utilG.Response(401, utils.ERROR, "用户不存在")
			c.Abort()
			return
		}

		// 将用户信息存储到上下文
		c.Set("user", user)
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)

		c.Next()
	}
}

// JWTAuthOrApiKey 支持 JWT 或 API Key 认证
func JWTAuthOrApiKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 先尝试 JWT 认证
		token := c.GetHeader("Authorization")
		if token != "" {
			if len(token) > 7 && token[:7] == "Bearer " {
				token = token[7:]
			}
			claims, err := utils.ParseJWT(token)
			if err == nil {
				// JWT 认证成功
				opUser := operation.NewUserOperation()
				user, err := opUser.GetUserByID(claims.UserID)
				if err == nil {
					c.Set("user", user)
					c.Set("user_id", claims.UserID)
					c.Set("username", claims.Username)
					c.Next()
					return
				}
			}
		}

		// 如果 JWT 认证失败，尝试 API Key 认证
		apiKey := c.GetHeader("apikey")
		if apiKey == "" {
			apiKey = c.Query("apikey")
		}

		if apiKey != "" {
			opKey := operation.NewApikeyOperation()
			valid, err := opKey.ApiKeyIsValid(apiKey)
			if err == nil && valid {
				// API Key 认证成功，使用 GetUserFromContext 的逻辑
				opUser := operation.NewUserOperation()
				users, err := opUser.GetAllUsers()
				if err == nil && len(users) > 0 {
					// 优先查找 super 角色的用户
					for _, u := range users {
						roles, err := opUser.GetUserRoles(u.ID)
						if err == nil {
							for _, role := range roles {
								if role.Name == "super" {
									c.Set("user", u)
									c.Set("user_id", u.ID)
									c.Set("username", u.Username)
									c.Next()
									return
								}
							}
						}
					}
					// 如果没有 super 用户，返回第一个用户
					c.Set("user", users[0])
					c.Set("user_id", users[0].ID)
					c.Set("username", users[0].Username)
					c.Next()
					return
				}
			}
		}

		// 两种认证都失败
		utilG := utils.Gin{C: c}
		utilG.Response(401, utils.ERROR, "未认证")
		c.Abort()
	}
}
