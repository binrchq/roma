package configs

import (
	"fmt"

	"github.com/spf13/viper"
)

// LoadConfig 从 viper 加载配置
// 注意：此函数已废弃，请直接使用 viper.Unmarshal
// 保留此函数以保持向后兼容
func LoadConfig(cfgPath string) (*Config, error) {
	if cfgPath == "" {
		return nil, fmt.Errorf("config file path is required")
	}
	v := viper.New()
	v.SetConfigFile(cfgPath)
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("fatal error config file: %w", err)
	}
	conf := NewConfig()
	if err := v.Unmarshal(&conf); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	return conf, nil
}
