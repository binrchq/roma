package middleware

import (
	"os"
	"strings"
	"time"

	"binrc.com/roma/core/global"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORSMiddleware 配置 CORS 中间件
// 前后端分离时，前端直接请求后端 API，需要支持跨域
// 允许的域名列表优先级：配置文件 > 环境变量 > 默认值（允许所有）
// 配置文件：api.cors_allow_origins
// 环境变量：ROMA_API_CORS_ALLOW_ORIGINS 或 ROMA_CORS_ALLOW_ORIGINS
// 多个域名用逗号分隔，例如：https://roma.binrc.com,https://roma-demo.binrc.com
func CORSMiddleware() gin.HandlerFunc {
	var allowOriginsEnv string

	// 优先从配置文件读取
	if global.CONFIG != nil && global.CONFIG.Api != nil && global.CONFIG.Api.CorsAllowOrigins != "" {
		allowOriginsEnv = global.CONFIG.Api.CorsAllowOrigins
	} else {
		// 其次从环境变量读取（支持两种格式）
		allowOriginsEnv = os.Getenv("ROMA_API_CORS_ALLOW_ORIGINS")
		if allowOriginsEnv == "" {
			allowOriginsEnv = os.Getenv("ROMA_CORS_ALLOW_ORIGINS")
		}
	}

	var allowOrigins []string

	if allowOriginsEnv != "" {
		// 解析逗号分隔的域名列表
		origins := strings.Split(allowOriginsEnv, ",")
		for _, origin := range origins {
			origin = strings.TrimSpace(origin)
			if origin != "" {
				allowOrigins = append(allowOrigins, origin)
			}
		}
	}

	// 如果没有配置域名，默认允许所有来源（开发环境）
	if len(allowOrigins) == 0 {
		allowOrigins = []string{"*"}
	}

	return cors.New(cors.Config{
		// 允许的来源（从环境变量注入）
		AllowOrigins: allowOrigins,
		// 允许的 HTTP 方法
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		// 允许的请求头
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Content-Length",
			"Accept-Encoding",
			"X-CSRF-Token",
			"Authorization",
			"Accept",
			"apikey",
			"X-Requested-With",
		},
		// 暴露的响应头
		ExposeHeaders: []string{
			"Content-Length",
			"Content-Type",
		},
		// 允许携带凭证（cookies, authorization headers）
		// 注意：如果 AllowOrigins 包含 "*"，AllowCredentials 必须为 false
		AllowCredentials: len(allowOrigins) > 0 && allowOrigins[0] != "*",
		// 预检请求的缓存时间
		MaxAge: 12 * time.Hour,
	})
}
