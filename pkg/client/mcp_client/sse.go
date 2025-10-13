// pkg/mcp_client/http_sse.go
package mcp_client

import (
	"context"
	"errors"
	"fmt"
	"github.com/FantasyRL/go-mcp-demo/config"
	"time"

	consul "github.com/hashicorp/consul/api"
	mcpc "github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

// newSSEMCPClient 通过 SSE 连接
func newSSEMCPClient() (*MCPClient, error) {
	baseURL, err := resolveBaseURL()
	if err != nil {
		return nil, err
	}
	c, err := mcpc.NewSSEMCPClient(baseURL)
	if err != nil {
		return nil, fmt.Errorf("new sse client: %w", err)
	}

	initTO := config.MCP.HTTP.InitTimeout
	if initTO <= 0 {
		initTO = 10 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), initTO)
	defer cancel()

	// 关键：HTTP/SSE 需先 Start 再 Initialize
	if err := c.Start(ctx); err != nil {
		return nil, fmt.Errorf("sse start: %w", err)
	}

	_, err = c.Initialize(ctx, mcp.InitializeRequest{
		Params: mcp.InitializeParams{
			ClientInfo: mcp.Implementation{Name: "mcp-host", Version: "0.1.0"},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("initialize mcp (sse): %w", err)
	}

	res, err := c.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		return nil, fmt.Errorf("list tools: %w", err)
	}
	return &MCPClient{Client: c, Tools: res.Tools}, nil
}

// newHTTPMCPClient 通过 HTTP 连接
func newHTTPMCPClient() (*MCPClient, error) {
	baseURL, err := resolveBaseURL()
	if err != nil {
		return nil, err
	}
	c, err := mcpc.NewStreamableHttpClient(baseURL)
	if err != nil {
		return nil, fmt.Errorf("new http client: %w", err)
	}

	initTO := config.MCP.HTTP.InitTimeout
	if initTO <= 0 {
		initTO = 10 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), initTO)
	defer cancel()

	if err := c.Start(ctx); err != nil {
		return nil, fmt.Errorf("http start: %w", err)
	}
	_, err = c.Initialize(ctx, mcp.InitializeRequest{
		Params: mcp.InitializeParams{
			ClientInfo: mcp.Implementation{Name: "mcp-host", Version: "0.1.0"},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("initialize mcp (http): %w", err)
	}

	res, err := c.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		return nil, fmt.Errorf("list tools: %w", err)
	}
	return &MCPClient{Client: c, Tools: res.Tools}, nil
}

// 解析 baseURL：优先 Consul -> 否则使用配置里的 base_url
func resolveBaseURL() (string, error) {
	if config.MCP.HTTP.Consul.Enable {
		addr, err := discoverViaConsul()
		if err != nil {
			return "", err
		}
		return addr, nil
	}
	if config.MCP.HTTP.BaseURL == "" {
		return "", errors.New("http.base_url empty and consul disabled")
	}
	return config.MCP.HTTP.BaseURL, nil
}

// 用 Consul 发现一个健康实例并拼接 URL
func discoverViaConsul() (string, error) {
	if config.MCP.HTTP.Consul.Address == "" || config.MCP.HTTP.Consul.Service == "" {
		return "", errors.New("consul.address or consul.service is empty")
	}
	conf := consul.DefaultConfig()
	conf.Address = config.MCP.HTTP.Consul.Address
	if config.MCP.HTTP.Consul.Datacenter != "" {
		conf.Datacenter = config.MCP.HTTP.Consul.Datacenter
	}
	if config.MCP.HTTP.Consul.Token != "" {
		conf.Token = config.MCP.HTTP.Consul.Token
	}
	cl, err := consul.NewClient(conf)
	if err != nil {
		return "", fmt.Errorf("consul client: %w", err)
	}
	q := &consul.QueryOptions{Datacenter: config.MCP.HTTP.Consul.Datacenter, Token: config.MCP.HTTP.Consul.Token}
	entries, _, err := cl.Health().Service(config.MCP.HTTP.Consul.Service, config.MCP.HTTP.Consul.Tag, true, q)
	if err != nil {
		return "", fmt.Errorf("consul discover: %w", err)
	}
	if len(entries) == 0 {
		return "", fmt.Errorf("consul: no healthy instances for %s", config.MCP.HTTP.Consul.Service)
	}
	inst := entries[0]
	host := inst.Service.Address
	if host == "" {
		host = inst.Node.Address
	}
	port := inst.Service.Port
	scheme := config.MCP.HTTP.Consul.Scheme
	if scheme == "" {
		scheme = "http"
	}
	path := config.MCP.HTTP.Consul.Path
	if path == "" {
		path = "/mcp"
	}
	return fmt.Sprintf("%s://%s:%d%s", scheme, host, port, path), nil
}
