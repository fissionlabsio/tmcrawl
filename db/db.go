package db

import (
	"bytes"
	"fmt"
	"path/filepath"

	"github.com/boltdb/bolt"
)

var (
	nodesBucket = []byte("NODES")
)

type (
	// DB defines the persistence interface for tmcrawl.
	DB interface {
		Get(key []byte) ([]byte, error)
		Has(key []byte) bool
		Set(key, value []byte) error
		Delete(key []byte) error
		IteratePrefix(prefix []byte, cb func(k, v []byte) bool)
		Close() error
	}

	// BoltDB defines a wrapper type around a Bolt DB that implements the DB
	// interface. It mainly provides transaction abstractions.
	BoltDB struct {
		db *bolt.DB
	}
)

// NewBoltDB returns a wrapper around a Bolt DB that implements the DB interface.
// It will create all the necessary Bolt DB buckets if they don't already exist.
func NewBoltDB(dataDir, dbName string, dbOpts *bolt.Options) (DB, error) {
	dbPath := filepath.Join(dataDir, dbName)
	db, err := bolt.Open(dbPath, 0600, dbOpts)
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(nodesBucket)
		if err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}

		return nil
	})

	return &BoltDB{db: db}, err
}

// Get returns a value for a given key. An error will never be returned as
// gauranteed by the Bolt DB semantics.
func (bdb *BoltDB) Get(key []byte) (value []byte, err error) {
	err = bdb.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(nodesBucket)
		v := b.Get(key)

		copy(value, v)
		return nil
	})

	return value, err
}

// Has returns a boolean determining if the underlying Bolt DB has a given key
// or not.
func (bdb *BoltDB) Has(key []byte) bool {
	v, err := bdb.Get(key)
	return v != nil && err == nil
}

// Set attempts to set a key/value pair into Bolt DB returning an error upon
// failure.
func (bdb *BoltDB) Set(key, value []byte) error {
	return bdb.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(nodesBucket)
		return b.Put(key, value)
	})
}

// Delete attempts to remove a value by key from Bolt DB returning an error
// upon failure.
func (bdb *BoltDB) Delete(key []byte) error {
	return bdb.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(nodesBucket)
		return b.Delete(key)
	})
}

// IteratePrefix iterates over a series of key/value pairs where each key contains
// the provided prefix. For each key/value pair, a cb function is invoked. If
// cb returns true, iteration is halted.
func (bdb *BoltDB) IteratePrefix(prefix []byte, cb func(k, v []byte) bool) {
	_ = bdb.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(nodesBucket)
		c := b.Cursor()

		for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
			if cb(k, v) {
				return nil
			}
		}

		return nil
	})
}

// Close closes the Bolt DB instance and returns an error upon failure.
func (bdb *BoltDB) Close() error {
	return bdb.db.Close()
}
