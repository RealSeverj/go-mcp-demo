package config

import (
	"github.com/spf13/viper"
)

var Cfg Config

func Load(path string) error {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")

	// 默认值
	setDefaults(v)

	if err := v.ReadInConfig(); err != nil {
		return err
	}
	if err := v.Unmarshal(&Cfg); err != nil {
		return err
	}

	return nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("ollama.base_url", "http://127.0.0.1:11434")
	v.SetDefault("ollama.model", "qwen3:4b")
	v.SetDefault("ollama.options.keep_alive", "5m")
	v.SetDefault("ollama.options.request_timeout", "60s")
	v.SetDefault("cli.history", true)
	v.SetDefault("cli.max_turns", 32)
	v.SetDefault("mcp.server_name", "legend.mcp.demo")
	v.SetDefault("mcp.transport", "stdio")
	v.SetDefault("mcp.http_tool.timeout", "10s")
}
