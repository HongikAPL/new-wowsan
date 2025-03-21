package broker

import (
	"sync"

	"github.com/igeon510/new-wowsan/pkg/proto"
)

type Subscription struct {
	Node  *proto.NodeInfo
	Topic string
}

type Broker struct {
	ID            string
	IP            string
	Port          string
	Peers         map[string]*proto.NodeInfo
	Subscriptions []Subscription
	mu            sync.RWMutex
}

func NewBroker(id, ip, port string) *Broker {
	return &Broker{
		ID:            id,
		IP:            ip,
		Port:          port,
		Peers:         make(map[string]*proto.NodeInfo),
		Subscriptions: []Subscription{},
	}
}

func (b *Broker) AddBroker(node *proto.NodeInfo) {
	b.mu.Lock()
	defer b.mu.Unlock()
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
