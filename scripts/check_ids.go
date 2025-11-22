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

	fmt.Println("🔍 检查各资源类型的 ID...")

	var databases []model.DatabaseConfig
	db.Where("database_nick LIKE 'test-%'").Find(&databases)
	fmt.Printf("\n📦 Database 资源:\n")
	for _, d := range databases {
		fmt.Printf("   ID=%d, Name=%s\n", d.ID, d.DatabaseNick)
	}

	var routers []model.RouterConfig
	db.Where("router_name LIKE 'test-%'").Find(&routers)
	fmt.Printf("\n📦 Router 资源:\n")
	for _, r := range routers {
		fmt.Printf("   ID=%d, Name=%s\n", r.ID, r.RouterName)
	}

	var switches []model.SwitchConfig
	db.Where("switch_name LIKE 'test-%'").Find(&switches)
	fmt.Printf("\n📦 Switch 资源:\n")
	for _, s := range switches {
		fmt.Printf("   ID=%d, Name=%s\n", s.ID, s.SwitchName)
	}

	fmt.Println("\n现在手动添加缺失的 ResourceRole 绑定...")
	
	for _, d := range databases {
		rr := model.ResourceRole{
			ResourceID:   d.ID,
			ResourceType: "database",
			RoleID:       1,
		}
		if err := db.Create(&rr).Error; err != nil {
			fmt.Printf("   ❌ Database ID=%d 绑定失败: %v\n", d.ID, err)
		} else {
			fmt.Printf("   ✅ Database ID=%d 绑定成功\n", d.ID)
		}
	}

	for _, r := range routers {
		rr := model.ResourceRole{
			ResourceID:   r.ID,
			ResourceType: "router",
			RoleID:       1,
		}
		if err := db.Create(&rr).Error; err != nil {
			fmt.Printf("   ❌ Router ID=%d 绑定失败: %v\n", r.ID, err)
		} else {
			fmt.Printf("   ✅ Router ID=%d 绑定成功\n", r.ID)
		}
	}

	for _, s := range switches {
		rr := model.ResourceRole{
			ResourceID:   s.ID,
			ResourceType: "switch",
			RoleID:       1,
		}
		if err := db.Create(&rr).Error; err != nil {
			fmt.Printf("   ❌ Switch ID=%d 绑定失败: %v\n", s.ID, err)
		} else {
			fmt.Printf("   ✅ Switch ID=%d 绑定成功\n", s.ID)
		}
	}

	fmt.Println("\n✅ 修复完成！现在重启 ROMA 服务即可看到所有资源")
}


