package interceptor

import (
	"context"

	promopb "gitlab.com/pisya-dev/promo-code-service/pkg/api/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type RequestWithCompanyID interface {
	GetCompanyId() string
}

func AuthInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {

	if _, ok := req.(*promopb.ActivatePromoRequest); ok {
		return handler(ctx, req)
	}

	if _, ok := req.(*promopb.PromoPingRequest); ok {
		return handler(ctx, req)
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "metadata missing")
	}

	companyIDs := md.Get("company_id")
	if len(companyIDs) == 0 {

		if reqWithCompanyID, ok := req.(RequestWithCompanyID); ok {
			companyIDs = []string{reqWithCompanyID.GetCompanyId()}
		}

	}

	if len(companyIDs) == 0 {
		return nil, status.Error(codes.Unauthenticated, "company_id missing")
	}

	ctx = context.WithValue(ctx, "company_id", companyIDs[0])
	return handler(ctx, req)
}
