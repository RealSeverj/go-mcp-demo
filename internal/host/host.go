package host

import (
	"context"
	"github.com/FantasyRL/go-mcp-demo/pkg/base"
	"github.com/FantasyRL/go-mcp-demo/pkg/base/mcp_client"
	"github.com/FantasyRL/go-mcp-demo/pkg/base/ollama"
	"github.com/openai/openai-go/v2"
)

// 简单的内存存储用户对话历史
var history = make(map[int64][]ollama.Message)
var historyOpenAI = make(map[int64][]openai.ChatCompletionMessageParamUnion)

type Host struct {
	ctx       context.Context
	mcpCli    *mcp_client.MCPClient
	ollamaCli *ollama.Client
}

func NewHost(ctx context.Context, clientSet *base.ClientSet) *Host {
	return &Host{
		ctx:       ctx,
		mcpCli:    clientSet.MCPCli,
		ollamaCli: clientSet.OllamaCli,
	}
}
