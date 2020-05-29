package lb

type Balance interface {
	/**
	 *负载均衡算法
	 */
	DoBalance([]*Instance, ...string) (*Instance, error)
}
type Instance struct {
	host string
	port int
}

func (p *Instance) GetHost() string {
	return p.host
}

func (p *Instance) GetPort() int {
	return p.port
}
