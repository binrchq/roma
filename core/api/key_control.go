package api

import (
	"net/http"
	"strconv"
	"time"

	"bitrec.ai/roma/core/model"
	"bitrec.ai/roma/core/operation"
	"bitrec.ai/roma/core/utils"
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
