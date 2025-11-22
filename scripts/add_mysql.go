package main

import (
	"fmt"
	"log"
	"flag"
	"errors"

	"binrc.com/roma/configs"
	"binrc.com/roma/core/constants"
	"binrc.com/roma/core/global"
	"binrc.com/roma/core/initialize"
	"binrc.com/roma/core/model"
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
	LoadConfig()
}

func main() {
	db, err := initialize.InitCDB()
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	global.CDB = db

	// MySQL 数据库配置
	dbResource := model.DatabaseConfig{
		DatabaseNick: "links-mysql",
		DatabaseType: "mysql",
		DatabaseName: "links",
		IPv4Pub:      "10.2.43.187",
		IPv4Priv:     "10.2.43.187",
		Port:         30298,
		Username:     "root",
		Password:     "GkCyITfV2ncn0cSrnw9rxA",
		Description:  "Links MySQL 数据库",
	}

	fmt.Println("正在添加 MySQL 数据库资源...")
	if err := db.Create(&dbResource).Error; err != nil {
		log.Fatal("创建失败:", err)
	}

	fmt.Printf("✓ 创建成功: %s (ID: %d)\n", dbResource.DatabaseNick, dbResource.ID)

	// 绑定到 super 角色
	var superRole model.Role
	db.Where("name = ?", "super").First(&superRole)
	
	if superRole.ID == 0 {
		log.Fatal("未找到 super 角色")
	}

	resourceRole := model.ResourceRole{
		ResourceID:   dbResource.ID,
		ResourceType: "database",
		RoleID:       int64(superRole.ID),
	}

	if err := db.Create(&resourceRole).Error; err != nil {
		log.Fatal("绑定角色失败:", err)
	}

	fmt.Printf("✓ 已绑定到 super 角色 (Role ID: %d)\n", superRole.ID)
	fmt.Println("\n现在可以使用了:")
	fmt.Println("  ssh super@localhost -p 2222")
	fmt.Println("  密码: 123456")
	fmt.Println("\n  在 TUI 中执行:")
	fmt.Println("  use database")
	fmt.Println("  ls")
	fmt.Println("  links-mysql")
}


