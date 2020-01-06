package crawl

import (
	"fmt"
	"time"

	"github.com/fissionlabsio/tmcrawl/config"
	"github.com/fissionlabsio/tmcrawl/db"
	"github.com/harwoeck/ipstack"
	"github.com/rs/zerolog/log"
)

// TODO: ...
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
		pool:            NewNodePool(),
		ipClient:        ipstack.NewClient(cfg.IPStackKey, false, 5),
	}
}

// TODO: ...
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
	}
}

// TODO: ...
func (c *Crawler) CrawlNode(nodeAddr string) {
	client := newRPCClient(nodeAddr)

	status, err := client.Status()
	if err != nil {
		log.Info().Err(err).Str("node", nodeAddr).Msg("failed to get node status; removing")
		c.removeNodeIfExist(nodeAddr)
		return
	}

	netInfo, err := client.NetInfo()
	if err != nil {
		log.Info().Err(err).Str("node", nodeAddr).Msg("failed to get node net info")
		return
	}

	for _, p := range netInfo.Peers {
		port := parsePort(p.NodeInfo.Other.RPCAddress)
		peer := fmt.Sprintf("tcp://%s:%s", p.RemoteIP, port)

		// only add peer to the pool if we haven't (re)discovered it
		if !c.db.Has([]byte(peer)) {
			c.pool.AddNode(peer)
		}
	}

	nodeIP := parseHostname(nodeAddr)

	// TODO: CACHE OR STORE!
	ipResp, err := c.ipClient.Check(nodeIP)
	if err != nil {
		log.Info().Err(err).Str("node", nodeIP).Msg("failed to get node geolocation info")
		return
	}

	node := Node{
		RemoteIP: nodeIP,
		RPCPort:  parsePort(nodeAddr),
		Moniker:  status.NodeInfo.Moniker,
		ID:       string(status.NodeInfo.ID()),
		Network:  status.NodeInfo.Network,
		Version:  status.NodeInfo.Version,
		TxIndex:  status.SyncInfo.CatchingUp,
		LastSync: time.Now().UTC(),
		Location: locationFromIPResp(ipResp),
	}

	bz, err := node.Marshal()
	if err != nil {
		log.Info().Err(err).Str("node", nodeAddr).Msg("failed to encode node")
		return
	}

	if err := c.db.Set([]byte(nodeAddr), bz); err != nil {
		log.Info().Err(err).Str("node", nodeAddr).Msg("failed to persist node")
	}

	log.Info().Str("node", nodeAddr).Msg("successfully crawled and persisted node")
}

func (c *Crawler) removeNodeIfExist(nodeAddr string) {
	nodeKey := []byte(nodeAddr)
	if c.db.Has(nodeKey) {
		bz, _ := c.db.Get(nodeKey)

		node := new(Node)
		if err := node.Unmarshal(bz); err != nil {
			log.Info().Err(err).Str("node", nodeAddr).Msg("failed to decode node")
			return
		}
	}
}
