package api

import (
	"net/http"
	"strconv"

	"binrc.com/roma/core/model"
	"binrc.com/roma/core/operation"
	"binrc.com/roma/core/permissions"
	"binrc.com/roma/core/utils"
	"github.com/gin-gonic/gin"
)

type SpaceController struct{}

func NewSpaceController() *SpaceController {
	return &SpaceController{}
}

// CreateSpace 创建空间（需要 admin 权限）
// @Summary Create a new space
// @Description Create a new space (requires admin permission)
// @Tags spaces
// @Accept json
// @Produce json
// @Param space body CreateSpaceRequest true "Space Data"
// @Success 200 {object} utils.Response{data=model.Space}
// @Failure 400 {object} utils.Response{data=""}
// @Failure 403 {object} utils.Response{data=""}
// @Failure 500 {object} utils.Response{data=""}
// @Router /api/v1/spaces [post]
func (sc *SpaceController) CreateSpace(c *gin.Context) {
	utilG := utils.Gin{C: c}

	// 检查权限：需要 user.add 权限（admin）
	user, exists := c.Get("user")
	if !exists {
		utilG.Response(http.StatusUnauthorized, utils.ERROR, "Authentication required")
		c.Abort()
		return
	}

	currentUser, ok := user.(*model.User)
	if !ok {
		utilG.Response(http.StatusInternalServerError, utils.ERROR, "Invalid user context")
		c.Abort()
		return
	}

	// 检查用户是否有创建空间的权限（通过权限描述符判断，不硬编码角色名称）
	opUser := operation.NewUserOperation()
	roles, err := opUser.GetUserRoles(currentUser.ID)
	if err != nil {
		utilG.Response(http.StatusForbidden, utils.ERROR, "无法获取用户角色")
		c.Abort()
		return
	}

	hasPermission := false
	for _, role := range roles {
		if role != nil {
			// 检查是否是 super 角色或拥有所有权限的角色
			if permissions.IsSuperRole(role) || permissions.HasAllPermissions(role) {
				hasPermission = true
				break
			}
			// 检查是否有 user.add 权限（用于创建空间）
			desc, err := permissions.ParseRoleDescriptor(role.Desc)
			if err == nil && desc != nil {
				if permissions.HasPermission(desc, "user", "add", "") {
					hasPermission = true
					break
				}
			}
		}
	}

	if !hasPermission {
		utilG.Response(http.StatusForbidden, utils.ERROR, "需要管理员权限才能创建空间")
		c.Abort()
		return
	}

	var req CreateSpaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utilG.Response(http.StatusBadRequest, utils.ERROR, "无效的输入数据")
		return
	}

	space := &model.Space{
		Name:        req.Name,
		Description: req.Description,
		IsActive:    true,
		CreatedBy:   currentUser.ID,
	}

	opSpace := operation.NewSpaceOperation()
	space, err = opSpace.CreateSpace(space)
	if err != nil {
		utilG.Response(http.StatusInternalServerError, utils.ERROR, "创建空间失败: "+err.Error())
		return
	}

	utilG.Response(http.StatusOK, utils.SUCCESS, space)
}

type CreateSpaceRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

// GetAllSpaces 获取所有空间
func (sc *SpaceController) GetAllSpaces(c *gin.Context) {
	utilG := utils.Gin{C: c}
	opSpace := operation.NewSpaceOperation()
	opUser := operation.NewUserOperation()

	// 获取当前用户
	user, exists := c.Get("user")
	if !exists {
		utilG.Response(http.StatusUnauthorized, utils.ERROR, "Authentication required")
		c.Abort()
		return
	}
	currentUser := user.(*model.User)

	// 检查用户是否是 super 角色或拥有所有权限的角色（通过权限描述符判断）
	roles, err := opUser.GetUserRoles(currentUser.ID)
	if err == nil {
		hasAdminRole := false
		for _, role := range roles {
			if role != nil {
				if permissions.IsSuperRole(role) || permissions.HasAllPermissions(role) {
					hasAdminRole = true
					break
				}
			}
		}
		// 如果是管理员，返回所有空间
		if hasAdminRole {
			spaces, err := opSpace.GetAllSpaces()
			if err != nil {
				utilG.Response(http.StatusInternalServerError, utils.ERROR, "获取空间列表失败")
				return
			}
			utilG.Response(http.StatusOK, utils.SUCCESS, spaces)
			return
		}
	}

	// 普通用户只返回所属的空间
	spaces, err := opSpace.GetUserSpaces(currentUser.ID)
	if err != nil {
		utilG.Response(http.StatusInternalServerError, utils.ERROR, "获取空间列表失败")
		return
	}

	utilG.Response(http.StatusOK, utils.SUCCESS, spaces)
}

// GetSpaceByID 根据ID获取空间
func (sc *SpaceController) GetSpaceByID(c *gin.Context) {
	utilG := utils.Gin{C: c}
	spaceID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utilG.Response(http.StatusBadRequest, utils.ERROR, "无效的空间ID")
		return
	}

	opSpace := operation.NewSpaceOperation()
	space, err := opSpace.GetSpaceByID(uint(spaceID))
	if err != nil {
		utilG.Response(http.StatusNotFound, utils.ERROR, "空间未找到")
		return
	}

	utilG.Response(http.StatusOK, utils.SUCCESS, space)
}

// AddSpaceMember 添加空间成员
func (sc *SpaceController) AddSpaceMember(c *gin.Context) {
	utilG := utils.Gin{C: c}
	spaceID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utilG.Response(http.StatusBadRequest, utils.ERROR, "无效的空间ID")
		return
	}

	var req AddSpaceMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utilG.Response(http.StatusBadRequest, utils.ERROR, "无效的输入数据")
		return
	}

	opSpace := operation.NewSpaceOperation()
	member, err := opSpace.AddSpaceMember(uint(spaceID), req.UserID, req.RoleID)
	if err != nil {
		utilG.Response(http.StatusInternalServerError, utils.ERROR, "添加成员失败: "+err.Error())
		return
	}

	utilG.Response(http.StatusOK, utils.SUCCESS, member)
}

type AddSpaceMemberRequest struct {
	UserID uint `json:"user_id" binding:"required"`
	RoleID uint `json:"role_id" binding:"required"`
}

// RemoveSpaceMember 移除空间成员
func (sc *SpaceController) RemoveSpaceMember(c *gin.Context) {
	utilG := utils.Gin{C: c}
	spaceID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utilG.Response(http.StatusBadRequest, utils.ERROR, "无效的空间ID")
		return
	}

	var req struct {
		UserID uint `json:"user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utilG.Response(http.StatusBadRequest, utils.ERROR, "无效的输入数据")
		return
	}

	opSpace := operation.NewSpaceOperation()
	err = opSpace.RemoveSpaceMember(uint(spaceID), req.UserID)
	if err != nil {
		utilG.Response(http.StatusInternalServerError, utils.ERROR, "移除成员失败: "+err.Error())
		return
	}

	utilG.Response(http.StatusOK, utils.SUCCESS, "成员移除成功")
}
