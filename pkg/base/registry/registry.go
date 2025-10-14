package registry

// Resolver 服务发现接口 做成接口可以扩展不同的服务发现实现(consul/etcd)
type Resolver interface {
	// Resolve 解析服务端地址，返回格式如 ip:port 的字符串
	Resolve() (string, error)
}
