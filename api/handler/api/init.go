package api

import (
	"github.com/FantasyRL/go-mcp-demo/pkg/client"
)

var clientSet *client.ClientSet

func Init() {
	clientSet = client.NewClientSet(client.WithMCPClient(), client.WithOllamaClient())
}
