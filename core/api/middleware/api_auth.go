package middleware

import (
	"bitrec.ai/roma/core/operation"
	"bitrec.ai/roma/core/utils"
	"github.com/gin-gonic/gin"
)

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
		c.Next()
	}
}

func isValidApiKey(apiKey string) bool {
	op := operation.NewApikeyOperation()
	exists, err := op.ApiKeyExists(apiKey)
	if err != nil {
		return false
	}
	return exists
}
