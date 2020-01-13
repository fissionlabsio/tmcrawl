package crawl

import (
	"math/rand"
	"time"
)

// NodePool implements an abstraction over a pool of nodes for which to crawl.
// It also contains a collection of nodes for which to reseed the pool when it's
// empty. Once the reseed list has reached capacity, a random node is removed
// when another is added. Note, it is not thread-safe.
type NodePool struct {
	nodes       map[string]struct{}
	reseedNodes []string
	rng         *rand.Rand
}

func NewNodePool(reseedCap uint) *NodePool {
	return &NodePool{
		nodes:       make(map[string]struct{}),
		reseedNodes: make([]string, 0, reseedCap),
		rng:         rand.New(rand.NewSource(time.Now().Unix())),
	}
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
	for nodeRPCAddr := range p.nodes {
		return nodeRPCAddr, true
	}

	return "", false
}

// AddNode adds a node RPC address to the node pool. In addition, it adds the
// node to the reseed list. If the reseed list is full, it replaces a random node.
func (p *NodePool) AddNode(nodeRPCAddr string) {
	p.nodes[nodeRPCAddr] = struct{}{}

	if len(p.reseedNodes) < cap(p.reseedNodes) {
		p.reseedNodes = append(p.reseedNodes, nodeRPCAddr)
	} else {
		// replace random node with the new node
		i := p.rng.Intn(len(p.reseedNodes))
		p.reseedNodes[i] = nodeRPCAddr
	}
}

// HasNode returns a boolean based on if a node RPC address exists in the node pool.
func (p *NodePool) HasNode(nodeRPCAddr string) bool {
	_, ok := p.nodes[nodeRPCAddr]
	return ok
}

// DeleteNode removes a node from the node pool if it exists.
func (p *NodePool) DeleteNode(nodeRPCAddr string) {
	delete(p.nodes, nodeRPCAddr)
}

// Reseed seeds the node pool with all the nodes found in the internal reseed
// list.
func (p *NodePool) Reseed() {
	p.Seed(p.reseedNodes)
}
