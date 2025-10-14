package mcp_client

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/FantasyRL/go-mcp-demo/config"
	mcpc "github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

type MCPClient struct {
	Client *mcpc.Client
	Tools  []mcp.Tool
}

// NewMCPClient 启动 MCP Server 并建立连接
func NewMCPClient(url string) (*MCPClient, error) {
	switch config.MCP.Transport {
	case "stdio", "":
		return newStdioMCPClient()
	case "sse":
		return newSSEMCPClientWithConn(url)
	case "http":
		return newHTTPMCPClientWithConn(url)
	default:
		return nil, fmt.Errorf("unknown MCP transport: %s", config.MCP.Transport)
	}
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
