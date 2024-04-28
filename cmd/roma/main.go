package main

import (
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"bitrec.ai/roma/configs"
	"bitrec.ai/roma/core/constants"
	"bitrec.ai/roma/core/global"
	"bitrec.ai/roma/core/initialize"
	"bitrec.ai/roma/core/operation"
	"bitrec.ai/roma/core/pkg/i18n"
	"bitrec.ai/roma/core/routers"
	"bitrec.ai/roma/core/services"
	"bitrec.ai/roma/core/sshd"
	"bitrec.ai/roma/core/utils/logger"

	"github.com/brckubo/ssh"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

var (
	cfgFile string
)

func init() {
	flag.StringVar(&cfgFile, "c", constants.BASE_DIR+"/configs/config.toml", "path of config file.")
	flag.Parse()
	// 加载配置文件
	LoadConfig()
	// 加载数据库
	LoadDatabase()
	// 加载i18n
	LoadI18n()
	// 初始化数据
	services.InitData()
}

func main() {
	go func() {
		go StartApiService()
		go StartSshdService()
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		log.Printf("roma get a signal %s\n", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGSTOP, syscall.SIGINT:
			log.Println("roma exit")
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}

func StartApiService() {
	if global.CONFIG.Api.GinMode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := routers.SetupRouter()
	pprof.Register(r)

	log.Printf("starting api server on port %s...\n", global.CONFIG.Api.Port)
	err := r.Run(global.CONFIG.Api.Host + ":" + global.CONFIG.Api.Port)
	if err != nil {
		panic(err)
	}

}

func StartSshdService() {
	//如果主机密钥为空
	op := operation.NewHostKeyOperation()
	if !op.HostKeyIsExist() {
		privateKeyBase64, publicKeyBase64, err := sshd.GenKey()
		if err != nil {
			logger.Logger.Panic(err)
		}
		op.SaveHostKey(privateKeyBase64, publicKeyBase64)
	}

	ssh.Handle(func(sess ssh.Session) {
		defer func() {
			if e, ok := recover().(error); ok {
				logger.Logger.Panic(e)
			}
		}()
		services.SessionHandler(&sess)
	})
	log.Printf("starting ssh server on port %s...\n", global.CONFIG.Common.Port)
	hostKey, err := op.GetLatestHostKey()
	if err != nil {
		logger.Logger.Panic("Get host key error:", err)
	}
	privateKeyBytes, err := base64.StdEncoding.DecodeString(string(hostKey.PrivateKey))
	if err != nil {
		log.Fatalf("Failed to decode base64 encoded private key: %s", err)
	}
	log.Fatal(ssh.ListenAndServe(
		fmt.Sprintf(":%s", global.CONFIG.Common.Port),
		nil,
		// ssh.PasswordAuth(services.PasswordAuth),
		ssh.PublicKeyAuth(services.PublicKeyAuth),
		ssh.HostKeyPEM(privateKeyBytes),
	),
	)
}

func LoadI18n() {
	i18n.LoadTranslations()
}

func LoadDatabase() {
	cdb, err := initialize.InitCDB()
	if err != nil {
		log.Fatal(err)
	}
	global.CDB = cdb
	// global.RDB = initialize.InitRDB()
}

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
	// if err := checkCfg(conf); err != nil {
	// 	return err
	// }
	global.CONFIG = conf
	// 热加载
	// v.OnConfigChange(func(e fsnotify.Event) {
	// 	fmt.Println("config file changed:", e.Name)
	// 	if err := v.Unmarshal(&conf); err != nil {
	// 		log.Println(err)
	// 	}
	// 	if err := checkCfg(conf); err != nil {
	// 		log.Println(err)
	// 	} else {
	// 		global.CONFIG = conf
	// 	}
	// })
	// v.WatchConfig()
	return nil
}

// func checkCfg(c *configs.Config) error {
// 	if c == nil {
// 		return errors.New("config nil")
// 	}
// 	if c.Common == nil {
// 		return errors.New("common section not config")
// 	}
// 	if c.Common.SshdListenPort == "" {
// 		return errors.New("common listen not set")
// 	}

// 	return nil
// }
