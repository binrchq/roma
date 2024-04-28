package initialize

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"bitrec.ai/roma/core/global"
	"bitrec.ai/roma/core/model"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var cdb *gorm.DB

func InitCDB() (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	// Determine the database type based on the URL
	if strings.Contains(global.CONFIG.Database.CdbUrl, ".db") {
		db, err = initSQLite()
	} else {
		db, err = initMySQL()
	}

	if err != nil {
		return nil, err
	}

	if err := migrateTables(db, &model.HostKey{}, &model.User{}, &model.Passport{}, &model.Role{}, &model.Apikey{}, &model.LinuxConfig{}, &model.WindowsConfig{}, &model.DatabaseConfig{}, &model.RouterConfig{}, &model.SwitchConfig{}, &model.ResourceRole{}, &model.Tag{}, &model.CredentialAccessLog{}, &model.AccessLog{}); err != nil {
		return nil, err
	}

	cdb = db
	return cdb, nil
}

func initSQLite() (*gorm.DB, error) {
	if _, err := os.Stat(global.CONFIG.Database.CdbUrl); os.IsNotExist(err) {
		directory := filepath.Dir(global.CONFIG.Database.CdbUrl)
		if err := os.MkdirAll(directory, 0755); err != nil {
			// 创建目录失败，输出错误信息并退出程序
			fmt.Printf("无法创建目录：%s\n", err)
			os.Exit(1)
		}
		if _, err := os.Create(global.CONFIG.Database.CdbUrl); err != nil {
			return nil, fmt.Errorf("failed to create SQLite database: %s", err)
		}
	}
	db, err := gorm.Open(sqlite.Open(global.CONFIG.Database.CdbUrl), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite database: %s", err)
	}
	return db, nil
}

func initMySQL() (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(global.CONFIG.Database.CdbUrl), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open MySQL database: %s", err)
	}
	return db, nil
}

func getTableName(model interface{}) string {
	// 通过反射获取结构体的类型
	t := reflect.TypeOf(model)
	// 如果是指针类型，取其指向的类型
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	// 返回结构体类型的名称作为表名
	return t.Name()
}

func migrateTables(db *gorm.DB, models ...interface{}) error {
	for _, model := range models {
		tableName := getTableName(model)
		if err := db.AutoMigrate(model); err != nil {
			return fmt.Errorf("failed to migrate table[%s]: %s", tableName, err)
		}
	}
	return nil
}
