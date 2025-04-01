package broker

import (
	"context"
	"log"
	"net"

	"github.com/HongikAPL/new-wowsan/pkg/proto"
	"google.golang.org/grpc"
)

// BrokerServer는 BrokerService를 구현한 서버 구조체입니다.
// 이 서버는 브로커 노드의 역할을 수행하며, 다른 브로커 노드와 통신하고 구독자에게 메시지를 전달합니다.
type BrokerServer struct {
	proto.UnimplementedBrokerServiceServer
	broker *Broker
}

// StartBrokerServer는 새로운 브로커 서버를 시작하는 함수입니다.
// 주어진 브로커 객체를 사용하여 gRPC 서버를 설정하고 실행합니다.
func StartBrokerServer(b *Broker) {
	lis, err := net.Listen("tcp", ":"+b.NodeInfo.Port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	service := &BrokerServer{broker: b}
	proto.RegisterBrokerServiceServer(grpcServer, service)

	log.Printf("BrokerService listening on port %s", b.NodeInfo.Port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

// AddBroker는 새로운 브로커 노드를 추가하는 함수입니다.
// 주어진 노드 정보를 사용하여 브로커 구조체에 새로운 브로커를 추가합니다.
func (s *BrokerServer) AddBroker(ctx context.Context, req *proto.AddBrokerRequest) (*proto.AddBrokerResponse, error) {
	// mu 락 추가 필요?
	if _, exists := s.broker.Peers[req.Node.Id]; exists {

		return &proto.AddBrokerResponse{Success: false, ErrorMessage: "Broker already exists"}, nil
	}
	s.broker.AddBroker(req.Node)

	log.Printf("Added peer from request: %s", req.Node.Id)
	return &proto.AddBrokerResponse{Success: true}, nil
}

// SendSubscription은 새로운 구독을 추가하는 함수입니다.
// 주어진 노드 정보와 토픽을 사용하여 브로커 구조체에 새로운 구독을 추가합니다.
// brokerclient.go의 SendSubscription 함수를 호출해야함. (이웃에게 전파)
func (s *BrokerServer) SendSubscription(ctx context.Context, req *proto.SendSubscriptionRequest) (*proto.SendSubscriptionResponse, error) {
	s.broker.mu.Lock()
	defer s.broker.mu.Unlock()

	// 구독자 추가
	s.broker.Subscriptions = append(s.broker.Subscriptions, Subscription{
		Node:  req.Node,
		Topic: req.Topic,
	})

	for _, peer := range s.broker.Peers {
		if peer.Id != req.Node.Id {
			go func(peer *proto.NodeInfo) {
				client, err := NewBrokerClient(peer)
				if err != nil {
					log.Printf("Failed to connect to peer %s: %v", peer.Id, err)
					return
				}
				defer client.Close()

				newReq := &proto.SendSubscriptionRequest{
					Node:  s.broker.NodeInfo,
					Topic: req.Topic,
				}

				_, err = client.SendSubscription(context.Background(), newReq)
				if err != nil {
					log.Printf("Failed to send subscription to peer %s: %v", peer.Id, err)
				}
			}(peer)
		}
	}

	return &proto.SendSubscriptionResponse{Success: true}, nil
}

// SendUnsubscription은 구독을 취소하는 함수입니다.
// 주어진 노드 정보와 구독 ID를 사용하여 브로커 구조체에서 구독을 취소합니다.
// brokerclient.go의 SendUnsubscription 함수를 호출해야함.
func (s *BrokerServer) SendUnsubscription(ctx context.Context, req *proto.SendUnsubscriptionRequest) (*proto.SendUnsubscriptionResponse, error) {
	s.broker.mu.Lock()
	defer s.broker.mu.Unlock()

	// 구독 취소
	s.broker.SendUnsubscription(req.Node, req.SubscriptionId)

	// 구독 취소 정보 전파
	for _, peer := range s.broker.Peers {
		if peer.Id != req.Node.Id {
			go func(peer *proto.NodeInfo) {
				client, err := NewBrokerClient(peer)
				if err != nil {
					log.Printf("Failed to connect to peer %s: %v", peer.Id, err)
					return
				}
				defer client.Close()

				newReq := &proto.SendUnsubscriptionRequest{
					Node:           s.broker.NodeInfo,
					SubscriptionId: req.SubscriptionId,
				}

				_, err = client.SendUnsubscription(context.Background(), newReq)
				if err != nil {
					log.Printf("Failed to send unsubscription to peer %s: %v", peer.Id, err)
				}
			}(peer)
		}
	}

	return &proto.SendUnsubscriptionResponse{Success: true}, nil
}

// SendPublication은 새로운 메시지를 전파하는 함수입니다.
// 주어진 노드 정보와 토픽을 사용하여 브로커 구조체에서 메시지를 전파합니다.
// brokerclient.go의 SendPublication 함수를 호출해야함.
// 전체 피어가 아닌, 구독정보가 있는 애들한테만 전파하면 됨.
func (s *BrokerServer) SendPublication(ctx context.Context, req *proto.SendPublicationRequest) (*proto.SendPublicationResponse, error) {
	s.broker.mu.RLock()
	defer s.broker.mu.RUnlock()

	// 메시지 전파
	for _, sub := range s.broker.Subscriptions {
		if sub.Topic == req.Topic {
			go func(subscriber *proto.NodeInfo) {
				client, err := NewBrokerClient(subscriber)
				if err != nil {
					log.Printf("Failed to connect to subscriber %s: %v", subscriber.Id, err)
					return
				}
				defer client.Close()

				newReq := &proto.SendPublicationRequest{
					Node:    s.broker.NodeInfo,
					Topic:   req.Topic,
					Content: req.Content,
				}

				_, err = client.SendPublication(context.Background(), newReq)
				if err != nil {
					log.Printf("Failed to send publication to subscriber %s: %v", subscriber.Id, err)
				}
			}(sub.Node)
		}
	}

	return &proto.SendPublicationResponse{Success: true}, nil
}
