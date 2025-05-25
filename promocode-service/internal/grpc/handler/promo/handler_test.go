package promo

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	adaptergrpc "gitlab.com/pisya-dev/promo-code-service/internal/adapter/grpc"
	"gitlab.com/pisya-dev/promo-code-service/internal/domain/dto/promo"
	"gitlab.com/pisya-dev/promo-code-service/internal/domain/dto/target"
	domainerrors "gitlab.com/pisya-dev/promo-code-service/internal/domain/errors"
	promoservice "gitlab.com/pisya-dev/promo-code-service/internal/service/promo"
	promopb "gitlab.com/pisya-dev/promo-code-service/pkg/api/pb"
	"gitlab.com/pisya-dev/promo-code-service/pkg/pointer"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestHandler_CreatePromo(t *testing.T) {
	type fields struct {
		promoService *MockpromoService
	}

	activeFrom, err := time.Parse(time.DateOnly, "2025-01-01")
	require.NoError(t, err)
	activeUntil, err := time.Parse(time.DateOnly, "2026-01-01")
	require.NoError(t, err)

	wantCreatedId := "someId"
	companyId := "someCompanyId"

	rq := &promopb.CreatePromoRequest{
		Mode: promopb.Mode_COMMON,

		PromoCommon: pointer.To("promo"),
		Description: "description",
		ImageUrl:    pointer.To("https://google.com"),
		Target: pointer.To(promopb.Target{
			AgeFrom:    pointer.ToInt64(10),
			AgeUntil:   pointer.ToInt64(30),
			Country:    pointer.To("RU"),
			Categories: []string{"pen", "city"},
		}),
		MaxCount:    5,
		ActiveFrom:  timestamppb.New(activeFrom),
		ActiveUntil: timestamppb.New(activeUntil),
	}

	tests := []struct {
		name    string
		prepare func(f *fields)
		want    *promopb.CreatePromoResponse
		wantErr bool
	}{
		{
			name: "success",
			prepare: func(f *fields) {

				rqActiveFrom := rq.GetActiveFrom().AsTime()

				rqActiveUntil := rq.GetActiveUntil().AsTime()

				mode, err := adaptergrpc.MapPbModeToDomain(rq.GetMode())
				require.NoError(t, err)

				dto := &promo.CreatePromoDTO{
					CompanyId:   companyId,
					Mode:        mode,
					PromoCommon: rq.GetPromoCommon(),
					Description: rq.GetDescription(),
					ImageUrl:    rq.GetImageUrl(),
					Target: target.DTO{
						AgeFrom:    rq.Target.GetAgeFrom(),
						AgeUntil:   rq.Target.GetAgeUntil(),
						Country:    rq.Target.GetCountry(),
						Categories: rq.Target.GetCategories(),
					},
					MaxCount:    rq.GetMaxCount(),
					ActiveFrom:  rqActiveFrom,
					ActiveUntil: rqActiveUntil,
				}

				f.promoService.EXPECT().Create(gomock.Any(), dto).Return(wantCreatedId, nil)

			},
			want: &promopb.CreatePromoResponse{
				Id: wantCreatedId,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)

			f := &fields{
				promoService: NewMockpromoService(ctrl),
			}

			tt.prepare(f)

			h := &Handler{
				promoService: f.promoService,
			}

			ctx := context.WithValue(context.Background(), "company_id", companyId)

			got, err := h.CreatePromo(ctx, rq)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreatePromo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			expected := &promopb.CreatePromoResponse{Id: wantCreatedId}

			require.Equal(t, expected, got)
		})
	}
}

func TestHandler_Activate(t *testing.T) {
	type fields struct {
		promoService *MockpromoService
	}

	promoId := "testPromo123"
	companyId := "company123"
	testCode := "TESTCODE123"

	tests := []struct {
		name        string
		prepare     func(f *fields)
		want        *promopb.ActivatePromoResponse
		wantErr     bool
		wantErrCode codes.Code
	}{
		{
			name: "successful activation",
			prepare: func(f *fields) {
				f.promoService.EXPECT().
					Activate(gomock.Any(), promoId).
					Return(testCode, nil)
			},
			want: &promopb.ActivatePromoResponse{
				Code:              testCode,
				SuccessActivation: true,
				Reason:            promopb.Reason_OK,
			},
			wantErr: false,
		},
		{
			name: "fraud detected",
			prepare: func(f *fields) {
				f.promoService.EXPECT().
					Activate(gomock.Any(), promoId).
					Return("", promoservice.ErrFraudDetected)
			},
			want: &promopb.ActivatePromoResponse{
				SuccessActivation: false,
				Reason:            promopb.Reason_ANTIFRAUD,
			},
			wantErr: false,
		},
		{
			name: "no activations left",
			prepare: func(f *fields) {
				f.promoService.EXPECT().
					Activate(gomock.Any(), promoId).
					Return("", promoservice.ErrNoActivations)
			},
			want: &promopb.ActivatePromoResponse{
				SuccessActivation: false,
				Reason:            promopb.Reason_NO_ACTIVATIONS_LEFT,
			},
			wantErr: false,
		},
		{
			name: "permission denied",
			prepare: func(f *fields) {
				f.promoService.EXPECT().
					Activate(gomock.Any(), promoId).
					Return("", promoservice.ErrPermissionDenied)
			},
			want:        nil,
			wantErr:     true,
			wantErrCode: codes.PermissionDenied,
		},
		{
			name: "promo not found",
			prepare: func(f *fields) {
				f.promoService.EXPECT().
					Activate(gomock.Any(), promoId).
					Return("", promoservice.ErrNotFound)
			},
			want:        nil,
			wantErr:     true,
			wantErrCode: codes.NotFound,
		},
		{
			name: "internal error",
			prepare: func(f *fields) {
				f.promoService.EXPECT().
					Activate(gomock.Any(), promoId).
					Return("", errors.New("some internal error"))
			},
			want:        nil,
			wantErr:     true,
			wantErrCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			f := &fields{
				promoService: NewMockpromoService(ctrl),
			}
			tt.prepare(f)

			h := &Handler{
				promoService: f.promoService,
			}

			ctx := context.WithValue(context.Background(), "company_id", companyId)
			got, err := h.Activate(ctx, &promopb.ActivatePromoRequest{PromoId: promoId})

			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, tt.wantErrCode, st.Code())
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestHandler_GetById(t *testing.T) {
	type fields struct {
		promoService *MockpromoService
	}
	companyId := "someCompanyId"
	promoId := "promoId"
	wantPromo := &promopb.Promo{
		Mode: promopb.Mode_UNIQUE,
	}
	imageURL := "https://google.com"

	promoCategories := []string{"1", "2"}

	promoDTO := &promo.DTO{
		PromoId:     "dvfkjldfvjhb",
		CompanyId:   companyId,
		CompanyName: "dfvdfbf",
		Mode:        promo.COMMON,
		Description: "ded",
		ImageURL:    imageURL,
		Target: &target.DTO{
			AgeFrom:    5,
			AgeUntil:   15,
			Country:    "RU",
			Categories: promoCategories,
		},
		ActiveFrom:       time.Unix(0, 0),
		ActiveUntil:      time.Unix(10, 0),
		Codes:            nil,
		ActivationsCount: 0,
	}

	tests := []struct {
		name    string
		prepare func(f *fields)
		want    *promopb.GetPromoResponse
		wantErr bool
	}{
		{
			name: "success",
			prepare: func(f *fields) {
				f.promoService.EXPECT().
					GetById(gomock.Any(), promoId, companyId).
					Return(promoDTO, nil)
			},
			want:    &promopb.GetPromoResponse{Promo: wantPromo},
			wantErr: false,
		},
		{
			name: "not found",
			prepare: func(f *fields) {
				f.promoService.EXPECT().
					GetById(gomock.Any(), promoId, companyId).
					Return(nil, promoservice.ErrNotFound)
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "permission denied",
			prepare: func(f *fields) {
				f.promoService.EXPECT().
					GetById(gomock.Any(), promoId, companyId).
					Return(nil, promoservice.ErrPermissionDenied)
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "internal error",
			prepare: func(f *fields) {
				f.promoService.EXPECT().
					GetById(gomock.Any(), promoId, companyId).
					Return(nil, errors.New("some internal error"))
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			f := &fields{promoService: NewMockpromoService(ctrl)}
			tt.prepare(f)
			h := &Handler{promoService: f.promoService}
			ctx := context.WithValue(context.Background(), "company_id", companyId)
			_, err := h.GetById(ctx, &promopb.GetPromoRequest{PromoId: promoId})
			if (err != nil) != tt.wantErr {
				t.Errorf("GetById() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestHandler_ListPromo(t *testing.T) {
	type fields struct {
		promoService *MockpromoService
	}

	companyId := "someCompanyId"
	countries := []string{"RU", "US"}
	limit := int64(10)
	offset := int64(0)
	sortBy := promopb.PromoSortBy_ACTIVE_UNTIL

	activeFrom := time.Unix(0, 0)
	activeUntil := time.Unix(10, 0)
	promoCategories := []string{"category1", "category2"}

	promoDTOs := []promo.DTO{
		{
			PromoId:     "promo1",
			CompanyId:   companyId,
			CompanyName: "Company1",
			Mode:        promo.COMMON,
			Codes: []promo.Code{
				{
					Code:        "CODE1",
					Activations: 5,
					MaxCount:    10,
				},
			},
			Description: "Description1",
			ImageURL:    "https://example.com/image1.jpg",
			Target: &target.DTO{
				AgeFrom:    18,
				AgeUntil:   30,
				Country:    "RU",
				Categories: promoCategories,
			},
			ActiveFrom:  activeFrom,
			ActiveUntil: activeUntil,
		},
		{
			PromoId:     "promo2",
			CompanyId:   companyId,
			CompanyName: "Company2",
			Mode:        promo.UNIQUE,
			Codes: []promo.Code{
				{
					Code:        "CODE2",
					Activations: 3,
					MaxCount:    5,
				},
			},
			Description: "Description2",
			ImageURL:    "https://example.com/image2.jpg",
			Target: &target.DTO{
				AgeFrom:    21,
				AgeUntil:   40,
				Country:    "US",
				Categories: promoCategories,
			},
			ActiveFrom:  activeFrom,
			ActiveUntil: activeUntil,
		},
	}

	tests := []struct {
		name    string
		prepare func(f *fields)
		want    *promopb.ListPromoResponse
		wantErr bool
	}{
		{
			name: "success",
			prepare: func(f *fields) {
				f.promoService.EXPECT().
					List(gomock.Any(), companyId, countries, int(limit), int(offset), adaptergrpc.MapPbSortByToDomain(sortBy)).
					Return(promoDTOs, nil)
				f.promoService.EXPECT().
					Count(gomock.Any(), companyId, countries).
					Return(len(promoDTOs), nil)
			},
			want: &promopb.ListPromoResponse{
				XTotalCount: int64(len(promoDTOs)),
				Promo: []*promopb.Promo{
					{
						PromoId:     "promo1",
						CompanyId:   companyId,
						CompanyName: "Company1",
						Mode:        promopb.Mode_COMMON,
						Codes: []*promopb.PromoCode{
							{
								Code:        "CODE1",
								Activations: 5,
								MaxCount:    10,
							},
						},
						Description: "Description1",
						ImageUrl:    pointer.To("https://example.com/image1.jpg"),
						Target: &promopb.Target{
							AgeFrom:    pointer.ToInt64(18),
							AgeUntil:   pointer.ToInt64(30),
							Country:    pointer.To("RU"),
							Categories: promoCategories,
						},
						ActiveFrom:  timestamppb.New(activeFrom),
						ActiveUntil: timestamppb.New(activeUntil),
					},
					{
						PromoId:     "promo2",
						CompanyId:   companyId,
						CompanyName: "Company2",
						Mode:        promopb.Mode_UNIQUE,
						Codes: []*promopb.PromoCode{
							{
								Code:        "CODE2",
								Activations: 3,
								MaxCount:    5,
							},
						},
						Description: "Description2",
						ImageUrl:    pointer.To("https://example.com/image2.jpg"),
						Target: &promopb.Target{
							AgeFrom:    pointer.ToInt64(21),
							AgeUntil:   pointer.ToInt64(40),
							Country:    pointer.To("US"),
							Categories: promoCategories,
						},
						ActiveFrom:  timestamppb.New(activeFrom),
						ActiveUntil: timestamppb.New(activeUntil),
					},
				},
			},
			wantErr: false,
		},
		{
			name: "empty list",
			prepare: func(f *fields) {
				f.promoService.EXPECT().
					List(gomock.Any(), companyId, countries, int(limit), int(offset), adaptergrpc.MapPbSortByToDomain(sortBy)).
					Return([]promo.DTO{}, nil)
				f.promoService.EXPECT().
					Count(gomock.Any(), companyId, countries).
					Return(0, nil)
			},
			want: &promopb.ListPromoResponse{
				XTotalCount: 0,
				Promo:       []*promopb.Promo{},
			},
			wantErr: false,
		},
		{
			name: "list error",
			prepare: func(f *fields) {
				f.promoService.EXPECT().
					List(gomock.Any(), companyId, countries, int(limit), int(offset), adaptergrpc.MapPbSortByToDomain(sortBy)).
					Return(nil, errors.New("some error"))
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "count error",
			prepare: func(f *fields) {
				f.promoService.EXPECT().
					List(gomock.Any(), companyId, countries, int(limit), int(offset), adaptergrpc.MapPbSortByToDomain(sortBy)).
					Return(promoDTOs, nil)
				f.promoService.EXPECT().
					Count(gomock.Any(), companyId, countries).
					Return(0, errors.New("some error"))
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			f := &fields{
				promoService: NewMockpromoService(ctrl),
			}

			tt.prepare(f)

			h := &Handler{
				promoService: f.promoService,
			}

			ctx := context.WithValue(context.Background(), "company_id", companyId)
			rq := &promopb.ListPromoRequest{
				Countries: countries,
				Limit:     pointer.ToInt64(limit),
				Offset:    pointer.ToInt64(offset),
				SortBy:    pointer.To(sortBy),
			}

			_, err := h.ListPromo(ctx, rq)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListPromo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

		})
	}
}

func TestHandler_Delete(t *testing.T) {
	type fields struct {
		promoService *MockpromoService
	}

	companyId := "testCompany"
	promoId := "testPromo"

	tests := []struct {
		name      string
		prepare   func(f *fields)
		wantError codes.Code
	}{
		{
			name: "successful deletion",
			prepare: func(f *fields) {
				f.promoService.EXPECT().
					Delete(gomock.Any(), promoId, companyId).
					Return(nil)
			},
			wantError: codes.OK,
		},
		{
			name: "permission denied",
			prepare: func(f *fields) {
				f.promoService.EXPECT().
					Delete(gomock.Any(), promoId, companyId).
					Return(promoservice.ErrPermissionDenied)
			},
			wantError: codes.PermissionDenied,
		},
		{
			name: "not found",
			prepare: func(f *fields) {
				f.promoService.EXPECT().
					Delete(gomock.Any(), promoId, companyId).
					Return(promoservice.ErrNotFound)
			},
			wantError: codes.NotFound,
		},
		{
			name: "internal error",
			prepare: func(f *fields) {
				f.promoService.EXPECT().
					Delete(gomock.Any(), promoId, companyId).
					Return(errors.New("database error"))
			},
			wantError: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			f := &fields{
				promoService: NewMockpromoService(ctrl),
			}
			tt.prepare(f)

			h := &Handler{promoService: f.promoService}
			ctx := context.WithValue(context.Background(), "company_id", companyId)

			_, err := h.Delete(ctx, &promopb.DeletePromoRequest{PromoId: promoId})

			if tt.wantError == codes.OK {
				require.NoError(t, err)
				return
			}

			if err == nil {
				t.Fatal("expected error, got nil")
			}

			st, ok := status.FromError(err)
			if !ok {
				t.Fatalf("invalid error type: %T", err)
			}

			require.Equal(t, tt.wantError, st.Code())
		})
	}
}

func TestHandler_Update(t *testing.T) {
	type fields struct {
		promoService *MockpromoService
	}

	companyId := "testCompany"
	promoId := "testPromo"
	description := "new description"
	imageUrl := "https://new-image.com"
	targetAgeFrom := int64(18)
	targetAgeUntil := int64(65)
	targetCountry := "US"
	targetCategories := []string{"cat1", "cat2"}
	activeFrom := time.Now().UTC()
	activeUntil := time.Now().AddDate(0, 1, 0).UTC()

	tests := []struct {
		name        string
		prepare     func(f *fields)
		wantErr     bool
		wantErrCode codes.Code
	}{
		{
			name: "successful update",
			prepare: func(f *fields) {
				f.promoService.EXPECT().Update(
					gomock.Any(),
					promoId,
					companyId,
					description,
					imageUrl,
					targetAgeFrom,
					targetAgeUntil,
					targetCountry,
					targetCategories,
					activeFrom,
					activeUntil,
				).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "permission denied",
			prepare: func(f *fields) {
				f.promoService.EXPECT().Update(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(promoservice.ErrPermissionDenied)
			},
			wantErr:     true,
			wantErrCode: codes.PermissionDenied,
		},
		{
			name: "not found",
			prepare: func(f *fields) {
				f.promoService.EXPECT().Update(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(promoservice.ErrNotFound)
			},
			wantErr:     true,
			wantErrCode: codes.NotFound,
		},
		{
			name: "validation error",
			prepare: func(f *fields) {
				f.promoService.EXPECT().Update(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(domainerrors.ValidationError{
						Field:   "field",
						Message: "invalid",
					})
			},
			wantErr:     true,
			wantErrCode: codes.InvalidArgument,
		},
		{
			name: "internal error",
			prepare: func(f *fields) {
				f.promoService.EXPECT().Update(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("internal error"))
			},
			wantErr:     true,
			wantErrCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			f := &fields{
				promoService: NewMockpromoService(ctrl),
			}
			tt.prepare(f)

			h := &Handler{promoService: f.promoService}
			ctx := context.WithValue(context.Background(), "company_id", companyId)

			req := &promopb.UpdatePromoRequest{
				PromoId:     promoId,
				Description: description,
				ImageUrl:    imageUrl,
				Target: &promopb.Target{
					AgeFrom:    &targetAgeFrom,
					AgeUntil:   &targetAgeUntil,
					Country:    &targetCountry,
					Categories: targetCategories,
				},
				ActiveFrom:  timestamppb.New(activeFrom),
				ActiveUntil: timestamppb.New(activeUntil),
			}

			_, err := h.Update(ctx, req)

			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, tt.wantErrCode, st.Code())
			} else {
				require.NoError(t, err)
			}
		})
	}
}
