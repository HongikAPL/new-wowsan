package broker

import (
	"log"
	"sync"

	"github.com/HongikAPL/new-wowsan/pkg/proto"
)

type Subscription struct {
	Node  *proto.NodeInfo
	Topic string
}

type Broker struct {
	NodeInfo      *proto.NodeInfo
	Peers         map[string]*proto.NodeInfo
	Subscriptions []Subscription
	mu            sync.RWMutex
}

func NewBroker(nodeInfo *proto.NodeInfo) *Broker {
	return &Broker{
		NodeInfo:      nodeInfo,
		Peers:         make(map[string]*proto.NodeInfo),
		Subscriptions: []Subscription{},
	}
}

func (b *Broker) AddBroker(node *proto.NodeInfo) {

	b.Peers[node.Id] = node
}

func (b *Broker) SendSubscription(node *proto.NodeInfo, topic string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, sub := range b.Subscriptions {
		if sub.Node.Id == node.Id && sub.Topic == topic {
			return
		}
	}

	b.Subscriptions = append(b.Subscriptions, Subscription{
		Node:  node,
		Topic: topic,
	})
}

func (b *Broker) SendUnsubscription(node *proto.NodeInfo, topic string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for i, sub := range b.Subscriptions {
		if sub.Node.Id == node.Id && sub.Topic == topic {
			b.Subscriptions = append(b.Subscriptions[:i], b.Subscriptions[i+1:]...)
			return
		}
	}
}

func (b *Broker) SendPublication(node *proto.NodeInfo, topic string, content string) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for _, sub := range b.Subscriptions {
		if sub.Topic == topic {
			go func(subscriber *proto.NodeInfo) {
				println("Sending message to", subscriber.Id, ":", content)
			}(sub.Node)
		}
	}
}

// 피어 출력
func (b *Broker) PrintPeers() {

	log.Println("Current peers:")
	for id, peer := range b.Peers {
		log.Printf("Peer ID: %s, IP: %s, Port: %s", id, peer.Ip, peer.Port)
	}
}
