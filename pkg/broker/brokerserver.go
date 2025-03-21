package broker

import (
	"context"
	"log"
	"net"

	"github.com/igeon510/new-wowsan/pkg/proto"
	"google.golang.org/grpc"
)

type BrokerServer struct {
	proto.UnimplementedBrokerServiceServer
	broker *Broker
}

func StartBrokerServer(b *Broker) {
	lis, err := net.Listen("tcp", ":"+b.Port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	service := &BrokerServer{broker: b}
	proto.RegisterBrokerServiceServer(grpcServer, service)

	log.Printf("BrokerService listening on port %s", b.Port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

func (s *BrokerServer) AddBroker(ctx context.Context, req *proto.AddBrokerRequest) (*proto.AddBrokerResponse, error) {
	s.broker.mu.Lock()
	defer s.broker.mu.Unlock()

	// 이미 존재하는지 확인
	if _, exists := s.broker.Peers[req.Node.Id]; exists {
		return &proto.AddBrokerResponse{Success: false, ErrorMessage: "Broker already exists"}, nil
	}

	// Broker 구조체에 새로운 브로커 추가
	s.broker.AddBroker(req.Node)

	return &proto.AddBrokerResponse{Success: true}, nil
}

// server addbroker까지만 구현함.
