package crawl

import (
	"fmt"
	"time"

	"github.com/fissionlabsio/tmcrawl/config"
	"github.com/fissionlabsio/tmcrawl/db"
	"github.com/harwoeck/ipstack"
	"github.com/rs/zerolog/log"
)

// Crawler implements the Tendermint p2p network crawler.
type Crawler struct {
	db       db.DB
	seeds    []string
	pool     *NodePool
	ipClient *ipstack.Client

	crawlInterval   uint
	recheckInterval uint
}

func NewCrawler(cfg config.Config, db db.DB) *Crawler {
	return &Crawler{
		db:              db,
		seeds:           cfg.Seeds,
		crawlInterval:   cfg.CrawlInterval,
		recheckInterval: cfg.RecheckInterval,
		pool:            NewNodePool(cfg.ReseedSize),
		ipClient:        ipstack.NewClient(cfg.IPStackKey, false, 5),
	}
}

// Crawl starts a blocking process in which a random node is selected from the
// node pool and crawled. For each successful crawl, it'll be persisted or updated
// and its peers will be added to the node pool if they do not already exist.
// This process continues indefinitely until all nodes are exhausted from the pool.
// When the pool is empty and after crawlInterval seconds since the last complete
// crawl, a random set of nodes from the DB are added to reseed the pool.
func (c *Crawler) Crawl() {
	// seed the pool with the initial set of seeds before crawling
	c.pool.Seed(c.seeds)

	for {
		nodeAddr, ok := c.pool.RandomNode()
		for ok {
			c.CrawlNode(nodeAddr)
			c.pool.DeleteNode(nodeAddr)

			nodeAddr, ok = c.pool.RandomNode()
		}

		log.Info().Uint("duration", c.crawlInterval).Msg("waiting until next crawl attempt...")
		time.Sleep(time.Duration(c.crawlInterval) * time.Second)
		c.pool.Reseed()
	}
}

// CrawlNode attempts to lookup a node's status and network info by it's node RPC
// address. Upon success, it will add all of its peers to the node pool if they
// do not already exist in the DB. If the node's status cannot be queried, it will
// be removed from the DB if it exists.
func (c *Crawler) CrawlNode(nodeAddr string) {
	client := newRPCClient(nodeAddr)

	status, err := client.Status()
	if err != nil {
		log.Info().Err(err).Str("node_address", nodeAddr).Msg("failed to get node status; removing")
		if err := c.deleteNodeIfExist(nodeAddr); err != nil {
			log.Info().Err(err).Str("node_address", nodeAddr).Msg("failed to delete node")
		}

		return
	}

	nodeID := string(status.NodeInfo.ID())

	netInfo, err := client.NetInfo()
	if err != nil {
		log.Info().Err(err).Str("node_address", nodeAddr).Str("node_id", nodeID).Msg("failed to get node net info")
		return
	}

	for _, p := range netInfo.Peers {
		port := parsePort(p.NodeInfo.Other.RPCAddress)
		peer := fmt.Sprintf("http://%s:%s", p.RemoteIP, port)

		// only add peer to the pool if we haven't (re)discovered it
		if !c.db.Has([]byte(peer)) {
			c.pool.AddNode(peer)
		}
	}

	nodeIP := parseHostname(nodeAddr)
	loc, err := c.getGeolocation(nodeIP)
	if err != nil {
		log.Info().Err(err).Str("node_address", nodeAddr).Str("node_id", nodeID).Msg("failed to get node geolocation info")
		return
	}

	node := Node{
		Address:  nodeAddr,
		RemoteIP: nodeIP,
		RPCPort:  parsePort(nodeAddr),
		Moniker:  status.NodeInfo.Moniker,
		ID:       nodeID,
		Network:  status.NodeInfo.Network,
		Version:  status.NodeInfo.Version,
		TxIndex:  status.NodeInfo.Other.TxIndex,
		LastSync: time.Now().UTC(),
		Location: loc,
	}
	c.saveNode(node)
}

func (c *Crawler) saveNode(n Node) {
	bz, err := n.Marshal()
	if err != nil {
		log.Info().Err(err).Str("node_address", n.Address).Str("node_id", n.ID).Msg("failed to encode node")
		return
	}

	if err := c.db.Set(NodeKey(n.ID), bz); err != nil {
		log.Info().Err(err).Str("node_address", n.Address).Str("node_id", n.ID).Msg("failed to persist node")
	}

	log.Info().Str("node_address", n.Address).Str("node_id", n.ID).Msg("successfully crawled and persisted node")
}

func (c *Crawler) deleteNodeIfExist(nodeAddr string) error {
	nodeKey := NodeKey(nodeAddr)
	if c.db.Has(nodeKey) {
		return c.db.Delete(nodeKey)
	}

	return nil
}

func (c *Crawler) getGeolocation(nodeIP string) (Location, error) {
	locKey := LocationKey(nodeIP)

	// attempt to query for location in the DB first
	if c.db.Has(locKey) {
		bz, err := c.db.Get(locKey)
		if err != nil {
			return Location{}, err
		}

		loc := new(Location)
		if err := loc.Unmarshal(bz); err != nil {
			return Location{}, err
		}

		return *loc, nil
	}

	// query for the location and persist it
	ipResp, err := c.ipClient.Check(nodeIP)
	if err != nil {
		return Location{}, err
	}

	loc := locationFromIPResp(ipResp)
	bz, err := loc.Marshal()
	if err != nil {
		return Location{}, err
	}

	err = c.db.Set(locKey, bz)
	return loc, err
}
