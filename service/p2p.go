package service

import (
	"bufio"
	"context"
	"errors"
	"fmt"

	log "github.com/ChainSafe/log15"
	ds "github.com/ipfs/go-datastore"
	libp2p "github.com/libp2p/go-libp2p"
	core "github.com/libp2p/go-libp2p-core"
	host "github.com/libp2p/go-libp2p-core/host"
	net "github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	kaddht "github.com/libp2p/go-libp2p-kad-dht"
	rhost "github.com/libp2p/go-libp2p/p2p/host/routed"
	ma "github.com/multiformats/go-multiaddr"
)

const ProtocolPrefix = "/dotcon/0.5"

// Service describes a p2p service, including host and dht
type Service struct {
	ctx            context.Context
	host           core.Host
	hostAddr       ma.Multiaddr
	dht            *kaddht.IpfsDHT
	bootstrapNodes []peer.AddrInfo
	noBootstrap    bool
}

// NewService creates a new p2p.Service using the service config. It initializes the host and dht
func NewService(conf *Config) (*Service, error) {
	ctx := context.Background()
	opts, err := conf.buildOpts()
	if err != nil {
		return nil, err
	}
	
	h, err := libp2p.New(ctx, opts...)
	if err != nil {
		return nil, err
	}

	dstore := ds.NewMapDatastore()
	dht := kaddht.NewDHT(ctx, h, dstore)

	// wrap the host with routed host so we can look up peers in DHT
	h = rhost.Wrap(h, dht)

	h.SetStreamHandler(ProtocolPrefix, handleStream)

	bootstrapNodes, err := stringsToPeerInfos(conf.BootstrapNodes)
	s := &Service{
		ctx:            ctx,
		host:           h,
		dht:            dht,
		bootstrapNodes: bootstrapNodes,
		noBootstrap:    conf.NoBootstrap,
	}
	return s, err
}

// Start begins the p2p Service, including discovery
func (s *Service) Start() (<-chan bool, <-chan error) {
	e := make(chan error)
	done := make(chan bool)
	go s.start(done, e)
	return done, e
}

// start begins the p2p Service, including discovery. start does not terminate once called.
func (s *Service) start(done chan bool, e chan error) {
	if len(s.bootstrapNodes) == 0 && !s.noBootstrap {
		e <- errors.New("no peers to bootstrap to")
	}

	if !s.noBootstrap {
		// connect to the bootstrap nodes
		err := s.bootstrapConnect()
		if err != nil {
			e <- err
		}
	}

	err := s.dht.Bootstrap(s.ctx)
	if err != nil {
		e <- err
	}

	hostAddr, err := ma.NewMultiaddr(fmt.Sprintf("/p2p/%s", s.host.ID().Pretty()))
	if err != nil {
		log.Error("start", "error", err)
	}

	// Now we can build a full multiaddress to reach this host
	// by encapsulating both addresses:
	addrs := s.host.Addrs()
	for _, addr := range addrs {
		log.Info("address can be reached", "hostAddr", addr.Encapsulate(hostAddr))
	}

	log.Info("listening for connections...")
	//e <- nil
	done <- true
}

// Stop stops the p2p service
func (s *Service) Stop() <-chan error {
	e := make(chan error)

	//Stop the host & IpfsDHT
	err := s.host.Close()
	if err != nil {
		e <- err
	}

	err = s.dht.Close()
	if err != nil {
		e <- err
	}

	return e
}

// Send sends a message to a specific peer
func (s *Service) Send(peer peer.AddrInfo, msg []byte) (err error) {
	log.Info("sending stream", "to", peer.ID, "msg", fmt.Sprintf("0x%x", msg))

	stream := s.getExistingStream(peer.ID)
	if stream == nil {
		stream, err = s.host.NewStream(s.ctx, peer.ID, ProtocolPrefix)
		log.Info("stream", "opening new stream to peer", peer.ID)
		if err != nil {
			log.Error("new stream", "error", err)
			return err
		}
	} else {
		log.Info("stream", "using existing stream for peer", peer.ID)
	}

	_, err = stream.Write(msg)
	if err != nil {
		log.Error("sending stream", "error", err)
		return err
	}

	return nil
}

func (s *Service) getExistingStream(p peer.ID) net.Stream {
	conns := s.host.Network().ConnsToPeer(p)
	for _, conn := range conns {
		streams := conn.GetStreams()
		for _, stream := range streams {
			if stream.Protocol() == ProtocolPrefix {
				return stream
			}
		}
	}

	return nil
}

func handleStream(stream net.Stream) {
	log.Info("stream handler", "got stream from", stream.Conn().RemotePeer())

	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
	msg, err := rw.Reader.ReadString('\n')
	if err != nil {
		log.Error("stream handler", "got stream from", stream.Conn().RemotePeer(), "err", err)
		return
	}

	fmt.Printf("got message: %s", msg)
}

func (s *Service) Host() host.Host {
	return s.host
}

// DHT returns the service's dht
func (s *Service) DHT() *kaddht.IpfsDHT {
	return s.dht
}

// Ctx returns the service's ctx
func (s *Service) Ctx() context.Context {
	return s.ctx
}
