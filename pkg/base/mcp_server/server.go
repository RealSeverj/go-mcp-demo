package mcp_server

import (
	"context"
	"fmt"
	"github.com/FantasyRL/go-mcp-demo/pkg/constant"
	"github.com/FantasyRL/go-mcp-demo/pkg/logger"
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
	s.AddTool(mcp.NewTool("long_running_tool",
		mcp.WithDescription("A long running tool that reports progress"),
		mcp.WithNumber("duration",
			mcp.Description("Total duration of the operation in seconds"),
			mcp.Required(),
		),
		mcp.WithNumber("steps",
			mcp.Description("Number of steps to complete the operation"),
			mcp.Required(),
		),
	), handleLongRunningOperationTool)

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

// handleLongRunningOperationTool 示例长时间运行的工具，支持进度汇报
// https://github.com/mark3labs/mcp-go/blob/main/examples/everything/main.go 413
func handleLongRunningOperationTool(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	// 从请求中提取工具参数
	arguments := request.GetArguments()
	// 从请求元数据中提取进度标识符
	progressToken := request.Params.Meta.ProgressToken

	// 获取任务总持续时间和步骤数
	duration, _ := arguments["duration"].(float64) // 任务总持续时间（秒）
	steps, _ := arguments["steps"].(float64)       // 任务步骤数

	// 计算每一步的持续时间（每步的时长）
	stepDuration := duration / steps

	// 获取服务器上下文
	server := server.ServerFromContext(ctx)

	// 执行任务：模拟长时间操作，并在每一步发送进度通知
	for i := 1; i < int(steps)+1; i++ {
		// 每步执行完成后，等待相应的时间（模拟耗时操作）
		time.Sleep(time.Duration(stepDuration * float64(time.Second)))

		// 如果有进度令牌（progressToken），则发送进度通知
		if progressToken != nil {
			// 构造进度通知消息
			err := server.SendNotificationToClient(
				ctx,
				"notifications/progress", // 通知类型
				map[string]any{
					"progress":      i,                                                              // 当前进度
					"total":         int(steps),                                                     // 总步骤数
					"progressToken": progressToken,                                                  // 进度令牌，标识该操作
					"message":       fmt.Sprintf("Server progress %v%%", int(float64(i)*100/steps)), // 进度消息
				},
			)
			// 错误处理：如果通知发送失败，返回错误
			if err != nil {
				logger.Errorf("Failed to send progress notification: %v", err)
				return nil, fmt.Errorf("failed to send notification: %w", err)
			}
		}
	}
	time.Sleep(time.Second)

	// 返回工具执行的最终结果（任务完成）
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text", // 内容类型：文本
				Text: fmt.Sprintf(
					"Long running operation completed. Duration: %f seconds, Steps: %d.",
					duration,   // 任务总持续时间
					int(steps), // 总步骤数
				),
			},
		},
	}, nil
}
