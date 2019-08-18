package service

import (
	"fmt"
	"github.com/libp2p/go-libp2p"
	ma "github.com/multiformats/go-multiaddr"
)

// Config is used to configure a p2p service
type Config struct {
	BootstrapNodes []string
	Port           int
	RandSeed       int64
	NoBootstrap    bool
}

func (sc *Config) buildOpts() ([]libp2p.Option, error) {
	ip := "0.0.0.0"

	priv, err := generateKey(sc.RandSeed)
	if err != nil {
		return nil, err
	}

	addr, err := ma.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%d", ip, sc.Port))
	if err != nil {
		return nil, err
	}

	return []libp2p.Option{
		libp2p.ListenAddrs(addr),
		libp2p.Identity(priv),
		libp2p.NATPortMap(),
	}, nil
}
