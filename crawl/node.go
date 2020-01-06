package crawl

import (
	"time"

	"github.com/vmihailenco/msgpack/v4"
)

type (
	// Node represents a full-node in a Tendermint-based network that contains
	// relevant p2p data.
	Node struct {
		RemoteIP string    `json:"remote_ip" yaml:"remote_ip"`
		RPCPort  string    `json:"rpc_port" yaml:"rpc_port"`
		Moniker  string    `json:"moniker" yaml:"moniker"`
		ID       string    `json:"id" yaml:"id"`
		Network  string    `json:"network" yaml:"network"`
		Version  string    `json:"version" yaml:"version"`
		TxIndex  bool      `json:"tx_index" yaml:"tx_index"`
		LastSync time.Time `json:"last_sync" yaml:"last_sync"`
		Location Location  `json:"location" yaml:"location"`
	}

	// Location defines geolocation information of a Tendermint node.
	Location struct {
		Country   string `json:"country" yaml:"country"`
		Region    string `json:"region" yaml:"region"`
		City      string `json:"city" yaml:"city"`
		Latitude  string `json:"latitude" yaml:"latitude"`
		Longitude string `json:"longitude" yaml:"longitude"`
	}
)

// Marshal returns the MessagePack encoding of a node.
func (n Node) Marshal() ([]byte, error) {
	bz, err := msgpack.Marshal(n)
	if err != nil {
		return nil, err
	}

	return bz, nil
}

// Unmarshal unmarshals a MessagePack encoding of a node.
func (n *Node) Unmarshal(bz []byte) error {
	if err := msgpack.Unmarshal(bz, n); err != nil {
		return err
	}

	return nil
}
