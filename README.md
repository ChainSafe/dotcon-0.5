# dotcon0.5 libp2p workshop

The purpose of this workshop code is to demonstrate peer discovery through the kademlia DHT, as well as opening streams and sending messages.  In this code, we show how to start a libp2p node and connect it to another libp2p node through the kademlia DHT and send a message.

### Connections and streams

To create a p2p network, we need to connect our nodes.  When we attempt to connect a node to another node, libp2p dials the other node and asks for certain information (ie. what transports they support, public key).  If the node responds in a satisfactory way, the connection is established. 

Once a connection is established, one of the nodes can open a stream to the other node.  A stream is a two-way buffer than each node can read or write to.  Streams are how nodes send messages.  A stream stays open until one side closes it.

### Creating a new libp2p host

In the libp2p world, a node is called a **host**.  Each host has its peer ID, which is a hash of it's public key and looks something like this: `Qmeq45rCLjFt573aFKgLrcAmAMSmYy9WXTuetDsELM2r8m`. 

In our code, we create a service struct which contains a host, as well as a DHT and other relevant info. It looks like this:
```
type Service struct {
	ctx            context.Context
	host           core.Host
	hostAddr       ma.Multiaddr
	dht            *kaddht.IpfsDHT
	dhtConfig      kaddht.BootstrapConfig
	bootstrapNodes []peer.AddrInfo
	noBootstrap    bool
}
```

To create a new service, we invoke `service.NewService(config)` in main.go.  Inside `NewService`, we 
create a new host using `libp2p.New()`:
```
	h, err := libp2p.New(ctx, opts...)
	if err != nil {
		return nil, err
	}
```
 
 This creates a host with the options we've specified in `config.go`, namely to listen at the addresses we've specified and create an identity using the specified private key.

To get the peer ID of the host, we can do `h.ID()`, or from a service s, `s.Host().ID()`. For more info relating to the host, see here: [libp2p host godoc](https://godoc.org/github.com/libp2p/go-libp2p-core/host). 

In `main.go`, we create 3 hosts. This is so we can connect host A to host B directly, connect host A to host C directly, then find host B from host C through the DHT.

### Bootstrapping

Bootstrapping is the process of connecting a node to a known set of other nodes. This is so that we can connect to an existing network and populate the DHT.

To connect to another node, we need to know its full multiaddress, which looks something like this:
`/ip4/192.168.0.133/tcp/5000/ipfs/QmTPAVze7YxdrSJL1vb7SfsoAPrCRmomwSbRKH8kpnGugK`.  It contains the IP, transport, port, and protocol of the node, as well as its peer ID.  

In `main.go`, when we create node A, we can gets its multiaddress string by getting its listening address, knowing the protocol ("ipfs"), and getting its peer ID: `fmt.Sprintf("%s/ipfs/%s", srvcA.Host().Addrs()[1].String(), srvcA.Host().ID())`  We then pass it into the bootstrapNodes array of our config.  Using `stringToPeerInfo` in `utils.go`, it gets turned into a peer.AddrInfo which is what is needed to connect directly to the other node.

To connect, we do `service.Host().Connect(s.ctx, p)`. Upon success, this creates a direct connection with the peer.

### Instantiating the DHT

In `NewService`, after creating a new host, we also create a new DHT. First, we use go-ipfs's [datastore](github.com/ipfs/go-datastore) to create a new datastore.  Then, we make a [new libp2p kad-dht](https://godoc.org/github.com/libp2p/go-libp2p-kad-dht#NewDHT) passing in the datastore.

Finally, we [wrap the host with the DHT](https://godoc.org/github.com/libp2p/go-libp2p/p2p/host/routed#Wrap) so that peer discovery is routed through the DHT. 

At this point, the host and DHT are connected, but there isn't anything in the DHT. After we bootstrap to known nodes, we can bootstrap the DHT. In `service.start()`, after doing `bootstrapConnect` which connects us to known nodes, we do `s.dht.Bootstrap(s.ctx)`.

Full DHT docs can be found [here](https://godoc.org/github.com/libp2p/go-libp2p-kad-dht).

### Kademlia DHT discovery

To discover another node, we need to know its peer ID. In `main.go`, we use the peer ID of node B to discover it: `peerB, _ := srvcC.DHT().FindPeer(srvcC.Ctx(), srvcB.Host().ID())`.  This gives us the full peer.AddrInfo of node B.

Now, knowing the AddrInfo, we can connect: `srvcC.Host().Connect(srvcC.Ctx(), peerB)`. Now they are directly connected, yay! From this point on, we can open a stream using `host.NewStream` with the peer and send them bytes.

Note: we don't need to explicitly connect to the peer before opening a stream, since opening a stream implicitly calls connect.

### Issues

You may get `could not find peerB="routing: not found"`, if this happens, please re-run the code. 

### Running the code

```go run main.go```

### More resources

Check out the [libp2p godocs](https://godoc.org/github.com/libp2p/go-libp2p-core) for details on the API.

Here's the [official documentation](https://docs.libp2p.io/) for libp2p.

If you have specific questions, head over to the [libp2p forums](https://discuss.libp2p.io).

For more in-depth info on how the kademlia DHT works, check out the [paper](https://pdos.csail.mit.edu/~petar/papers/maymounkov-kademlia-lncs.pdf)