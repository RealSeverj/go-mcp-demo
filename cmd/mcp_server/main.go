package main

import (
	"flag"
	"github.com/FantasyRL/go-mcp-demo/config"
	"github.com/FantasyRL/go-mcp-demo/pkg/base/mcp_server"
	"github.com/FantasyRL/go-mcp-demo/pkg/constant"
	"github.com/FantasyRL/go-mcp-demo/pkg/logger"
	"github.com/FantasyRL/go-mcp-demo/pkg/utils"
)

var (
	serviceName = "mcp_server"
	configPath  = flag.String("cfg", "config/config.yaml", "config file path")
)

func init() {
	flag.Parse()
	config.Load(*configPath, serviceName)
	logger.Init(serviceName, config.GetLoggerLevel())
}

func main() {
	coreServer := mcp_server.NewCoreServer(config.MCP.ServerName, config.MCP.Transport)
	switch config.MCP.Transport {
	case constant.MCPTransportStdio:
		if err := mcp_server.ServeStdio(coreServer); err != nil {
			logger.Errorf("serve stdio: %v", err)
			return
		}
	case constant.MCPTransportHTTP:
		addr, err := utils.GetAvailablePort()
		if err != nil {
			logger.Errorf("mcp_server: get available port failed, err: %v", err)
			return
		}
		if err := mcp_server.NewStreamableHTTPServer(coreServer).Start(addr); err != nil {
			logger.Errorf("serve http: %v", err)
			return
		}
	case constant.MCPTransportSSE:
		addr, err := utils.GetAvailablePort()
		if err != nil {
			logger.Errorf("mcp_server: get available port failed, err: %v", err)
			return
		}
		// http与http-sse统一都可以用这个来init
		if err := mcp_server.NewStreamableHTTPServer(coreServer).Start(addr); err != nil {
			logger.Errorf("serve http: %v", err)
			return
		}
	}
}
