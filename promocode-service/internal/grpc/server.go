package promogrpc

import (
	"context"

	handlerPromo "gitlab.com/pisya-dev/promo-code-service/internal/grpc/handler/promo"
	promopb "gitlab.com/pisya-dev/promo-code-service/pkg/api/pb"
)

type ServerAPI struct {
	promopb.UnimplementedPromoServiceServer
	promoHandler *handlerPromo.Handler
}

func New(
	promoHandler *handlerPromo.Handler,
) *ServerAPI {
	return &ServerAPI{
		UnimplementedPromoServiceServer: promopb.UnimplementedPromoServiceServer{},
		promoHandler:                    promoHandler,
	}
}

func (s *ServerAPI) PromoPing(ctx context.Context, request *promopb.PromoPingRequest) (*promopb.PromoPingResponse, error) {
	return &promopb.PromoPingResponse{Ok: true}, nil
}

func (s *ServerAPI) CreatePromo(ctx context.Context, request *promopb.CreatePromoRequest) (*promopb.CreatePromoResponse, error) {
	return s.promoHandler.CreatePromo(ctx, request)
}

func (s *ServerAPI) ListPromo(ctx context.Context, request *promopb.ListPromoRequest) (*promopb.ListPromoResponse, error) {
	return s.promoHandler.ListPromo(ctx, request)
}

func (s *ServerAPI) GetPromo(ctx context.Context, request *promopb.GetPromoRequest) (*promopb.GetPromoResponse, error) {
	return s.promoHandler.GetById(ctx, request)
}

func (s *ServerAPI) UpdatePromo(ctx context.Context, request *promopb.UpdatePromoRequest) (*promopb.UpdatePromoResponse, error) {
	return s.promoHandler.Update(ctx, request)
}

func (s *ServerAPI) DeletePromo(ctx context.Context, request *promopb.DeletePromoRequest) (*promopb.DeletePromoResponse, error) {
	return s.promoHandler.Delete(ctx, request)
}

func (s *ServerAPI) ActivatePromo(ctx context.Context, r *promopb.ActivatePromoRequest) (*promopb.ActivatePromoResponse, error) {
	return s.promoHandler.Activate(ctx, r)
}
