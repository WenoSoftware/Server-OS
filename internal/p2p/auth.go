package p2p

import (
	"fmt"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
)

type SignedPayload struct {
	Data      []byte `json:"data"`
	Signature []byte `json:"signature"`
	SenderID  string `json:"sender_id"`
}

// SignData signs arbitrary data using the host's private key
func SignData(privKey crypto.PrivKey, data []byte) ([]byte, error) {
	return privKey.Sign(data)
}

// VerifyData checks if a signature matches the sender's public key & peer ID
func VerifyData(senderIDStr string, data []byte, signature []byte) (bool, error) {
	peerID, err := peer.Decode(senderIDStr)
	if err != nil {
		return false, fmt.Errorf("invalid peer ID: %v", err)
	}

	pubKey, err := peerID.ExtractPublicKey()
	if err != nil || pubKey == nil {
		return false, fmt.Errorf("could not extract public key for peer: %v", err)
	}

	return pubKey.Verify(data, signature)
}