package mcp_server

import (
	"context"
	"github.com/FantasyRL/go-mcp-demo/pkg/constant"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// HTTPOpts：Streamable HTTP(含 SSE) 选项
type HTTPOpts struct {
	// EndpointPath 仅对 shttp.Start(":8080") 的一行启动生效；
	// 若作为 http.Handler 挂到 mux，路由由 mux 决定，该字段不生效。
	EndpointPath      string
	HeartbeatInterval time.Duration // 建议 20~30s，降低中间件 idle 断开
}

// NewCoreServer 在此注册 tools/prompts/resources
func NewCoreServer(name, version string) *server.MCPServer {
	s := server.NewMCPServer(
		name,
		version,
		server.WithRecovery(),
		server.WithToolCapabilities(false),
	)

	// 示例工具：time_now
	tool := mcp.NewTool("time_now", mcp.WithDescription("返回当前时间（RFC3339）"))
	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		now := time.Now().Format(time.RFC3339)
		return mcp.NewToolResultText(now), nil
	})

	return s
}

// NewStreamableHTTPServer 基于核心 Server 创建 HTTP/SSE 服务器组件
func NewStreamableHTTPServer(core *server.MCPServer) *server.StreamableHTTPServer {
	var httpOpts []server.StreamableHTTPOption
	httpOpts = append(httpOpts, server.WithHeartbeatInterval(constant.MCPServerHeartbeatInterval))
	return server.NewStreamableHTTPServer(core, httpOpts...)
}

// ServeStdio stdio
func ServeStdio(core *server.MCPServer) error {
	return server.ServeStdio(core)
}
