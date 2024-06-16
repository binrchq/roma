package routers

import (
	"bitrec.ai/roma/core/api"
	"bitrec.ai/roma/core/api/middleware"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()
	// 使用中间件
	r.Use(gin.Recovery())          // 恢复从任何恐慌中恢复，如果有的话
	r.Use(middleware.ApiKeyAuth()) // API Key 鉴权中间件

	// 用户相关路由
	userController := api.NewUserController()
	r.GET("/api/users", userController.GetAllUsers)
	r.POST("/api/users", userController.CreateUser)
	r.GET("/api/users/:id", userController.GetUserByID)
	r.PUT("/api/users/:id", userController.UpdateUserByID)
	r.DELETE("/api/users/:id", userController.DeleteUserByID)

	// 角色相关路由
	roleController := api.NewRoleController()
	r.GET("/api/roles", roleController.GetAllRoles)
	r.POST("/api/roles", roleController.CreateRole)
	r.GET("/api/roles/:id", roleController.GetRoleByID)
	r.PUT("/api/roles/:id", roleController.UpdateRoleByID)
	r.DELETE("/api/roles/:id", roleController.DeleteRoleByID)

	// 资源相关路由
	resourceController := api.NewResourceControl()
	r.GET("/api/resources", resourceController.GetAllResource)
	r.POST("/api/resources", resourceController.AddResource)
	r.PUT("/api/resources/:id", resourceController.UpdateResource)
	r.DELETE("/api/resources/:id", resourceController.DeleteResource)

	// API Key 相关路由
	apiKeyController := api.NewApikeyController()
	r.POST("/api/apikeys", apiKeyController.CreateApikey)
	r.DELETE("/api/apikeys/:id", apiKeyController.DeleteApikeyByID)

	return r
}
