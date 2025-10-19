package base

import (
	"github.com/FantasyRL/go-mcp-demo/config"
	"github.com/FantasyRL/go-mcp-demo/pkg/base/ai_provider"
	"github.com/FantasyRL/go-mcp-demo/pkg/base/mcp_client"
	"github.com/FantasyRL/go-mcp-demo/pkg/base/registry/consul"
	"github.com/FantasyRL/go-mcp-demo/pkg/constant"
	"log"
)

func WithMCPClient() Option {
	return func(clientSet *ClientSet) {
		switch {
		// stdio 启动
		case config.MCP.Transport == constant.MCPTransportStdio:
			mcpCli, err := mcp_client.NewMCPClient("")
			if err != nil {
				log.Fatalf("failed to create mcp client: %s", err)
			}
			clientSet.MCPCli = mcpCli
		// 单点通信
		case config.Registry.Provider == constant.RegistryProviderNone:
			mcpCli, err := mcp_client.NewMCPClient(config.MCP.HTTP.BaseURL)
			if err != nil {
				log.Fatalf("failed to create mcp client: %s", err)
			}
			clientSet.MCPCli = mcpCli
		// 服务发现模式
		// todo: 做成真正的服务发现，目前仍然是单点通信
		case config.Registry.Provider == constant.RegistryProviderConsul:
			resolver := consul.NewResolver()
			if resolver == nil {
				log.Fatalf("consul config invalid,can't create MCP client")
			}
			url, err := resolver.Resolve()
			if err != nil {
				log.Fatalf("failed to resolve mcp url from consul: %s", err)
			}
			mcpCli, err := mcp_client.NewMCPClient(url)
			if err != nil {
				log.Fatalf("failed to create mcp client: %s", err)
			}
			clientSet.RegistryResolver = resolver
			clientSet.MCPCli = mcpCli
		default:
			log.Fatalf("unknown registry provider: %s,can't create MCP client", config.Registry.Provider)
		}

	}
}

func WithAiProviderClient() Option {
	return func(clientSet *ClientSet) {
		cli := ai_provider.NewAiProviderClient()
		clientSet.AiProviderCli = cli
	}
}
