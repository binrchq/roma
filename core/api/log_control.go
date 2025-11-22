package api

import (
	"net/http"
	"strconv"

	"binrc.com/roma/core/operation"
	"binrc.com/roma/core/utils"
	"github.com/gin-gonic/gin"
)

type LogController struct{}

func NewLogController() *LogController {
	return &LogController{}
}

// GetAccessLogs 获取访问日志
func (l *LogController) GetAccessLogs(c *gin.Context) {
	utilG := utils.Gin{C: c}

	username := c.Query("username")
	resourceType := c.Query("resource_type")
	limit := 50

	if l := c.Query("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil {
			limit = val
		}
	}

	opAccess := operation.NewAccessOperation()
	logs, err := opAccess.GetAccessLogs(username, resourceType, limit)
	if err != nil {
		utilG.Response(http.StatusInternalServerError, utils.ERROR, "获取访问日志失败: "+err.Error())
		return
	}

	// 为每条日志添加用户名
	logsWithUsername := make([]map[string]interface{}, len(logs))
	opUser := operation.NewUserOperation()
	for i, log := range logs {
		logMap := map[string]interface{}{
			"id":            log.ID,
			"user_id":       log.UserID,
			"resource_type": log.ResourceType,
			"resource_id":   log.ResourceID,
			"action":        log.Action,
			"action_level":  log.ActionLevel,
			"source":        log.Source,
			"ip_pub":        log.IPPub,
			"ip_priv":       log.IPPriv,
			"status":        log.Status,
			"timestamp":     log.Timestamp,
			"ip_address":    log.IPPub,
			"ip":            log.IPPub,
		}
		// 获取用户名
		if user, err := opUser.GetUserByID(log.UserID); err == nil {
			logMap["username"] = user.Username
		} else {
			logMap["username"] = ""
		}
		logsWithUsername[i] = logMap
	}

	utilG.Response(http.StatusOK, utils.SUCCESS, map[string]interface{}{
		"count":    len(logsWithUsername),
		"username": username,
		"type":     resourceType,
		"logs":     logsWithUsername,
	})
}

// GetCredentialLogs 获取凭证日志
func (l *LogController) GetCredentialLogs(c *gin.Context) {
	utilG := utils.Gin{C: c}

	username := c.Query("username")
	limit := 50

	if l := c.Query("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil {
			limit = val
		}
	}

	opAccess := operation.NewAccessOperation()
	logs, err := opAccess.GetCredentialLogs(username, limit)
	if err != nil {
		utilG.Response(http.StatusInternalServerError, utils.ERROR, "获取凭证日志失败: "+err.Error())
		return
	}

	// 为每条日志添加用户名
	logsWithUsername := make([]map[string]interface{}, len(logs))
	opUser := operation.NewUserOperation()
	for i, log := range logs {
		logMap := map[string]interface{}{
			"id":            log.ID,
			"credential_id": log.CredentialID,
			"user_id":       log.UserID,
			"action":        log.Action,
			"operation":     log.Action,
			"ip":            log.IP,
			"ip_address":    log.IP,
			"status":        log.Status,
			"timestamp":     log.Timestamp,
		}
		// 获取用户名
		if user, err := opUser.GetUserByID(log.UserID); err == nil {
			logMap["username"] = user.Username
		} else {
			logMap["username"] = ""
		}
		logsWithUsername[i] = logMap
	}

	utilG.Response(http.StatusOK, utils.SUCCESS, map[string]interface{}{
		"count":    len(logsWithUsername),
		"username": username,
		"logs":     logsWithUsername,
	})
}

// GetAuditLogs 获取审计日志
func (l *LogController) GetAuditLogs(c *gin.Context) {
	utilG := utils.Gin{C: c}

	page := 1
	pageSize := 50

	if p := c.Query("page"); p != "" {
		if val, err := strconv.Atoi(p); err == nil && val > 0 {
			page = val
		}
	}

	if ps := c.Query("page_size"); ps != "" {
		if val, err := strconv.Atoi(ps); err == nil && val > 0 {
			pageSize = val
		}
	}

	filters := make(map[string]interface{})
	if username := c.Query("username"); username != "" {
		filters["username"] = username
	}
	if action := c.Query("action"); action != "" {
		filters["action"] = action
	}
	if actionType := c.Query("action_type"); actionType != "" {
		filters["action_type"] = actionType
	}
	if resourceType := c.Query("resource_type"); resourceType != "" {
		filters["resource_type"] = resourceType
	}
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}

	opAudit := operation.NewAuditOperation()
	logs, total, err := opAudit.GetAuditLogs(page, pageSize, filters)
	if err != nil {
		utilG.Response(http.StatusInternalServerError, utils.ERROR, "获取审计日志失败: "+err.Error())
		return
	}

	utilG.Response(http.StatusOK, utils.SUCCESS, map[string]interface{}{
		"logs":      logs,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}
