package grpc_client

import (
	"context"
	"fmt"

	"gitlab.com/pisya-dev/auth-service/internal/config"
	pb "gitlab.com/pisya-dev/auth-service/pkg/api/promopb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type PromoSvcClient struct {
	conn   *grpc.ClientConn
	client pb.PromoServiceClient
}

func NewPromoServiceClient(config config.GrpcConfig) (*PromoSvcClient, error) {
	conn, err := grpc.NewClient(config.PromoClientAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to account service: %w", err)
	}
	client := pb.NewPromoServiceClient(conn)
	return &PromoSvcClient{
		conn:   conn,
		client: client,
	}, nil
}

func (p *PromoSvcClient) CreatePromo(ctx context.Context, req *pb.CreatePromoRequest) (*pb.CreatePromoResponse, error) {
	return p.client.CreatePromo(ctx, req)
}
