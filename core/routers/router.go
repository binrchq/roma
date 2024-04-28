package routers

import (
	"bitrec.ai/roma/core/api"
	"bitrec.ai/roma/core/api/middleware"
	"github.com/gin-gonic/gin"
	// ginSwagger "github.com/swaggo/gin-swagger"
	// "github.com/swaggo/gin-swagger/swaggerFiles"
)

func SetupRouter() *gin.Engine {
	g := gin.Default()
	apiRes := g.Group("/api/resource")
	apiRes.Use(middleware.ApiKeyAuth())
	{
		apiRes.POST("/add", api.AddResource)
	}
	return g
}
