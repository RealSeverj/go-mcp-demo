package host

import (
	"context"
	"github.com/FantasyRL/go-mcp-demo/config"
	"github.com/FantasyRL/go-mcp-demo/pkg/base"
	"github.com/FantasyRL/go-mcp-demo/pkg/base/mcp_client"
	"github.com/FantasyRL/go-mcp-demo/pkg/base/ollama"
	"github.com/FantasyRL/go-mcp-demo/pkg/logger"
)

// 简单的内存存储用户对话历史
var history = make(map[int64][]ollama.Message)

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

func (h *Host) Chat(id int64, msg string) (string, error) {
	// 获取当前用户的对话历史（如果没有则初始化为空切片）
	userHistory := history[id]
	if userHistory == nil {
		userHistory = []ollama.Message{}
	}

	// 将当前用户消息加入历史
	userHistory = append(userHistory, ollama.Message{Role: "user", Content: msg})

	// 转换工具定义
	ollamaTools := h.mcpCli.ConvertToolsToOllama()
	ollamaOptions := ollama.BuildOptions()

	// 第一次调用模型（带历史）
	resp, err := h.ollamaCli.Chat(h.ctx, ollama.ChatRequest{
		Model:     config.Ollama.Model,
		Messages:  userHistory, // 使用完整历史
		Options:   ollamaOptions,
		Tools:     ollamaTools,
		KeepAlive: config.Ollama.Options.KeepAlive,
	})
	if err != nil {
		return "", err
	}

	// 更新历史：添加模型回复
	userHistory = append(userHistory, ollama.Message{Role: "assistant", Content: resp.Message.Content})

	// 如果有工具调用
	if len(resp.Message.ToolCalls) > 0 {
		for _, c := range resp.Message.ToolCalls {
			args, err := ollama.ParseToolArguments(c.Function.Arguments)
			if err != nil {
				args = map[string]any{"_error": err.Error()}
			}

			out, err := h.mcpCli.CallTool(context.Background(), c.Function.Name, args)
			if err != nil {
				out = "tool error: " + err.Error()
			}

			// 添加工具执行结果到历史
			userHistory = append(userHistory, ollama.Message{
				Role:     "tool",
				ToolName: c.Function.Name,
				Content:  out,
			})
			logger.Infof("[tool] %s executed\n", c.Function.Name)
		}

		// 再次调用模型，传入完整历史（包含工具返回）
		resp2, err := h.ollamaCli.Chat(h.ctx, ollama.ChatRequest{
			Model:    config.Ollama.Model,
			Messages: userHistory, // 包含工具返回的新历史
			Options:  ollamaOptions,
			Tools:    ollamaTools,
		})
		if err != nil {
			return "", err
		}

		// 更新历史：添加最终模型回复
		userHistory = append(userHistory, ollama.Message{Role: "assistant", Content: resp2.Message.Content})

		// 保存回 map
		history[id] = userHistory

		return resp2.Message.Content, nil
	}

	// 无工具调用，直接返回模型回复
	history[id] = userHistory // 保存更新后的历史
	return resp.Message.Content, nil
}
