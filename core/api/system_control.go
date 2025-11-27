package api

import (
	"os"
	"runtime"
	"strings"
	"time"

	"binrc.com/roma/core/constants"
	"binrc.com/roma/core/global"
	"binrc.com/roma/core/operation"
	"binrc.com/roma/core/pkg/k8s"
	"binrc.com/roma/core/utils"
	"github.com/gin-gonic/gin"
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

	// 获取主机密钥公钥
	opHostKey := operation.NewHostKeyOperation()
	var hostKeyPublicKey string
	if hostKey, err := opHostKey.GetLatestHostKey(); err == nil && hostKey != nil {
		hostKeyPublicKey = string(hostKey.PublicKey)
	}

	// 获取 SSH 服务端口
	// 优先级：
	// 1. 环境变量 ROMA_SSH_ADDRESS (格式: host:port) - 提取端口
	// 2. 通过 Kubernetes API 查询 LoadBalancer Service 的 NodePort
	// 3. 环境变量或配置中的端口
	var sshPort string

	sshAddress := os.Getenv("ROMA_SSH_ADDRESS")
	if sshAddress != "" {
		// 解析 host:port 格式，提取端口
		parts := strings.Split(sshAddress, ":")
		if len(parts) == 2 {
			sshPort = parts[1]
		} else {
			sshPort = "2200"
		}
	} else {
		// 尝试通过 Kubernetes API 查询 NodePort
		if nodePort, err := k8s.GetNodePortFromEnv(); err == nil {
			sshPort = nodePort
		} else {
			// 如果 Kubernetes 查询失败，使用环境变量或默认值
			sshPort = "2200"
			if global.CONFIG != nil && global.CONFIG.Common != nil && global.CONFIG.Common.Port != "" {
				sshPort = global.CONFIG.Common.Port
			}
		}
	}

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
		"ssh_service": map[string]interface{}{
			"port":       sshPort,
			"public_key": hostKeyPublicKey,
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
