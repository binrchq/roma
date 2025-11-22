package api

import (
	"net/http"

	"binrc.com/roma/core/connector"
	"binrc.com/roma/core/global"
	"binrc.com/roma/core/model"
	"github.com/gin-gonic/gin"
)

// ResourceConnectorController 资源连接器控制器
type ResourceConnectorController struct{}

// NewResourceConnectorController 创建资源连接器控制器
func NewResourceConnectorController() *ResourceConnectorController {
	return &ResourceConnectorController{}
}

// GetDatabaseConnectionInfo 获取数据库连接信息
// @Summary 获取数据库连接信息
// @Tags ResourceConnector
// @Param id path int true "数据库 ID"
// @Success 200 {object} map[string]interface{}
// @Router /resources/database/{id}/connection [get]
func (c *ResourceConnectorController) GetDatabaseConnectionInfo(ctx *gin.Context) {
	id := ctx.Param("id")

	var dbConfig model.DatabaseConfig
	if err := global.CDB.Where("id = ?", id).First(&dbConfig).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "数据库配置不存在"})
		return
	}

	conn := connector.NewDatabaseConnector(&dbConfig)
	info := conn.GetConnectionInfo()

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    info,
	})
}

// ExecuteDatabaseQuery 执行数据库查询
// @Summary 执行数据库查询
// @Tags ResourceConnector
// @Param id path int true "数据库 ID"
// @Param query body string true "SQL 查询"
// @Success 200 {object} map[string]interface{}
// @Router /resources/database/{id}/query [post]
func (c *ResourceConnectorController) ExecuteDatabaseQuery(ctx *gin.Context) {
	id := ctx.Param("id")

	var req struct {
		Query string `json:"query" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var dbConfig model.DatabaseConfig
	if err := global.CDB.Where("id = ?", id).First(&dbConfig).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "数据库配置不存在"})
		return
	}

	conn := connector.NewDatabaseConnector(&dbConfig)
	result, err := conn.ExecuteQuery(req.Query)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// GetDockerConnectionInfo 获取 Docker 连接信息
// @Summary 获取 Docker 连接信息
// @Tags ResourceConnector
// @Param id path int true "Docker ID"
// @Success 200 {object} map[string]interface{}
// @Router /resources/docker/{id}/connection [get]
func (c *ResourceConnectorController) GetDockerConnectionInfo(ctx *gin.Context) {
	id := ctx.Param("id")

	var dockerConfig model.DockerConfig
	if err := global.CDB.Where("id = ?", id).First(&dockerConfig).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Docker 配置不存在"})
		return
	}

	conn := connector.NewDockerConnector(&dockerConfig)
	info := conn.GetConnectionInfo()

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    info,
	})
}

// ExecuteDockerCommand 执行 Docker 命令
// @Summary 执行 Docker 命令
// @Tags ResourceConnector
// @Param id path int true "Docker ID"
// @Param command body string true "Docker 命令"
// @Success 200 {object} map[string]interface{}
// @Router /resources/docker/{id}/command [post]
func (c *ResourceConnectorController) ExecuteDockerCommand(ctx *gin.Context) {
	id := ctx.Param("id")

	var req struct {
		Command string `json:"command" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var dockerConfig model.DockerConfig
	if err := global.CDB.Where("id = ?", id).First(&dockerConfig).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Docker 配置不存在"})
		return
	}

	conn := connector.NewDockerConnector(&dockerConfig)
	defer conn.Close()

	output, err := conn.ExecuteDockerCommand(req.Command)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "output": output})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"output":  output,
	})
}

// GetWindowsConnectionInfo 获取 Windows 连接信息
// @Summary 获取 Windows 连接信息（RDP）
// @Tags ResourceConnector
// @Param id path int true "Windows ID"
// @Success 200 {object} map[string]interface{}
// @Router /resources/windows/{id}/connection [get]
func (c *ResourceConnectorController) GetWindowsConnectionInfo(ctx *gin.Context) {
	id := ctx.Param("id")

	var winConfig model.WindowsConfig
	if err := global.CDB.Where("id = ?", id).First(&winConfig).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Windows 配置不存在"})
		return
	}

	conn := connector.NewWindowsConnector(&winConfig)
	info := conn.GetConnectionInfo()

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    info,
	})
}

// GetRouterConnectionInfo 获取路由器连接信息
// @Summary 获取路由器连接信息（Web + SSH）
// @Tags ResourceConnector
// @Param id path int true "路由器 ID"
// @Success 200 {object} map[string]interface{}
// @Router /resources/router/{id}/connection [get]
func (c *ResourceConnectorController) GetRouterConnectionInfo(ctx *gin.Context) {
	id := ctx.Param("id")

	var routerConfig model.RouterConfig
	if err := global.CDB.Where("id = ?", id).First(&routerConfig).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "路由器配置不存在"})
		return
	}

	conn := connector.NewRouterConnector(&routerConfig)
	info := conn.GetConnectionInfo()

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    info,
	})
}

// ExecuteRouterCommand 执行路由器命令
// @Summary 执行路由器命令
// @Tags ResourceConnector
// @Param id path int true "路由器 ID"
// @Param command body string true "路由器命令"
// @Success 200 {object} map[string]interface{}
// @Router /resources/router/{id}/command [post]
func (c *ResourceConnectorController) ExecuteRouterCommand(ctx *gin.Context) {
	id := ctx.Param("id")

	var req struct {
		Command string `json:"command" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var routerConfig model.RouterConfig
	if err := global.CDB.Where("id = ?", id).First(&routerConfig).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "路由器配置不存在"})
		return
	}

	conn := connector.NewRouterConnector(&routerConfig)
	defer conn.Close()

	output, err := conn.ExecuteCommand(req.Command)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "output": output})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"output":  output,
	})
}

// GetSwitchConnectionInfo 获取交换机连接信息
// @Summary 获取交换机连接信息（SSH）
// @Tags ResourceConnector
// @Param id path int true "交换机 ID"
// @Success 200 {object} map[string]interface{}
// @Router /resources/switch/{id}/connection [get]
func (c *ResourceConnectorController) GetSwitchConnectionInfo(ctx *gin.Context) {
	id := ctx.Param("id")

	var switchConfig model.SwitchConfig
	if err := global.CDB.Where("id = ?", id).First(&switchConfig).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "交换机配置不存在"})
		return
	}

	conn := connector.NewSwitchConnector(&switchConfig)
	info := conn.GetConnectionInfo()

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    info,
	})
}

// ExecuteSwitchCommand 执行交换机命令
// @Summary 执行交换机命令
// @Tags ResourceConnector
// @Param id path int true "交换机 ID"
// @Param command body string true "交换机命令"
// @Success 200 {object} map[string]interface{}
// @Router /resources/switch/{id}/command [post]
func (c *ResourceConnectorController) ExecuteSwitchCommand(ctx *gin.Context) {
	id := ctx.Param("id")

	var req struct {
		Command string `json:"command" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var switchConfig model.SwitchConfig
	if err := global.CDB.Where("id = ?", id).First(&switchConfig).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "交换机配置不存在"})
		return
	}

	conn := connector.NewSwitchConnector(&switchConfig)
	defer conn.Close()

	output, err := conn.ExecuteCommand(req.Command)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "output": output})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"output":  output,
	})
}
