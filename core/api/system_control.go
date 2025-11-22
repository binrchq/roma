package api

import (
	"runtime"

	"binrc.com/roma/core/constants"
	"binrc.com/roma/core/operation"
	"binrc.com/roma/core/utils"
	"github.com/gin-gonic/gin"
	"time"
)

type SystemController struct{}

func NewSystemController() *SystemController {
	return &SystemController{}
}

// GetSystemInfo 获取系统信息
func (s *SystemController) GetSystemInfo(c *gin.Context) {
	utilG := utils.Gin{C: c}

	opRes := operation.NewResourceOperation()
	opUser := operation.NewUserOperation()
	opRole := operation.NewRoleOperation()

	// 统计资源
	resourceTypes := []string{
		constants.ResourceTypeLinux,
		constants.ResourceTypeWindows,
		constants.ResourceTypeDocker,
		constants.ResourceTypeDatabase,
		constants.ResourceTypeRouter,
		constants.ResourceTypeSwitch,
	}

	resourceCounts := make(map[string]int)
	totalResources := 0

	roles, _ := opRole.GetAllRoles()
	for _, resourceType := range resourceTypes {
		count := 0
		seen := make(map[int64]bool)
		
		for _, role := range roles {
			resources, err := opRes.GetResourceListByRoleId(role.ID, resourceType)
			if err == nil {
				for _, res := range resources {
					id := res.GetID()
					if !seen[id] {
						seen[id] = true
						count++
					}
				}
			}
		}
		resourceCounts[resourceType] = count
		totalResources += count
	}

	// 统计用户
	users, _ := opUser.GetAllUsers()
	userCount := len(users)

	// 统计角色
	roleCount := len(roles)

	result := map[string]interface{}{
		"system": map[string]interface{}{
			"name":       "ROMA Bastion Host",
			"version":    "1.0.0",
			"go_version": runtime.Version(),
			"os":         runtime.GOOS,
			"arch":       runtime.GOARCH,
		},
		"statistics": map[string]interface{}{
			"total_resources": totalResources,
			"resources":       resourceCounts,
			"total_users":     userCount,
			"total_roles":     roleCount,
		},
		"roles": roles,
	}

	utilG.Response(utils.SUCCESS, utils.SUCCESS, result)
}

// GetHealth 健康检查
func (s *SystemController) GetHealth(c *gin.Context) {
	utilG := utils.Gin{C: c}
	
	utilG.Response(utils.SUCCESS, utils.SUCCESS, map[string]interface{}{
		"status": "healthy",
		"time":   time.Now().Format("2006-01-02 15:04:05"),
	})
}


