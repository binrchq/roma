package constants

const (
	AccessLogActionLogin       = "login"  // 登录
	AccessLogActionLogout      = "logout" // 登出
	AccessLogActionCreate      = "create" // 创建
	AccessLogActionUpdate      = "update" // 更新
	AccessLogActionDelete      = "delete" // 删除
	AccessLogActionOther       = "other"  // 其他
	AccessLogActionLevelInfo   = "info"   // 信息
	AccessLogActionLevelWarn   = "warn"   // 警告
	AccessLogActionLevelError  = "error"  // 错误
	AccessLogActionLevelFatal  = "fatal"  // 致命
	AccessLogActionLevelDebug  = "debug"  // 调试
	AccessLogActionLevelTrace  = "trace"  // 跟踪
	AccessLogActionLevelCrit   = "crit"   // 严重
	AccessLogActionLevelAlert  = "alert"  // 警报
	AccessLogActionLevelEmerg  = "emerg"  // 紧急
	AccessLogActionLevelNotice = "notice" // 注意
	AccessLogActionLevelAll    = "all"    // 所有
	AccessLogSourceWeb         = "web"    // Web
	AccessLogSourceApi         = "api"    // API
	AccessLogSourceCli         = "cli"    // CLI
)
