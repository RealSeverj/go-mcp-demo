package client

import (
	"github.com/FantasyRL/go-mcp-demo/pkg/client/mcp_client"
	"github.com/FantasyRL/go-mcp-demo/pkg/client/ollama"
	"github.com/FantasyRL/go-mcp-demo/pkg/logger"
)

func WithMCPClient() Option {
	return func(clientSet *ClientSet) {
		mcpCli, err := mcp_client.NewMCPClient()
		if err != nil {
			logger.Fatalf("failed to create mcp client: %s", err)
		}
		clientSet.MCPCli = mcpCli
	}
}

func WithOllamaClient() Option {
	return func(clientSet *ClientSet) {
		ollamaCli := ollama.NewOllamaClient()
		clientSet.OllamaCli = ollamaCli
	}
}
