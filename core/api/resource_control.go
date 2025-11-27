package api

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	"binrc.com/roma/core/constants"
	"binrc.com/roma/core/global"
	"binrc.com/roma/core/model"
	"binrc.com/roma/core/operation"
	"binrc.com/roma/core/permissions"
	"binrc.com/roma/core/utils"
	"github.com/gin-gonic/gin"
)

type ResourceControl struct{}

func NewResourceControl() *ResourceControl {
	return &ResourceControl{}
}

func (r *ResourceControl) AddResource(c *gin.Context) {
	utilG := utils.Gin{C: c}
	var resourceData struct {
		Type    string            `json:"type"`
		Data    []json.RawMessage `json:"data"`     // 使用 json.RawMessage 保存未解码的 JSON 字符串
		Role    string            `json:"role"`     // 可选的角色名称
		SpaceID *uint             `json:"space_id"` // 可选的空间ID
	}
	if err := c.ShouldBindJSON(&resourceData); err != nil {
		utilG.Response(utils.ERROR, utils.ERROR, err.Error())
		return
	}
	// 检查 role 参数是否为空，如果为空，则设置默认值为 "ops"
	roleName := "ops"
	if resourceData.Role != "" {
		roleName = resourceData.Role
	}
	// 开启事务
	tx := global.GetDB().Begin()
	if tx.Error != nil {
		utilG.Response(utils.ERROR, utils.ERROR, "服务器错误,数据库异常Q4A")
		return
	}
	var failedCount int // 记录失败的条目数
	var failedMsgs []string

	opRes := operation.NewResourceOperation()
	opRole := operation.NewRoleOperation()
	opSpace := operation.NewSpaceOperation()

	// 确定要使用的空间ID
	var targetSpaceID uint
	if resourceData.SpaceID != nil && *resourceData.SpaceID > 0 {
		targetSpaceID = *resourceData.SpaceID
	} else {
		// 如果没有指定空间，使用 default 空间
		defaultSpace, err := opSpace.GetSpaceByName("default")
		if err != nil {
			utilG.Response(utils.ERROR, utils.ERROR, "无法找到默认空间，请先创建 default 空间")
			tx.Rollback()
			return
		}
		targetSpaceID = defaultSpace.ID
	}

	for id, r := range resourceData.Data {
		var resModel model.Resource
		// 将 r 转换为相应的资源类型并创建资源
		switch resourceData.Type {
		case constants.ResourceTypeLinux:
			resModel = new(model.LinuxConfig)
		case constants.ResourceTypeRouter:
			resModel = new(model.RouterConfig)
		case constants.ResourceTypeWindows:
			resModel = new(model.WindowsConfig)
		case constants.ResourceTypeDocker:
			resModel = new(model.DockerConfig)
		case constants.ResourceTypeDatabase:
			resModel = new(model.DatabaseConfig)
		case constants.ResourceTypeSwitch:
			resModel = new(model.SwitchConfig)
		default:
			utilG.Response(utils.ERROR, utils.ERROR, "未知的资源类型")
			return
		}
		// 解码 JSON 数据到资源模型
		if err := json.Unmarshal(r, resModel); err != nil {
			errMsg := fmt.Sprintf("JSON解析失败s2:原因.%s 数据No.%d", err.Error(), id)
			failedMsgs = append(failedMsgs, errMsg)
			log.Println(errMsg) // 记录错误到日志
			failedCount++
			continue // 继续处理下一个数据
		}
		// 创建资源
		resModel, err := opRes.CreateResource(resModel, resourceData.Type)
		if err != nil {
			errMsg := fmt.Sprintf("写入数据库失败:原因.%s 数据No.%d", err.Error(), id)
			failedMsgs = append(failedMsgs, errMsg)
			log.Println(errMsg) // 记录错误到日志
			failedCount++
			tx.Rollback() // 回滚事务
			continue
		}
		// 绑定资源角色
		// role, err := opRole.GetRoleByName(resourceData.Role)
		role, err := opRole.GetRoleByName(roleName)
		if err != nil {
			errMsg := fmt.Sprintf("资源赋值失败1:原因.%s 数据No.%d", err.Error(), id)
			failedMsgs = append(failedMsgs, errMsg)
			log.Println(errMsg) // 记录错误到日志
			failedCount++
			tx.Rollback() // 回滚事务
			continue
		}
		err = opRes.CreateResourceAndAssociate(int64(role.ID), resModel.GetID(), resourceData.Type)
		if err != nil {
			errMsg := fmt.Sprintf("资源赋值失败2:原因.%s 数据No.%d", err.Error(), id)
			failedMsgs = append(failedMsgs, errMsg)
			log.Println(errMsg) // 记录错误到日志
			failedCount++
			tx.Rollback() // 回滚事务
			continue
		}

		// 将资源分配到空间
		err = opSpace.AssignResourceToSpace(targetSpaceID, resModel.GetID(), resourceData.Type)
		if err != nil {
			errMsg := fmt.Sprintf("资源空间分配失败:原因.%s 数据No.%d", err.Error(), id)
			failedMsgs = append(failedMsgs, errMsg)
			log.Println(errMsg) // 记录错误到日志
			failedCount++
			tx.Rollback() // 回滚事务
			continue
		}
	}

	if failedCount > 0 {
		utilG.Response(utils.ERROR, utils.ERROR, fmt.Sprintf("%d 个资源创建失败(%s)", failedCount, strings.Join(failedMsgs, ";")))
		tx.Rollback() // 回滚事务
		return
	}
	// 提交事务
	tx.Commit()

	utilG.Response(utils.SUCCESS, utils.SUCCESS, "资源创建成功")
}

func (r *ResourceControl) UpdateResource(c *gin.Context) {
	utilG := utils.Gin{C: c}
	var resourceData struct {
		Type    string            `json:"type"`
		Data    []json.RawMessage `json:"data"`     // 使用 json.RawMessage 保存未解码的 JSON 字符串
		Role    string            `json:"role"`     // 可选的角色名称
		SpaceID *uint             `json:"space_id"` // 可选的空间ID
	}
	if err := c.ShouldBindJSON(&resourceData); err != nil {
		utilG.Response(utils.ERROR, utils.ERROR, err.Error())
		return
	}
	// 开启事务
	tx := global.GetDB().Begin()
	if tx.Error != nil {
		utilG.Response(utils.ERROR, utils.ERROR, "服务器错误,数据库异常Q4A")
		return
	}
	var failedCount int // 记录失败的条目数
	var failedMsgs []string

	opRes := operation.NewResourceOperation()
	opRole := operation.NewRoleOperation()
	opSpace := operation.NewSpaceOperation()

	for id, r := range resourceData.Data {
		var resModel model.Resource
		// 将 r 转换为相应的资源类型并创建资源
		switch resourceData.Type {
		case constants.ResourceTypeLinux:
			resModel = new(model.LinuxConfig)
		case constants.ResourceTypeRouter:
			resModel = new(model.RouterConfig)
		case constants.ResourceTypeWindows:
			resModel = new(model.WindowsConfig)
		case constants.ResourceTypeDocker:
			resModel = new(model.DockerConfig)
		case constants.ResourceTypeDatabase:
			resModel = new(model.DatabaseConfig)
		case constants.ResourceTypeSwitch:
			resModel = new(model.SwitchConfig)
		default:
			utilG.Response(utils.ERROR, utils.ERROR, "未知的资源类型")
			return
		}

		// 解码 JSON 数据到资源模型
		if err := json.Unmarshal(r, resModel); err != nil {
			errMsg := fmt.Sprintf("JSON解析失败:原因.%s 数据No.%d", err.Error(), id)
			failedMsgs = append(failedMsgs, errMsg)
			log.Println(errMsg) // 记录错误到日志
			failedCount++
			continue // 继续处理下一个数据
		}

		// 更新资源
		resModel, err := opRes.UpdateResource(resModel, resourceData.Type)
		if err != nil {
			errMsg := fmt.Sprintf("更新数据库失败:原因.%s 数据No.%d", err.Error(), id)
			failedMsgs = append(failedMsgs, errMsg)
			log.Println(errMsg) // 记录错误到日志
			failedCount++
			tx.Rollback() // 回滚事务
			continue
		}

		// 如果提供了角色信息，则更新资源角色关联（可选）
		if resourceData.Role != "" {
			role, err := opRole.GetRoleByName(resourceData.Role)
			if err != nil {
				errMsg := fmt.Sprintf("获取角色失败:原因.%s 数据No.%d", err.Error(), id)
				failedMsgs = append(failedMsgs, errMsg)
				log.Println(errMsg) // 记录错误到日志
				failedCount++
				tx.Rollback() // 回滚事务
				continue
			}

			err = opRes.CreateResourceAndAssociate(int64(role.ID), resModel.GetID(), resourceData.Type)
			if err != nil {
				errMsg := fmt.Sprintf("资源角色关联失败:原因.%s 数据No.%d", err.Error(), id)
				failedMsgs = append(failedMsgs, errMsg)
				log.Println(errMsg) // 记录错误到日志
				failedCount++
				tx.Rollback() // 回滚事务
				continue
			}
		}
		// 如果提供了空间ID，更新资源的空间分配
		if resourceData.SpaceID != nil && *resourceData.SpaceID > 0 {
			// 先删除旧的关联
			global.GetDB().Where("resource_id = ? AND resource_type = ?", resModel.GetID(), resourceData.Type).
				Delete(&model.ResourceSpace{})
			// 创建新的关联
			err = opSpace.AssignResourceToSpace(*resourceData.SpaceID, resModel.GetID(), resourceData.Type)
			if err != nil {
				errMsg := fmt.Sprintf("资源空间分配失败:原因.%s 数据No.%d", err.Error(), id)
				failedMsgs = append(failedMsgs, errMsg)
				log.Println(errMsg) // 记录错误到日志
				failedCount++
				tx.Rollback() // 回滚事务
				continue
			}
		}
		// 如果没有提供角色信息，则保持现有关联不变
	}

	if failedCount > 0 {
		utilG.Response(utils.ERROR, utils.ERROR, fmt.Sprintf("%d 个资源更新失败(%s)", failedCount, strings.Join(failedMsgs, ";")))
		tx.Rollback() // 回滚事务
		return
	}

	// 提交事务
	tx.Commit()
	utilG.Response(utils.SUCCESS, utils.SUCCESS, "资源更新成功")
}

func (r *ResourceControl) DeleteResource(c *gin.Context) {
	utilG := utils.Gin{C: c}
	var resourceData struct {
		Type string `json:"type"`
		Data []struct {
			ID int64 `json:"id"` // Assuming ID is of type int64
		} `json:"data"`
	}
	if err := c.ShouldBindJSON(&resourceData); err != nil {
		utilG.Response(utils.ERROR, utils.ERROR, err.Error())
		return
	}
	// 检查 role 参数是否为空，如果为空，则设置默认值为 "ops"
	roleName := "ops"
	// 开启事务
	tx := global.GetDB().Begin()
	if tx.Error != nil {
		utilG.Response(utils.ERROR, utils.ERROR, "服务器错误,数据库异常Q4A")
		return
	}
	var failedCount int // 记录失败的条目数
	var failedMsgs []string

	opRes := operation.NewResourceOperation()
	opRole := operation.NewRoleOperation()
	for id, r := range resourceData.Data {
		// 删除资源
		err := opRes.DeleteResource(strconv.Itoa(int(r.ID)), resourceData.Type)
		if err != nil {
			errMsg := fmt.Sprintf("删除数据库失败:原因.%s 数据No.%d", err.Error(), id)
			failedMsgs = append(failedMsgs, errMsg)
			log.Println(errMsg) // 记录错误到日志
			// 记录审计日志（失败）
			RecordAuditLog(c, "delete_resource", "high_risk", resourceData.Type, uint(r.ID), "",
				fmt.Sprintf("删除资源失败: %s", err.Error()), "failed", err.Error())
			failedCount++
			tx.Rollback() // 回滚事务
			continue
		}

		// 记录审计日志（成功）
		RecordAuditLog(c, "delete_resource", "high_risk", resourceData.Type, uint(r.ID), "",
			fmt.Sprintf("删除资源: 类型=%s, ID=%d", resourceData.Type, r.ID), "success", "")

		// 如果需要，可以根据业务需求，解除资源与角色之间的关联
		// 示例：opRes.DeleteResourceAndRoleAssociation(r.ID, resourceData.Type)

		// 绑定资源角色（这部分根据具体逻辑来处理，如果删除时需要额外操作）
		_, err = opRole.GetRoleByName(roleName)
		if err != nil {
			errMsg := fmt.Sprintf("获取角色失败:原因.%s 数据No.%d", err.Error(), id)
			failedMsgs = append(failedMsgs, errMsg)
			log.Println(errMsg) // 记录错误到日志
			failedCount++
			tx.Rollback() // 回滚事务
			continue
		}

		// 如果需要，可以根据业务需求，进行资源与角色的关联操作
		// 示例：opRes.CreateResourceAndAssociate(int64(role.ID), r.ID, resourceData.Type)
	}

	if failedCount > 0 {
		utilG.Response(utils.ERROR, utils.ERROR, fmt.Sprintf("%d 个资源删除失败(%s)", failedCount, strings.Join(failedMsgs, ";")))
		tx.Rollback() // 回滚事务
		return
	}

	// 提交事务
	tx.Commit()
	utilG.Response(utils.SUCCESS, utils.SUCCESS, "资源删除成功")
}

func (r *ResourceControl) GetAllResource(c *gin.Context) {
	utilG := utils.Gin{C: c}

	// 从查询参数或请求体获取资源类型
	resourceType := c.Query("type")
	if resourceType == "" {
		var resourceData struct {
			Type string `json:"type"`
		}
		if err := c.ShouldBindJSON(&resourceData); err == nil {
			resourceType = resourceData.Type
		}
	}

	// 获取当前用户（从上下文）
	userInterface, exists := c.Get("user")
	if !exists {
		utilG.Response(utils.ERROR, utils.ERROR, "User not found in context")
		return
	}
	user, ok := userInterface.(*model.User)
	if !ok {
		utilG.Response(utils.ERROR, utils.ERROR, "Invalid user type")
		return
	}

	opRes := operation.NewResourceOperation()
	opUser := operation.NewUserOperation()

	// 获取用户角色
	roles, err := opUser.GetUserRoles(user.ID)
	if err != nil {
		utilG.Response(utils.ERROR, utils.ERROR, "Failed to get user roles")
		return
	}

	// 检查是否有 super 角色或拥有所有权限的角色（可以查看所有资源）
	isSuperOrSystem := false
	for _, role := range roles {
		if permissions.IsSuperRole(role) || permissions.HasAllPermissions(role) {
			isSuperOrSystem = true
			break
		}
	}

	var resList []model.Resource

	if isSuperOrSystem {
		// super/system 角色：返回所有资源
		if resourceType != "" {
			// 根据资源类型获取所有资源
			allRoles, err := operation.NewRoleOperation().GetAllRoles()
			if err != nil {
				utilG.Response(utils.ERROR, utils.ERROR, err.Error())
				return
			}
			for _, role := range allRoles {
				resources, err := opRes.GetResourceListByRoleId(role.ID, resourceType)
				if err != nil {
					continue
				}
				resList = append(resList, resources...)
			}
		} else {
			// 获取所有类型的资源
			resourceTypes := []string{
				constants.ResourceTypeLinux,
				constants.ResourceTypeWindows,
				constants.ResourceTypeDocker,
				constants.ResourceTypeDatabase,
				constants.ResourceTypeRouter,
				constants.ResourceTypeSwitch,
			}
			allRoles, err := operation.NewRoleOperation().GetAllRoles()
			if err != nil {
				utilG.Response(utils.ERROR, utils.ERROR, err.Error())
				return
			}
			for _, resType := range resourceTypes {
				for _, role := range allRoles {
					resources, err := opRes.GetResourceListByRoleId(role.ID, resType)
					if err != nil {
						continue
					}
					resList = append(resList, resources...)
				}
			}
		}
	} else {
		// 其他角色：只返回分配给该用户角色的资源
		for _, role := range roles {
			if resourceType != "" {
				resources, err := opRes.GetResourceListByRoleId(role.ID, resourceType)
				if err != nil {
					continue
				}
				resList = append(resList, resources...)
			} else {
				// 获取所有类型的资源
				resourceTypes := []string{
					constants.ResourceTypeLinux,
					constants.ResourceTypeWindows,
					constants.ResourceTypeDocker,
					constants.ResourceTypeDatabase,
					constants.ResourceTypeRouter,
					constants.ResourceTypeSwitch,
				}
				for _, resType := range resourceTypes {
					resources, err := opRes.GetResourceListByRoleId(role.ID, resType)
					if err != nil {
						continue
					}
					resList = append(resList, resources...)
				}
			}
		}
	}

	// 去重（基于资源 ID 和类型）并过滤已删除的资源，同时进行权限检查
	uniqueResources := make(map[string]model.Resource)
	for _, res := range resList {
		// 检查资源是否已删除
		isDeleted := false
		var resType string
		switch v := res.(type) {
		case *model.LinuxConfig:
			isDeleted = v.DeletedAt.Valid
			resType = constants.ResourceTypeLinux
		case *model.WindowsConfig:
			isDeleted = v.DeletedAt.Valid
			resType = constants.ResourceTypeWindows
		case *model.DockerConfig:
			isDeleted = v.DeletedAt.Valid
			resType = constants.ResourceTypeDocker
		case *model.DatabaseConfig:
			isDeleted = v.DeletedAt.Valid
			resType = constants.ResourceTypeDatabase
		case *model.RouterConfig:
			isDeleted = v.DeletedAt.Valid
			resType = constants.ResourceTypeRouter
		case *model.SwitchConfig:
			isDeleted = v.DeletedAt.Valid
			resType = constants.ResourceTypeSwitch
		}
		// 跳过已删除的资源
		if isDeleted {
			continue
		}

		// 使用 CheckResourceAccess 进行权限检查（角色 + 空间）
		// 对于 super/system 角色，如果配置允许绕过，则跳过检查
		policy := global.CONFIG.PermissionPolicy
		skipCheck := false
		if policy != nil && policy.SuperBypassAll && isSuperOrSystem {
			skipCheck = true
		}

		if !skipCheck {
			// 传入已获取的用户角色，避免重复查询
			allowed, _ := permissions.CheckResourceAccessWithRoles(user, roles, res.GetID(), resType, "list")
			if !allowed {
				continue
			}
		}

		// 使用资源名称和 ID 作为唯一键
		key := fmt.Sprintf("%s-%d", res.GetName(), res.GetID())
		uniqueResources[key] = res
	}

	// 批量获取所有资源的角色和空间信息（优化性能，避免 N+1 查询）
	opResourceRole := operation.NewResourceRoleOperation()
	opSpace := operation.NewSpaceOperation()

	// 构建资源键映射：resourceKey -> (resource, resType)
	type resourceInfo struct {
		resource model.Resource
		resType  string
	}
	resourceMap := make(map[string]resourceInfo)
	resourceKeys := make([]string, 0, len(uniqueResources))

	for _, res := range uniqueResources {
		var resType string
		switch res.(type) {
		case *model.LinuxConfig:
			resType = constants.ResourceTypeLinux
		case *model.WindowsConfig:
			resType = constants.ResourceTypeWindows
		case *model.DockerConfig:
			resType = constants.ResourceTypeDocker
		case *model.DatabaseConfig:
			resType = constants.ResourceTypeDatabase
		case *model.RouterConfig:
			resType = constants.ResourceTypeRouter
		case *model.SwitchConfig:
			resType = constants.ResourceTypeSwitch
		}

		key := fmt.Sprintf("%s:%d", resType, res.GetID())
		resourceMap[key] = resourceInfo{resource: res, resType: resType}
		resourceKeys = append(resourceKeys, key)
	}

	// 批量获取当前资源的角色信息（优化：只查询当前资源的角色）
	roleMap := make(map[string][]*model.Role) // key -> roles
	if len(resourceKeys) > 0 {
		// 构建查询条件：收集所有 (resource_type, resource_id) 对
		type resourceKey struct {
			ResourceType string
			ResourceID   int64
		}
		resourceKeysList := make([]resourceKey, 0, len(resourceMap))
		for key, info := range resourceMap {
			_ = key // 避免未使用变量警告
			resourceKeysList = append(resourceKeysList, resourceKey{
				ResourceType: info.resType,
				ResourceID:   info.resource.GetID(),
			})
		}

		// 批量查询这些资源的角色
		var allResourceRoles []model.ResourceRole
		query := opResourceRole.DB.Preload("Role")
		for _, rk := range resourceKeysList {
			query = query.Or("resource_type = ? AND resource_id = ?", rk.ResourceType, rk.ResourceID)
		}
		query.Find(&allResourceRoles)

		for _, rr := range allResourceRoles {
			key := fmt.Sprintf("%s:%d", rr.ResourceType, rr.ResourceID)
			if _, exists := resourceMap[key]; exists && rr.Role != nil {
				roleMap[key] = append(roleMap[key], rr.Role)
			}
		}
	}

	// 批量获取当前资源的空间信息（优化：只查询当前资源的空间）
	spaceMap := make(map[string]*model.Space) // key -> space
	if len(resourceKeys) > 0 {
		// 构建查询条件：收集所有 (resource_type, resource_id) 对
		type resourceKey struct {
			ResourceType string
			ResourceID   int64
		}
		resourceKeysList := make([]resourceKey, 0, len(resourceMap))
		for key, info := range resourceMap {
			_ = key // 避免未使用变量警告
			resourceKeysList = append(resourceKeysList, resourceKey{
				ResourceType: info.resType,
				ResourceID:   info.resource.GetID(),
			})
		}

		// 批量查询这些资源的空间
		var allResourceSpaces []model.ResourceSpace
		query := opSpace.DB.Preload("Space")
		for _, rk := range resourceKeysList {
			query = query.Or("resource_type = ? AND resource_id = ?", rk.ResourceType, rk.ResourceID)
		}
		query.Find(&allResourceSpaces)

		for _, rs := range allResourceSpaces {
			key := fmt.Sprintf("%s:%d", rs.ResourceType, rs.ResourceID)
			if _, exists := resourceMap[key]; exists && rs.Space != nil {
				spaceMap[key] = rs.Space
			}
		}
	}

	// 构建最终列表
	finalList := make([]interface{}, 0, len(uniqueResources))
	for _, key := range resourceKeys {
		info := resourceMap[key]
		res := info.resource

		// 获取该资源的角色和空间
		roles := roleMap[key]
		space := spaceMap[key]

		// 将资源转换为 map，添加角色和空间信息
		resMap := make(map[string]interface{})
		// 使用 JSON 序列化/反序列化来转换资源对象
		resJSON, _ := json.Marshal(res)
		json.Unmarshal(resJSON, &resMap)

		if len(roles) > 0 {
			rolesJSON, _ := json.Marshal(roles)
			var rolesArray []map[string]interface{}
			json.Unmarshal(rolesJSON, &rolesArray)
			resMap["roles"] = rolesArray
		}

		if space != nil {
			spaceJSON, _ := json.Marshal(space)
			var spaceMapData map[string]interface{}
			json.Unmarshal(spaceJSON, &spaceMapData)
			resMap["space"] = spaceMapData
		}

		finalList = append(finalList, resMap)
	}

	utilG.Response(utils.SUCCESS, utils.SUCCESS, finalList)
}

// GetResourceByID 根据 ID 获取单个资源
func (r *ResourceControl) GetResourceByID(c *gin.Context) {
	utilG := utils.Gin{C: c}

	resourceID := c.Param("id")
	resourceType := c.Query("type")
	if resourceType == "" {
		utilG.Response(utils.ERROR, utils.ERROR, "Resource type is required")
		return
	}

	// 获取当前用户
	userInterface, exists := c.Get("user")
	if !exists {
		utilG.Response(utils.ERROR, utils.ERROR, "User not found in context")
		return
	}
	user, ok := userInterface.(*model.User)
	if !ok {
		utilG.Response(utils.ERROR, utils.ERROR, "Invalid user type")
		return
	}

	opRes := operation.NewResourceOperation()
	opUser := operation.NewUserOperation()

	// 获取用户角色
	roles, err := opUser.GetUserRoles(user.ID)
	if err != nil {
		utilG.Response(utils.ERROR, utils.ERROR, "Failed to get user roles")
		return
	}

	// 检查是否有 super 角色或拥有所有权限的角色
	isSuperOrSystem := false
	for _, role := range roles {
		if permissions.IsSuperRole(role) || permissions.HasAllPermissions(role) {
			isSuperOrSystem = true
			break
		}
	}

	// 根据资源类型获取资源
	var resource model.Resource
	idInt, err := strconv.ParseInt(resourceID, 10, 64)
	if err != nil {
		utilG.Response(utils.ERROR, utils.ERROR, "Invalid resource ID")
		return
	}

	switch resourceType {
	case constants.ResourceTypeLinux:
		var linuxConfig model.LinuxConfig
		if err := opRes.DB.Where("id = ?", idInt).First(&linuxConfig).Error; err != nil {
			utilG.Response(utils.ERROR, utils.ERROR, "Resource not found")
			return
		}
		resource = &linuxConfig
	case constants.ResourceTypeWindows:
		var windowsConfig model.WindowsConfig
		if err := opRes.DB.Where("id = ?", idInt).First(&windowsConfig).Error; err != nil {
			utilG.Response(utils.ERROR, utils.ERROR, "Resource not found")
			return
		}
		resource = &windowsConfig
	case constants.ResourceTypeDocker:
		var dockerConfig model.DockerConfig
		if err := opRes.DB.Where("id = ?", idInt).First(&dockerConfig).Error; err != nil {
			utilG.Response(utils.ERROR, utils.ERROR, "Resource not found")
			return
		}
		resource = &dockerConfig
	case constants.ResourceTypeDatabase:
		var databaseConfig model.DatabaseConfig
		if err := opRes.DB.Where("id = ?", idInt).First(&databaseConfig).Error; err != nil {
			utilG.Response(utils.ERROR, utils.ERROR, "Resource not found")
			return
		}
		resource = &databaseConfig
	case constants.ResourceTypeRouter:
		var routerConfig model.RouterConfig
		if err := opRes.DB.Where("id = ?", idInt).First(&routerConfig).Error; err != nil {
			utilG.Response(utils.ERROR, utils.ERROR, "Resource not found")
			return
		}
		resource = &routerConfig
	case constants.ResourceTypeSwitch:
		var switchConfig model.SwitchConfig
		if err := opRes.DB.Where("id = ?", idInt).First(&switchConfig).Error; err != nil {
			utilG.Response(utils.ERROR, utils.ERROR, "Resource not found")
			return
		}
		resource = &switchConfig
	default:
		utilG.Response(utils.ERROR, utils.ERROR, "Unknown resource type")
		return
	}

	// 检查用户是否有权限访问该资源（使用多维度权限检查）
	if !isSuperOrSystem {
		allowed, reason := permissions.CheckResourceAccessWithRoles(user, roles, resource.GetID(), resourceType, "get")
		if !allowed {
			utilG.Response(utils.ERROR, utils.ERROR, "Permission denied: "+reason)
			return
		}
	}

	// 获取资源的角色信息（所有角色）
	opResourceRole := operation.NewResourceRoleOperation()
	resourceRoles, _ := opResourceRole.GetResourceRoles(idInt, resourceType)
	resourceRoleList := make([]*model.Role, 0)
	for _, rr := range resourceRoles {
		if rr.Role != nil {
			resourceRoleList = append(resourceRoleList, rr.Role)
		}
	}

	// 获取资源的空间信息
	opSpace := operation.NewSpaceOperation()
	var space *model.Space
	resourceSpace, spaceErr := opSpace.GetResourceSpace(idInt, resourceType)
	if spaceErr == nil && resourceSpace != nil && resourceSpace.Space != nil {
		space = resourceSpace.Space
	}

	// 构建响应，包含资源信息、角色信息和空间信息
	response := map[string]interface{}{
		"resource": resource,
	}
	if len(resourceRoleList) > 0 {
		response["roles"] = resourceRoleList
		// 为了向后兼容，也保留单个 role 字段（使用第一个角色）
		response["role"] = resourceRoleList[0]
	}
	if space != nil {
		response["space"] = space
	}

	utilG.Response(utils.SUCCESS, utils.SUCCESS, response)
}
