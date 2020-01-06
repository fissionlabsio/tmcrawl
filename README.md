# tmcrawl

> A Tendermint p2p crawling utility and API.

The `tmcrawl` utility will capture geolocation information and node metadata such as network
name, node version, RPC information, and node ID for each crawled node. The utility
will first start with a set of seeds and attempt to crawl as many nodes as possible
from those seeds. When there are no nodes left to crawl, `tmcrawl` will pick a random
node from the known list of nodes to reseed the crawl every `crawl_interval` seconds
from the last attempted crawl finish.

Nodes will also be periodically checked every `recheck_interval`. If any node cannot
be reached, it'll be removed from the known set of nodes.

Note, `tmcrawl` is a Tendermint p2p network crawler, it does not operate as a seed
node or any other type of node. However, it can be used to gather a set of peers.

## Install

`tmcrawl` takes a simple configuration. It needs to only know about a an
[ipstack](https://ipstack.com/) API access key and an initial set of seed nodes.
See `config.toml` for reference.

To install the binary:

```shell
$ make install
```

**Note**: Requires [Go 1.13+](https://golang.org/dl/)

## Usage

`tmcrawl` runs as a daemon process and exposes a RESTful JSON API.

To start the binary:

```shell
$ tmcrawl </path/to/config.toml> [flags]
```

The RESTful JSON API is served over `listen_addr` as provided in the configuration.
See `--help` for further documentation.

## API

All API documentation is hosted via Swagger UI under path `/swagger/`.

## Future Improvements

- Unit and integration tests
- Support `recheck_interval`
- Front-end visualization

## Contributing

Contributions are welcome! Please open an Issues or Pull Request for any changes.

## License

[CCC0 1.0 Universal](https://creativecommons.org/share-your-work/public-domain/cc0/)
