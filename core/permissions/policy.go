package permissions

import (
	"binrc.com/roma/core/global"
	"binrc.com/roma/core/model"
	"binrc.com/roma/core/operation"
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
			// 检查1: 用户必须是空间成员
			isMember, err := opSpace.IsUserInSpace(user.ID, resourceSpace.SpaceID)
			if err != nil || !isMember {
				return false, "用户不是空间成员，无法访问空间资源"
			}

			// 检查2: 如果资源有角色要求，用户必须同时拥有匹配的角色
			if policy.EnableResourceRole {
				opResourceRole := operation.NewResourceRoleOperation()
				resourceRoles, err := opResourceRole.GetResourceRoles(resourceID, resourceType)
				if err == nil && len(resourceRoles) > 0 {
					hasMatchingRole := false
					for _, rr := range resourceRoles {
						// 只检查属于该空间的资源角色
						if rr.SpaceID != nil && *rr.SpaceID == resourceSpace.SpaceID {
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
						// 如果没有匹配的角色，继续检查空间角色权限
					}
				}
			}

			// 检查3: 用户在空间中的角色是否有权限
			member, err := opSpace.GetSpaceMember(user.ID, resourceSpace.SpaceID)
			if err == nil && member != nil && member.RoleID > 0 {
				// 检查空间角色权限
				opRole := operation.NewRoleOperation()
				spaceRole, err := opRole.GetRoleByID(uint64(member.RoleID))
				if err == nil && spaceRole != nil {
					desc, err := ParseRoleDescriptor(spaceRole.Desc)
					if err == nil && desc != nil {
						if !HasPermission(desc, "resource", action, "") {
							return false, "空间角色权限不足"
						}
					}
				}
			}
		} else if resourceSpace == nil || resourceSpace.SpaceID == 0 {
			// 资源没有空间归属，检查是否允许访问全局资源
			// 如果启用了空间隔离，没有空间归属的资源默认不允许访问（除非有全局权限）
			if policy.EnableSpaceIsolation {
				// 继续检查全局角色权限（第4步）
			}
		}
	}

	// 4. 检查用户全局角色权限（传统方式）
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
