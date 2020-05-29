package lb

import (
	"errors"
	"sync"
)

type randomLb struct {
	index  uint8
	length uint8
	m      *sync.RWMutex
	nodes  []*BlNode
}

func (r *randomLb) DoBalance() (node *BlNode, err error) {
	r.m.RLock()
	defer r.m.RUnlock()
	if r.length == 0 {
		return nil, errors.New("not found available node ")
	}
	r.index++
	return r.nodes[r.index%r.length], nil
}

func (r *randomLb) AppendNode(node *BlNode) {
	r.m.Lock()
	defer r.m.Unlock()
	r.nodes = append(r.nodes, node)
	r.length++
}

func (r *randomLb) RemoveNode(id string) {
	r.m.Lock()
	defer r.m.Unlock()
	for i, n := range r.nodes {
		if n.Id == id {
			r.nodes = append(r.nodes[:i], r.nodes[i+1:]...)
		}
	}
}

func newRandomLb() *randomLb {
	return &randomLb{
		m:      &sync.RWMutex{},
		index:  0,
		length: 0,
	}
}
