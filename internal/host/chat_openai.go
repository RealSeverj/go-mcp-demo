package host

import (
	"context"
	"encoding/json"
	"github.com/FantasyRL/go-mcp-demo/config"
	"github.com/FantasyRL/go-mcp-demo/pkg/constant"
	"github.com/FantasyRL/go-mcp-demo/pkg/errno"
	openai "github.com/openai/openai-go/v2"
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

const maxToolRounds = 10 // 防御性上限，避免死循环

func (h *Host) StreamChatOpenAI(
	ctx context.Context,
	id int64,
	userMsg string,
	emit func(event string, v any) error,
) error {
	// 历史（OpenAI）
	hist := historyOpenAI[id]
	if hist == nil {
		hist = []openai.ChatCompletionMessageParamUnion{}
	}
	// 用户消息
	hist = append(hist, openai.UserMessage(userMsg))

	// 工具（OpenAI 版）
	tools := h.mcpCli.ConvertToolsToOpenAI()

	round := 0
	for {
		round++
		if round > maxToolRounds {
			historyOpenAI[id] = hist
			_ = emit(constant.SSEEventDone, map[string]any{"reason": "tool_round_limit"})
			return nil
		}

		// 一轮生成：边流边推，若需要工具则中断本轮
		var assistantBuf string
		var acc openai.ChatCompletionAccumulator
		var needTools bool

		err := h.aiProviderCli.ChatStreamOpenAI(ctx, openai.ChatCompletionNewParams{
			Model:    openai.ChatModel(config.AiProvider.Model),
			Messages: hist,
			Tools:    tools,
		}, func(chunk *openai.ChatCompletionChunk) error {
			acc.AddChunk(*chunk)
			if len(chunk.Choices) > 0 {
				if s := chunk.Choices[0].Delta.Content; s != "" {
					assistantBuf += s
					_ = emit(constant.SSEEventDelta, map[string]any{"text": s})
				}
				// 工具调用结束标志（OpenAI：最后一帧 finish_reason = "tool_calls"）
				if chunk.Choices[0].FinishReason == "tool_calls" {
					needTools = true
					if len(acc.Choices) > 0 && len(acc.Choices[0].Message.ToolCalls) > 0 {
						_ = emit(constant.SSEEventStartToolCall, map[string]any{
							"tool_calls": acc.Choices[0].Message.ToolCalls,
							"round":      round,
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

		// 把已产生的 assistant 文本落历史
		if assistantBuf != "" && !needTools {
			hist = append(hist, openai.AssistantMessage(assistantBuf))
		}

		// 如果本轮不需要工具，说明模型已经给出最终答案
		if !needTools {
			historyOpenAI[id] = hist
			_ = emit(constant.SSEEventDone, map[string]any{"reason": "completed"})
			return nil
		}

		// 执行（可能多个）工具调用，然后将每个工具结果以 ToolMessage 落历史
		if len(acc.Choices) == 0 || len(acc.Choices[0].Message.ToolCalls) == 0 {
			// 偶发兜底：标记需要工具但没聚合到（理论上不会发生）
			historyOpenAI[id] = hist
			_ = emit(constant.SSEEventDone, map[string]any{"reason": "no_tool_details"})
			return nil
		}

		toolCallsParam := make([]openai.ChatCompletionMessageToolCallUnionParam, 0, len(acc.Choices[0].Message.ToolCalls))
		for _, tc := range acc.Choices[0].Message.ToolCalls {
			toolCallsParam = append(toolCallsParam, openai.ChatCompletionMessageToolCallUnionParam{
				OfFunction: &openai.ChatCompletionMessageFunctionToolCallParam{
					ID:   tc.ID,
					Type: "function",
					Function: openai.ChatCompletionMessageFunctionToolCallFunctionParam{
						Name:      tc.Function.Name,
						Arguments: tc.Function.Arguments, // 注意：这里是字符串
					},
				},
			})
		}
		// 根据openAI规范，tool_call前需要一条assistantMsg
		assistantWithCalls := openai.ChatCompletionAssistantMessageParam{
			Role:      "assistant",
			ToolCalls: toolCallsParam,
		}
		hist = append(hist, openai.ChatCompletionMessageParamUnion{OfAssistant: &assistantWithCalls})

		for _, tc := range acc.Choices[0].Message.ToolCalls {
			name := tc.Function.Name

			// OpenAI 的 arguments 是字符串，需要解成 map[string]any
			var args map[string]any
			if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
				args = map[string]any{"_parse_error": err.Error(), "_raw": tc.Function.Arguments}
			}

			_ = emit(constant.SSEEventToolCall, map[string]any{
				"round": round,
				"name":  name,
				"args":  args,
			})

			out, callErr := h.mcpCli.CallTool(ctx, name, args)
			if callErr != nil {
				out = "tool error: " + callErr.Error()
			}

			_ = emit(constant.SSEEventToolResult, map[string]any{
				"round":  round,
				"name":   name,
				"result": out,
			})

			// 工具结果回模型（重要）：OpenAI 规范用 ToolMessage，必须带 tool_call_id
			hist = append(hist, openai.ToolMessage(out, tc.ID))
			//logger.Infof("[tool round %d] %s executed", round, name)
		}

		// 循环进入下一轮：模型会在新的上下文（含工具结果）上继续生成
	}
}
