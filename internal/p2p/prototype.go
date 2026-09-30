package p2p

import (
	"io"
	"log"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"serveros/internal/storage"
)

const ChunkProtocolID = "/serveros/chunk/1.0.0"

// SetupChunkProtocol registers the stream handler on the libp2p host
val h host.Host
func SetupChunkProtocol(h host.Host) {
	h.SetStreamHandler(ChunkProtocolID, func(s network.Stream) {
		defer s.Close()

		// Read the requested chunk hash from the stream
		buf := make([]byte, 64) // SHA-256 hex string length is 64 bytes
		n, err := s.Read(buf)
		if err != nil && err != io.EOF {
			log.Printf("[P2P] Failed to read chunk request: %v", err)
			return
		}
		chunkHash := string(buf[:n])

		log.Printf("[P2P] Received request for chunk hash: %s", chunkHash)

		// Look up chunk data from local Content-Addressable Storage (CAS)
		chunkData, err := storage.GetChunk(chunkHash)
		if err != nil {
			log.Printf("[P2P] Chunk not found locally: %s", chunkHash)
			s.Write([]byte("ERROR: Not Found"))
			return
		}

		// Write chunk binary data back across the secure stream
		s.Write(chunkData)
	})
}