package mcp_server

import (
	"github.com/FantasyRL/go-mcp-demo/pkg/base/tool_set"
	"github.com/FantasyRL/go-mcp-demo/pkg/constant"
	"time"

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
func NewCoreServer(name, version string, toolSet *tool_set.ToolSet) *server.MCPServer {
	s := server.NewMCPServer(
		name,
		version,
		server.WithRecovery(),
		server.WithToolCapabilities(false),
	)

	for _, t := range toolSet.Tools {
		s.AddTool(*t, toolSet.HandlerFunc[t.Name])
	}

	return s
}

// NewStreamableHTTPServer 基于核心 Server 创建StreamableHTTP服务器组件
func NewStreamableHTTPServer(core *server.MCPServer) *server.StreamableHTTPServer {
	var httpOpts []server.StreamableHTTPOption
	httpOpts = append(httpOpts, server.WithHeartbeatInterval(constant.MCPServerHeartbeatInterval))
	return server.NewStreamableHTTPServer(core, httpOpts...)
}

// ServeStdio stdio
func ServeStdio(core *server.MCPServer) error {
	return server.ServeStdio(core)
}

// NewHTTPSSEServer [MCP规范已废弃]基于核心 Server 创建 SSE 服务器组件
func NewHTTPSSEServer(core *server.MCPServer) *server.SSEServer {
	var sseOpts []server.SSEOption
	sseOpts = append(sseOpts, server.WithKeepAliveInterval(constant.MCPServerHeartbeatInterval))
	return server.NewSSEServer(core, sseOpts...)
}
