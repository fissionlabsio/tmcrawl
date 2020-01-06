package crawl

import (
	"fmt"
	"net/url"
	"time"

	"github.com/harwoeck/ipstack"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
	libclient "github.com/tendermint/tendermint/rpc/lib/client"
)

var (
	clientTimeout = 2 * time.Second
)

func newRPCClient(remote string) *rpcclient.HTTP {
	httpClient := libclient.DefaultHTTPClient(remote)
	httpClient.Timeout = clientTimeout
	return rpcclient.NewHTTPWithClient(remote, "/websocket", httpClient)
}

func parsePort(nodeAddr string) string {
	u, err := url.Parse(nodeAddr)
	if err != nil {
		return ""
	}

	return u.Port()
}

func parseHostname(nodeAddr string) string {
	u, err := url.Parse(nodeAddr)
	if err != nil {
		return ""
	}

	return u.Hostname()
}

func locationFromIPResp(r *ipstack.Response) Location {
	return Location{
		Country:   r.CountryName,
		Region:    r.RegionName,
		City:      r.City,
		Latitude:  fmt.Sprintf("%f", r.Latitude),
		Longitude: fmt.Sprintf("%f", r.Longitude),
	}
}
