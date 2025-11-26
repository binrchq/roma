package routers

import (
	"binrc.com/roma/core/api"
	"binrc.com/roma/core/api/middleware"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()
	// 使用中间件
	r.Use(gin.Recovery()) // 恢复从任何恐慌中恢复，如果有的话
	r.Use(middleware.CORSMiddleware()) // CORS 支持，前后端分离时需要

	// 健康检查端点 - 不需要认证，用于 Docker/K8s 健康检查
	systemController := api.NewSystemController()
	r.GET("/health", systemController.GetHealth)

	// API v1 路由组
	v1 := r.Group("/api/v1")
	{
		// 认证相关路由 - 不需要认证
		authController := api.NewAuthController()
		auth := v1.Group("/auth")
		{
			auth.POST("/login", authController.Login)   // 登录
			auth.POST("/logout", authController.Logout) // 登出
		}

		// 其他路由需要 JWT 或 API Key 认证
		v1.Use(middleware.JWTAuthOrApiKey())
		// 用户相关路由 - 需要 user.add/delete/update/get/list 权限（super 角色）
		userController := api.NewUserController()
		users := v1.Group("/users")
		{
			users.GET("", middleware.RequirePermission("user", "list"), userController.GetAllUsers)
			users.POST("", middleware.RequirePermission("user", "add"), userController.CreateUser)
			users.GET("/:id", middleware.RequirePermission("user", "get"), userController.GetUserByID)
			users.PUT("/:id", middleware.RequirePermission("user", "update"), userController.UpdateUserByID)
			users.DELETE("/:id", middleware.RequirePermission("user", "delete"), userController.DeleteUserByID)

			// 用户自己的资料管理 - 需要认证
			users.GET("/me", middleware.RequirePermission("user", "get"), userController.GetCurrentUser)
			users.PUT("/me", middleware.RequirePermission("user", "update"), userController.UpdateProfile)
		}

		// 角色相关路由 - 需要 user 管理权限（super 角色）
		roleController := api.NewRoleController()
		roles := v1.Group("/roles")
		{
			roles.GET("", middleware.RequirePermission("user", "list"), roleController.GetAllRoles)
			roles.POST("", middleware.RequirePermission("user", "add"), roleController.CreateRole)
			roles.GET("/:id", middleware.RequirePermission("user", "get"), roleController.GetRoleByID)
			roles.PUT("/:id", middleware.RequirePermission("user", "update"), roleController.UpdateRoleByID)
			roles.DELETE("/:id", middleware.RequirePermission("user", "delete"), roleController.DeleteRoleByID)
		}

		// 资源相关路由 - 需要 resource 权限
		resourceController := api.NewResourceControl()
		resources := v1.Group("/resources")
		{
			// 列出资源 - 需要 list 权限
			resources.GET("", middleware.RequirePermission("resource", "list"), resourceController.GetAllResource)
			// 添加资源 - 需要 add 权限（super/system 角色）
			resources.POST("", middleware.RequirePermission("resource", "add"), resourceController.AddResource)
			// 获取单个资源 - 需要 get 权限
			resources.GET("/:id", middleware.RequirePermission("resource", "get"), resourceController.GetResourceByID)
			// 更新资源 - 需要 update 权限（super/system 角色）
			resources.PUT("/:id", middleware.RequirePermission("resource", "update"), resourceController.UpdateResource)
			// 删除资源 - 需要 delete 权限（super/system 角色）
			resources.DELETE("/:id", middleware.RequirePermission("resource", "delete"), resourceController.DeleteResource)
		}

		// 资源连接和执行相关路由 - 需要 use 权限
		resourceConnectorController := api.NewResourceConnectorController()
		connectors := v1.Group("/connectors")
		{
			// 数据库连接
			connectors.GET("/database/:id", middleware.RequirePermission("resource", "use"), resourceConnectorController.GetDatabaseConnectionInfo)
			connectors.POST("/database/:id/query", middleware.RequirePermission("resource", "use"), resourceConnectorController.ExecuteDatabaseQuery)

			// Docker 连接
			connectors.GET("/docker/:id", middleware.RequirePermission("resource", "use"), resourceConnectorController.GetDockerConnectionInfo)
			connectors.POST("/docker/:id/command", middleware.RequirePermission("resource", "use"), resourceConnectorController.ExecuteDockerCommand)

			// Windows 连接
			connectors.GET("/windows/:id", middleware.RequirePermission("resource", "use"), resourceConnectorController.GetWindowsConnectionInfo)

			// 路由器连接
			connectors.GET("/router/:id", middleware.RequirePermission("resource", "use"), resourceConnectorController.GetRouterConnectionInfo)
			connectors.POST("/router/:id/command", middleware.RequirePermission("resource", "use"), resourceConnectorController.ExecuteRouterCommand)

			// 交换机连接
			connectors.GET("/switch/:id", middleware.RequirePermission("resource", "use"), resourceConnectorController.GetSwitchConnectionInfo)
			connectors.POST("/switch/:id/command", middleware.RequirePermission("resource", "use"), resourceConnectorController.ExecuteSwitchCommand)
		}

		// 日志相关路由 - 需要 list 权限（所有角色都可以查看）
		logController := api.NewLogController()
		logs := v1.Group("/logs")
		{
			logs.GET("/access", middleware.RequirePermission("resource", "list"), logController.GetAccessLogs)
			logs.GET("/credential", middleware.RequirePermission("resource", "list"), logController.GetCredentialLogs)
			logs.GET("/audit", middleware.RequirePermission("resource", "list"), logController.GetAuditLogs)
		}

		// 系统相关路由 - 所有角色都可以访问
		systemController := api.NewSystemController()
		system := v1.Group("/system")
		{
			system.GET("/info", systemController.GetSystemInfo)
			system.GET("/health", systemController.GetHealth)
		}

		// API Key 相关路由 - 需要 user 管理权限（super 角色，仅管理员）
		apiKeyController := api.NewApikeyController()
		apiKeys := v1.Group("/apikeys")
		{
			apiKeys.GET("", middleware.RequirePermission("user", "list"), apiKeyController.GetAllApikeys)
			apiKeys.POST("", middleware.RequirePermission("user", "add"), apiKeyController.CreateApikey)
			apiKeys.GET("/:id", middleware.RequirePermission("user", "get"), apiKeyController.GetApikeyByID)
			apiKeys.DELETE("/:id", middleware.RequirePermission("user", "delete"), apiKeyController.DeleteApikeyByID)
		}

		// SSH 密钥管理路由 - 用户自己的 SSH 密钥
		sshKeyController := api.NewSSHKeyController()
		sshKeys := v1.Group("/ssh-keys")
		{
			sshKeys.GET("/me", middleware.RequirePermission("user", "get"), sshKeyController.GetMySSHKey)                 // 获取当前用户的 SSH 公钥
			sshKeys.POST("/me/upload", middleware.RequirePermission("user", "update"), sshKeyController.UploadSSHKey)     // 上传 SSH 密钥
			sshKeys.POST("/me/generate", middleware.RequirePermission("user", "update"), sshKeyController.GenerateSSHKey) // 重新生成 SSH 密钥
		}
	}

	return r
}
