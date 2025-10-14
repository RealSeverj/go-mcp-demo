package api

import (
	"github.com/FantasyRL/go-mcp-demo/pkg/base"
)

var clientSet *base.ClientSet

func Init() {
	clientSet = base.NewClientSet(base.WithMCPClient(), base.WithOllamaClient())
}
