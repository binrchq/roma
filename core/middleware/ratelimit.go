package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"binrc.com/roma/core/utils"
	"github.com/gin-gonic/gin"
)

// RateLimiter 速率限制器
type RateLimiter struct {
	// 每个IP的连接数限制
	maxConnectionsPerIP int
	// 每个IP的连接速率限制（每秒）
	maxConnectionsPerSecond int
	// IP连接计数
	ipConnections map[string]int
	// IP最后连接时间
	ipLastConnection map[string][]time.Time
	// 互斥锁
	mu sync.RWMutex
	// 清理定时器
	cleanupTicker *time.Ticker
}

var globalRateLimiter *RateLimiter

// InitRateLimiter 初始化速率限制器
// 输入: maxConnectionsPerIP - 每个IP的最大并发连接数；maxConnectionsPerSecond - 每个IP每秒最大连接数
// 输出: 无
// 必要性: 防止DDoS攻击，限制单个IP的连接数和连接速率
func InitRateLimiter(maxConnectionsPerIP, maxConnectionsPerSecond int) {
	globalRateLimiter = &RateLimiter{
		maxConnectionsPerIP:     maxConnectionsPerIP,
		maxConnectionsPerSecond: maxConnectionsPerSecond,
		ipConnections:           make(map[string]int),
		ipLastConnection:        make(map[string][]time.Time),
		cleanupTicker:           time.NewTicker(1 * time.Minute),
	}

	// 启动清理协程，定期清理过期的连接记录
	go func() {
		for range globalRateLimiter.cleanupTicker.C {
			globalRateLimiter.cleanup()
		}
	}()
}

// cleanup 清理过期的连接记录
func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	// 清理1分钟前的连接记录
	cutoff := now.Add(-1 * time.Minute)

	for ip, times := range rl.ipLastConnection {
		validTimes := []time.Time{}
		for _, t := range times {
			if t.After(cutoff) {
				validTimes = append(validTimes, t)
			}
		}
		if len(validTimes) == 0 {
			delete(rl.ipLastConnection, ip)
		} else {
			rl.ipLastConnection[ip] = validTimes
		}
	}
}

// AllowConnection 检查是否允许新连接
// 输入: ip - 客户端IP地址
// 输出: bool - 是否允许连接；string - 拒绝原因（如果不允许）
// 必要性: 实现连接数限制和连接速率限制
func (rl *RateLimiter) AllowConnection(ip string) (bool, string) {
	if rl == nil {
		return true, ""
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	// 检查并发连接数
	currentConnections := rl.ipConnections[ip]
	if currentConnections >= rl.maxConnectionsPerIP {
		return false, fmt.Sprintf("too many connections from %s (max: %d)", ip, rl.maxConnectionsPerIP)
	}

	// 检查连接速率
	now := time.Now()
	lastConnections := rl.ipLastConnection[ip]
	// 清理1秒前的连接记录
	recentConnections := []time.Time{}
	for _, t := range lastConnections {
		if now.Sub(t) < 1*time.Second {
			recentConnections = append(recentConnections, t)
		}
	}

	if len(recentConnections) >= rl.maxConnectionsPerSecond {
		return false, fmt.Sprintf("connection rate limit exceeded for %s (max: %d/sec)", ip, rl.maxConnectionsPerSecond)
	}

	// 允许连接，更新计数
	rl.ipConnections[ip] = currentConnections + 1
	recentConnections = append(recentConnections, now)
	rl.ipLastConnection[ip] = recentConnections

	return true, ""
}

// ReleaseConnection 释放连接
// 输入: ip - 客户端IP地址
// 输出: 无
// 必要性: 连接关闭时减少连接计数
func (rl *RateLimiter) ReleaseConnection(ip string) {
	if rl == nil {
		return
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	if count, exists := rl.ipConnections[ip]; exists && count > 0 {
		rl.ipConnections[ip] = count - 1
		if rl.ipConnections[ip] == 0 {
			delete(rl.ipConnections, ip)
		}
	}
}

// RateLimitMiddleware 速率限制中间件
// 用途: 限制API请求速率，防止DDoS攻击
// 输入: c - Gin上下文
// 输出: 无
// 必要性: 保护API服务免受DDoS攻击
func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if ip == "" {
			ip = c.GetHeader("X-Forwarded-For")
		}
		if ip == "" {
			ip = c.GetHeader("X-Real-IP")
		}

		allowed, reason := globalRateLimiter.AllowConnection(ip)
		if !allowed {
			utilG := utils.Gin{C: c}
			utilG.Response(http.StatusTooManyRequests, utils.ERROR, reason)
			c.Abort()
			return
		}

		// 请求完成后释放连接计数
		c.Next()
		globalRateLimiter.ReleaseConnection(ip)
	}
}
