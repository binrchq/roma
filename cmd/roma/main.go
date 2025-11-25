package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"binrc.com/roma/configs"
	"binrc.com/roma/core/constants"
	"binrc.com/roma/core/global"
	"binrc.com/roma/core/initialize"
	"binrc.com/roma/core/operation"
	"binrc.com/roma/core/pkg/i18n"
	"binrc.com/roma/core/routers"
	"binrc.com/roma/core/services"
	"binrc.com/roma/core/sshd"
	"binrc.com/roma/core/utils/logger"

	"github.com/fatih/color"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/loganchef/ssh"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	rootCmd = &cobra.Command{
		Use:   "roma",
		Short: "Roma - 远程运维管理工具",
		Long:  "Roma 是一个功能强大的远程运维管理工具",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			initConfig()
			bindFlags(cmd)
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := &configs.Config{}
			if err := viper.Unmarshal(cfg); err != nil {
				return fmt.Errorf("failed to unmarshal config: %w", err)
			}

			global.CONFIG = cfg

			// 强制启用颜色输出（在 Docker 容器中也需要颜色）
			// 检查环境变量，如果明确设置了 NO_COLOR，则禁用颜色
			if os.Getenv("NO_COLOR") == "" {
				color.NoColor = false
			}

			LoadDatabase()
			LoadI18n()
			services.InitData()

			startServices()
			return nil
		},
	}
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", constants.BASE_DIR+"/configs/config.toml", "配置文件路径")

	// API 配置
	rootCmd.PersistentFlags().String("api-host", "", "API 服务主机地址")
	rootCmd.PersistentFlags().String("api-port", "", "API 服务端口")
	rootCmd.PersistentFlags().String("api-gin-mode", "", "Gin 运行模式 (debug, release, test)")

	// Common 配置
	rootCmd.PersistentFlags().String("common-port", "", "SSH 服务端口")
	rootCmd.PersistentFlags().String("common-language", "", "语言设置 (zh, en, ru)")
	rootCmd.PersistentFlags().String("common-prompt", "", "SSH 提示符")
	rootCmd.PersistentFlags().String("common-history-tmp-dir", "", "历史文件存储目录")
	rootCmd.PersistentFlags().Int("common-history-tmp-max-line", 0, "最大历史记录行数")
	rootCmd.PersistentFlags().Int("common-history-tmp-max-size", 0, "最大历史文件大小（字节）")

	// Database 配置
	rootCmd.PersistentFlags().String("database-cdb-url", "", "数据库 CDB URL")
	rootCmd.PersistentFlags().String("database-rdb-url", "", "数据库 RDB URL")
	rootCmd.PersistentFlags().String("database-rdb-passwd", "", "数据库 RDB 密码")

	// Log 配置
	rootCmd.PersistentFlags().String("log-level", "", "日志级别 (debug, info, warn, error)")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
		if err := viper.ReadInConfig(); err != nil {
			log.Printf("Warning: failed to read config file %s: %v\n", cfgFile, err)
		}
	}

	viper.SetEnvPrefix("ROMA")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	viper.AutomaticEnv()

	bindEnvVars()
}

func bindFlags(cmd *cobra.Command) {
	viper.BindPFlag("config", cmd.PersistentFlags().Lookup("config"))

	// API 配置
	viper.BindPFlag("api.host", cmd.PersistentFlags().Lookup("api-host"))
	viper.BindPFlag("api.port", cmd.PersistentFlags().Lookup("api-port"))
	viper.BindPFlag("api.gin_mode", cmd.PersistentFlags().Lookup("api-gin-mode"))

	// Common 配置
	viper.BindPFlag("common.port", cmd.PersistentFlags().Lookup("common-port"))
	viper.BindPFlag("common.language", cmd.PersistentFlags().Lookup("common-language"))
	viper.BindPFlag("common.prompt", cmd.PersistentFlags().Lookup("common-prompt"))
	viper.BindPFlag("common.history_tmp_dir", cmd.PersistentFlags().Lookup("common-history-tmp-dir"))
	viper.BindPFlag("common.history_tmp_max_line", cmd.PersistentFlags().Lookup("common-history-tmp-max-line"))
	viper.BindPFlag("common.history_tmp_max_size", cmd.PersistentFlags().Lookup("common-history-tmp-max-size"))

	// Database 配置
	viper.BindPFlag("database.cdb_url", cmd.PersistentFlags().Lookup("database-cdb-url"))
	viper.BindPFlag("database.rdb_url", cmd.PersistentFlags().Lookup("database-rdb-url"))
	viper.BindPFlag("database.rdb_passwd", cmd.PersistentFlags().Lookup("database-rdb-passwd"))

	// Log 配置
	viper.BindPFlag("log.level", cmd.PersistentFlags().Lookup("log-level"))
}

func bindEnvVars() {
	// API 配置
	viper.BindEnv("api.gin_mode", "ROMA_API_GIN_MODE")
	viper.BindEnv("api.host", "ROMA_API_HOST")
	viper.BindEnv("api.port", "ROMA_API_PORT")

	// Common 配置
	viper.BindEnv("common.history_tmp_dir", "ROMA_COMMON_HISTORY_TMP_DIR")
	viper.BindEnv("common.history_tmp_max_line", "ROMA_COMMON_HISTORY_TMP_MAX_LINE")
	viper.BindEnv("common.history_tmp_max_size", "ROMA_COMMON_HISTORY_TMP_MAX_SIZE")
	viper.BindEnv("common.language", "ROMA_COMMON_LANGUAGE")
	viper.BindEnv("common.port", "ROMA_COMMON_PORT")
	viper.BindEnv("common.prompt", "ROMA_COMMON_PROMPT")

	// Database 配置
	viper.BindEnv("database.cdb_url", "ROMA_DATABASE_CDB_URL")
	viper.BindEnv("database.rdb_passwd", "ROMA_DATABASE_RDB_PASSWD")
	viper.BindEnv("database.rdb_url", "ROMA_DATABASE_RDB_URL")

	// Log 配置
	viper.BindEnv("log.level", "ROMA_LOG_LEVEL")

	// ApiKey 配置
	viper.BindEnv("apikey.prefix", "ROMA_APIKEY_PREFIX")
	viper.BindEnv("apikey.key", "ROMA_APIKEY_KEY")

	// User1st 配置
	viper.BindEnv("user_1st.email", "ROMA_USER_1ST_EMAIL")
	viper.BindEnv("user_1st.name", "ROMA_USER_1ST_NAME")
	viper.BindEnv("user_1st.nickname", "ROMA_USER_1ST_NICKNAME")
	viper.BindEnv("user_1st.password", "ROMA_USER_1ST_PASSWORD")
	viper.BindEnv("user_1st.public_key", "ROMA_USER_1ST_PUBLIC_KEY")
	viper.BindEnv("user_1st.username", "ROMA_USER_1ST_USERNAME")
	viper.BindEnv("user_1st.roles", "ROMA_USER_1ST_ROLES")

	// ControlPassport 配置
	viper.BindEnv("control_passport.service_user", "ROMA_CONTROL_PASSPORT_SERVICE_USER")
	viper.BindEnv("control_passport.password", "ROMA_CONTROL_PASSPORT_PASSWORD")
	viper.BindEnv("control_passport.resource_type", "ROMA_CONTROL_PASSPORT_RESOURCE_TYPE")
	viper.BindEnv("control_passport.passport_pub", "ROMA_CONTROL_PASSPORT_PASSPORT_PUB")
	viper.BindEnv("control_passport.passport", "ROMA_CONTROL_PASSPORT_PASSPORT")
	viper.BindEnv("control_passport.description", "ROMA_CONTROL_PASSPORT_DESCRIPTION")

	// Banner 配置
	viper.BindEnv("banner.show", "ROMA_BANNER_SHOW")
	viper.BindEnv("banner.banner", "ROMA_BANNER_BANNER")

	// Title 配置
	viper.BindEnv("title", "ROMA_TITLE")
}

func startServices() {
	go func() {
		go StartApiService()
		go StartSshdService()
		// MCP 服务器应该独立运行，由 AI 工具按需启动
		// 不需要嵌入到 roma 主程序中
		// go StartMCPService()
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
	// GenKey() 返回的是原始 PEM 格式的私钥，不需要 Base64 解码
	// 直接使用 hostKey.PrivateKey 即可
	privateKeyBytes := hostKey.PrivateKey
	if len(privateKeyBytes) == 0 {
		log.Fatalf("Host key private key is empty")
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
