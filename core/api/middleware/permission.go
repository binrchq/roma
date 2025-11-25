package middleware

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"binrc.com/roma/core/model"
	"binrc.com/roma/core/operation"
	"binrc.com/roma/core/utils"
	"github.com/gin-gonic/gin"
)

// PermissionRule 权限规则
type PermissionRule struct {
	Target     string   // "user" 或 "resource"
	Scope      string   // 范围过滤，如 "*peripheral", "*trial", "trial"
	ScopeType  string   // "exclude" 或 "include" 或 ""
	Operations []string // 允许的操作列表
}

// ParsePermissionRule 从角色描述中解析权限规则
// 格式: operation:target-scope.(op1|op2|op3)
// 示例: operation:user.(add|delete|update|get|list)
//
//	operation:resource-(*peripheral).(get|list)
//	operation:resource-(*trial).(get|list|use)
func ParsePermissionRule(desc string) []PermissionRule {
	var rules []PermissionRule

	// 提取 operation: 规则
	re := regexp.MustCompile(`operation:([^\.]+)\.\(([^\)]+)\)`)
	matches := re.FindAllStringSubmatch(desc, -1)

	for _, match := range matches {
		if len(match) != 3 {
			continue
		}

		targetScope := strings.TrimSpace(match[1])
		operationsStr := strings.TrimSpace(match[2])

		rule := PermissionRule{
			Operations: strings.Split(operationsStr, "|"),
		}

		// 解析 target 和 scope
		if strings.Contains(targetScope, "-") {
			parts := strings.SplitN(targetScope, "-", 2)
			rule.Target = strings.TrimSpace(parts[0])
			scopeStr := strings.TrimSpace(parts[1])

			// 处理范围过滤
			if strings.HasPrefix(scopeStr, "(*") && strings.HasSuffix(scopeStr, ")") {
				// 排除规则: (*peripheral)
				rule.ScopeType = "exclude"
				rule.Scope = strings.TrimPrefix(strings.TrimSuffix(scopeStr, ")"), "(*")
			} else if strings.HasPrefix(scopeStr, "(") && strings.HasSuffix(scopeStr, ")") {
				// 包含规则: (trial)
				rule.ScopeType = "include"
				rule.Scope = strings.TrimPrefix(strings.TrimSuffix(scopeStr, ")"), "(")
			} else {
				rule.Scope = scopeStr
			}
		} else {
			rule.Target = targetScope
		}

		// 清理操作列表
		for i, op := range rule.Operations {
			rule.Operations[i] = strings.TrimSpace(op)
		}

		rules = append(rules, rule)
	}

	return rules
}

// CheckPermission 检查用户是否有权限执行指定操作
func CheckPermission(user *model.User, target string, opName string, resourceScope string) bool {
	opUser := operation.NewUserOperation()
	roles, err := opUser.GetUserRoles(user.ID)
	if err != nil {
		return false
	}

	// super 角色拥有所有权限
	for _, role := range roles {
		if role.Name == "super" {
			return true
		}
	}

	// 检查每个角色的权限规则
	for _, role := range roles {
		rules := ParsePermissionRule(role.Desc)
		for _, rule := range rules {
			// 检查目标类型是否匹配
			if rule.Target != target {
				continue
			}

			// 检查操作是否允许
			hasOperation := false
			for _, op := range rule.Operations {
				if op == opName {
					hasOperation = true
					break
				}
			}
			if !hasOperation {
				continue
			}

			// 检查范围过滤
			if rule.Scope != "" {
				if rule.ScopeType == "exclude" {
					// 排除规则: 如果资源在排除范围内，则不允许
					if strings.Contains(strings.ToLower(resourceScope), strings.ToLower(rule.Scope)) {
						continue
					}
				} else if rule.ScopeType == "include" {
					// 包含规则: 只允许指定范围的资源
					if !strings.Contains(strings.ToLower(resourceScope), strings.ToLower(rule.Scope)) {
						continue
					}
				}
			}

			// 权限通过
			return true
		}
	}

	return false
}

// GetUserFromContext 从上下文中获取用户信息
// GetUserFromContext 从上下文获取用户，支持 JWT 和 API Key 两种认证方式
func GetUserFromContext(c *gin.Context) (*model.User, error) {
	// 优先从上下文获取用户（由 JWT 或 API Key 中间件设置）
	if user, exists := c.Get("user"); exists {
		if u, ok := user.(*model.User); ok {
			return u, nil
		}
	}

	// 尝试从 JWT token 获取用户
	token := c.GetHeader("Authorization")
	if token != "" {
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}
		claims, err := utils.ParseJWT(token)
		if err == nil {
			opUser := operation.NewUserOperation()
			user, err := opUser.GetUserByID(claims.UserID)
			if err == nil {
				return user, nil
			}
		}
	}

	// 从 API Key 获取用户
	apiKey := c.GetHeader("apikey")
	if apiKey == "" {
		apiKey = c.Query("apikey")
	}

	if apiKey != "" {
		opKey := operation.NewApikeyOperation()
		// 验证 API Key 是否有效
		valid, err := opKey.ApiKeyIsValid(apiKey)
		if err != nil || !valid {
			return nil, fmt.Errorf("invalid API key")
		}

		// API Key 默认拥有 super 权限
		// 查找 super 角色的用户作为代表
		opUser := operation.NewUserOperation()
		users, err := opUser.GetAllUsers()
		if err == nil && len(users) > 0 {
			// 优先查找 super 角色的用户
			for _, u := range users {
				roles, err := opUser.GetUserRoles(u.ID)
				if err == nil {
					for _, role := range roles {
						if role.Name == "super" {
							return u, nil
						}
					}
				}
			}
			// 如果没有 super 用户，返回第一个用户（临时方案）
			return users[0], nil
		}

		return nil, fmt.Errorf("no user found for API key")
	}

	return nil, fmt.Errorf("no user found in context")
}

// RequirePermission 权限检查中间件
func RequirePermission(target string, opName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		utilG := utils.Gin{C: c}

		// 获取用户
		user, err := GetUserFromContext(c)
		if err != nil {
			utilG.Response(http.StatusUnauthorized, utils.ERROR, "Authentication required")
			c.Abort()
			return
		}

		// 检查是否是访问自己的信息（/me 路径）
		// 用户总是可以访问和更新自己的信息
		path := c.Request.URL.Path
		if strings.HasSuffix(path, "/me") {
			// 访问自己的信息，直接允许
			c.Set("user", user)
			c.Next()
			return
		}

		// 检查是否是更新自己的信息（通过 ID 参数判断）
		if target == "user" && (opName == "get" || opName == "update") {
			userID := c.Param("id")
			if userID != "" {
				// 如果请求的是自己的 ID，允许访问
				if fmt.Sprintf("%d", user.ID) == userID {
					c.Set("user", user)
					c.Next()
					return
				}
			}
		}

		// 获取资源范围（如果有）
		resourceScope := c.GetString("resource_scope")
		if resourceScope == "" {
			// 尝试从请求参数中获取
			resourceScope = c.Query("scope")
		}

		// 检查权限
		if !CheckPermission(user, target, opName, resourceScope) {
			utilG.Response(http.StatusForbidden, utils.ERROR, fmt.Sprintf("Permission denied: %s.%s", target, opName))
			c.Abort()
			return
		}

		// 将用户信息存储到上下文
		c.Set("user", user)
		c.Next()
	}
}
