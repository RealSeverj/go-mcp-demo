package config

import "time"

type server struct {
	Secret   string `mapstructure:"private-key"`
	Version  string
	Name     string
	LogLevel string `mapstructure:"log-level"`
}

type OllamaOptions struct {
	Temperature   *float64       `mapstructure:"temperature"`
	TopP          *float64       `mapstructure:"top_p"`
	TopK          *int           `mapstructure:"top_k"`
	MaxTokens     *int           `mapstructure:"max_tokens"`
	Extra         map[string]any `mapstructure:"extra"` // 透传到 options
	KeepAlive     string         `mapstructure:"keep_alive"`
	RequestTimout time.Duration  `mapstructure:"request_timeout"`
}

type ollamaConfig struct {
	BaseURL string        `mapstructure:"base_url"` // e.g. http://127.0.0.1:11434
	Model   string        `mapstructure:"model"`    // e.g. qwen3:1.7b
	Options OllamaOptions `mapstructure:"options"`
}

type cliConfig struct {
	SystemPrompt string `mapstructure:"system_prompt"`
	History      bool   `mapstructure:"history"`
	MaxTurns     int    `mapstructure:"max_turns"`
}

/************ MCP（仅关注自身传输及超时，不再包含 Consul） ************/

type mcpStdio struct {
	ServerCmd  string   `mapstructure:"server_cmd"`  // 如 ./bin/mcp-server
	ServerArgs []string `mapstructure:"server_args"` // 可为空
}

type mcpHTTP struct {
	Mode    string `mapstructure:"mode"`     // "sse" | "http"
	BaseURL string `mapstructure:"base_url"` // 直连时使用，如 "http://127.0.0.1:8080/mcp"
}

type mcpConfig struct {
	ServerName string   `mapstructure:"server_name"`
	Transport  string   `mapstructure:"transport"` // "stdio" | "sse" | "http"
	Stdio      mcpStdio `mapstructure:"stdio"`
	HTTP       mcpHTTP  `mapstructure:"http"`
}

type consulConfig struct {
	Enable     bool   `mapstructure:"enable"`
	Address    string `mapstructure:"address"`    // 例如 "127.0.0.1:8500"
	Datacenter string `mapstructure:"datacenter"` // 可空
	Token      string `mapstructure:"token"`      // 可空
	Service    string `mapstructure:"service"`    // 要发现的服务名
	Tag        string `mapstructure:"tag"`        // 可空
	Scheme     string `mapstructure:"scheme"`     // "http" | "https"
	Path       string `mapstructure:"path"`       // 例如 "/mcp"
}

// registryConfig 是“注册中心/地址解析”的顶层入口
// Provider: "consul" | "static" | "none"（或留空）
// - 当 Provider=consul 时使用 Consul 子配置
// - 当 Provider=static 时使用 services.* 的静态地址（见 services 配置）
// - 当 BaseURL 非空时，优先使用 BaseURL（由 MCP 自身指定）
type registryConfig struct {
	Provider string       `mapstructure:"provider"` // "consul" | "none"
	Consul   consulConfig `mapstructure:"consul"`
}

type service struct {
	Name     string
	AddrList []string
	LB       bool `mapstructure:"load-balance"`
}

type Config struct {
	Server   server         `mapstructure:"server"`
	Ollama   ollamaConfig   `mapstructure:"ollama"`
	CLI      cliConfig      `mapstructure:"cli"`
	MCP      mcpConfig      `mapstructure:"mcp"`
	Registry registryConfig `mapstructure:"registry"`
}
