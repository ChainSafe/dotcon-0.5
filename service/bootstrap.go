package service

import (
	"errors"
	"fmt"
	"sync"

	log "github.com/ChainSafe/log15"

	"github.com/libp2p/go-libp2p-core/peer"
	ps "github.com/libp2p/go-libp2p-core/peerstore"
)

// this code is borrowed from the go-ipfs bootstrap process
func (s *Service) bootstrapConnect() error {
	peers := s.bootstrapNodes
	if len(peers) < 1 {
		return errors.New("not enough bootstrap peers")
	}

	// begin bootstrapping
	errs := make(chan error, len(peers))
	var wg sync.WaitGroup

	var err error
	for _, p := range peers {

		// performed asynchronously because when performed synchronously, if
		// one `Connect` call hangs, subsequent calls are more likely to
		// fail/abort due to an expiring context.

		wg.Add(1)
		go func(p peer.AddrInfo) {
			defer wg.Done()
			log.Info("bootstrap attempt", "host", s.host.ID(), "peer", p.ID)

			s.host.Peerstore().AddAddrs(p.ID, p.Addrs, ps.PermanentAddrTTL)
			if err = s.host.Connect(s.ctx, p); err != nil {
				log.Error("bootstrap error", "peer", p.ID, "error", err)
				errs <- err
				return
			}
			log.Info("bootstrap success", "peer", p.ID)
		}(p)
	}
	wg.Wait()

	// our failure condition is when no connection attempt succeeded.
	// drain the errs channel, counting the results.
	close(errs)
	count := 0
	for err = range errs {
		if err != nil {
			count++
		}
	}
	if count == len(peers) {
		return fmt.Errorf("failed to bootstrap. %s", err)
	}
	return err
}
