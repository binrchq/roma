package global

import (
	"os"

	"binrc.com/roma/configs"
	"github.com/redis/go-redis/v9"
	"github.com/shirou/gopsutil/process"
	"gorm.io/gorm"
)

var (
	CDB     *gorm.DB         // 重复创建db会有内存泄露
	RDB     *redis.Client    // 重复创建redis会有内存泄露
	CONFIG  *configs.Config  // 全局配置
	Process *process.Process //进程号

)

func init() {
	p, err := process.NewProcess(int32(os.Getpid()))
	if err != nil {
		panic(err)
	}
	Process = p
}

func GetDB() *gorm.DB {
	return CDB
}

func GetRDB() *redis.Client {
	return RDB
}
