package permissions

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"binrc.com/roma/configs"
	"binrc.com/roma/core/model"
)

const descriptorVersion = "1.0"

// RoleDescriptor represents the structured permission definition stored inside role.Desc.
type RoleDescriptor struct {
	Version     string                 `json:"version"`
	Description string                 `json:"description,omitempty"`
	IsSuper     bool                   `json:"is_super,omitempty"`
	Permissions []PermissionDefinition `json:"permissions,omitempty"`
}

type PermissionDefinition struct {
	Target  string           `json:"target"`
	Actions []string         `json:"actions"`
	Scope   *ScopeDefinition `json:"scope,omitempty"`
}

type ScopeDefinition struct {
	Type  string `json:"type,omitempty"`
	Value string `json:"value,omitempty"`
}

// BuildRoleDescriptor converts a RoleConfig into a JSON descriptor string.
func BuildRoleDescriptor(cfg *configs.RoleConfig) (string, error) {
	if cfg == nil {
		return "", errors.New("role config is nil")
	}

	if len(cfg.Permissions) == 0 && !cfg.IsDefaultSuper {
		// fall back to legacy desc if provided
		return strings.TrimSpace(cfg.Desc), nil
	}

	desc := RoleDescriptor{
		Version:     descriptorVersion,
		Description: fallbackDescription(cfg),
		IsSuper:     cfg.IsDefaultSuper,
	}

	for _, permCfg := range cfg.Permissions {
		if permCfg == nil {
			continue
		}
		target := strings.TrimSpace(strings.ToLower(permCfg.Target))
		if target == "" {
			return "", fmt.Errorf("role %s permission target is empty", cfg.Name)
		}
		if len(permCfg.Actions) == 0 && !cfg.IsDefaultSuper {
			return "", fmt.Errorf("role %s permission actions missing", cfg.Name)
		}
		def := PermissionDefinition{
			Target:  target,
			Actions: normalizeActions(permCfg.Actions),
		}
		if permCfg.Scope != nil && strings.TrimSpace(permCfg.Scope.Value) != "" {
			def.Scope = &ScopeDefinition{
				Type:  strings.ToLower(strings.TrimSpace(permCfg.Scope.Type)),
				Value: strings.TrimSpace(permCfg.Scope.Value),
			}
		}
		desc.Permissions = append(desc.Permissions, def)
	}

	if len(desc.Permissions) == 0 && !desc.IsSuper {
		return "", fmt.Errorf("role %s has no valid permissions", cfg.Name)
	}

	payload, err := json.Marshal(desc)
	if err != nil {
		return "", err
	}
	return string(payload), nil
}

func fallbackDescription(cfg *configs.RoleConfig) string {
	if cfg == nil {
		return ""
	}
	if strings.TrimSpace(cfg.Description) != "" {
		return cfg.Description
	}
	return strings.TrimSpace(cfg.Desc)
}

func normalizeActions(actions []string) []string {
	if len(actions) == 0 {
		return []string{"*"}
	}
	seen := make(map[string]struct{})
	var result []string
	for _, a := range actions {
		key := strings.TrimSpace(strings.ToLower(a))
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, key)
	}
	if len(result) == 0 {
		return []string{"*"}
	}
	return result
}

// ParseRoleDescriptor attempts to parse a descriptor JSON string.
func ParseRoleDescriptor(raw string) (*RoleDescriptor, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, errors.New("empty descriptor")
	}
	var desc RoleDescriptor
	if err := json.Unmarshal([]byte(raw), &desc); err != nil {
		return nil, err
	}
	if desc.Version == "" {
		desc.Version = descriptorVersion
	}
	return &desc, nil
}

// HasPermission determines whether the descriptor grants the given action on target.
func HasPermission(desc *RoleDescriptor, target, action, resourceScope string) bool {
	if desc == nil {
		return false
	}
	if desc.IsSuper {
		return true
	}
	target = strings.ToLower(strings.TrimSpace(target))
	action = strings.ToLower(strings.TrimSpace(action))
	scope := strings.ToLower(strings.TrimSpace(resourceScope))

	for _, perm := range desc.Permissions {
		if perm.Target != "*" && perm.Target != target {
			continue
		}
		if !hasActionMatch(perm.Actions, action) {
			continue
		}
		if perm.Scope != nil && perm.Scope.Value != "" {
			switch perm.Scope.Type {
			case "exclude":
				if scope == "" {
					// no scope specified, treat as allowed
				} else if strings.Contains(scope, strings.ToLower(perm.Scope.Value)) {
					continue
				}
			case "include":
				if scope == "" || !strings.Contains(scope, strings.ToLower(perm.Scope.Value)) {
					continue
				}
			}
		}
		return true
	}
	return false
}

func hasActionMatch(actions []string, action string) bool {
	if len(actions) == 0 {
		return false
	}
	for _, a := range actions {
		if a == "*" || a == action {
			return true
		}
	}
	return false
}

// IsSuperRole 检查角色是否是 super 角色（通过权限描述符判断，不硬编码角色名称）
func IsSuperRole(role *model.Role) bool {
	if role == nil {
		return false
	}
	desc, err := ParseRoleDescriptor(role.Desc)
	if err == nil && desc != nil {
		return desc.IsSuper
	}
	return false
}

// HasAllPermissions 检查角色是否拥有所有权限（通过权限描述符判断）
func HasAllPermissions(role *model.Role) bool {
	if role == nil {
		return false
	}
	desc, err := ParseRoleDescriptor(role.Desc)
	if err == nil && desc != nil {
		if desc.IsSuper {
			return true
		}
		// 检查是否有 target="*" 且 actions=["*"] 的权限
		for _, perm := range desc.Permissions {
			if perm.Target == "*" {
				for _, action := range perm.Actions {
					if action == "*" {
						return true
					}
				}
			}
		}
	}
	return false
}
