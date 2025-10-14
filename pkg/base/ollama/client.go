package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/FantasyRL/go-mcp-demo/config"
	"io"
	"net"
	"net/http"
	"time"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

type ClientOptions struct {
	BaseURL       string
	RequestTimout time.Duration
}

// NewOllamaClient 创建一个 Ollama 客户端
func NewOllamaClient() *Client {
	to := config.Ollama.Options.RequestTimout
	if to <= 0 {
		to = 60 * time.Second
	}
	return &Client{
		baseURL: config.Ollama.BaseURL,
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
	}
}

// Chat 调用 /api/chat，非流式
func (c *Client) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	endpoint := fmt.Sprintf("%s/api/chat", c.baseURL)
	req.Stream = false

	b, _ := json.Marshal(req)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama chat failed: %s - %s", resp.Status, string(body))
	}
	var cr ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&cr); err != nil {
		return nil, err
	}
	return &cr, nil
}
