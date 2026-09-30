package p2p

import (
	"context"
	"fmt"
	"log"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
)

type PubSubManager struct {
	PubSub *pubsub.PubSub
	Topics map[string]*pubsub.Topic
	Host   host.Host
}

// InitPubSub initializes GossipSub with peer exchange enabled
func InitPubSub(ctx context.Context, h host.Host) (*PubSubManager, error) {
	ps, err := pubsub.NewGossipSub(ctx, h, pubsub.WithPeerExchange(true))
	if err != nil {
		return nil, err
	}

	return &PubSubManager{
		PubSub: ps,
		Topics: make(map[string]*pubsub.Topic),
		Host:   h,
	}, nil
}

// JoinTopic joins a PubSub topic and handles incoming messages
func (psm *PubSubManager) JoinTopic(ctx context.Context, topicName string, onMessage func([]byte, string)) error {
	topic, err := psm.PubSub.Join(topicName)
	if err != nil {
		return err
	}
	psm.Topics[topicName] = topic

	sub, err := topic.Subscribe()
	if err != nil {
		return err
	}

	go func() {
		for {
			msg, err := sub.Next(ctx)
			if err != nil {
				log.Printf("[PubSub] Error reading from topic %s: %v", topicName, err)
				break
			}

			// Ignore messages sent by ourselves
			if msg.GetFrom() == psm.Host.ID() {
				continue
			}

			senderID := msg.GetFrom()
			log.Printf("[PubSub Auth] Verified authentic message from Peer: %s", senderID.String()[:12])
			onMessage(msg.Data, senderID.String())
		}
	}()

	log.Printf("[PubSub] Successfully joined secure topic: %s", topicName)
	return nil
}

// Broadcast sends a message across the topic swarm
func (psm *PubSubManager) Broadcast(ctx context.Context, topicName string, data []byte) error {
	topic, exists := psm.Topics[topicName]
	if !exists {
		return fmt.Errorf("not joined to topic: %s", topicName)
	}

	return topic.Publish(ctx, data)
}