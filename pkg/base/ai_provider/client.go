package ai_provider

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/FantasyRL/go-mcp-demo/config"
	"github.com/FantasyRL/go-mcp-demo/pkg/constant"
	"github.com/FantasyRL/go-mcp-demo/pkg/errno"
	"github.com/FantasyRL/go-mcp-demo/pkg/logger"
	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/option"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	mode         string
	baseURL      string
	httpClient   *http.Client
	openaiClient *openai.Client
}

type ClientOptions struct {
	BaseURL       string
	RequestTimout time.Duration
}

// NewAiProviderClient 创建一个 AiProvider 客户端
func NewAiProviderClient() *Client {
	to := config.AiProvider.Options.RequestTimout
	if to <= 0 {
		to = 60 * time.Second
	}
	switch config.AiProvider.Mode {
	case constant.AiProviderModeLocal:
		// 本地 AiProvider
		base := strings.TrimRight(config.AiProvider.BaseURL, "/") + "/v1" // AiProvider 的 OpenAI 兼容层
		openaiCli := openai.NewClient(
			option.WithAPIKey("ollama"),
			option.WithBaseURL(base),
		)
		return &Client{
			mode:    constant.AiProviderModeLocal,
			baseURL: config.AiProvider.BaseURL,
			httpClient: &http.Client{
				Timeout: to,
				Transport: &http.Transport{
					Proxy: http.ProxyFromEnvironment,
					DialContext: (&net.Dialer{
						Timeout:   10 * time.Second,
						KeepAlive: 30 * time.Second,
					}).DialContext,
					ForceAttemptHTTP2:     true,
					MaxIdleConns:          100,
					IdleConnTimeout:       90 * time.Second,
					TLSHandshakeTimeout:   10 * time.Second,
					ExpectContinueTimeout: 1 * time.Second,
				},
			},
			openaiClient: &openaiCli,
		}
	case constant.AiProviderModeRemote:
		// 远程 openAI-API
		openaiCli := openai.NewClient(
			option.WithAPIKey(config.AiProvider.Remote.APIKey),
			option.WithBaseURL(config.AiProvider.Remote.BaseURL))
		return &Client{
			mode:         constant.AiProviderModeRemote,
			baseURL:      config.AiProvider.Remote.BaseURL,
			openaiClient: &openaiCli,
		}
	default:
		logger.Errorf("unsupported mode: %s", config.AiProvider.Mode)
		return nil
	}

}

// Chat 调用 /api/chat，非流式
func (c *Client) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	endpoint := fmt.Sprintf("%s/api/chat", c.baseURL)
	req.Stream = false

	b, _ := json.Marshal(req)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(b))
	if err != nil {
		logger.Errorf("ollama.Chat NewRequestWithContext error: %v", err)
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		logger.Errorf("ollama.Chat Do request error: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logger.Errorf("ollama.Chat error response: %s", string(body))
		return nil, fmt.Errorf("ollama chat failed: %s - %s", resp.Status, string(body))
	}
	var cr ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&cr); err != nil {
		return nil, err
	}
	return &cr, nil
}

// ChatStream api/chat，流式
func (c *Client) ChatStream(ctx context.Context, req ChatRequest, onChunk func(*ChatResponse) error) error {
	endpoint := fmt.Sprintf("%s/api/chat", c.baseURL)
	req.Stream = true

	b, _ := json.Marshal(req)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(b))
	if err != nil {
		logger.Errorf("ollama.ChatStream NewRequestWithContext error: %v", err)
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		logger.Errorf("ollama.ChatStream Do request error: %v", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		all, _ := io.ReadAll(resp.Body)
		logger.Errorf("ollama chat stream error response: %s", string(all))
		return fmt.Errorf("ollama chat stream failed: %s - %s", resp.Status, string(all))
	}

	sc := bufio.NewScanner(resp.Body)
	// 适当放宽单行大小，以免长 JSON 被截断
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, 10*1024*1024)

	for sc.Scan() {
		line := sc.Bytes()
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		chunk := new(ChatResponse)
		if err := json.Unmarshal(line, chunk); err != nil {
			logger.Errorf("ollama.ChatStream json unmarshal chunk error: %v", err)
			return err
		}
		if err := onChunk(chunk); err != nil {
			if errors.Is(err, errno.OllamaInternalStopStream) {
				return nil
			}
			return err
		}
		if chunk.Done {
			break
		}
	}
	return sc.Err()
}

// ChatStreamOpenAI 使用 OpenAI 兼容层流式聊天
func (c *Client) ChatStreamOpenAI(
	ctx context.Context,
	req openai.ChatCompletionNewParams,
	onChunk func(*openai.ChatCompletionChunk) error,
) error {
	stream := c.openaiClient.Chat.Completions.NewStreaming(ctx, req)
	defer stream.Close()
	for stream.Next() {
		chunk := stream.Current()
		if err := onChunk(&chunk); err != nil {
			if errors.Is(err, errno.OllamaInternalStopStream) {
				return nil
			}
			return err
		}
	}
	if err := stream.Err(); err != nil {
		logger.Errorf("openai.ChatStreamOpenAI stream error: %v", err)
		return err
	}
	return nil
}

func (c *Client) ChatOpenAI(
	ctx context.Context,
	req openai.ChatCompletionNewParams,
) (*openai.ChatCompletion, error) {
	resp, err := c.openaiClient.Chat.Completions.New(ctx, req)
	if err != nil {
		logger.Errorf("openai.ChatOpenAI error: %v", err)
		return nil, err
	}
	return resp, nil
}
