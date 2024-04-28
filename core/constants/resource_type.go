package constants

// 资源类型枚举
const (
	ResourceTypeLinux    = "linux"
	ResourceTypeRouter   = "router"
	ResourceTypeWindows  = "windows"
	ResourceTypeDocker   = "docker"
	ResourceTypeDatabase = "database"
	ResourceTypeSwitch   = "switch"
)

// GetResourceType 返回所有资源类型的切片
func GetResourceType() []string {
	// 直接返回资源类型的切片
	return []string{
		ResourceTypeLinux,
		ResourceTypeRouter,
		ResourceTypeWindows,
		ResourceTypeDocker,
		ResourceTypeDatabase,
		ResourceTypeSwitch,
	}
}
