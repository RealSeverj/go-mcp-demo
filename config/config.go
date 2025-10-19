package config

import (
	"github.com/spf13/viper"
	"github.com/west2-online/fzuhelper-server/pkg/constants"
	"log"
)

var (
	AiProvider   *AiProviderConfig
	CLI          *cliConfig
	MCP          *mcpConfig
	Server       *server
	Registry     *registryConfig
	Service      *service
	runtimeViper = viper.New()
)

func Load(path string, srv string) {
	runtimeViper.SetConfigFile(path)
	runtimeViper.SetConfigType("yaml")

	if err := runtimeViper.ReadInConfig(); err != nil {
		log.Fatal(err)
		return
	}
	cfg := new(Config)
	if err := runtimeViper.Unmarshal(&cfg); err != nil {
		log.Fatal(err)
		return
	}

	AiProvider = &cfg.AiProvider
	CLI = &cfg.CLI
	MCP = &cfg.MCP
	Server = &cfg.Server
	Registry = &cfg.Registry
	Service = getService(srv)
}

// GetLoggerLevel 会返回服务的日志等级
func GetLoggerLevel() string {
	if Server == nil {
		return constants.DefaultLogLevel
	}
	return Server.LogLevel
}

func getService(name string) *service {
	addrList := runtimeViper.GetStringSlice("services." + name + ".addr")

	return &service{
		Name:     runtimeViper.GetString("services." + name + ".name"),
		AddrList: addrList,
		LB:       runtimeViper.GetBool("services." + name + ".load-balance"),
	}
}
