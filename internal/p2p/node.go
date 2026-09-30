package p2p

import (
	"log"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
)

// InitP2PNode initializes a new libp2p host node
func InitP2PNode() (host.Host, error) {
	// Create a new libp2p Host with default options (random port, TCP/QUIC)
	h, err := libp2p.New(
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0"),
	)
	if err != nil {
		return nil, err
	}

	log.Printf("[P2P] Node initialized with ID: %s", h.ID().String())
	return h, nil
}