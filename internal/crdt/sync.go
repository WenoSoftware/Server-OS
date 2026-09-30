package crdt

import (
	"context"
	"log"
	"time"

	"serveros/internal/p2p"
)

const SyncTopic = "serveros-crdt-sync-v1"

// StartCRDTSync initializes local state, subscribes to updates, and broadcasts changes
func StartCRDTSync(ctx context.Context, psManager *p2p.PubSubManager) *DistributedState {
	state := NewState()

	// 1. Handle incoming state updates from the swarm
	err := psManager.JoinTopic(ctx, SyncTopic, func(data []byte, sender string) {
		if err := state.Merge(data); err != nil {
			log.Printf("[CRDT] Failed to merge state from peer %s: %v", sender[:8], err)
			return
		}
		log.Printf("[CRDT] Successfully merged state update from peer: %s", sender[:8])
	})
	if err != nil {
		log.Fatalf("[CRDT] Failed to join sync topic: %v", err)
	}

	// 2. Periodically broadcast local state to the swarm (e.g., every 5 seconds)
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				payload, err := state.Export()
				if err != nil {
					log.Printf("[CRDT] Failed to export state for broadcast: %v", err)
					continue
				}

				if err := psManager.Broadcast(ctx, SyncTopic, payload); err != nil {
					// Safe to log and ignore temporary network partition drops
					log.Printf("[CRDT] Broadcast tick skipped (no peers connected yet?): %v", err)
				}
			}
		}
	}()

	log.Println("[CRDT] State synchronization background worker active.")
	return state
}