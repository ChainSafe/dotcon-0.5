package main

import (
	"fmt"
	"os"
	"time"

	log "github.com/ChainSafe/log15"

	"github.com/ansermino/dotcon0.5/service"
)

// This program, once finished, should start 3 nodes A, B, and C, and connect B->A and C->A
// then try to discover B through node C using a DHT.
// Once nodes B and C are connected, node C sends node B a message.
//
// TODO: finish implementation of NewService in service/p2p.go
// You will see the lines that need to be done marked with TODO with links to relevant godocs.
func main() {
	log.Info("Starting DHT client...")
	log.New()

	// create service A
	cfgA := &service.Config{
		BootstrapNodes: nil,
		Port:           5000,
		RandSeed:       0,
		NoBootstrap:    true,
	}

	srvcA, err := service.NewService(cfgA)
	if err != nil {
		log.Error("error starting service A", "err", err)
		os.Exit(1)
	}

	// form full address of service A
	srvcAaddr := fmt.Sprintf("%s/ipfs/%s", srvcA.Host().Addrs()[1].String(), srvcA.Host().ID())

	// create service B, bootstrap it to service A
	cfgB := &service.Config{
		BootstrapNodes: []string{srvcAaddr},
		Port:           5001,
		RandSeed:       1,
		NoBootstrap:    false,
	}

	srvcB, err := service.NewService(cfgB)
	if err != nil {
		log.Error("error starting service B", "err", err)
		os.Exit(1)
	}

	// create service C, bootstrap it to service A but not service B
	cfgC := &service.Config{
		BootstrapNodes: []string{srvcAaddr},
		Port:           5002,
		RandSeed:       1,
		NoBootstrap:    false,
	}

	srvcC, err := service.NewService(cfgC)
	if err != nil {
		log.Error("error starting service C", "err", err)
		os.Exit(1)
	}

	// uncomment this block of code when you think you're done implementing NewService :)

	// // start all services
	// done, _ := srvcA.Start()
	// defer srvcA.Stop()
	// <-done

	// done, _ = srvcB.Start()
	// defer srvcB.Stop()
	// <-done

	// done, _ = srvcC.Start()
	// defer srvcC.Stop()
	// <-done

	// // timeout, just to make sure everything is connected
	// time.Sleep(5 * time.Second)

	// // find node B's peer info through node C's DHT
	// peerB, err := srvcC.DHT().FindPeer(srvcC.Ctx(), srvcB.Host().ID())
	// if err != nil {
	// 	log.Error("find peer", "could not find peerB", err)
	// 	os.Exit(1)
	// }

	// // connect to peer B from peer C
	// err = srvcC.Host().Connect(srvcC.Ctx(), peerB)
	// if err != nil {
	// 	log.Error("connect", "could not connect to peerB", err)
	// 	os.Exit(1)
	// }

	// // send a message
	// msg := []byte("hello friend \n")
	// err = srvcC.Send(peerB, msg)
	// if err != nil {
	// 	log.Error("send", "error", err)
	// }

	// // continue to run the program (ie. wait until message is sent)
	// // ctrl+c to exit
	// select {}
}
