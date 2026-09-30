package p2p

import (
	"context"
	"fmt"
	"log"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
)

// InitP2PNode boots up a local libp2p network host for Server OS
func InitP2PNode() (host.Host, error) {
	ctx := context.Background()

	// Create a new libp2p Host with default security and transport layers
	h, err := libp2p.New(
		libp2p.ListenAddrStrings(
			"/ip4/0.0.0.0/tcp/0", // Listen on any available TCP port globally
		),
	)
	if err != nil {
		return nil, err
	}

	log.Printf("[ServerOS P2P] Node booted successfully!")
	log.Printf("[ServerOS P2P] Local Peer ID: %s", h.ID())
	
	// Print multiaddresses so other peers know how to connect
	for _, addr := range h.Addrs() {
		log.Printf("[ServerOS P2P] Listening on: %s/p2p/%s", addr, h.ID())
	}

	return h, nil
}