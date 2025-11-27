package permissions

import (
	"binrc.com/roma/core/global"
	"binrc.com/roma/core/model"
	"binrc.com/roma/core/operation"
	"github.com/rs/zerolog/log"
)

// CheckResourceAccess 检查用户是否有权限访问资源（多维度权限检查）
// 如果 userRoles 为 nil，会自动获取用户角色
func CheckResourceAccess(user *model.User, resourceID int64, resourceType, action string) (bool, string) {
	return CheckResourceAccessWithRoles(user, nil, resourceID, resourceType, action)
}

// CheckResourceAccessWithRoles 检查用户是否有权限访问资源（多维度权限检查）
// 允许传入已获取的用户角色，避免重复查询
func CheckResourceAccessWithRoles(user *model.User, userRoles []*model.Role, resourceID int64, resourceType, action string) (bool, string) {
	var err error
	if userRoles == nil {
		opUser := operation.NewUserOperation()
		userRoles, err = opUser.GetUserRoles(user.ID)
		if err != nil {
			return false, "无法获取用户角色"
		}
	}

	// 1. 检查是否是 super 角色（如果配置允许绕过）
	policy := global.CONFIG.PermissionPolicy
	if policy != nil && policy.SuperBypassAll {
		for _, role := range userRoles {
			if IsSuperRole(role) {
				return true, ""
			}
		}
	}

	// 2. 检查资源角色（如果启用）
	if policy != nil && policy.EnableResourceRole {
		opResourceRole := operation.NewResourceRoleOperation()
		resourceRoles, err := opResourceRole.GetResourceRoles(resourceID, resourceType)
		if err == nil && len(resourceRoles) > 0 {
			hasMatchingRole := false
			for _, rr := range resourceRoles {
				// 检查用户是否有匹配的角色
				for _, ur := range userRoles {
					if ur != nil && ur.ID == rr.RoleID {
						hasMatchingRole = true
						break
					}
				}
				if hasMatchingRole {
					break
				}
			}

			if !hasMatchingRole {
				// 如果要求完全匹配且没有匹配的角色，拒绝访问
				if policy.RequireExactRoleMatch {
					return false, "用户角色与资源要求的角色不匹配"
				}
				// 否则继续检查其他维度
			}
		}
	}

	// 3. 检查空间隔离（如果启用）- 必须同时满足：空间成员 AND 资源角色
	if policy != nil && policy.EnableSpaceIsolation {
		opSpace := operation.NewSpaceOperation()

		// 获取资源所属的空间
		resourceSpace, err := opSpace.GetResourceSpace(resourceID, resourceType)
		if err == nil && resourceSpace != nil && resourceSpace.SpaceID > 0 {
			// 检查1: 用户必须是空间成员（这是强制要求，如果不满足直接拒绝）
			isMember, err := opSpace.IsUserInSpace(user.ID, resourceSpace.SpaceID)
			if err != nil || !isMember {
				log.Debug().
					Uint("user_id", user.ID).
					Uint("space_id", resourceSpace.SpaceID).
					Int64("resource_id", resourceID).
					Str("resource_type", resourceType).
					Str("action", action).
					Msg("权限检查失败: 用户不是空间成员")
				return false, "用户不是空间成员，无法访问空间资源"
			}

			// 检查2: 如果资源有角色要求，用户必须同时拥有匹配的角色
			if policy.EnableResourceRole {
				opResourceRole := operation.NewResourceRoleOperation()
				resourceRoles, err := opResourceRole.GetResourceRoles(resourceID, resourceType)
				if err == nil && len(resourceRoles) > 0 {
					hasMatchingRole := false
					for _, rr := range resourceRoles {
						// 检查资源角色：如果资源角色有 SpaceID，必须与资源空间匹配；如果为 nil，则视为全局角色
						roleSpaceMatch := true
						if rr.SpaceID != nil {
							// 资源角色有空间限制，必须与资源空间匹配
							roleSpaceMatch = (*rr.SpaceID == resourceSpace.SpaceID)
						}
						// 如果空间匹配，检查用户是否有该角色
						if roleSpaceMatch {
							for _, ur := range userRoles {
								if ur != nil && ur.ID == rr.RoleID {
									hasMatchingRole = true
									break
								}
							}
							if hasMatchingRole {
								break
							}
						}
					}

					if !hasMatchingRole {
						if policy.RequireExactRoleMatch {
							return false, "用户角色与资源要求的角色不匹配"
						}
						// 如果没有匹配的角色且不要求完全匹配，继续检查用户全局角色权限（第4步）
					} else {
						// 如果用户有匹配的角色且是空间成员，允许访问
						return true, ""
					}
				}
			}

			// 检查3: 如果资源没有角色要求，但用户在空间中，检查用户全局角色权限
			// 继续到第4步检查全局角色权限
		} else if resourceSpace == nil || resourceSpace.SpaceID == 0 {
			// 资源没有空间归属
			// 如果启用了空间隔离，没有空间归属的资源应该被视为在 default 空间中
			if policy.EnableSpaceIsolation {
				// 如果配置了默认空间，尝试获取 default 空间
				if policy.DefaultSpace != nil && *policy.DefaultSpace != "" {
					defaultSpace, err := opSpace.GetSpaceByName(*policy.DefaultSpace)
					if err == nil && defaultSpace != nil {
						// 检查用户是否是 default 空间的成员
						isMember, err := opSpace.IsUserInSpace(user.ID, defaultSpace.ID)
						if err != nil || !isMember {
							// 用户不是 default 空间成员，拒绝访问
							log.Debug().
								Uint("user_id", user.ID).
								Uint("default_space_id", defaultSpace.ID).
								Int64("resource_id", resourceID).
								Str("resource_type", resourceType).
								Str("action", action).
								Msg("权限检查失败: 资源没有空间归属，且用户不是默认空间成员")
							return false, "资源没有空间归属，且用户不是默认空间成员"
						}
						// 用户是 default 空间成员，继续检查角色权限（第4步）
						log.Debug().
							Uint("user_id", user.ID).
							Uint("default_space_id", defaultSpace.ID).
							Int64("resource_id", resourceID).
							Str("resource_type", resourceType).
							Msg("用户是默认空间成员，继续检查角色权限")
					} else {
						// 如果 default 空间不存在，且启用了空间隔离，拒绝访问
						log.Debug().
							Uint("user_id", user.ID).
							Int64("resource_id", resourceID).
							Str("resource_type", resourceType).
							Str("action", action).
							Str("default_space", *policy.DefaultSpace).
							Msg("权限检查失败: 资源没有空间归属，且默认空间不存在")
						return false, "资源没有空间归属，且默认空间不存在"
					}
				} else {
					// 如果未配置默认空间，且启用了空间隔离，拒绝访问
					log.Debug().
						Uint("user_id", user.ID).
						Int64("resource_id", resourceID).
						Str("resource_type", resourceType).
						Str("action", action).
						Msg("权限检查失败: 资源没有空间归属，且未配置默认空间")
					return false, "资源没有空间归属，且未配置默认空间"
				}
			}
			// 如果没有启用空间隔离，继续检查全局角色权限（第4步）
		}
	}

	// 4. 检查用户全局角色权限（传统方式）
	// 注意：只有在以下情况下才会到达这里：
	// 1. 资源有空间归属，用户是空间成员，且（资源没有角色要求 OR 用户有匹配的角色 OR 不要求完全匹配）
	// 2. 资源没有空间归属，但（未启用空间隔离 OR 用户在 default 空间中）
	for _, role := range userRoles {
		if role == nil {
			continue
		}
		desc, err := ParseRoleDescriptor(role.Desc)
		if err == nil && desc != nil {
			if HasPermission(desc, "resource", action, "") {
				return true, ""
			}
		}
	}

	return false, "权限不足"
}
