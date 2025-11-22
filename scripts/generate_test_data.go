package main

import (
	"fmt"
	"log"

	"binrc.com/roma/configs"
	"binrc.com/roma/core/global"
	"binrc.com/roma/core/initialize"
	"binrc.com/roma/core/model"
	"binrc.com/roma/core/operation"
	"binrc.com/roma/core/constants"

	
	"errors"
	"flag"
	"github.com/spf13/viper"
)
var (
	cfgFile string
)

func LoadConfig() {
	if err := readCfg(cfgFile); err != nil {
		panic(err)
	}
}
func readCfg(cfgPath string) error {
	if cfgPath == "" {
		return errors.New("config file is not given")
	}
	v := viper.New()
	v.SetConfigFile(cfgPath)
	err := v.ReadInConfig()
	if err != nil {
		return fmt.Errorf("Fatal error config file: %s \n", err)
	}
	conf := configs.NewConfig()
	if err := v.Unmarshal(&conf); err != nil {
		log.Println(err)
	}
	global.CONFIG = conf
	return nil
}
func init() {
	flag.StringVar(&cfgFile, "c", constants.BASE_DIR+"/configs/config.toml", "path of config file.")
	flag.Parse()
	// 加载配置文件
	LoadConfig()
}
func main() {


	// 初始化数据库连接（会自动创建表）
	fmt.Println("🔌 连接数据库并创建表结构...")
	db, err := initialize.InitCDB()
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	global.CDB = db

	fmt.Println("🚀 开始生成测试数据...")

	// 清理旧的测试数据
	fmt.Println("📝 清理旧测试数据...")
	db.Exec("DELETE FROM linux_configs WHERE hostname LIKE 'test-%'")
	db.Exec("DELETE FROM windows_configs WHERE hostname LIKE 'test-%'")
	db.Exec("DELETE FROM docker_configs WHERE ContainerName LIKE 'test-%'")
	db.Exec("DELETE FROM database_configs WHERE database_nick LIKE 'test-%'")
	db.Exec("DELETE FROM router_configs WHERE router_name LIKE 'test-%'")
	db.Exec("DELETE FROM switch_configs WHERE switch_name LIKE 'test-%'")
	db.Exec("DELETE FROM resource_roles WHERE resource_id IN (SELECT id FROM linux_configs WHERE hostname LIKE 'test-%')")

	// 获取角色 ID（假设你已经有角色数据）
	var superRole, adminRole, devRole model.Role
	db.Where("name = ?", "super").First(&superRole)
	db.Where("name = ?", "admin").First(&adminRole)
	db.Where("name = ?", "developer").First(&devRole)

	if superRole.ID == 0 {
		fmt.Println("⚠️  未找到角色，先创建角色...")
		superRole = model.Role{Name: "super", Desc: "operation:*.*"}
		db.Create(&superRole)
		adminRole = model.Role{Name: "admin", Desc: "operation:*-(*peripheral).*"}
		db.Create(&adminRole)
		devRole = model.Role{Name: "developer", Desc: "operation:*-(*peripheral).get"}
		db.Create(&devRole)
	}

	fmt.Println("✅ 找到角色:")
	fmt.Printf("   - Super: ID=%d\n", superRole.ID)
	fmt.Printf("   - Admin: ID=%d\n", adminRole.ID)
	fmt.Printf("   - Developer: ID=%d\n", devRole.ID)

	// 1. 创建 Linux 测试资源
	fmt.Println("\n🐧 生成 Linux 测试资源...")
	linuxResources := []model.LinuxConfig{
		{
			Hostname:    "test-linux-web-001",
			Port:        22,
			IPv4Pub:     "192.168.1.10",
			IPv4Priv:    "10.0.1.10",
			Username:    "root",
			Password:    "test123",
			Description: "测试Web服务器",
		},
		{
			Hostname:    "test-linux-db-001",
			Port:        22,
			IPv4Pub:     "192.168.1.11",
			IPv4Priv:    "10.0.1.11",
			Username:    "root",
			Password:    "test123",
			Description: "测试数据库服务器",
		},
		{
			Hostname:    "test-linux-app-001",
			Port:        22,
			IPv4Pub:     "192.168.1.12",
			IPv4Priv:    "10.0.1.12",
			Username:    "ubuntu",
			Password:    "test123",
			Description: "测试应用服务器",
		},
	}

	for _, res := range linuxResources {
		if err := db.Create(&res).Error; err != nil {
			fmt.Printf("   ❌ 创建失败: %s - %v\n", res.Hostname, err)
		} else {
			fmt.Printf("   ✅ 创建成功: %s (ID: %d)\n", res.Hostname, res.ID)
			// 绑定到 super 角色
			resourceRole := model.ResourceRole{
				ResourceID:   res.ID,
				ResourceType: "linux",
				RoleID:       int64(superRole.ID),
			}
			db.Create(&resourceRole)
		}
	}

	// 2. 创建 Windows 测试资源
	fmt.Println("\n🪟 生成 Windows 测试资源...")
	windowsResources := []model.WindowsConfig{
		{
			Hostname:    "test-win-srv-001",
			Port:        3389,
			IPv4Pub:     "192.168.2.10",
			IPv4Priv:    "10.0.2.10",
			Username:    "Administrator",
			Password:    "Test@123",
			Description: "测试Windows服务器",
		},
		{
			Hostname:    "test-win-dev-001",
			Port:        3389,
			IPv4Pub:     "192.168.2.11",
			IPv4Priv:    "10.0.2.11",
			Username:    "Developer",
			Password:    "Test@123",
			Description: "测试Windows开发机",
		},
	}

	for _, res := range windowsResources {
		if err := db.Create(&res).Error; err != nil {
			fmt.Printf("   ❌ 创建失败: %s - %v\n", res.Hostname, err)
		} else {
			fmt.Printf("   ✅ 创建成功: %s (ID: %d)\n", res.Hostname, res.ID)
			resourceRole := model.ResourceRole{
				ResourceID:   res.ID,
				ResourceType: "windows",
				RoleID:       int64(superRole.ID),
			}
			db.Create(&resourceRole)
		}
	}

	// 3. 创建 Docker 测试资源
	fmt.Println("\n🐳 生成 Docker 测试资源...")
	dockerResources := []model.DockerConfig{
		{
			ContainerName: "test-docker-nginx",
			Port:          22,
			IPv4Priv:      "10.0.3.10",
			Username:      "root",
			Password:      "test123",
			Description:   "测试Nginx容器",
		},
		{
			ContainerName: "test-docker-redis",
			Port:          22,
			IPv4Priv:      "10.0.3.11",
			Username:      "root",
			Password:      "test123",
			Description:   "测试Redis容器",
		},
	}

	for _, res := range dockerResources {
		if err := db.Create(&res).Error; err != nil {
			fmt.Printf("   ❌ 创建失败: %s - %v\n", res.ContainerName, err)
		} else {
			fmt.Printf("   ✅ 创建成功: %s (ID: %d)\n", res.ContainerName, res.ID)
			resourceRole := model.ResourceRole{
				ResourceID:   res.ID,
				ResourceType: "docker",
				RoleID:       int64(superRole.ID),
			}
			db.Create(&resourceRole)
		}
	}

	// 4. 创建 Database 测试资源
	fmt.Println("\n🗄️  生成 Database 测试资源...")
	databaseResources := []model.DatabaseConfig{
		{
			DatabaseNick: "test-db-mysql",
			DatabaseType: "mysql",
			DatabaseName: "test_db",
			IPv4Pub:      "192.168.4.10",
			IPv4Priv:     "10.0.4.10",
			Port:         3306,
			Username:     "root",
			Password:     "test123",
			Description:  "测试MySQL数据库",
		},
		{
			DatabaseNick: "test-db-postgres",
			DatabaseType: "postgresql",
			DatabaseName: "test_db",
			IPv4Pub:      "192.168.4.11",
			IPv4Priv:     "10.0.4.11",
			Port:         5432,
			Username:     "postgres",
			Password:     "test123",
			Description:  "测试PostgreSQL数据库",
		},
		{
			DatabaseNick: "test-db-redis",
			DatabaseType: "redis",
			DatabaseName: "0",
			IPv4Pub:      "192.168.4.12",
			IPv4Priv:     "10.0.4.12",
			Port:         6379,
			Password:     "test123",
			Description:  "测试Redis数据库",
		},
	}

	for _, res := range databaseResources {
		if err := db.Create(&res).Error; err != nil {
			fmt.Printf("   ❌ 创建失败: %s - %v\n", res.DatabaseNick, err)
		} else {
			fmt.Printf("   ✅ 创建成功: %s (ID: %d)\n", res.DatabaseNick, res.ID)
			resourceRole := model.ResourceRole{
				ResourceID:   res.ID,
				ResourceType: "database",
				RoleID:       int64(superRole.ID),
			}
			db.Create(&resourceRole)
		}
	}

	// 5. 创建 Router 测试资源
	fmt.Println("\n🌐 生成 Router 测试资源...")
	routerResources := []model.RouterConfig{
		{
			RouterName:  "test-router-core-001",
			IPv4Pub:     "192.168.5.1",
			IPv4Priv:    "10.0.5.1",
			Port:        22,
			WebPort:     80,
			Username:    "admin",
			Password:    "test123",
			WebUsername: "admin",
			WebPassword: "test123",
			Description: "测试核心路由器",
		},
		{
			RouterName:  "test-router-edge-001",
			IPv4Pub:     "192.168.5.2",
			IPv4Priv:    "10.0.5.2",
			Port:        22,
			WebPort:     80,
			Username:    "admin",
			Password:    "test123",
			WebUsername: "admin",
			WebPassword: "test123",
			Description: "测试边缘路由器",
		},
	}

	for _, res := range routerResources {
		if err := db.Create(&res).Error; err != nil {
			fmt.Printf("   ❌ 创建失败: %s - %v\n", res.RouterName, err)
		} else {
			fmt.Printf("   ✅ 创建成功: %s (ID: %d)\n", res.RouterName, res.ID)
			resourceRole := model.ResourceRole{
				ResourceID:   res.ID,
				ResourceType: "router",
				RoleID:       int64(superRole.ID),
			}
			db.Create(&resourceRole)
		}
	}

	// 6. 创建 Switch 测试资源
	fmt.Println("\n🔀 生成 Switch 测试资源...")
	switchResources := []model.SwitchConfig{
		{
			SwitchName:  "test-switch-access-001",
			IPv4Pub:     "192.168.6.10",
			IPv4Priv:    "10.0.6.10",
			Port:        22,
			Username:    "admin",
			Password:    "test123",
			Description: "测试接入交换机",
		},
		{
			SwitchName:  "test-switch-core-001",
			IPv4Pub:     "192.168.6.11",
			IPv4Priv:    "10.0.6.11",
			Port:        22,
			Username:    "admin",
			Password:    "test123",
			Description: "测试核心交换机",
		},
	}

	for _, res := range switchResources {
		if err := db.Create(&res).Error; err != nil {
			fmt.Printf("   ❌ 创建失败: %s - %v\n", res.SwitchName, err)
		} else {
			fmt.Printf("   ✅ 创建成功: %s (ID: %d)\n", res.SwitchName, res.ID)
			resourceRole := model.ResourceRole{
				ResourceID:   res.ID,
				ResourceType: "switch",
				RoleID:       int64(superRole.ID),
			}
			db.Create(&resourceRole)
		}
	}

	// 统计信息
	fmt.Println("\n📊 测试数据生成完成！")
	fmt.Println("=====================================")
	
	var counts []struct {
		Type  string
		Count int64
	}

	db.Raw("SELECT 'Linux' as type, COUNT(*) as count FROM linux_configs WHERE hostname LIKE 'test-%'").Scan(&counts)
	for _, c := range counts {
		fmt.Printf("  %s: %d 条\n", c.Type, c.Count)
	}
	
	db.Raw("SELECT 'Windows' as type, COUNT(*) as count FROM windows_configs WHERE hostname LIKE 'test-%'").Scan(&counts)
	for _, c := range counts {
		fmt.Printf("  %s: %d 条\n", c.Type, c.Count)
	}
	
	db.Raw("SELECT 'Docker' as type, COUNT(*) as count FROM docker_configs WHERE container_name LIKE 'test-%'").Scan(&counts)
	for _, c := range counts {
		fmt.Printf("  %s: %d 条\n", c.Type, c.Count)
	}
	
	db.Raw("SELECT 'Database' as type, COUNT(*) as count FROM database_configs WHERE database_nick LIKE 'test-%'").Scan(&counts)
	for _, c := range counts {
		fmt.Printf("  %s: %d 条\n", c.Type, c.Count)
	}
	
	db.Raw("SELECT 'Router' as type, COUNT(*) as count FROM router_configs WHERE router_name LIKE 'test-%'").Scan(&counts)
	for _, c := range counts {
		fmt.Printf("  %s: %d 条\n", c.Type, c.Count)
	}
	
	db.Raw("SELECT 'Switch' as type, COUNT(*) as count FROM switch_configs WHERE switch_name LIKE 'test-%'").Scan(&counts)
	for _, c := range counts {
		fmt.Printf("  %s: %d 条\n", c.Type, c.Count)
	}

	fmt.Println("=====================================")
	fmt.Println("\n🎉 现在可以测试 TUI 了:")
	fmt.Println("   ssh super@localhost -p 2222")
	fmt.Println("   密码: 123456")
	fmt.Println("\n   然后在 TUI 中执行:")
	fmt.Println("   - use linux && ls")
	fmt.Println("   - use windows && ls")
	fmt.Println("   - use docker && ls")
	fmt.Println("   - use database && ls")
	fmt.Println("   - use router && ls")
	fmt.Println("   - use switch && ls")

	// 验证操作
	fmt.Println("\n🔍 验证数据...")
	op := operation.NewResourceOperation()
	
	linuxList, _ := op.GetResourceListByRoleId(superRole.ID, "linux")
	fmt.Printf("   Super 角色的 Linux 资源: %d 个\n", len(linuxList))
	
	windowsList, _ := op.GetResourceListByRoleId(superRole.ID, "windows")
	fmt.Printf("   Super 角色的 Windows 资源: %d 个\n", len(windowsList))
	
	dockerList, _ := op.GetResourceListByRoleId(superRole.ID, "docker")
	fmt.Printf("   Super 角色的 Docker 资源: %d 个\n", len(dockerList))
	
	databaseList, _ := op.GetResourceListByRoleId(superRole.ID, "database")
	fmt.Printf("   Super 角色的 Database 资源: %d 个\n", len(databaseList))
	
	routerList, _ := op.GetResourceListByRoleId(superRole.ID, "router")
	fmt.Printf("   Super 角色的 Router 资源: %d 个\n", len(routerList))
	
	switchList, _ := op.GetResourceListByRoleId(superRole.ID, "switch")
	fmt.Printf("   Super 角色的 Switch 资源: %d 个\n", len(switchList))
}

