package crawl

// NodePool implements an abstraction over a pool of nodes for which to crawl.
// Note, it is not thread-safe.
type NodePool struct {
	nodes map[string]struct{}
}

func NewNodePool() *NodePool {
	return &NodePool{nodes: make(map[string]struct{})}
}

// Size returns the size of the pool.
func (p *NodePool) Size() int {
	return len(p.nodes)
}

// Seed seeds the node pool with a given set of node IPs.
func (p *NodePool) Seed(seeds []string) {
	for _, s := range seeds {
		p.AddNode(s)
	}
}

// RandomNode returns a random node, based on Golang's map semantics, from the pool.
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
