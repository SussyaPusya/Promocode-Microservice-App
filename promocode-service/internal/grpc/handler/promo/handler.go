package promo

import (
	"context"
	"errors"
	"log"

	adaptergrpc "gitlab.com/pisya-dev/promo-code-service/internal/adapter/grpc"
	promodto "gitlab.com/pisya-dev/promo-code-service/internal/domain/dto/promo"
	"gitlab.com/pisya-dev/promo-code-service/internal/domain/dto/target"
	domainerrors "gitlab.com/pisya-dev/promo-code-service/internal/domain/errors"
	"gitlab.com/pisya-dev/promo-code-service/internal/pkg/functional"
	promoservice "gitlab.com/pisya-dev/promo-code-service/internal/service/promo"
	promopb "gitlab.com/pisya-dev/promo-code-service/pkg/api/pb"
	"gitlab.com/pisya-dev/promo-code-service/pkg/pointer"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Handler struct {
	promoService promoService
}

func New(promoService promoService) *Handler {
	return &Handler{
		promoService: promoService,
	}
}

func (h *Handler) CreatePromo(ctx context.Context, r *promopb.CreatePromoRequest) (*promopb.CreatePromoResponse, error) {

	dto := &promodto.CreatePromoDTO{
		CompanyId:   ctx.Value("company_id").(string),
		Mode:        promodto.Mode(r.GetMode().String()),
		PromoCommon: r.GetPromoCommon(),
		PromoUnique: r.GetPromoUnique(),
		Description: r.GetDescription(),
		ImageUrl:    r.GetImageUrl(),
		Target: target.DTO{
			AgeFrom:    r.Target.GetAgeFrom(),
			AgeUntil:   r.Target.GetAgeUntil(),
			Country:    r.Target.GetCountry(),
			Categories: r.Target.GetCategories(),
		},
		MaxCount:    r.GetMaxCount(),
		ActiveFrom:  r.GetActiveFrom().AsTime(),
		ActiveUntil: r.GetActiveUntil().AsTime(),
	}
	promoId, err := h.promoService.Create(ctx, dto)
	if err != nil {
		log.Println(err)

		if errors.As(err, &domainerrors.ValidationError{}) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		return nil, status.Error(codes.Internal, "internal error")
	}

	resp := &promopb.CreatePromoResponse{
		Id: promoId,
	}
	return resp, nil
}

func (h *Handler) ListPromo(ctx context.Context, r *promopb.ListPromoRequest) (*promopb.ListPromoResponse, error) {
	promoDTOs, err := h.promoService.List(ctx, ctx.Value("company_id").(string), r.GetCountries(), int(r.GetLimit()), int(r.GetOffset()), adaptergrpc.MapPbSortByToDomain(r.GetSortBy()))
	if err != nil {
		log.Println(err)
		return nil, status.Error(codes.Internal, "internal server error")
	}

	promosGRPC := make([]*promopb.Promo, len(promoDTOs))

	for idx, promoDTO := range promoDTOs {

		promoMode, err := adaptergrpc.MapDomainModeToPb(promoDTO.Mode)
		if err != nil {
			log.Println(err)

			if errors.As(err, &domainerrors.ValidationError{}) {
				return nil, status.Error(codes.InvalidArgument, err.Error())
			}
			return nil, status.Error(codes.Internal, "internal server error")
		}

		promosGRPC[idx] = &promopb.Promo{
			PromoId:     promoDTO.PromoId,
			CompanyId:   promoDTO.CompanyId,
			CompanyName: promoDTO.CompanyName,
			Mode:        promoMode,
			Codes: functional.Map(promoDTO.Codes, func(code promodto.Code) *promopb.PromoCode {
				return &promopb.PromoCode{
					Code:        code.Code,
					Activations: code.Activations,
					MaxCount:    code.MaxCount,
				}
			}),
			Description: promoDTO.Description,
			ImageUrl:    pointer.To(promoDTO.ImageURL),
			Target: &promopb.Target{
				AgeFrom:    pointer.ToInt64(promoDTO.Target.AgeFrom),
				AgeUntil:   pointer.To(promoDTO.Target.AgeUntil),
				Country:    pointer.To(promoDTO.Target.Country),
				Categories: promoDTO.Target.Categories,
			},
			ActiveFrom:  timestamppb.New(promoDTO.ActiveFrom),
			ActiveUntil: timestamppb.New(promoDTO.ActiveUntil),
		}

	}

	promosCount, err := h.promoService.Count(ctx, ctx.Value("company_id").(string), r.GetCountries())
	if err != nil {
		log.Println(err)

		return nil, status.Error(codes.Internal, "internal server error")
	}

	resp := &promopb.ListPromoResponse{
		XTotalCount: int64(promosCount),
		Promo:       promosGRPC,
	}

	return resp, nil
}

func (h *Handler) GetById(ctx context.Context, r *promopb.GetPromoRequest) (*promopb.GetPromoResponse, error) {
	promoDTO, err := h.promoService.GetById(ctx, r.GetPromoId(), ctx.Value("company_id").(string))

	if err != nil {
		log.Println(err)

		if errors.Is(err, promoservice.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "promo not found")
		}
		if errors.Is(err, promoservice.ErrPermissionDenied) {
			return nil, status.Error(codes.PermissionDenied, "permission denied")
		}
		return nil, status.Error(codes.Internal, "internal server error")
	}

	if promoDTO == nil {
		return nil, status.Error(codes.NotFound, "promo not found")
	}

	promoMode, err := adaptergrpc.MapDomainModeToPb(promoDTO.Mode)
	if err != nil {
		if errors.As(err, &domainerrors.ValidationError{}) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		return nil, status.Error(codes.Internal, "internal server error")
	}

	promoGRPC := &promopb.Promo{
		PromoId:     promoDTO.PromoId,
		CompanyId:   promoDTO.CompanyId,
		CompanyName: promoDTO.CompanyName,
		Mode:        promoMode,
		Codes: functional.Map(promoDTO.Codes, func(code promodto.Code) *promopb.PromoCode {
			return &promopb.PromoCode{
				Code:        code.Code,
				Activations: code.Activations,
				MaxCount:    code.MaxCount,
			}
		}),
		Description: promoDTO.Description,
		ImageUrl:    pointer.To(promoDTO.ImageURL),
		Target: &promopb.Target{
			AgeFrom:    pointer.ToInt64(promoDTO.Target.AgeFrom),
			AgeUntil:   pointer.To(promoDTO.Target.AgeUntil),
			Country:    pointer.To(promoDTO.Target.Country),
			Categories: promoDTO.Target.Categories,
		},
		ActiveFrom:  timestamppb.New(promoDTO.ActiveFrom),
		ActiveUntil: timestamppb.New(promoDTO.ActiveUntil),
	}

	return &promopb.GetPromoResponse{Promo: promoGRPC}, nil

}

func (h *Handler) Update(ctx context.Context, r *promopb.UpdatePromoRequest) (*promopb.UpdatePromoResponse, error) {
	err := h.promoService.Update(
		ctx,
		r.GetPromoId(),
		ctx.Value("company_id").(string),
		r.GetDescription(),
		r.GetImageUrl(),
		r.Target.GetAgeFrom(),
		r.Target.GetAgeUntil(),
		r.Target.GetCountry(),
		r.Target.GetCategories(),
		r.GetActiveFrom().AsTime(),
		r.GetActiveUntil().AsTime(),
	)
	if err != nil {
		log.Println(err)

		if errors.Is(err, promoservice.ErrPermissionDenied) {
			return nil, status.Error(codes.PermissionDenied, "permission denied")
		}
		if errors.Is(err, promoservice.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "promo not found")
		}
		if errors.As(err, &domainerrors.ValidationError{}) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &promopb.UpdatePromoResponse{}, nil

}

func (h *Handler) Delete(ctx context.Context, r *promopb.DeletePromoRequest) (*promopb.DeletePromoResponse, error) {

	err := h.promoService.Delete(ctx, r.GetPromoId(), ctx.Value("company_id").(string))
	if err != nil {
		log.Println(err)

		if errors.Is(err, promoservice.ErrPermissionDenied) {
			return nil, status.Error(codes.PermissionDenied, "permission denied")
		}
		if errors.Is(err, promoservice.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "promo not found")
		}
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &promopb.DeletePromoResponse{}, nil
}

func (h *Handler) Activate(ctx context.Context, r *promopb.ActivatePromoRequest) (*promopb.ActivatePromoResponse, error) {

	code, err := h.promoService.Activate(ctx, r.GetPromoId())
	if err != nil {
		log.Println(err)

		if errors.Is(err, promoservice.ErrFraudDetected) {
			return &promopb.ActivatePromoResponse{
				SuccessActivation: false,
				Reason:            promopb.Reason_ANTIFRAUD,
			}, nil
		}
		if errors.Is(err, promoservice.ErrPermissionDenied) {
			return nil, status.Error(codes.PermissionDenied, "permission denied")
		}
		if errors.Is(err, promoservice.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "promo not found")
		}
		if errors.Is(err, promoservice.ErrNoActivations) {
			return &promopb.ActivatePromoResponse{
				SuccessActivation: false,
				Reason:            promopb.Reason_NO_ACTIVATIONS_LEFT,
			}, nil
		}
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &promopb.ActivatePromoResponse{
		Code:              code,
		SuccessActivation: true,
		Reason:            promopb.Reason_OK,
	}, nil

}
