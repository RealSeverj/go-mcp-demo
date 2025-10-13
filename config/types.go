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
	Model   string        `mapstructure:"model"`    // e.g. qwen3:latest
	Options OllamaOptions `mapstructure:"options"`
}

type cliConfig struct {
	SystemPrompt string `mapstructure:"system_prompt"`
	History      bool   `mapstructure:"history"`
	MaxTurns     int    `mapstructure:"max_turns"`
}

type mcpStdio struct {
	ServerCmd  string   `mapstructure:"server_cmd"`  // 如 ./bin/mcp-server
	ServerArgs []string `mapstructure:"server_args"` // 可为空
}

type mcpConsul struct {
	Enable     bool   `mapstructure:"enable"`
	Address    string `mapstructure:"address"`    // 例如 "127.0.0.1:8500"
	Datacenter string `mapstructure:"datacenter"` // 可空
	Token      string `mapstructure:"token"`      // 可空
	Service    string `mapstructure:"service"`    // 要发现的服务名
	Tag        string `mapstructure:"tag"`        // 可空，筛选用
	Scheme     string `mapstructure:"scheme"`     // "http" or "https"
	Path       string `mapstructure:"path"`       // 例如 "/mcp" 或 "/mcp/sse"
}

type mcpHTTP struct {
	Mode        string        `mapstructure:"mode"`         // "sse" 或 "http"
	BaseURL     string        `mapstructure:"base_url"`     // 直接指定URL时使用，如 "http://host:8080/mcp"
	InitTimeout time.Duration `mapstructure:"init_timeout"` // 默认 10s
	CallTimeout time.Duration `mapstructure:"call_timeout"` // 默认 30s
	Consul      mcpConsul     `mapstructure:"consul"`       // 如果启用则优先生效
}

type mcpConfig struct {
	ServerName string        `mapstructure:"server_name"`
	Transport  string        `mapstructure:"transport"` // "stdio" | "sse" | "http"
	Stdio      mcpStdio      `mapstructure:"stdio"`
	HTTP       mcpHTTP       `mapstructure:"http"`
	InitTO     time.Duration `mapstructure:"init_timeout"` // stdio时用，默认10s
	CallTO     time.Duration `mapstructure:"call_timeout"` // stdio时用，默认30s
}

type service struct {
	Name     string
	AddrList []string
	LB       bool `mapstructure:"load-balance"`
}

type Config struct {
	Server server       `mapstructure:"server"`
	Ollama ollamaConfig `mapstructure:"ollama"`
	CLI    cliConfig    `mapstructure:"cli"`
	MCP    mcpConfig    `mapstructure:"mcp"`
}
