package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Configuration sentinel errors
var (
	ErrEmptyConfigPath = errors.New("empty configuration file path")
	ErrNoIPStackKey    = errors.New("no ipstack access key provided in configuration")
	ErrNoSeeds         = errors.New("no seeds provided in configuration")
)

var (
	defaultListenAddr           = "0.0.0.0:27758"
	defaultCrawlInterval   uint = 15
	defaultRecheckInterval uint = 3600
	defaultReseedSize      uint = 100
)

// Config defines all necessary tmcrawl configuration parameters.
type Config struct {
	DataDir    string   `toml:"data_dir"`
	ListenAddr string   `toml:"listen_addr"`
	Seeds      []string `toml:"seeds"`
	ReseedSize uint     `toml:"reseed_size"`
	IPStackKey string   `toml:"ipstack_key"`

	CrawlInterval   uint `toml:"crawl_interval"`
	RecheckInterval uint `toml:"recheck_interval"`
}

// ParseConfig attempts to read and parse a tmcrawl config from the given file
// path. An error is returned if reading or parsing the config fails.
func ParseConfig(configPath string) (Config, error) {
	var cfg Config

	if configPath == "" {
		return cfg, ErrEmptyConfigPath
	}

	configData, err := ioutil.ReadFile(configPath)
	if err != nil {
		return cfg, fmt.Errorf("failed to read config: %w", err)
	}

	if _, err := toml.Decode(string(configData), &cfg); err != nil {
		return cfg, fmt.Errorf("failed to decode config: %w", err)
	}

	if len(cfg.Seeds) == 0 {
		return cfg, ErrNoSeeds
	}
	if cfg.IPStackKey == "" {
		return cfg, ErrNoIPStackKey
	}
	if cfg.ListenAddr == "" {
		cfg.ListenAddr = defaultListenAddr
	}
	if cfg.ReseedSize == 0 {
		cfg.ReseedSize = defaultReseedSize
	}
	if cfg.CrawlInterval == 0 {
		cfg.CrawlInterval = defaultCrawlInterval
	}
	if cfg.RecheckInterval == 0 {
		cfg.RecheckInterval = defaultRecheckInterval
	}
	if cfg.DataDir == "" {
		cfg.DataDir = filepath.Join(os.Getenv("HOME"), ".tmcrawl")
	}

	return cfg, nil
}
