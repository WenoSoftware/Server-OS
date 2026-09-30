package p2p

import (
	"context"
	"fmt"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	dhtopts "github.com/libp2p/go-libp2p-kad-dht/opts"
	"github.com/libp2p/go-libp2p/p2p/discovery/routing"
	dutil "github.com/libp2p/go-libp2p/p2p/discovery/util"
)

// InitP2PNode initializes a libp2p host with Kademlia DHT routing enabled
func InitP2PNode() (host.Host, error) {
	// 1. Create standard libp2p host
	h, err := libp2p.New(
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0"),
	)
	if err != nil {
		return nil, err
	}

	// 2. Start the Kademlia DHT in client/server mode
	ctx := context.Background()
	kademliaDHT, err := dht.New(ctx, h, dhtopts.Mode(dht.ModeServer))
	if err != nil {
		return nil, fmt.Errorf("failed to start Kademlia DHT: %v", err)
	}

	// 3. Bootstrap the DHT (connects to default IPFS/libp2p bootstrap nodes)
	if err := kademliaDHT.Bootstrap(ctx); err != nil {
		return nil, fmt.Errorf("failed to bootstrap DHT: %v", err)
	}

	// 4. Set up routing discovery so peers can find each other via the DHT
	routingDiscovery := routing.NewRoutingDiscovery(kademliaDHT)
	dutil.Advertise(ctx, routingDiscovery, "weno-server-os-net")

	return h, nil
}