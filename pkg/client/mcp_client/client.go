package mcp_client

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/FantasyRL/go-mcp-demo/pkg/logger"
	"time"

	mcpc "github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

type MCPClient struct {
	Client *mcpc.Client
	Tools  []mcp.Tool
}

// NewMCPClient 启动 MCP Server 并建立 stdio 连接
func NewMCPClient() (*MCPClient, error) {
	cmd := "./bin/mcp-server"

	client, err := mcpc.NewStdioMCPClient(cmd, nil)
	if err != nil {
		return nil, fmt.Errorf("start stdio client: %w", err)
	}

	timeout := 10 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// 初始化
	_, err = client.Initialize(ctx, mcp.InitializeRequest{
		Params: mcp.InitializeParams{
			ClientInfo: mcp.Implementation{
				Name:    "mcp-host",
				Version: "0.1.0",
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("initialize mcp: %w", err)
	}

	// 获取工具列表
	res, err := client.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		return nil, fmt.Errorf("list tools: %w", err)
	}

	logger.Infof("[mcp-client] discovered %d tools\n", len(res.Tools))
	return &MCPClient{Client: client, Tools: res.Tools}, nil
}

// ConvertToolsToOllama 转换 MCP 工具定义到 Ollama 工具格式
func (m *MCPClient) ConvertToolsToOllama() []map[string]any {
	var out []map[string]any
	for _, t := range m.Tools {
		var params map[string]any
		b, _ := json.Marshal(t.InputSchema)
		_ = json.Unmarshal(b, &params)

		out = append(out, map[string]any{
			"type": "function",
			"function": map[string]any{
				"name":        t.Name,
				"description": t.Description,
				"parameters":  params,
			},
		})
	}
	return out
}

// CallTool 调用 MCP 工具
func (m *MCPClient) CallTool(ctx context.Context, name string, args any) (string, error) {
	timeout := 30 * time.Second
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	res, err := m.Client.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      name,
			Arguments: args,
		},
	})
	if err != nil {
		return "", fmt.Errorf("call tool %s: %w", name, err)
	}

	// 提取返回文本
	var text string
	for _, c := range res.Content {
		if tc, ok := mcp.AsTextContent(c); ok {
			text += tc.Text + "\n"
		}
	}
	if text == "" && res.StructuredContent != nil {
		b, _ := json.Marshal(res.StructuredContent)
		text = string(b)
	}
	if text == "" {
		text = "(no content)"
	}
	return text, nil
}

// Close 关闭连接
func (m *MCPClient) Close() {
	if m.Client != nil {
		_ = m.Client.Close()
	}
}
