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
		Type string            `json:"type"`
		Data []json.RawMessage `json:"data"` // 使用 json.RawMessage 保存未解码的 JSON 字符串
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
		Type string            `json:"type"`
		Data []json.RawMessage `json:"data"` // 使用 json.RawMessage 保存未解码的 JSON 字符串
		Role string            `json:"role"` // 可选的角色名称
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
				errMsg := fmt.Sprintf("资源赋值失败:原因.%s 数据No.%d", err.Error(), id)
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

	// 检查是否是 super 或 system 角色（可以查看所有资源）
	isSuperOrSystem := false
	for _, role := range roles {
		if role.Name == "super" || role.Name == "system" {
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

	// 去重（基于资源 ID 和类型）并过滤已删除的资源
	uniqueResources := make(map[string]model.Resource)
	for _, res := range resList {
		// 检查资源是否已删除
		isDeleted := false
		switch v := res.(type) {
		case *model.LinuxConfig:
			isDeleted = v.DeletedAt.Valid
		case *model.WindowsConfig:
			isDeleted = v.DeletedAt.Valid
		case *model.DockerConfig:
			isDeleted = v.DeletedAt.Valid
		case *model.DatabaseConfig:
			isDeleted = v.DeletedAt.Valid
		case *model.RouterConfig:
			isDeleted = v.DeletedAt.Valid
		case *model.SwitchConfig:
			isDeleted = v.DeletedAt.Valid
		}
		// 跳过已删除的资源
		if isDeleted {
			continue
		}
		// 使用资源名称和 ID 作为唯一键
		key := fmt.Sprintf("%s-%d", res.GetName(), res.GetID())
		uniqueResources[key] = res
	}

	finalList := make([]model.Resource, 0, len(uniqueResources))
	for _, res := range uniqueResources {
		finalList = append(finalList, res)
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

	// 检查是否是 super 或 system 角色
	isSuperOrSystem := false
	for _, role := range roles {
		if role.Name == "super" || role.Name == "system" {
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

	// 如果不是 super/system 角色，检查资源是否分配给该用户
	if !isSuperOrSystem {
		hasAccess := false
		for _, role := range roles {
			var resourceRole model.ResourceRole
			if err := opRes.DB.Where("resource_id = ? AND resource_type = ? AND role_id = ?",
				idInt, resourceType, role.ID).First(&resourceRole).Error; err == nil {
				hasAccess = true
				break
			}
		}
		if !hasAccess {
			utilG.Response(utils.ERROR, utils.ERROR, "Access denied: resource not assigned to your role")
			return
		}
	}

	// 获取资源的角色信息
	var resourceRole model.ResourceRole
	var roleInfo *model.Role
	if err := opRes.DB.Where("resource_id = ? AND resource_type = ?", idInt, resourceType).First(&resourceRole).Error; err == nil {
		opRole := operation.NewRoleOperation()
		role, err := opRole.GetRoleByID(uint64(resourceRole.RoleID))
		if err == nil {
			roleInfo = role
		}
	}

	// 构建响应，包含资源信息和角色信息
	response := map[string]interface{}{
		"resource": resource,
	}
	if roleInfo != nil {
		response["role"] = roleInfo
	}

	utilG.Response(utils.SUCCESS, utils.SUCCESS, response)
}
