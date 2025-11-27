package middleware

import (
	"net/http"
	"time"

	"binrc.com/roma/core/model"
	"binrc.com/roma/core/operation"
	"binrc.com/roma/core/utils"
	"github.com/gin-gonic/gin"
)

var globalBlacklistOp *operation.BlacklistOperation

// InitIPBlacklist 初始化IP黑名单（使用数据库）
// 输入: 无
// 输出: 无
// 必要性: 管理恶意IP，防止DDoS攻击和暴力破解，使用数据库持久化
func InitIPBlacklist() {
	globalBlacklistOp = operation.NewBlacklistOperation()

	// 启动清理协程，定期清理过期的黑名单
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			globalBlacklistOp.CleanExpired()
		}
	}()
}

// AddToBlacklist 添加IP到全局黑名单（导出函数，供其他包调用）
// 输入: ip - IP地址；duration - 封禁时长（0表示永久封禁）；reason - 封禁原因；source - 封禁来源
// 输出: error - 错误信息
// 必要性: 允许其他包（如SSH安全模块）添加IP到黑名单，保存到数据库
func AddToBlacklist(ip string, duration time.Duration, reason, source string) error {
	if globalBlacklistOp == nil {
		return nil
	}

	blacklist := &model.Blacklist{
		IP:     ip,
		Reason: reason,
		Source: source,
	}

	if duration > 0 {
		banUntil := time.Now().Add(duration)
		blacklist.BanUntil = &banUntil
	}

	// 获取IP信息
	ipInfo, err := GetIPInfo(ip)
	if err == nil && ipInfo != "" {
		blacklist.IPInfo = ipInfo
	}

	_, err = globalBlacklistOp.CreateOrUpdate(blacklist)
	return err
}

// RemoveFromBlacklist 从全局黑名单移除IP（导出函数，供其他包调用）
// 输入: ip - IP地址
// 输出: error - 错误信息
// 必要性: 允许其他包手动解封IP
func RemoveFromBlacklist(ip string) error {
	if globalBlacklistOp == nil {
		return nil
	}
	return globalBlacklistOp.Delete(ip)
}

// IPBlacklistMiddleware IP黑名单中间件
// 用途: 阻止黑名单IP访问系统
// 输入: c - Gin上下文
// 输出: 无
// 必要性: 防止恶意IP访问系统
func IPBlacklistMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if ip == "" {
			ip = c.GetHeader("X-Forwarded-For")
		}
		if ip == "" {
			ip = c.GetHeader("X-Real-IP")
		}

		if globalBlacklistOp != nil {
			isBlacklisted, blacklist := globalBlacklistOp.IsBlacklisted(ip)
			if isBlacklisted {
				utilG := utils.Gin{C: c}
				reason := "IP address is blacklisted"
				if blacklist != nil && blacklist.Reason != "" {
					reason = blacklist.Reason
				}
				utilG.Response(http.StatusForbidden, utils.ERROR, reason)
				c.Abort()
				return
			}
		}

		c.Next()
	}
}
