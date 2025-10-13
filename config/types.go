package config

import "time"

type OllamaOptions struct {
	Temperature   *float64       `mapstructure:"temperature"`
	TopP          *float64       `mapstructure:"top_p"`
	TopK          *int           `mapstructure:"top_k"`
	MaxTokens     *int           `mapstructure:"max_tokens"`
	Extra         map[string]any `mapstructure:"extra"` // 透传到 options
	KeepAlive     string         `mapstructure:"keep_alive"`
	RequestTimout time.Duration  `mapstructure:"request_timeout"`
}

type OllamaConfig struct {
	BaseURL string        `mapstructure:"base_url"` // e.g. http://127.0.0.1:11434
	Model   string        `mapstructure:"model"`    // e.g. qwen3:latest
	Options OllamaOptions `mapstructure:"options"`
}

type CLIConfig struct {
	SystemPrompt string `mapstructure:"system_prompt"`
	History      bool   `mapstructure:"history"`
	MaxTurns     int    `mapstructure:"max_turns"`
}

type MCPConfig struct {
	ServerName string `mapstructure:"server_name"`
	Transport  string `mapstructure:"transport"` // stdio | grpc | http
}

type MCPStdio struct {
	ServerCmd  string   `mapstructure:"server_cmd"`  // 如 ./bin/mcp-server
	ServerArgs []string `mapstructure:"server_args"` // 可为空
}

type Config struct {
	Ollama OllamaConfig `mapstructure:"ollama"`
	CLI    CLIConfig    `mapstructure:"cli"`
	MCP    MCPConfig    `mapstructure:"mcp"`
}
