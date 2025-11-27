package sshd

import (
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"binrc.com/roma/core/operation"
	"binrc.com/roma/core/utils/logger"
	"github.com/loganchef/ssh"
)

// sshBlacklist SSH专用的黑名单（与API黑名单独立，避免循环导入）
var sshBlacklist = struct {
	permanent map[string]bool
	temporary map[string]time.Time
	mu        sync.RWMutex
}{
	permanent: make(map[string]bool),
	temporary: make(map[string]time.Time),
}

// addToSSHBlacklist 添加IP到SSH黑名单
func addToSSHBlacklist(ip string, duration time.Duration) {
	sshBlacklist.mu.Lock()
	defer sshBlacklist.mu.Unlock()

	if duration == 0 {
		sshBlacklist.permanent[ip] = true
	} else {
		sshBlacklist.temporary[ip] = time.Now().Add(duration)
	}
}

// isSSHBlacklisted 检查IP是否在SSH黑名单中
func isSSHBlacklisted(ip string) bool {
	sshBlacklist.mu.RLock()
	defer sshBlacklist.mu.RUnlock()

	if sshBlacklist.permanent[ip] {
		return true
	}

	if unbanTime, exists := sshBlacklist.temporary[ip]; exists && time.Now().Before(unbanTime) {
		return true
	}

	return false
}

// SSHSecurityManager SSH安全管理器
type SSHSecurityManager struct {
	// IP连接计数
	ipConnections map[string]int
	// IP最后连接时间
	ipLastConnection map[string][]time.Time
	// IP认证失败计数
	ipAuthFailures map[string]int
	// IP最后失败时间
	ipLastFailure map[string]time.Time
	// IP封禁到期时间
	ipBanUntil map[string]time.Time
	// 互斥锁
	mu sync.RWMutex
	// 清理定时器
	cleanupTicker *time.Ticker
	// 配置
	maxConnectionsPerIP     int
	maxConnectionsPerSecond int
	maxAuthFailures         int
	banDuration             time.Duration
	failureWindow           time.Duration
}

var globalSSHSecurityManager *SSHSecurityManager

// InitSSHSecurity 初始化SSH安全管理器
// 输入: maxConnectionsPerIP - 每个IP最大并发连接数；maxConnectionsPerSecond - 每秒最大连接数；maxAuthFailures - 最大认证失败次数；banDuration - 封禁时长；failureWindow - 失败计数窗口
// 输出: 无
// 必要性: 保护SSH服务器免受DDoS攻击和暴力破解
func InitSSHSecurity(maxConnectionsPerIP, maxConnectionsPerSecond, maxAuthFailures int, banDuration, failureWindow time.Duration) {
	globalSSHSecurityManager = &SSHSecurityManager{
		ipConnections:           make(map[string]int),
		ipLastConnection:        make(map[string][]time.Time),
		ipAuthFailures:          make(map[string]int),
		ipLastFailure:           make(map[string]time.Time),
		ipBanUntil:              make(map[string]time.Time),
		cleanupTicker:           time.NewTicker(1 * time.Minute),
		maxConnectionsPerIP:     maxConnectionsPerIP,
		maxConnectionsPerSecond: maxConnectionsPerSecond,
		maxAuthFailures:         maxAuthFailures,
		banDuration:             banDuration,
		failureWindow:           failureWindow,
	}

	// 启动清理协程
	go func() {
		for range globalSSHSecurityManager.cleanupTicker.C {
			globalSSHSecurityManager.cleanup()
		}
	}()
}

// cleanup 清理过期的记录
func (sm *SSHSecurityManager) cleanup() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-1 * time.Minute)

	// 清理连接记录
	for ip, times := range sm.ipLastConnection {
		validTimes := []time.Time{}
		for _, t := range times {
			if t.After(cutoff) {
				validTimes = append(validTimes, t)
			}
		}
		if len(validTimes) == 0 {
			delete(sm.ipLastConnection, ip)
		} else {
			sm.ipLastConnection[ip] = validTimes
		}
	}

	// 清理失败记录
	for ip, lastTime := range sm.ipLastFailure {
		if now.Sub(lastTime) > sm.failureWindow {
			delete(sm.ipAuthFailures, ip)
			delete(sm.ipLastFailure, ip)
		}
	}

	// 清理过期的封禁记录
	for ip, banTime := range sm.ipBanUntil {
		if now.After(banTime) {
			delete(sm.ipBanUntil, ip)
			delete(sm.ipAuthFailures, ip)
			delete(sm.ipLastFailure, ip)
		}
	}
}

// AllowConnection 检查是否允许新连接
// 输入: ip - 客户端IP地址
// 输出: bool - 是否允许连接；string - 拒绝原因（如果不允许）
// 必要性: 实现连接数限制和连接速率限制
func (sm *SSHSecurityManager) AllowConnection(ip string) (bool, string) {
	if sm == nil {
		return true, ""
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	// 检查是否被封禁
	if banTime, exists := sm.ipBanUntil[ip]; exists && time.Now().Before(banTime) {
		return false, fmt.Sprintf("IP %s is temporarily banned until %s", ip, banTime.Format("2006-01-02 15:04:05"))
	}

	// 检查并发连接数
	currentConnections := sm.ipConnections[ip]
	if currentConnections >= sm.maxConnectionsPerIP {
		return false, fmt.Sprintf("too many connections from %s (max: %d)", ip, sm.maxConnectionsPerIP)
	}

	// 检查连接速率
	now := time.Now()
	lastConnections := sm.ipLastConnection[ip]
	recentConnections := []time.Time{}
	for _, t := range lastConnections {
		if now.Sub(t) < 1*time.Second {
			recentConnections = append(recentConnections, t)
		}
	}

	if len(recentConnections) >= sm.maxConnectionsPerSecond {
		return false, fmt.Sprintf("connection rate limit exceeded for %s (max: %d/sec)", ip, sm.maxConnectionsPerSecond)
	}

	// 允许连接，更新计数
	sm.ipConnections[ip] = currentConnections + 1
	recentConnections = append(recentConnections, now)
	sm.ipLastConnection[ip] = recentConnections

	return true, ""
}

// ReleaseConnection 释放连接
// 输入: ip - 客户端IP地址
// 输出: 无
// 必要性: 连接关闭时减少连接计数
func (sm *SSHSecurityManager) ReleaseConnection(ip string) {
	if sm == nil {
		return
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	if count, exists := sm.ipConnections[ip]; exists && count > 0 {
		sm.ipConnections[ip] = count - 1
		if sm.ipConnections[ip] == 0 {
			delete(sm.ipConnections, ip)
		}
	}
}

// RecordAuthFailure 记录认证失败
// 输入: ip - 客户端IP地址
// 输出: bool - 是否应该封禁该IP
// 必要性: 记录认证失败，达到阈值后自动封禁
func (sm *SSHSecurityManager) RecordAuthFailure(ip string) bool {
	if sm == nil {
		return false
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	now := time.Now()

	// 检查是否在封禁期内
	if banTime, exists := sm.ipBanUntil[ip]; exists && now.Before(banTime) {
		return true
	}

	// 检查失败计数窗口
	lastTime, exists := sm.ipLastFailure[ip]
	if !exists || now.Sub(lastTime) > sm.failureWindow {
		// 重置计数
		sm.ipAuthFailures[ip] = 1
		sm.ipLastFailure[ip] = now
		return false
	}

	// 增加失败计数
	sm.ipAuthFailures[ip]++
	sm.ipLastFailure[ip] = now

	// 检查是否达到封禁阈值
	if sm.ipAuthFailures[ip] >= sm.maxAuthFailures {
		sm.ipBanUntil[ip] = now.Add(sm.banDuration)
		logger.Logger.Warning(fmt.Sprintf("SSH: IP %s banned for %v due to %d failed authentication attempts", ip, sm.banDuration, sm.ipAuthFailures[ip]))
		// 添加到SSH黑名单（内存中）
		addToSSHBlacklist(ip, sm.banDuration)
		// 同时保存到数据库（通过middleware，但避免循环导入，使用异步方式）
		go func() {
			// 注意：这里不能直接导入middleware，避免循环导入
			// 实际保存到数据库的操作应该在API层统一处理
			// 或者通过事件/消息队列实现
		}()
		return true
	}

	return false
}

// RecordAuthSuccess 记录认证成功
// 输入: ip - 客户端IP地址
// 输出: 无
// 必要性: 认证成功时清除失败计数
func (sm *SSHSecurityManager) RecordAuthSuccess(ip string) {
	if sm == nil {
		return
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	delete(sm.ipAuthFailures, ip)
	delete(sm.ipLastFailure, ip)
	delete(sm.ipBanUntil, ip)
}

// GetClientIP 从SSH连接中获取客户端IP
// 输入: ctx - SSH上下文（可以是ssh.Context或ssh.Session）
// 输出: string - 客户端IP地址
// 必要性: 从SSH连接中提取IP地址用于安全控制
func GetClientIP(ctx interface{}) string {
	var remoteAddr net.Addr
	switch v := ctx.(type) {
	case ssh.Context:
		remoteAddr = v.RemoteAddr()
	case ssh.Session:
		remoteAddr = v.RemoteAddr()
	default:
		return ""
	}

	if remoteAddr == nil {
		return ""
	}
	host, _, err := net.SplitHostPort(remoteAddr.String())
	if err != nil {
		return remoteAddr.String()
	}
	return host
}

// publicKeyAuth 公钥认证函数（内部使用，避免循环导入）
// 输入: ctx - SSH上下文；key - 客户端提供的公钥
// 输出: bool - 是否认证成功
// 必要性: 这是公钥认证的核心逻辑，由SecurePublicKeyAuth包装后使用
func publicKeyAuth(ctx ssh.Context, key ssh.PublicKey) bool {
	username := ctx.User()
	op := operation.NewUserOperation()
	user, err := op.GetUserByUsername(username)
	if err != nil {
		log.Printf("PublicKeyAuth: 用户 %s 不存在: %v", username, err)
		return false
	}

	// 如果用户没有设置公钥，拒绝认证
	if user.PublicKey == "" {
		log.Printf("PublicKeyAuth: 用户 %s 没有设置公钥", username)
		return false
	}

	// 清理公钥字符串（去除前后空白和换行符）
	pub := strings.TrimSpace(user.PublicKey)
	if pub == "" {
		log.Printf("PublicKeyAuth: 用户 %s 的公钥为空", username)
		return false
	}

	// 解析数据库中存储的公钥
	allowed, _, _, _, err := ssh.ParseAuthorizedKey([]byte(pub))
	if err != nil {
		log.Printf("PublicKeyAuth: 用户 %s 的公钥解析失败: %v, 公钥内容: %s", username, err, pub[:min(len(pub), 50)])
		return false
	}

	// 比较客户端提供的公钥和数据库中存储的公钥
	isEqual := ssh.KeysEqual(key, allowed)
	if !isEqual {
		log.Printf("PublicKeyAuth: 用户 %s 的公钥不匹配", username)
	}
	return isEqual
}

// SecurePublicKeyAuth 安全的公钥认证包装器
// 输入: ctx - SSH上下文；key - 客户端提供的公钥
// 输出: bool - 是否认证成功
// 必要性: 在公钥认证中添加安全检查和失败记录
func SecurePublicKeyAuth(ctx ssh.Context, key ssh.PublicKey) bool {
	ip := GetClientIP(ctx)
	if ip == "" {
		ip = ctx.RemoteAddr().String()
	}

	// 检查是否在SSH黑名单中
	if isSSHBlacklisted(ip) {
		logger.Logger.Warning(fmt.Sprintf("SSH: Blocked connection attempt from blacklisted IP %s", ip))
		return false
	}

	// 检查是否被封禁（由SSH安全管理器管理）
	if globalSSHSecurityManager != nil {
		sm := globalSSHSecurityManager
		sm.mu.RLock()
		if banTime, exists := sm.ipBanUntil[ip]; exists && time.Now().Before(banTime) {
			sm.mu.RUnlock()
			logger.Logger.Warning(fmt.Sprintf("SSH: Blocked connection attempt from banned IP %s", ip))
			return false
		}
		sm.mu.RUnlock()
	}

	// 执行实际的公钥认证（避免循环导入，直接在这里实现）
	success := publicKeyAuth(ctx, key)

	// 记录认证结果
	if globalSSHSecurityManager != nil {
		if success {
			globalSSHSecurityManager.RecordAuthSuccess(ip)
		} else {
			shouldBan := globalSSHSecurityManager.RecordAuthFailure(ip)
			if shouldBan {
				logger.Logger.Warning(fmt.Sprintf("SSH: IP %s banned due to too many failed authentication attempts", ip))
			}
		}
	}

	return success
}

// SecureConnectionHandler 安全的连接处理器包装器
// 输入: handler - 原始连接处理器
// 输出: ssh.Handler - 包装后的连接处理器
// 必要性: 在连接处理中添加连接数限制和安全检查
func SecureConnectionHandler(handler func(ssh.Session)) func(ssh.Session) {
	return func(sess ssh.Session) {
		ip := GetClientIP(sess)
		if ip == "" {
			ip = sess.RemoteAddr().String()
		}

		// 检查是否允许连接
		if globalSSHSecurityManager != nil {
			allowed, reason := globalSSHSecurityManager.AllowConnection(ip)
			if !allowed {
				logger.Logger.Warning(fmt.Sprintf("SSH: Connection rejected from %s: %s", ip, reason))
				sess.Close()
				return
			}

			// 连接关闭时释放计数
			defer globalSSHSecurityManager.ReleaseConnection(ip)
		}

		// 执行原始处理器
		handler(sess)
	}
}
