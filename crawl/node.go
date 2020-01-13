package crawl

import (
	"github.com/vmihailenco/msgpack/v4"
)

// Node persistence prefix keys
var (
	NodeKeyPrefix     = []byte("node/")
	LocationKeyPrefix = []byte("location/")
)

type (
	// Node represents a full-node in a Tendermint-based network that contains
	// relevant p2p data.
	Node struct {
		Address  string   `json:"address" yaml:"address"`
		RPCPort  string   `json:"rpc_port" yaml:"rpc_port"`
		P2PPort  string   `json:"p2p_port" yaml:"p2p_port"`
		Moniker  string   `json:"moniker" yaml:"moniker"`
		ID       string   `json:"id" yaml:"id"`
		Network  string   `json:"network" yaml:"network"`
		Version  string   `json:"version" yaml:"version"`
		TxIndex  string   `json:"tx_index" yaml:"tx_index"`
		LastSync string   `json:"last_sync" yaml:"last_sync"`
		Location Location `json:"location" yaml:"location"`
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

// Key returns the addressable persistence key of a Node.
func (n Node) Key() []byte {
	return NodeKey(n.Address)
}

// Marshal returns the MessagePack encoding of a Node.
func (n Node) Marshal() ([]byte, error) {
	bz, err := msgpack.Marshal(n)
	if err != nil {
		return nil, err
	}

	return bz, nil
}

// Unmarshal unmarshals a MessagePack encoding of a Node.
func (n *Node) Unmarshal(bz []byte) error {
	if err := msgpack.Unmarshal(bz, n); err != nil {
		return err
	}

	return nil
}

// Marshal returns the MessagePack encoding of a Location.
func (l Location) Marshal() ([]byte, error) {
	bz, err := msgpack.Marshal(l)
	if err != nil {
		return nil, err
	}

	return bz, nil
}

// Unmarshal unmarshals a MessagePack encoding of a Location.
func (l *Location) Unmarshal(bz []byte) error {
	if err := msgpack.Unmarshal(bz, l); err != nil {
		return err
	}

	return nil
}

// NodeKey constructs the DB key for node persistence.
func NodeKey(addressable string) []byte {
	return append(NodeKeyPrefix, []byte(addressable)...)
}

// LocationKey constructs the DB key for location persistence/caching.
func LocationKey(addressable string) []byte {
	return append(LocationKeyPrefix, []byte(addressable)...)
}
