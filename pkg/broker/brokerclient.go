package broker

import (
	"context"
	"fmt"

	"github.com/HongikAPL/new-wowsan/pkg/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type BrokerClient struct {
	conn   *grpc.ClientConn
	client proto.BrokerServiceClient
}

func NewBrokerClient(n *proto.NodeInfo) (*BrokerClient, error) {
	addr := fmt.Sprintf("%s:%s", n.Ip, n.Port)

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &BrokerClient{
		conn:   conn,
		client: proto.NewBrokerServiceClient(conn),
	}, nil
}
func (bc *BrokerClient) AddBroker(ctx context.Context, req *proto.AddBrokerRequest) (*proto.AddBrokerResponse, error) {
	return bc.client.AddBroker(ctx, req)
}

func (c *BrokerClient) SendSubscription(ctx context.Context, req *proto.SendSubscriptionRequest) (*proto.SendSubscriptionResponse, error) {
	return c.client.SendSubscription(ctx, req)
}

func (c *BrokerClient) SendUnsubscription(ctx context.Context, req *proto.SendUnsubscriptionRequest) (*proto.SendUnsubscriptionResponse, error) {
	return c.client.SendUnsubscription(ctx, req)
}

func (c *BrokerClient) SendPublication(ctx context.Context, req *proto.SendPublicationRequest) (*proto.SendPublicationResponse, error) {
	return c.client.SendPublication(ctx, req)
}

func (c *BrokerClient) Close() error {
	return c.conn.Close()
}
