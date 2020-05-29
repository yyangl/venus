package lb

type Balance interface {
	/**
	 *负载均衡算法
	 */
	DoBalance() (node *BlNode, err error)
	AppendNode(node *BlNode)
	RemoveNode(id string)
}

type BlNode struct {
	Id   string
	Ip   string
	Port int
}

var DefaultRandom Balance = newRandomLb()
