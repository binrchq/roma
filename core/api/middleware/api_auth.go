package middleware

import (
	"binrc.com/roma/core/operation"
	"binrc.com/roma/core/utils"
	"github.com/gin-gonic/gin"
)

// ApiKeyAuth API Key 认证中间件
// 验证 API Key 的有效性，但不进行权限检查
func ApiKeyAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		utilG := utils.Gin{C: c}
		apiKey := c.GetHeader("apikey")
		if apiKey == "" {
			apiKey = c.Query("apikey")
			if apiKey == "" {
				utilG.Response(utils.ERROR, utils.ERROR, "API key is missing")
				c.Abort()
				return
			}
		}
		if !isValidApiKey(apiKey) {
			utilG.Response(utils.ERROR, utils.ERROR, "Invalid API key")
			c.Abort()
			return
		}
		// 将 API Key 存储到上下文，供后续使用
		c.Set("api_key", apiKey)
		c.Next()
	}
}

func isValidApiKey(apiKey string) bool {
	op := operation.NewApikeyOperation()
	exists, err := op.ApiKeyExists(apiKey)
	if err != nil {
		return false
	}
	valid, err := op.ApiKeyIsValid(apiKey)
	if err != nil {
		return false
	}
	if !exists || !valid {
		return false
	}
	return exists
}
