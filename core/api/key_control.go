package api

import (
	"net/http"
	"strconv"
	"time"

	"binrc.com/roma/core/model"
	"binrc.com/roma/core/operation"
	"binrc.com/roma/core/utils"
	"github.com/gin-gonic/gin"
)

type ApikeyController struct {
	// 可以添加其他依赖项
}

func NewApikeyController() *ApikeyController {
	return &ApikeyController{}
}

// 创建API Key
// CreateApikey generates a new API key and returns it
func (ac *ApikeyController) CreateApikey(c *gin.Context) {
	utilG := utils.Gin{C: c}

	// Generate a new API Key
	newApiKey := utils.GenerateKey()

	// Create a new Apikey model instance
	newApikey := model.Apikey{
		Apikey:      "apikey." + newApiKey,
		Description: "New API Key",
		ExpiresAt:   time.Now().AddDate(1, 0, 0), // Set expiration date to 1 year from now
	}

	// Save the new API Key to the database
	opKey := operation.NewApikeyOperation()
	_, err := opKey.Create(&newApikey)
	if err != nil {
		utilG.Response(http.StatusInternalServerError, utils.ERROR, "创建API Key失败")
		return
	}

	// Return the newly generated API Key
	utilG.Response(http.StatusOK, utils.SUCCESS, gin.H{"message": "API Key创建成功", "api_key": newApiKey})
}

// 获取所有API Keys
func (ac *ApikeyController) GetAllApikeys(c *gin.Context) {
	utilG := utils.Gin{C: c}
	opKey := operation.NewApikeyOperation()
	apikeys, err := opKey.GetAllApiKeys()
	if err != nil {
		utilG.Response(http.StatusInternalServerError, utils.ERROR, "获取API Key列表失败")
		return
	}
	utilG.Response(http.StatusOK, utils.SUCCESS, apikeys)
}

// 根据ID获取API Key
func (ac *ApikeyController) GetApikeyByID(c *gin.Context) {
	utilG := utils.Gin{C: c}
	apikeyID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utilG.Response(http.StatusBadRequest, utils.ERROR, "无效的API Key ID")
		return
	}
	opKey := operation.NewApikeyOperation()
	apikey, err := opKey.GetApiKeyById(uint(apikeyID))
	if err != nil {
		utilG.Response(http.StatusNotFound, utils.ERROR, "API Key未找到")
		return
	}
	utilG.Response(http.StatusOK, utils.SUCCESS, apikey)
}

// 根据ID删除API Key
func (ac *ApikeyController) DeleteApikeyByID(c *gin.Context) {
	utilG := utils.Gin{C: c}
	apikeyID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utilG.Response(http.StatusBadRequest, utils.ERROR, "无效的API Key ID")
		return
	}
	opKey := operation.NewApikeyOperation()
	err = opKey.ExpiresApikeyById(uint(apikeyID))
	if err != nil {
		utilG.Response(http.StatusInternalServerError, utils.ERROR, "删除API Key失败")
		return
	}
	utilG.Response(http.StatusOK, utils.SUCCESS, "API Key删除成功")
}

type CreateMyApikeyRequest struct {
	Description string `json:"description"`
	ExpiresDays int    `json:"expires_days"` // 过期天数，默认30天
}

// CreateMyApikey 创建当前用户自己的 API Key
func (ac *ApikeyController) CreateMyApikey(c *gin.Context) {
	utilG := utils.Gin{C: c}

	// 从上下文获取用户
	user, exists := c.Get("user")
	if !exists {
		utilG.Response(http.StatusUnauthorized, utils.ERROR, "未认证")
		return
	}

	currentUser := user.(*model.User)

	var req CreateMyApikeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// 如果没有提供，使用默认值
		req.Description = "用户自己创建的 API Key"
		req.ExpiresDays = 30
	}

	if req.Description == "" {
		req.Description = "用户自己创建的 API Key"
	}
	if req.ExpiresDays <= 0 {
		req.ExpiresDays = 30
	}

	// 生成新的 API Key
	newApiKey := utils.GenerateKey()
	fullApiKey := "apikey." + newApiKey

	// 创建 API Key
	newApikey := model.Apikey{
		Apikey:      fullApiKey,
		Description: req.Description + " (用户: " + currentUser.Username + ")",
		ExpiresAt:   time.Now().AddDate(0, 0, req.ExpiresDays),
	}

	opKey := operation.NewApikeyOperation()
	_, err := opKey.Create(&newApikey)
	if err != nil {
		utilG.Response(http.StatusInternalServerError, utils.ERROR, "创建API Key失败")
		return
	}

	utilG.Response(http.StatusOK, utils.SUCCESS, gin.H{
		"message":    "API Key创建成功",
		"api_key":    fullApiKey,
		"expires_at": newApikey.ExpiresAt,
	})
}

// GetMyApikeys 获取当前用户自己的 API Keys
// 注意：当前 Apikey 模型没有 user_id 字段，所以返回所有有效的 API Keys
// 后续可以添加 user_id 字段来关联用户
func (ac *ApikeyController) GetMyApikeys(c *gin.Context) {
	utilG := utils.Gin{C: c}

	// 从上下文获取用户
	_, exists := c.Get("user")
	if !exists {
		utilG.Response(http.StatusUnauthorized, utils.ERROR, "未认证")
		return
	}

	opKey := operation.NewApikeyOperation()
	apikeys, err := opKey.GetAllApiKeys()
	if err != nil {
		utilG.Response(http.StatusInternalServerError, utils.ERROR, "获取API Key列表失败")
		return
	}

	// 过滤出未过期的 API Keys
	now := time.Now()
	validApikeys := []model.Apikey{}
	for _, key := range apikeys {
		if key.ExpiresAt.After(now) {
			validApikeys = append(validApikeys, *key)
		}
	}

	utilG.Response(http.StatusOK, utils.SUCCESS, validApikeys)
}

// DeleteMyApikey 删除当前用户自己的 API Key
func (ac *ApikeyController) DeleteMyApikey(c *gin.Context) {
	utilG := utils.Gin{C: c}

	// 从上下文获取用户
	_, exists := c.Get("user")
	if !exists {
		utilG.Response(http.StatusUnauthorized, utils.ERROR, "未认证")
		return
	}

	apikeyID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utilG.Response(http.StatusBadRequest, utils.ERROR, "无效的API Key ID")
		return
	}

	opKey := operation.NewApikeyOperation()
	err = opKey.ExpiresApikeyById(uint(apikeyID))
	if err != nil {
		utilG.Response(http.StatusInternalServerError, utils.ERROR, "删除API Key失败")
		return
	}

	utilG.Response(http.StatusOK, utils.SUCCESS, "API Key删除成功")
}
