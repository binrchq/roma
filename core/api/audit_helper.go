package api

import (
	"fmt"
	"strings"

	"binrc.com/roma/core/model"
	"binrc.com/roma/core/operation"
	"github.com/gin-gonic/gin"
)

// RecordAuditLog 记录审计日志的辅助函数
func RecordAuditLog(c *gin.Context, action, actionType, resourceType string, resourceID uint, resourceName, description, status, errorMessage string) {
	// 获取当前用户信息
	user, exists := c.Get("user")
	if !exists {
		return
	}

	currentUser := user.(*model.User)

	// 获取客户端IP
	ipAddress := c.ClientIP()
	if ipAddress == "" {
		ipAddress = c.GetHeader("X-Forwarded-For")
	}
	if ipAddress == "" {
		ipAddress = c.GetHeader("X-Real-IP")
	}

	// 创建审计日志
	auditLog := &model.AuditLog{
		UserID:       currentUser.ID,
		Username:     currentUser.Username,
		Action:       action,
		ActionType:   actionType,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		ResourceName: resourceName,
		Description:  description,
		IPAddress:    ipAddress,
		Status:       status,
		ErrorMessage: errorMessage,
	}

	// 异步记录审计日志（不阻塞主流程）
	go func() {
		opAudit := operation.NewAuditOperation()
		if err := opAudit.CreateAuditLog(auditLog); err != nil {
			// 记录失败不影响主流程，只记录错误
			// log.Printf("Failed to create audit log: %v", err)
		}
	}()
}

// IsHighRiskCommand 检测是否为高危命令
func IsHighRiskCommand(command string) bool {
	command = strings.ToLower(strings.TrimSpace(command))

	// 高危命令关键词列表
	highRiskKeywords := []string{
		"rm -rf",
		"rm -r",
		"rm -f",
		"dd if=",
		"mkfs",
		"fdisk",
		"chmod 777",
		"chmod +x",
		"chown",
		"systemctl stop",
		"systemctl disable",
		"kill -9",
		"> /dev/null",
		"| sh",
		"| bash",
		"curl |",
		"wget |",
		"format",
		"del /f /s /q",
		"format c:",
		"shutdown",
		"reboot",
		"halt",
		"poweroff",
		"init 0",
		"init 6",
		"iptables -f",
		"iptables -x",
		"drop database",
		"truncate",
		"drop table",
		"delete from",
		"update.*set.*=",
		"alter table",
		"grant all",
		"revoke",
	}

	for _, keyword := range highRiskKeywords {
		if strings.Contains(command, keyword) {
			return true
		}
	}

	return false
}

// RecordCommandAuditLog 记录命令执行审计日志
func RecordCommandAuditLog(c *gin.Context, command, resourceType string, resourceID uint, resourceName string, status, errorMessage string) {
	// 获取当前用户信息
	user, exists := c.Get("user")
	if !exists {
		return
	}

	currentUser := user.(*model.User)

	// 获取客户端IP
	ipAddress := c.ClientIP()
	if ipAddress == "" {
		ipAddress = c.GetHeader("X-Forwarded-For")
	}
	if ipAddress == "" {
		ipAddress = c.GetHeader("X-Real-IP")
	}

	// 创建审计日志
	auditLog := &model.AuditLog{
		UserID:       currentUser.ID,
		Username:     currentUser.Username,
		Action:       "execute_command",
		ActionType:   "high_risk",
		ResourceType: resourceType,
		ResourceID:   resourceID,
		ResourceName: resourceName,
		Description:  fmt.Sprintf("执行命令: %s", command),
		IPAddress:    ipAddress,
		Status:       status,
		ErrorMessage: errorMessage,
	}

	// 异步记录审计日志
	go func() {
		opAudit := operation.NewAuditOperation()
		if err := opAudit.CreateAuditLog(auditLog); err != nil {
			// 记录失败不影响主流程
		}
	}()
}

// RecordTUICommandAuditLog 记录TUI中命令执行的审计日志（不依赖gin.Context）
func RecordTUICommandAuditLog(username, command, resourceType string, resourceID uint, resourceName, ipAddress, status, errorMessage string) {
	// 获取用户ID
	opUser := operation.NewUserOperation()
	user, err := opUser.GetUserByUsername(username)
	if err != nil {
		// 如果获取用户失败，仍然记录日志，但使用0作为用户ID
		recordTUIAuditLog(0, username, command, resourceType, resourceID, resourceName, ipAddress, status, errorMessage)
		return
	}

	recordTUIAuditLog(user.ID, username, command, resourceType, resourceID, resourceName, ipAddress, status, errorMessage)
}

// recordTUIAuditLog 内部函数，实际记录审计日志
func recordTUIAuditLog(userID uint, username, command, resourceType string, resourceID uint, resourceName, ipAddress, status, errorMessage string) {
	// 创建审计日志
	auditLog := &model.AuditLog{
		UserID:       userID,
		Username:     username,
		Action:       "execute_command",
		ActionType:   "high_risk",
		ResourceType: resourceType,
		ResourceID:   resourceID,
		ResourceName: resourceName,
		Description:  fmt.Sprintf("执行命令: %s", command),
		IPAddress:    ipAddress,
		Status:       status,
		ErrorMessage: errorMessage,
	}

	// 异步记录审计日志
	go func() {
		opAudit := operation.NewAuditOperation()
		if err := opAudit.CreateAuditLog(auditLog); err != nil {
			// 记录失败不影响主流程
		}
	}()
}
