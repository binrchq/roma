package api

import (
	"net/http"
	"strconv"

	"bitrec.ai/roma/core/model"
	"bitrec.ai/roma/core/operation"
	"bitrec.ai/roma/core/utils"
	"github.com/gin-gonic/gin"
)

type UserController struct {
	// 可以添加其他依赖项
}

func NewUserController() *UserController {
	return &UserController{}
}

type CreateUserRequest struct {
	Username  string   `json:"username" binding:"required"`
	Name      string   `json:"name" binding:"required"`
	Nickname  string   `json:"nickname" binding:"required"`
	Password  string   `json:"password" binding:"required"`
	PublicKey string   `json:"public_key"`
	Email     string   `json:"email" binding:"required"`
	RoleIDs   []uint64 `json:"role_ids"` // 角色ID列表
}

// CreateUser godoc
// @Summary Create a new user
// @Description Create a new user with the input payload
// @Tags users
// @Accept json
// @Produce json
// @Param user body model.User true "User Data"
// @Success 200 {object} utils.Response{data=model.User}
// @Failure 400 {object} utils.Response{data=""}
// @Failure 500 {object} utils.Response{data=""}
// @Router /api/resource/users [post]
func (uc *UserController) CreateUser(c *gin.Context) {
	utilG := utils.Gin{C: c}
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utilG.Response(http.StatusBadRequest, utils.ERROR, "无效的输入数据")
		return
	}

	// 查找请求中的角色
	var roles []model.Role
	if len(req.RoleIDs) > 0 {
		for _, roleID := range req.RoleIDs {
			opRole := operation.NewRoleOperation()
			role, err := opRole.GetRoleByID(roleID)
			if err != nil {
				utilG.Response(http.StatusInternalServerError, utils.ERROR, "获取角色失败")
				return
			}
			roles = append(roles, *role)
		}
	}
	// 创建新用户
	newUser := &model.User{
		Username:  req.Username,
		Name:      req.Name,
		Nickname:  req.Nickname,
		Password:  req.Password,
		PublicKey: req.PublicKey,
		Email:     req.Email,
		Roles:     roles, // 关联角色
	}

	opUser := operation.NewUserOperation()
	newUser, err := opUser.CreateUser(newUser)
	for _, role := range roles {
		err = opUser.AddRoleToUser(newUser.ID, role.ID)
		if err != nil {
			utilG.Response(http.StatusInternalServerError, utils.ERROR, "创建用户失败"+err.Error())
		}
	}
	if err != nil {
		utilG.Response(http.StatusInternalServerError, utils.ERROR, "创建用户失败"+err.Error())
		return
	}

	utilG.Response(http.StatusOK, utils.SUCCESS, "用户创建成功")
}

// GetAllUsers godoc
// @Summary Get all users
// @Description Get all users
// @Tags users
// @Produce json
// @Success 200 {array} model.User
// @Failure 500 {object} utils.Response{data=""}
// @Router /api/resource/users [get]
func (uc *UserController) GetAllUsers(c *gin.Context) {
	utilG := utils.Gin{C: c}
	opUser := operation.NewUserOperation()
	users, err := opUser.GetAllUsers()
	if err != nil {
		utilG.Response(http.StatusInternalServerError, utils.ERROR, "获取用户列表失败")
		return
	}
	utilG.Response(http.StatusOK, utils.SUCCESS, users)
}

// 根据ID获取用户
func (uc *UserController) GetUserByID(c *gin.Context) {
	utilG := utils.Gin{C: c}
	userID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utilG.Response(http.StatusBadRequest, utils.ERROR, "无效的用户ID")
		return
	}
	opUser := operation.NewUserOperation()
	user, err := opUser.GetUserByID(userID)
	if err != nil {
		utilG.Response(http.StatusNotFound, utils.ERROR, "用户未找到")
		return
	}
	utilG.Response(http.StatusOK, utils.SUCCESS, user)
}

// 更新用户信息
func (uc *UserController) UpdateUserByID(c *gin.Context) {
	utilG := utils.Gin{C: c}
	userID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utilG.Response(http.StatusBadRequest, utils.ERROR, "无效的用户ID")
		return
	}
	var updatedUser model.User
	if err := c.ShouldBindJSON(&updatedUser); err != nil {
		utilG.Response(http.StatusBadRequest, utils.ERROR, "无效的输入数据")
		return
	}
	updatedUser.ID = uint(userID) // 确保设置ID
	opUser := operation.NewUserOperation()
	_, err = opUser.UpdateUser(&updatedUser)
	if err != nil {
		utilG.Response(http.StatusInternalServerError, utils.ERROR, "更新用户信息失败")
		return
	}
	utilG.Response(http.StatusOK, utils.SUCCESS, "用户信息更新成功")
}

// 根据ID删除用户
func (uc *UserController) DeleteUserByID(c *gin.Context) {
	utilG := utils.Gin{C: c}
	userID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utilG.Response(http.StatusBadRequest, utils.ERROR, "无效的用户ID")
		return
	}
	opUser := operation.NewUserOperation()
	err = opUser.DisabledUser(userID)
	if err != nil {
		utilG.Response(http.StatusInternalServerError, utils.ERROR, "删除用户失败")
		return
	}
	utilG.Response(http.StatusOK, utils.SUCCESS, "用户删除成功")
}
