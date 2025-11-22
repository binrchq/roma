package main

import (
	"fmt"
	"log"

	"binrc.com/roma/configs"
	"binrc.com/roma/core/constants"
	"binrc.com/roma/core/global"
	"binrc.com/roma/core/initialize"
	"binrc.com/roma/core/model"
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
	LoadConfig()
}

func main() {
	db, err := initialize.InitCDB()
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	global.CDB = db

	fmt.Println("📊 检查数据库中的资源和角色绑定关系...")
	fmt.Println("=" + string(make([]byte, 70)) + "=")

	// 查询所有 ResourceRole
	var resourceRoles []model.ResourceRole
	db.Find(&resourceRoles)

	fmt.Printf("\n✅ 找到 %d 条资源-角色绑定记录:\n\n", len(resourceRoles))

	// 按资源类型分组
	typeMap := make(map[string][]model.ResourceRole)
	for _, rr := range resourceRoles {
		typeMap[rr.ResourceType] = append(typeMap[rr.ResourceType], rr)
	}

	for resType, roles := range typeMap {
		fmt.Printf("📦 %s: %d 个资源\n", resType, len(roles))
		for _, rr := range roles {
			fmt.Printf("   - ResourceID=%d, RoleID=%d\n", rr.ResourceID, rr.RoleID)
		}
	}

	fmt.Println("\n" + string(make([]byte, 72)) + "\n")

	// 检查每种资源类型的数量
	var count int64
	
	db.Model(&model.LinuxConfig{}).Count(&count)
	fmt.Printf("Linux 资源总数: %d\n", count)
	
	db.Model(&model.WindowsConfig{}).Count(&count)
	fmt.Printf("Windows 资源总数: %d\n", count)
	
	db.Model(&model.DockerConfig{}).Count(&count)
	fmt.Printf("Docker 资源总数: %d\n", count)
	
	db.Model(&model.DatabaseConfig{}).Count(&count)
	fmt.Printf("Database 资源总数: %d\n", count)
	
	db.Model(&model.RouterConfig{}).Count(&count)
	fmt.Printf("Router 资源总数: %d\n", count)
	
	db.Model(&model.SwitchConfig{}).Count(&count)
	fmt.Printf("Switch 资源总数: %d\n", count)
}


