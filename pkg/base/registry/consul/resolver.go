package consul

import (
	"fmt"
	"github.com/FantasyRL/go-mcp-demo/config"

	"github.com/hashicorp/consul/api"
)

type ConsulConfig struct {
	Address    string // 127.0.0.1:8500
	Datacenter string // 可空
	Token      string // 可空
	Service    string // 要发现的服务名（必填）
	Tag        string // 可空，按 tag 过滤
	Scheme     string // http / https，默认 http
	Path       string // 例如 "/mcp"，默认 "/mcp"
}

type Resolver struct {
	cfg *ConsulConfig
}

func NewResolver() *Resolver {
	cfg := &ConsulConfig{
		Address:    config.Registry.Consul.Address,
		Datacenter: config.Registry.Consul.Datacenter,
		Token:      config.Registry.Consul.Token,
		Service:    config.Registry.Consul.Service,
		Tag:        config.Registry.Consul.Tag,
		Scheme:     config.Registry.Consul.Scheme,
		Path:       config.Registry.Consul.Path,
	}
	if cfg.Address == "" || cfg.Service == "" {
		return nil
	}
	return &Resolver{cfg: cfg}
}

func (r *Resolver) Resolve() (string, error) {
	conf := api.DefaultConfig()
	conf.Address = r.cfg.Address
	if r.cfg.Datacenter != "" {
		conf.Datacenter = r.cfg.Datacenter
	}
	if r.cfg.Token != "" {
		conf.Token = r.cfg.Token
	}

	cl, err := api.NewClient(conf)
	if err != nil {
		return "", fmt.Errorf("consul client: %w", err)
	}
	q := &api.QueryOptions{Datacenter: r.cfg.Datacenter, Token: r.cfg.Token}
	entries, _, err := cl.Health().Service(r.cfg.Service, r.cfg.Tag, true, q)
	if err != nil {
		return "", fmt.Errorf("consul discover: %w", err)
	}
	if len(entries) == 0 {
		return "", fmt.Errorf("consul: no healthy instances for %s", r.cfg.Service)
	}

	inst := entries[0]
	host := inst.Service.Address
	if host == "" {
		host = inst.Node.Address
	}
	port := inst.Service.Port
	scheme := r.cfg.Scheme
	if scheme == "" {
		scheme = "http"
	}
	path := r.cfg.Path
	if path == "" {
		path = "/mcp"
	}
	return fmt.Sprintf("%s://%s:%d%s", scheme, host, port, path), nil
}
