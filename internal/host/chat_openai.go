package host

import (
	"context"
	"encoding/json"
	openai "github.com/openai/openai-go/v2"

	"github.com/FantasyRL/go-mcp-demo/config"
	"github.com/FantasyRL/go-mcp-demo/pkg/constant"
	"github.com/FantasyRL/go-mcp-demo/pkg/errno"
	"github.com/FantasyRL/go-mcp-demo/pkg/logger"
)

// 将 OpenAI 的 tool_calls[].function.arguments (string) 解成 map[string]any（与原逻辑一致）
func parseOpenAIToolArgs(argStr string) map[string]any {
	if argStr == "" {
		return map[string]any{}
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(argStr), &m); err == nil {
		return m
	}
	// 如果不是 JSON，就当成纯字符串包裹
	return map[string]any{"_": argStr}
}

// StreamChatOpenAI OpenAI 格式的流式聊天
func (h *Host) StreamChatOpenAI(
	ctx context.Context,
	id int64,
	userMsg string,
	emit func(event string, v any) error, // SSE: 事件名 + 任意 JSON
) error {
	// 历史
	hist := historyOpenAI[id]
	if hist == nil {
		hist = []openai.ChatCompletionMessageParamUnion{}
	}
	// 用户消息
	hist = append(hist, openai.UserMessage(userMsg))

	// 工具
	tools := h.mcpCli.ConvertToolsToOpenAI()

	var assistantBuf string
	// 用 Accumulator 汇总完整 tool_calls
	var acc openai.ChatCompletionAccumulator
	var toolCallsFinished bool

	err := h.ollamaCli.ChatStreamOpenAI(ctx, openai.ChatCompletionNewParams{
		Model:    openai.ChatModel(config.Ollama.Model),
		Messages: hist,
		Tools:    tools,
		// 其他参数可按需补充
	}, func(chunk *openai.ChatCompletionChunk) error {
		// 累加到 acc（用于拿完整的 tool_calls）
		acc.AddChunk(*chunk)

		if len(chunk.Choices) > 0 {
			if s := chunk.Choices[0].Delta.Content; s != "" {
				assistantBuf += s
				_ = emit(constant.SSEEventDelta, map[string]any{"text": s})
			}
			// 工具调用请求结束标志
			if chunk.Choices[0].FinishReason == "tool_calls" {
				toolCallsFinished = true
				// 取完整的 tool_calls
				if len(acc.Choices) > 0 && len(acc.Choices[0].Message.ToolCalls) > 0 {
					_ = emit(constant.SSEEventStartToolCall, map[string]any{
						"tool_calls": acc.Choices[0].Message.ToolCalls,
					})
				}
				return errno.OllamaInternalStopStream
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	// 存储历史
	if assistantBuf != "" {
		hist = append(hist, openai.AssistantMessage(assistantBuf))
	}

	// 没有工具调用：直接完成
	if !toolCallsFinished {
		historyOpenAI[id] = hist
		_ = emit(constant.SSEEventDone, map[string]any{"reason": "no_tool"})
		return nil
	}

	// 执行工具
	for _, tc := range acc.Choices[0].Message.ToolCalls {
		name := tc.Function.Name
		args := parseOpenAIToolArgs(tc.Function.Arguments)

		_ = emit(constant.SSEEventToolCall, map[string]any{
			"name": name,
			"args": args,
		})

		out, callErr := h.mcpCli.CallTool(ctx, name, args)
		if callErr != nil {
			out = "tool error: " + callErr.Error()
		}

		_ = emit(constant.SSEEventToolResult, map[string]any{
			"name":   name,
			"result": out,
		})

		// 工具结果落历史
		hist = append(hist, openai.ToolMessage(out, tc.ID))
		logger.Infof("[tool] %s executed", name)
	}

	// 二次流式：带工具结果，让模型给最终回答（与原逻辑一致）
	var finalBuf string
	err = h.ollamaCli.ChatStreamOpenAI(ctx, openai.ChatCompletionNewParams{
		Model:    config.Ollama.Model,
		Messages: hist,
		Tools:    tools,
	}, func(chunk *openai.ChatCompletionChunk) error {
		if len(chunk.Choices) > 0 {
			if s := chunk.Choices[0].Delta.Content; s != "" {
				finalBuf += s
				_ = emit(constant.SSEEventDelta, map[string]any{"text": s})
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	if finalBuf != "" {
		hist = append(hist, openai.AssistantMessage(finalBuf))
	}
	historyOpenAI[id] = hist
	_ = emit(constant.SSEEventDone, map[string]any{"reason": "completed"})
	return nil
}
