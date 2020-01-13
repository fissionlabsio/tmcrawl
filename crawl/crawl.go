package crawl

import (
	"fmt"
	"time"

	"github.com/fissionlabsio/tmcrawl/config"
	"github.com/fissionlabsio/tmcrawl/db"
	"github.com/harwoeck/ipstack"
	"github.com/rs/zerolog/log"
)

const defaultP2PPort = "26656"

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
		nodeRPCAddr, ok := c.pool.RandomNode()
		for ok {
			c.CrawlNode(nodeRPCAddr)
			c.pool.DeleteNode(nodeRPCAddr)

			nodeRPCAddr, ok = c.pool.RandomNode()
		}

		log.Info().Uint("duration", c.crawlInterval).Msg("waiting until next crawl attempt...")
		time.Sleep(time.Duration(c.crawlInterval) * time.Second)
		c.pool.Reseed()
	}
}

// CrawlNode performs the main crawling functionality for a Tendermint node. It
// accepts a node RPC address and attempts to ping that node's P2P address by
// using the RPC address and the default P2P port of 26656. If the P2P address
// cannot be reached, the node is deleted if it exists in the database. Otherwise,
// we attempt to get additional metadata aboout the node via it's RPC address
// and its set of peers. For every peer that doesn't exist in the node pool, it
// is added.
func (c *Crawler) CrawlNode(nodeRPCAddr string) {
	host := parseHostname(nodeRPCAddr)
	nodeP2PAddr := fmt.Sprintf("%s:%s", host, defaultP2PPort)

	node := Node{
		Address:  host,
		RPCPort:  parsePort(nodeRPCAddr),
		P2PPort:  defaultP2PPort,
		LastSync: time.Now().UTC().Format(time.RFC3339),
	}

	log.Debug().Str("p2p_address", nodeP2PAddr).Str("rpc_address", nodeRPCAddr).Msg("pinging node...")
	if ok := PingAddress(nodeP2PAddr, 5); !ok {
		log.Info().Str("p2p_address", nodeP2PAddr).Str("rpc_address", nodeRPCAddr).Msg("failed to ping node; deleting...")

		if err := c.DeleteNodeIfExist(node); err != nil {
			log.Info().Err(err).Str("p2p_address", nodeP2PAddr).Str("rpc_address", nodeRPCAddr).Msg("failed to delete node")
		}

		return
	}

	loc, err := c.GetGeolocation(host)
	if err != nil {
		log.Info().Err(err).Str("p2p_address", nodeP2PAddr).Str("rpc_address", nodeRPCAddr).Msg("failed to get node geolocation")
	} else {
		node.Location = loc
	}

	client := newRPCClient(nodeRPCAddr)

	status, err := client.Status()
	if err != nil {
		log.Info().Err(err).Str("p2p_address", nodeP2PAddr).Str("rpc_address", nodeRPCAddr).Msg("failed to get node status")
	} else {
		node.Moniker = status.NodeInfo.Moniker
		node.ID = string(status.NodeInfo.ID())
		node.Network = status.NodeInfo.Network
		node.Version = status.NodeInfo.Version
		node.TxIndex = status.NodeInfo.Other.TxIndex

		netInfo, err := client.NetInfo()
		if err != nil {
			log.Info().Err(err).Str("p2p_address", nodeP2PAddr).Str("rpc_address", nodeRPCAddr).Msg("failed to get node net info")
			return
		}

		for _, p := range netInfo.Peers {
			peerRPCPort := parsePort(p.NodeInfo.Other.RPCAddress)
			peerRPCAddress := fmt.Sprintf("http://%s:%s", p.RemoteIP, peerRPCPort)
			peer := Node{
				Address: p.RemoteIP,
			}

			// only add peer to the pool if we haven't (re)discovered it
			if !c.db.Has(peer.Key()) {
				log.Debug().Str("p2p_address", nodeP2PAddr).Str("rpc_address", nodeRPCAddr).Str("peer_rpc_address", peerRPCAddress).Msg("adding peer to node pool")
				c.pool.AddNode(peerRPCAddress)
			}
		}
	}

	if err := c.SaveNode(node); err != nil {
		log.Info().Err(err).Str("p2p_address", nodeP2PAddr).Str("rpc_address", nodeRPCAddr).Msg("failed to encode node")
	} else {
		log.Info().Str("p2p_address", nodeP2PAddr).Str("rpc_address", nodeRPCAddr).Msg("successfully crawled and persisted node")
	}
}

// SaveNode persists a node to the database by it's addressable key. An error is
// returned if it cannot be marshaled or the database operation fails.
func (c *Crawler) SaveNode(n Node) error {
	bz, err := n.Marshal()
	if err != nil {
		return err
	}

	if err := c.db.Set(n.Key(), bz); err != nil {
		return err
	}

	return nil
}

// DeleteNodeIfExist removes a node by it's addressable key from the database
// if it exists. An error is returned if it exists and cannot be deleted.
func (c *Crawler) DeleteNodeIfExist(n Node) error {
	key := n.Key()
	if c.db.Has(key) {
		return c.db.Delete(key)
	}

	return nil
}

// GetGeolocation returns a Location object containing geolocation information
// for a given node IP. It will first check to see if the location already exists
// in the database and return it if so. Otherwise, a query is made against IPStack
// and persisted. An error is returned if the location cannot be decoded or queried
// for.
func (c *Crawler) GetGeolocation(nodeIP string) (Location, error) {
	locKey := LocationKey(nodeIP)

	// return location if it exists in the database
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
