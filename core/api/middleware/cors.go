package middleware

import (
	"fmt"
	"os"
	"strings"
	"time"

	"binrc.com/roma/core/global"
	"binrc.com/roma/core/utils/logger"

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
	var configSource string

	// 优先从配置文件读取
	if global.CONFIG != nil && global.CONFIG.Api != nil && global.CONFIG.Api.CorsAllowOrigins != "" {
		allowOriginsEnv = global.CONFIG.Api.CorsAllowOrigins
		configSource = "config file"
	} else {
		// 其次从环境变量读取（支持两种格式）
		allowOriginsEnv = os.Getenv("ROMA_API_CORS_ALLOW_ORIGINS")
		if allowOriginsEnv != "" {
			configSource = "ROMA_API_CORS_ALLOW_ORIGINS"
		} else {
			allowOriginsEnv = os.Getenv("ROMA_CORS_ALLOW_ORIGINS")
			if allowOriginsEnv != "" {
				configSource = "ROMA_CORS_ALLOW_ORIGINS"
			}
		}
	}

	var allowOrigins []string
	var allowOriginMap = make(map[string]bool) // 用于快速查找

	if allowOriginsEnv != "" {
		// 解析逗号分隔的域名列表
		origins := strings.Split(allowOriginsEnv, ",")
		for _, origin := range origins {
			origin = strings.TrimSpace(origin)
			if origin != "" {
				// 移除尾随斜杠
				origin = strings.TrimSuffix(origin, "/")
				allowOrigins = append(allowOrigins, origin)
				allowOriginMap[origin] = true
			}
		}
	}

	// 如果没有配置域名，默认允许所有来源（开发环境）
	allowAll := len(allowOrigins) == 0
	if allowAll {
		allowOrigins = []string{"*"}
	}

	// 输出CORS配置日志
	logger.Logger.Info(fmt.Sprintf("CORS配置已加载 (来源: %s): %v", configSource, allowOrigins))

	// 使用 AllowOriginFunc 来更灵活地匹配域名
	allowOriginFunc := func(origin string) bool {
		if allowAll {
			return true
		}
		// 移除尾随斜杠
		origin = strings.TrimSuffix(origin, "/")
		// 精确匹配
		if allowOriginMap[origin] {
			return true
		}
		// 如果精确匹配失败，记录日志用于调试
		logger.Logger.Debug(fmt.Sprintf("CORS: Origin '%s' 不在允许列表中: %v", origin, allowOrigins))
		return false
	}

	return cors.New(cors.Config{
		// 使用 AllowOriginFunc 来更灵活地匹配
		AllowOriginFunc: allowOriginFunc,
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
			"X-Browser-Fingerprint", // 添加浏览器指纹头
		},
		// 暴露的响应头
		ExposeHeaders: []string{
			"Content-Length",
			"Content-Type",
		},
		// 允许携带凭证（cookies, authorization headers）
		// 注意：如果 AllowOriginFunc 返回 true 对所有来源，AllowCredentials 必须为 false
		AllowCredentials: !allowAll,
		// 预检请求的缓存时间
		MaxAge: 12 * time.Hour,
	})
}
