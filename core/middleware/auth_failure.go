package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"binrc.com/roma/core/utils"
	"github.com/gin-gonic/gin"
)

// AuthFailureTracker 认证失败追踪器
type AuthFailureTracker struct {
	// 标识符（浏览器指纹或IP） -> 失败次数
	failureCount map[string]int
	// 标识符 -> 最后失败时间
	lastFailureTime map[string]time.Time
	// 标识符 -> 封禁到期时间
	banUntil map[string]time.Time
	// 互斥锁
	mu sync.RWMutex
	// 清理定时器
	cleanupTicker *time.Ticker
	// 配置
	maxFailures        int           // 最大失败次数
	banDuration        time.Duration // 封禁时长
	failureWindow      time.Duration // 失败计数窗口
	exponentialBackoff bool          // 是否启用指数退避
}

var globalAuthFailureTracker *AuthFailureTracker

// InitAuthFailureTracker 初始化认证失败追踪器
// 输入: maxFailures - 最大失败次数；banDuration - 封禁时长；failureWindow - 失败计数窗口；exponentialBackoff - 是否启用指数退避
// 输出: 无
// 必要性: 防止暴力破解，自动封禁频繁认证失败的IP
func InitAuthFailureTracker(maxFailures int, banDuration, failureWindow time.Duration, exponentialBackoff bool) {
	globalAuthFailureTracker = &AuthFailureTracker{
		failureCount:       make(map[string]int),
		lastFailureTime:    make(map[string]time.Time),
		banUntil:           make(map[string]time.Time),
		cleanupTicker:      time.NewTicker(1 * time.Minute),
		maxFailures:        maxFailures,
		banDuration:        banDuration,
		failureWindow:      failureWindow,
		exponentialBackoff: exponentialBackoff,
	}

	// 启动清理协程
	go func() {
		for range globalAuthFailureTracker.cleanupTicker.C {
			globalAuthFailureTracker.cleanup()
		}
	}()
}

// cleanup 清理过期的失败记录
func (aft *AuthFailureTracker) cleanup() {
	aft.mu.Lock()
	defer aft.mu.Unlock()

	now := time.Now()
	// 清理过期的失败计数
	for ip, lastTime := range aft.lastFailureTime {
		if now.Sub(lastTime) > aft.failureWindow {
			delete(aft.failureCount, ip)
			delete(aft.lastFailureTime, ip)
		}
	}
	// 清理过期的封禁记录
	for ip, banTime := range aft.banUntil {
		if now.After(banTime) {
			delete(aft.banUntil, ip)
			delete(aft.failureCount, ip)
			delete(aft.lastFailureTime, ip)
		}
	}
}

// RecordFailure 记录认证失败
// 输入: identifier - 标识符（浏览器指纹或IP地址）
// 输出: bool - 是否应该封禁；time.Time - 解封时间（如果被封禁）
// 必要性: 记录认证失败，达到阈值后自动封禁
func (aft *AuthFailureTracker) RecordFailure(identifier string) (bool, time.Time) {
	if aft == nil {
		return false, time.Time{}
	}

	// 如果identifier是IP地址，跳过内网IP
	if utils.IsPrivateIP(identifier) {
		return false, time.Time{}
	}

	aft.mu.Lock()
	defer aft.mu.Unlock()

	now := time.Now()

	// 检查是否在封禁期内
	if banTime, exists := aft.banUntil[identifier]; exists && now.Before(banTime) {
		return true, banTime
	}

	// 检查失败计数窗口
	lastTime, exists := aft.lastFailureTime[identifier]
	if !exists || now.Sub(lastTime) > aft.failureWindow {
		// 重置计数
		aft.failureCount[identifier] = 1
		aft.lastFailureTime[identifier] = now
		return false, time.Time{}
	}

	// 增加失败计数
	aft.failureCount[identifier]++
	aft.lastFailureTime[identifier] = now

	// 检查是否达到封禁阈值
	if aft.failureCount[identifier] >= aft.maxFailures {
		// 计算封禁时长（指数退避）
		banDuration := aft.banDuration
		if aft.exponentialBackoff {
			// 指数退避：1次封禁15分钟，2次30分钟，3次1小时，4次2小时...
			banCount := (aft.failureCount[identifier] - aft.maxFailures) / aft.maxFailures
			for i := 0; i < banCount; i++ {
				banDuration *= 2
			}
			// 最大封禁24小时
			if banDuration > 24*time.Hour {
				banDuration = 24 * time.Hour
			}
		}

		banTime := now.Add(banDuration)
		aft.banUntil[identifier] = banTime
		return true, banTime
	}

	return false, time.Time{}
}

// RecordSuccess 记录认证成功
// 输入: identifier - 标识符（浏览器指纹或IP地址）
// 输出: 无
// 必要性: 认证成功时清除失败计数
func (aft *AuthFailureTracker) RecordSuccess(identifier string) {
	if aft == nil {
		return
	}

	aft.mu.Lock()
	defer aft.mu.Unlock()

	delete(aft.failureCount, identifier)
	delete(aft.lastFailureTime, identifier)
	delete(aft.banUntil, identifier)
}

// IsBanned 检查标识符是否被封禁
// 输入: identifier - 标识符（浏览器指纹或IP地址）
// 输出: bool - 是否被封禁；time.Time - 解封时间（如果被封禁）
// 必要性: 检查标识符是否在封禁期内
func (aft *AuthFailureTracker) IsBanned(identifier string) (bool, time.Time) {
	if aft == nil {
		return false, time.Time{}
	}

	aft.mu.RLock()
	defer aft.mu.RUnlock()

	if banTime, exists := aft.banUntil[identifier]; exists && time.Now().Before(banTime) {
		return true, banTime
	}

	return false, time.Time{}
}

// getIdentifier 获取客户端标识符（优先使用浏览器指纹，否则使用IP）
func getIdentifier(c *gin.Context) string {
	// 优先使用浏览器指纹
	fingerprint := c.GetHeader("X-Browser-Fingerprint")
	if fingerprint != "" {
		return fingerprint
	}

	// 如果没有浏览器指纹，使用IP作为后备
	ip := c.ClientIP()
	if ip == "" {
		ip = c.GetHeader("X-Forwarded-For")
	}
	if ip == "" {
		ip = c.GetHeader("X-Real-IP")
	}
	return ip
}

// AuthFailureMiddleware 认证失败处理中间件
// 用途: 记录认证失败，自动封禁恶意客户端（基于浏览器指纹或IP）
// 输入: c - Gin上下文
// 输出: 无
// 必要性: 防止暴力破解攻击
func AuthFailureMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 仅对认证相关路由生效
		if c.Request.URL.Path != "/api/v1/auth/login" {
			c.Next()
			return
		}

		identifier := getIdentifier(c)

		// 检查是否被封禁
		banned, unbanTime := globalAuthFailureTracker.IsBanned(identifier)
		if banned {
			utilG := utils.Gin{C: c}
			utilG.Response(http.StatusTooManyRequests, utils.ERROR,
				fmt.Sprintf("Too many failed login attempts. Unban time: %s", unbanTime.Format("2006-01-02 15:04:05")))
			c.Abort()
			return
		}

		c.Next()

		// 检查响应状态码，记录认证失败
		if c.Writer.Status() == http.StatusUnauthorized || c.Writer.Status() == http.StatusForbidden {
			shouldBan, banTime := globalAuthFailureTracker.RecordFailure(identifier)
			if shouldBan {
				// 如果identifier是IP地址，添加到黑名单（保存到数据库）
				// 浏览器指纹不添加到IP黑名单，只在内存中封禁
				if !utils.IsPrivateIP(identifier) {
					reason := fmt.Sprintf("Too many failed login attempts (banned after %d failures)", globalAuthFailureTracker.maxFailures)
					AddToBlacklist(identifier, globalAuthFailureTracker.banDuration, reason, "api_auth_failure")
				}

				// 如果响应还未发送，返回包含解封时间的错误信息
				if !c.Writer.Written() {
					utilG := utils.Gin{C: c}
					utilG.Response(http.StatusInternalServerError, utils.ERROR,
						fmt.Sprintf("Too many failed login attempts (banned after %d failures). Unban time: %s",
							globalAuthFailureTracker.maxFailures, banTime.Format("2006-01-02 15:04:05")))
					c.Abort()
					return
				}
			}
		} else if c.Writer.Status() == http.StatusOK {
			// 认证成功，清除失败记录
			globalAuthFailureTracker.RecordSuccess(identifier)
		}
	}
}
