package api

import (
	"net/http"
	"strconv"

	"bitrec.ai/roma/core/model"
	"bitrec.ai/roma/core/operation"
	"bitrec.ai/roma/core/utils"
	"github.com/gin-gonic/gin"
)

type RoleController struct {
	// 可以添加其他依赖项
}

func NewRoleController() *RoleController {
    return &RoleController{}
}
// 创建角色
func (rc *RoleController) CreateRole(c *gin.Context) {
	utilG := utils.Gin{C: c}
	var newRole model.Role
	if err := c.ShouldBindJSON(&newRole); err != nil {
		utilG.Response(http.StatusBadRequest, utils.ERROR, "无效的输入数据")
		return
	}
	opRole := operation.NewRoleOperation()
	_, err := opRole.Create(&newRole)
	if err != nil {
		utilG.Response(http.StatusInternalServerError, utils.ERROR, "创建角色失败")
		return
	}
	utilG.Response(http.StatusOK, utils.SUCCESS, "角色创建成功")
}

// 获取所有角色
func (rc *RoleController) GetAllRoles(c *gin.Context) {
	utilG := utils.Gin{C: c}
	opRole := operation.NewRoleOperation()
	roles, err := opRole.GetAllRoles()
	if err != nil {
		utilG.Response(http.StatusInternalServerError, utils.ERROR, "获取角色列表失败")
		return
	}
	utilG.Response(http.StatusOK, utils.SUCCESS, roles)
}

// 根据ID获取角色
func (rc *RoleController) GetRoleByID(c *gin.Context) {
	utilG := utils.Gin{C: c}
	roleID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utilG.Response(http.StatusBadRequest, utils.ERROR, "无效的角色ID")
		return
	}
	opRole := operation.NewRoleOperation()
	role, err := opRole.GetRoleByID(roleID)
	if err != nil {
		utilG.Response(http.StatusNotFound, utils.ERROR, "角色未找到")
		return
	}
	utilG.Response(http.StatusOK, utils.SUCCESS, role)
}

// 更新角色信息
func (rc *RoleController) UpdateRoleByID(c *gin.Context) {
	utilG := utils.Gin{C: c}
	roleID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utilG.Response(http.StatusBadRequest, utils.ERROR, "无效的角色ID")
		return
	}
	var updatedRole model.Role
	if err := c.ShouldBindJSON(&updatedRole); err != nil {
		utilG.Response(http.StatusBadRequest, utils.ERROR, "无效的输入数据")
		return
	}
	updatedRole.ID = uint(roleID) // 确保设置ID
	opRole := operation.NewRoleOperation()
	_, err = opRole.Update(&updatedRole)
	if err != nil {
		utilG.Response(http.StatusInternalServerError, utils.ERROR, "更新角色信息失败")
		return
	}
	utilG.Response(http.StatusOK, utils.SUCCESS, "角色信息更新成功")
}

// 根据ID删除角色
func (rc *RoleController) DeleteRoleByID(c *gin.Context) {
	utilG := utils.Gin{C: c}
	roleID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utilG.Response(http.StatusBadRequest, utils.ERROR, "无效的角色ID")
		return
	}
	opRole := operation.NewRoleOperation()
	err = opRole.DeleteByID(roleID)
	if err != nil {
		utilG.Response(http.StatusInternalServerError, utils.ERROR, "删除角色失败")
		return
	}
	utilG.Response(http.StatusOK, utils.SUCCESS, "角色删除成功")
}
