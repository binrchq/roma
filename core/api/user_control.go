package api

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"binrc.com/roma/core/model"
	"binrc.com/roma/core/operation"
	"binrc.com/roma/core/utils"
	"github.com/gin-gonic/gin"
	"github.com/loganchef/ssh"
)

// maskKey 掩码密钥，只显示头尾各20个字符
func maskKey(key string) string {
	if key == "" {
		return ""
	}
	if len(key) <= 40 {
		return key
	}
	// 显示前20个字符和后20个字符
	prefix := key[:20]
	suffix := key[len(key)-20:]
	return prefix + "..." + suffix
}

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
	// 使用 bcrypt 加密用户密码（不可逆）
	hashedPassword, encryptErr := utils.HashPassword(req.Password)
	if encryptErr != nil {
		utilG.Response(http.StatusInternalServerError, utils.ERROR, "密码加密失败: "+encryptErr.Error())
		return
	}

	// 标准化公钥格式（如果提供了公钥）
	publicKey := strings.TrimSpace(req.PublicKey)
	if publicKey != "" {
		// 验证公钥格式
		_, _, _, _, err := ssh.ParseAuthorizedKey([]byte(publicKey))
		if err != nil {
			utilG.Response(http.StatusBadRequest, utils.ERROR, "公钥格式无效: "+err.Error())
			return
		}
		// 保持原始格式，只去除前后空白
		publicKey = strings.TrimSpace(publicKey)
	}

	// 创建新用户
	newUser := &model.User{
		Username:  req.Username,
		Name:      req.Name,
		Nickname:  req.Nickname,
		Password:  hashedPassword,
		PublicKey: publicKey,
		Email:     req.Email,
		Roles:     roles, // 关联角色
	}

	opUser := operation.NewUserOperation()
	newUser, err := opUser.CreateUser(newUser)
	if err != nil {
		utilG.Response(http.StatusInternalServerError, utils.ERROR, "创建用户失败: "+err.Error())
		return
	}

	// 添加角色关联（如果失败，记录警告但不影响用户创建）
	var roleErrors []string
	for _, role := range roles {
		err = opUser.AddRoleToUser(newUser.ID, role.ID)
		if err != nil {
			// 记录错误但继续处理其他角色
			roleErrors = append(roleErrors, fmt.Sprintf("角色 %s 添加失败: %v", role.Name, err))
			log.Printf("添加用户角色失败: user_id=%d, role_id=%d, error=%v", newUser.ID, role.ID, err)
		}
	}

	// 如果有角色添加失败，返回警告信息，但用户已创建成功
	if len(roleErrors) > 0 {
		// 返回用户信息和警告
		responseData := map[string]interface{}{
			"user":    newUser,
			"warning": "用户创建成功，但部分角色添加失败: " + fmt.Sprintf("%v", roleErrors),
		}
		utilG.Response(http.StatusOK, utils.SUCCESS, responseData)
		return
	}

	// 全部成功，返回用户信息
	utilG.Response(http.StatusOK, utils.SUCCESS, newUser)
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
	user, err := opUser.GetUserByID(uint(userID))
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

	// 获取原用户信息用于审计日志
	opUser := operation.NewUserOperation()
	oldUser, _ := opUser.GetUserByID(uint(userID))

	var updatedUser model.User
	if err := c.ShouldBindJSON(&updatedUser); err != nil {
		utilG.Response(http.StatusBadRequest, utils.ERROR, "无效的输入数据")
		return
	}
	updatedUser.ID = uint(userID) // 确保设置ID
	_, err = opUser.UpdateUser(&updatedUser)
	if err != nil {
		RecordAuditLog(c, "update_user", "high_risk", "user", uint(userID), oldUser.Username,
			fmt.Sprintf("更新用户信息失败: %s", err.Error()), "failed", err.Error())
		utilG.Response(http.StatusInternalServerError, utils.ERROR, "更新用户信息失败")
		return
	}

	// 记录审计日志（成功）
	RecordAuditLog(c, "update_user", "high_risk", "user", uint(userID), oldUser.Username,
		fmt.Sprintf("更新用户信息: %s (ID: %d)", oldUser.Username, userID), "success", "")

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

	// 获取用户信息用于审计日志
	opUser := operation.NewUserOperation()
	user, _ := opUser.GetUserByID(uint(userID))
	username := ""
	if user != nil {
		username = user.Username
	}

	err = opUser.DisabledUser(userID)
	if err != nil {
		RecordAuditLog(c, "delete_user", "high_risk", "user", uint(userID), username,
			fmt.Sprintf("删除用户失败: %s", err.Error()), "failed", err.Error())
		utilG.Response(http.StatusInternalServerError, utils.ERROR, "删除用户失败")
		return
	}

	// 记录审计日志（成功）
	RecordAuditLog(c, "delete_user", "high_risk", "user", uint(userID), username,
		fmt.Sprintf("删除用户: %s (ID: %d)", username, userID), "success", "")

	utilG.Response(http.StatusOK, utils.SUCCESS, "用户删除成功")
}

// GetCurrentUser 获取当前登录用户信息
func (uc *UserController) GetCurrentUser(c *gin.Context) {
	utilG := utils.Gin{C: c}

	// 从上下文获取用户（由中间件设置）
	user, exists := c.Get("user")
	if !exists {
		utilG.Response(http.StatusUnauthorized, utils.ERROR, "未认证")
		return
	}

	currentUser, ok := user.(*model.User)
	if !ok {
		log.Printf("Failed to convert user to *model.User, type: %T", user)
		utilG.Response(http.StatusInternalServerError, utils.ERROR, "用户信息格式错误")
		return
	}

	// 重新查询用户并预加载角色信息（确保角色被正确加载）
	opUser := operation.NewUserOperation()
	fullUser := &model.User{}
	if err := opUser.DB.Preload("Roles").First(fullUser, currentUser.ID).Error; err != nil {
		log.Printf("Failed to load user with roles: %v", err)
		// 如果查询失败，使用原始用户信息，但返回空角色数组
		responseData := map[string]interface{}{
			"user":  currentUser,
			"roles": []*model.Role{},
		}
		utilG.Response(http.StatusOK, utils.SUCCESS, responseData)
		return
	}

	log.Printf("Loaded user ID: %d, Roles count: %d", fullUser.ID, len(fullUser.Roles))
	for i, role := range fullUser.Roles {
		log.Printf("Role %d: ID=%d, Name=%s", i, role.ID, role.Name)
	}

	// 将 []Role 转换为 []*model.Role
	roles := make([]*model.Role, len(fullUser.Roles))
	for i := range fullUser.Roles {
		roles[i] = &fullUser.Roles[i]
	}

	// 掩码公钥（只显示头尾）
	maskedUser := *fullUser
	if maskedUser.PublicKey != "" {
		maskedUser.PublicKey = maskKey(maskedUser.PublicKey)
	}

	// 构建返回数据，包含用户信息和角色信息
	responseData := map[string]interface{}{
		"user":  maskedUser,
		"roles": roles,
	}

	log.Printf("Response data: user ID=%d, roles count=%d", fullUser.ID, len(roles))
	utilG.Response(http.StatusOK, utils.SUCCESS, responseData)
}

type UpdateProfileRequest struct {
	Name     string `json:"name"`
	Nickname string `json:"nickname"`
	Email    string `json:"email"`
	Password string `json:"password"` // 可选，如果提供则更新密码
}

// UpdateProfile 更新当前用户自己的资料
func (uc *UserController) UpdateProfile(c *gin.Context) {
	utilG := utils.Gin{C: c}

	// 从上下文获取用户
	user, exists := c.Get("user")
	if !exists {
		utilG.Response(http.StatusUnauthorized, utils.ERROR, "未认证")
		return
	}

	currentUser := user.(*model.User)

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utilG.Response(http.StatusBadRequest, utils.ERROR, "无效的输入数据")
		return
	}

	// 更新允许修改的字段
	if req.Name != "" {
		currentUser.Name = req.Name
	}
	if req.Nickname != "" {
		currentUser.Nickname = req.Nickname
	}
	if req.Email != "" {
		currentUser.Email = req.Email
	}
	if req.Password != "" {
		// 使用 bcrypt 加密用户密码（不可逆）
		hashedPassword, err := utils.HashPassword(req.Password)
		if err != nil {
			utilG.Response(http.StatusInternalServerError, utils.ERROR, "密码加密失败: "+err.Error())
			return
		}
		currentUser.Password = hashedPassword
	}

	opUser := operation.NewUserOperation()
	_, err := opUser.UpdateUser(currentUser)
	if err != nil {
		utilG.Response(http.StatusInternalServerError, utils.ERROR, "更新资料失败")
		return
	}

	utilG.Response(http.StatusOK, utils.SUCCESS, "资料更新成功")
}
