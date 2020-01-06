package crawl

// TODO: ...
type NodePool struct {
	nodes map[string]struct{}
}

func NewNodePool() *NodePool {
	return &NodePool{nodes: make(map[string]struct{})}
}

// Seed seeds the node pool with a given set of node IPs.
func (p *NodePool) Seed(seeds []string) {
	for _, s := range seeds {
		p.AddNode(s)
	}
}

// TODO:
func (p *NodePool) RandomNode() (string, bool) {
	for nodeIP := range p.nodes {
		return nodeIP, true
	}

	return "", false
}

// AddNode adds a node IP to the node pool.
func (p *NodePool) AddNode(nodeIP string) {
	p.nodes[nodeIP] = struct{}{}
}

// HasNode returns a boolean based on if a node IP exists in the node pool.
func (p *NodePool) HasNode(nodeIP string) bool {
	_, ok := p.nodes[nodeIP]
	return ok
}

// DeleteNode removes a node from the node pool if it exists.
func (p *NodePool) DeleteNode(nodeIP string) {
	delete(p.nodes, nodeIP)
}
