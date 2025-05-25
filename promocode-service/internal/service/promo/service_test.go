package promo

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/pisya-dev/promo-code-service/internal/domain/dto/promo"
	"gitlab.com/pisya-dev/promo-code-service/internal/domain/dto/target"
	promoenum "gitlab.com/pisya-dev/promo-code-service/internal/domain/enum/promo"
	"gitlab.com/pisya-dev/promo-code-service/internal/storage/model"
	promoStorage "gitlab.com/pisya-dev/promo-code-service/internal/storage/promo"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

func TestService_List(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	logger := zap.NewNop()

	// Моки зависимостей
	mockRepo := NewMockpromoRepository(ctrl)
	mockPromoCodeRepo := NewMockpromoCodeRepository(ctrl)
	mockRedis := NewMockredisDb(ctrl)
	mockAccountClient := NewMockaccountServiceClient(ctrl)

	// Инициализация сервиса с моками
	service := &Service{
		log:                  logger,
		promoRepository:      mockRepo,
		promoCodeRepository:  mockPromoCodeRepo,
		redisDb:              mockRedis,
		accountServiceClient: mockAccountClient,
	}

	// Тестовые данные
	now := time.Now()
	testPromos := []promoStorage.PromoDetails{
		{
			Id:               "1",
			CompanyId:        "company1",
			Description:      "Test Promo 1",
			ImageUrl:         "http://example.com/1.jpg",
			Mode:             promo.COMMON,
			TargetAgeFrom:    18,
			TargetAgeUntil:   30,
			TargetCountry:    "US",
			TargetCategories: []string{"cat1"},
			ActiveFrom:       now.Add(-24 * time.Hour),
			ActiveUntil:      now.Add(24 * time.Hour),
			Codes: []promoStorage.CodeDTO{
				{Code: "CODE1", MaxCount: 10, Activations: 5},
			},
		},
		{
			Id:               "2",
			CompanyId:        "company2",
			Description:      "Test Promo 2",
			ImageUrl:         "http://example.com/2.jpg",
			Mode:             promo.COMMON,
			TargetAgeFrom:    25,
			TargetAgeUntil:   40,
			TargetCountry:    "RU",
			TargetCategories: []string{"cat2", "cat3"},
			ActiveFrom:       now.Add(-12 * time.Hour),
			ActiveUntil:      now.Add(12 * time.Hour),
			Codes: []promoStorage.CodeDTO{
				{Code: "UNIQUE1", MaxCount: 1, Activations: 1},
				{Code: "UNIQUE2", MaxCount: 1, Activations: 0},
			},
		},
	}

	t.Run("successful list", func(t *testing.T) {
		// Настройка моков
		mockRepo.EXPECT().List(
			ctx,
			"testCompany",
			[]string{"US", "RU"},
			10,
			promoenum.SortByActiveFrom,
			20,
		).Return(testPromos, nil)

		mockAccountClient.EXPECT().GetCompanyNameByCompanyID(ctx, "company1").
			Return("Company One", nil)
		mockAccountClient.EXPECT().GetCompanyNameByCompanyID(ctx, "company2").
			Return("Company Two", nil)

		// Вызов метода
		result, err := service.List(
			ctx,
			"testCompany",
			[]string{"US", "RU"},
			20,
			10,
			promoenum.SortByActiveFrom,
		)

		// Проверки
		assert.NoError(t, err)
		assert.Len(t, result, 2)

	})

	t.Run("repository error", func(t *testing.T) {
		expectedErr := errors.New("database error")

		mockRepo.EXPECT().List(
			gomock.Any(),
			gomock.Any(),
			gomock.Any(),
			gomock.Any(),
			gomock.Any(),
			gomock.Any(),
		).Return(nil, expectedErr)

		result, err := service.List(ctx, "", nil, 0, 0, promoenum.SortByCreatedAt)

		assert.ErrorIs(t, err, expectedErr)
		assert.Nil(t, result)
	})

	t.Run("account service error", func(t *testing.T) {
		testErr := errors.New("account service error")

		mockRepo.EXPECT().List(
			gomock.Any(),
			gomock.Any(),
			gomock.Any(),
			gomock.Any(),
			gomock.Any(),
			gomock.Any(),
		).Return(testPromos, nil)

		mockAccountClient.EXPECT().GetCompanyNameByCompanyID(gomock.Any(), "company1").
			Return("", testErr)

		result, err := service.List(ctx, "", nil, 0, 0, promoenum.SortByActiveUntil)

		assert.ErrorContains(t, err, "accountServiceClient.GetCompanyNameByCompanyID")
		assert.Nil(t, result)
	})

	t.Run("empty list", func(t *testing.T) {
		mockRepo.EXPECT().List(
			gomock.Any(),
			gomock.Any(),
			gomock.Any(),
			gomock.Any(),
			gomock.Any(),
			gomock.Any(),
		).Return([]promoStorage.PromoDetails{}, nil)

		result, err := service.List(ctx, "", nil, 0, 0, promoenum.SortByActiveUntil)

		assert.NoError(t, err)
		assert.Empty(t, result)
	})
}

func TestService_GetById(t *testing.T) {
	type mockRepo func(*MockpromoRepositoryMockRecorder)
	type mockRedis func(*MockredisDbMockRecorder)
	type mockAccountService func(*MockaccountServiceClientMockRecorder)

	now := time.Now()
	testPromo := &promoStorage.PromoDetails{
		Id:               "test-promo",
		CompanyId:        "test-company",
		Description:      "Test promo",
		ImageUrl:         "http://example.com/image.jpg",
		Mode:             promo.COMMON,
		TargetAgeFrom:    18,
		TargetAgeUntil:   30,
		TargetCountry:    "US",
		TargetCategories: []string{"cat1", "cat2"},
		ActiveFrom:       now.Add(-24 * time.Hour),
		ActiveUntil:      now.Add(24 * time.Hour),
		Codes: []promoStorage.CodeDTO{
			{Code: "CODE1", MaxCount: 10, Activations: 5},
		},
	}

	tests := []struct {
		name        string
		promoId     string
		companyId   string
		mockRepo    mockRepo
		mockRedis   mockRedis
		mockAccount mockAccountService
		want        *promo.DTO
		wantErr     error
	}{
		{
			name:      "success from cache",
			promoId:   "test-promo",
			companyId: "test-company",
			mockRedis: func(r *MockredisDbMockRecorder) {
				// Возвращаем валидные кэшированные данные
				data, _ := json.Marshal(testPromo)
				r.Get(gomock.Any(), "test-promo").Return(redis.NewStringResult(string(data), nil))
			},
			mockAccount: func(a *MockaccountServiceClientMockRecorder) {
				a.GetCompanyNameByCompanyID(gomock.Any(), "test-company").Return("Test Company", nil)
			},
			mockRepo: func(r *MockpromoRepositoryMockRecorder) {
			},
			want: &promo.DTO{
				PromoId:     "test-promo",
				CompanyId:   "test-company",
				CompanyName: "Test Company",
				Mode:        promo.COMMON,
				Description: "Test promo",
				ImageURL:    "http://example.com/image.jpg",
				Codes: []promo.Code{
					{Code: "CODE1", MaxCount: 10, Activations: 5},
				},
				Target: &target.DTO{
					AgeFrom:    18,
					AgeUntil:   30,
					Country:    "US",
					Categories: []string{"cat1", "cat2"},
				},
				ActiveFrom:  testPromo.ActiveFrom,
				ActiveUntil: testPromo.ActiveUntil,
			},
		},
		{
			name:      "success from db with cache set",
			promoId:   "test-promo",
			companyId: "test-company",
			mockRedis: func(r *MockredisDbMockRecorder) {
				// Пустой кэш
				r.Get(gomock.Any(), "test-promo").Return(redis.NewStringResult("", nil))
				// Ожидаем запись в Redis
				data, _ := json.Marshal(testPromo)
				r.Set(gomock.Any(), "test-promo", string(data), time.Nanosecond*0).Return(redis.NewStatusResult("", nil))
			},
			mockAccount: func(a *MockaccountServiceClientMockRecorder) {
				a.GetCompanyNameByCompanyID(gomock.Any(), "test-company").Return("Test Company", nil)
			},
			mockRepo: func(r *MockpromoRepositoryMockRecorder) {
				r.GetById(gomock.Any(), "test-promo").Return(testPromo, nil)
			},
			want: &promo.DTO{
				// ... аналогично предыдущему тесту ...
			},
		},
		{
			name:      "permission denied",
			promoId:   "test-promo",
			companyId: "another-company",
			mockRedis: func(r *MockredisDbMockRecorder) {
				data, _ := json.Marshal(testPromo)
				r.Get(gomock.Any(), "test-promo").Return(redis.NewStringResult(string(data), nil))
			},
			mockRepo:    func(r *MockpromoRepositoryMockRecorder) {},
			mockAccount: func(a *MockaccountServiceClientMockRecorder) {},
			wantErr:     ErrPermissionDenied,
		},
		{
			name:      "not found in db",
			promoId:   "non-existent",
			companyId: "test-company",
			mockRedis: func(r *MockredisDbMockRecorder) {
				r.Get(gomock.Any(), "non-existent").Return(redis.NewStringResult("", nil))
			},
			mockRepo: func(r *MockpromoRepositoryMockRecorder) {
				r.GetById(gomock.Any(), "non-existent").Return(nil, nil)
			},
			mockAccount: func(a *MockaccountServiceClientMockRecorder) {},
			wantErr:     ErrNotFound,
		},
		{
			name:      "account service error",
			promoId:   "test-promo",
			companyId: "test-company",
			mockRedis: func(r *MockredisDbMockRecorder) {
				data, _ := json.Marshal(testPromo)
				r.Get(gomock.Any(), "test-promo").Return(redis.NewStringResult(string(data), nil))
			},
			mockAccount: func(a *MockaccountServiceClientMockRecorder) {
				a.GetCompanyNameByCompanyID(gomock.Any(), "test-company").
					Return("", errors.New("service unavailable"))
			},
			mockRepo: func(r *MockpromoRepositoryMockRecorder) {},
			wantErr:  errors.New("accountServiceClient.GetCompanyNameByCompanyID"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := NewMockpromoRepository(ctrl)
			redisMock := NewMockredisDb(ctrl)
			accountMock := NewMockaccountServiceClient(ctrl)

			tt.mockRepo(repo.EXPECT())
			tt.mockRedis(redisMock.EXPECT())
			tt.mockAccount(accountMock.EXPECT())

			service := &Service{
				log:                  zaptest.NewLogger(t),
				promoRepository:      repo,
				redisDb:              redisMock,
				accountServiceClient: accountMock,
			}

			_, err := service.GetById(context.Background(), tt.promoId, tt.companyId)

			if tt.wantErr != nil {
				require.Error(t, err)
				return
			}

		})
	}
}

func TestService_Count(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	testCompanyID := "test-company"
	testCountries := []string{"RU", "US"}

	tests := []struct {
		name           string
		companyID      string
		countries      []string
		mockRepoCount  int
		mockRepoErr    error
		expectedCount  int
		expectedErrStr string
	}{
		{
			name:          "successful count",
			companyID:     testCompanyID,
			countries:     testCountries,
			mockRepoCount: 5,
			expectedCount: 5,
		},
		{
			name:           "repository error",
			companyID:      testCompanyID,
			countries:      testCountries,
			mockRepoErr:    errors.New("database error"),
			expectedErrStr: "promoRepository.Count: database error",
		},
		{
			name:          "empty countries filter",
			companyID:     testCompanyID,
			countries:     []string{},
			mockRepoCount: 10,
			expectedCount: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockRepo := NewMockpromoRepository(ctrl)
			mockRepo.EXPECT().
				Count(ctx, tt.companyID, tt.countries).
				Return(tt.mockRepoCount, tt.mockRepoErr)

			// Create service with mock dependencies
			service := &Service{
				log:             zap.NewNop(),
				promoRepository: mockRepo,
			}

			// Execute
			_, err := service.Count(ctx, tt.companyID, tt.countries)

			// Assert
			if tt.expectedErrStr != "" {
				assert.ErrorContains(t, err, tt.expectedErrStr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_Create(t *testing.T) {

	ctrl := gomock.NewController(t)

	type fields struct {
		log                  *zap.Logger
		promoRepository      *MockpromoRepository
		promoCodeRepository  *MockpromoCodeRepository
		redisDb              *MockredisDb
		accountServiceClient *MockaccountServiceClient
	}
	type args struct {
		ctx      context.Context
		promoDto *promo.CreatePromoDTO
	}

	promoId := "4eacc594-942f-482e-b0df-3c6a3f63ef33"
	promoCodeId := "8d812a2f-6001-4ca9-9f67-e6a78e069b04"

	tests := []struct {
		name    string
		fields  fields
		args    args
		prepare func(f *fields, a *args)
		want    string
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				log:                  zap.NewNop(),
				promoRepository:      NewMockpromoRepository(ctrl),
				promoCodeRepository:  NewMockpromoCodeRepository(ctrl),
				redisDb:              NewMockredisDb(ctrl),
				accountServiceClient: NewMockaccountServiceClient(ctrl),
			},
			args: args{
				ctx: context.Background(),
				promoDto: &promo.CreatePromoDTO{
					CompanyId:   "1113d90f-993c-4586-ab59-aa41b62ef792",
					Mode:        promo.COMMON,
					PromoCommon: "exhibition",
					PromoUnique: nil,
					Description: "Sale 10% for Ablabla exbibitions",
					ImageUrl:    "http://google.com",
					Target: target.DTO{
						AgeFrom:    10,
						AgeUntil:   60,
						Country:    "RU",
						Categories: []string{"young", "eat"},
					},
					MaxCount:    9000,
					ActiveFrom:  time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
					ActiveUntil: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
				},
			},
			prepare: func(f *fields, a *args) {
				f.promoRepository.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, p *model.Promo) (string, error) {
					require.Equal(t, a.promoDto.CompanyId, p.CompanyId)
					require.Equal(t, a.promoDto.Description, p.Description)
					require.Equal(t, a.promoDto.ImageUrl, p.ImageUrl)
					require.Equal(t, a.promoDto.ActiveFrom, p.ActiveFrom)
					require.Equal(t, a.promoDto.ActiveUntil, p.ActiveUntil)
					require.Equal(t, a.promoDto.Mode, p.Mode)
					require.Equal(t, a.promoDto.Target.AgeFrom, int64(p.TargetAgeFrom))
					require.Equal(t, a.promoDto.Target.AgeUntil, int64(p.TargetAgeUntil))
					require.Equal(t, a.promoDto.Target.Country, p.TargetCountry)
					require.Equal(t, a.promoDto.Target.Categories, p.TargetCategories)
					require.Equal(t, a.promoDto.PromoCommon, p.PromoCommon)
					require.Equal(t, a.promoDto.PromoUnique, p.PromoUnique)
					require.Equal(t, a.promoDto.MaxCount, p.MaxCount)

					if p.CreatedAt.IsZero() {
						t.Error("Created at must not be zero")
					}
					return promoId, nil
				})

				f.promoCodeRepository.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, code *model.PromoCode) (string, error) {
					require.Equal(t, promoId, code.PromoId)
					require.Equal(t, a.promoDto.MaxCount, code.MaxCount)
					require.Equal(t, a.promoDto.PromoCommon, code.Code)

					return promoCodeId, nil
				})

			},
			want:    promoId,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{
				log:                  tt.fields.log,
				promoRepository:      tt.fields.promoRepository,
				promoCodeRepository:  tt.fields.promoCodeRepository,
				redisDb:              tt.fields.redisDb,
				accountServiceClient: tt.fields.accountServiceClient,
			}

			tt.prepare(&tt.fields, &tt.args)

			gotId, err := s.Create(tt.args.ctx, tt.args.promoDto)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equalf(t, tt.want, gotId, "Create(%v, %v)", tt.args.ctx, tt.args.promoDto)
		})
	}
}

func TestService_Update(t *testing.T) {

	ctrl := gomock.NewController(t)

	type fields struct {
		log                  *zap.Logger
		promoRepository      *MockpromoRepository
		promoCodeRepository  *MockpromoCodeRepository
		redisDb              *MockredisDb
		accountServiceClient *MockaccountServiceClient
	}
	type args struct {
		ctx              context.Context
		promoId          string
		companyId        string
		description      string
		imageUrl         string
		targetAgeFrom    int64
		targetAgeUntil   int64
		targetCountry    string
		targetCategories []string
		activeFrom       time.Time
		activeUntil      time.Time
	}
	tests := []struct {
		name    string
		fields  fields
		prepare func(f *fields, a *args)
		args    args
		wantErr bool
		want    string
	}{
		{
			name: "success",
			fields: fields{
				log:                  zap.NewNop(),
				promoRepository:      NewMockpromoRepository(ctrl),
				promoCodeRepository:  NewMockpromoCodeRepository(ctrl),
				redisDb:              NewMockredisDb(ctrl),
				accountServiceClient: NewMockaccountServiceClient(ctrl),
			},
			args: args{
				ctx:              context.Background(),
				promoId:          "f5db5acc-03da-4215-bb0d-87078e422c45",
				companyId:        "8eb7064a-a899-4ad4-814f-deb2f660536b",
				description:      "test-description",
				imageUrl:         "test-image-url",
				targetAgeFrom:    18,
				targetAgeUntil:   30,
				targetCountry:    "US",
				targetCategories: []string{"cat1", "cat2"},
				activeFrom:       time.Now(),
				activeUntil:      time.Now().Add(24 * time.Hour),
			},
			prepare: func(f *fields, a *args) {
				//f.promoRepository.EXPECT().Update(
				//	gomock.Any(),
				//	a.promoId,
				//	a.description,
				//	a.imageUrl,
				//	a.targetAgeFrom,
				//	a.targetAgeUntil,
				//	a.targetCountry,
				//	a.targetCategories,
				//	a.activeFrom,
				//	a.activeUntil,
				//).Return(nil)

				f.promoRepository.EXPECT().GetById(gomock.Any(), a.promoId).Return(&promoStorage.PromoDetails{
					Id:        a.promoId,
					CompanyId: a.companyId,
				}, nil)

				f.promoRepository.EXPECT().Update(
					gomock.Any(),
					a.promoId,
					a.description,
					a.imageUrl,
					a.targetAgeFrom,
					a.targetAgeUntil,
					a.targetCountry,
					a.targetCategories,
					a.activeFrom,
					a.activeUntil,
				).Return(nil)

				f.redisDb.EXPECT().Get(gomock.Any(), gomock.Eq(a.promoId)).Return(redis.NewStringResult("", redis.Nil))

				f.redisDb.EXPECT().Set(gomock.Any(), a.promoId, gomock.Any(), gomock.Any()).Return(redis.NewStatusCmd(context.Background(), "OK"))

				f.redisDb.EXPECT().
					Del(gomock.Any(), gomock.Eq(a.promoId)).
					Return(redis.NewIntCmd(context.Background(), 1))

				f.accountServiceClient.EXPECT().GetCompanyNameByCompanyID(gomock.Any(), a.companyId).Return("companyName", nil)

			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{
				log:                  tt.fields.log,
				promoRepository:      tt.fields.promoRepository,
				promoCodeRepository:  tt.fields.promoCodeRepository,
				redisDb:              tt.fields.redisDb,
				accountServiceClient: tt.fields.accountServiceClient,
			}

			tt.prepare(&tt.fields, &tt.args)

			err := s.Update(
				tt.args.ctx,
				tt.args.promoId,
				tt.args.companyId,
				tt.args.description,
				tt.args.imageUrl,
				tt.args.targetAgeFrom,
				tt.args.targetAgeUntil,
				tt.args.targetCountry,
				tt.args.targetCategories,
				tt.args.activeFrom,
				tt.args.activeUntil,
			)

			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestService_Delete(t *testing.T) {

	ctrl := gomock.NewController(t)

	type fields struct {
		log                  *zap.Logger
		promoRepository      *MockpromoRepository
		promoCodeRepository  *MockpromoCodeRepository
		redisDb              *MockredisDb
		accountServiceClient *MockaccountServiceClient
	}
	type args struct {
		ctx       context.Context
		promoId   string
		companyId string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		prepare func(f *fields, a *args)
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				log:                  zap.NewNop(),
				promoRepository:      NewMockpromoRepository(ctrl),
				promoCodeRepository:  NewMockpromoCodeRepository(ctrl),
				redisDb:              NewMockredisDb(ctrl),
				accountServiceClient: NewMockaccountServiceClient(ctrl),
			},
			args: args{
				ctx:       context.Background(),
				promoId:   "4eacc594-942f-482e-b0df-3c6a3f63ef33",
				companyId: "8eb7064a-a899-4ad4-814f-deb2f660536b",
			},
			prepare: func(f *fields, a *args) {
				f.promoRepository.EXPECT().GetById(gomock.Any(), a.promoId).Return(&promoStorage.PromoDetails{
					Id:        a.promoId,
					CompanyId: a.companyId,
				}, nil)

				f.promoRepository.EXPECT().Delete(gomock.Any(), a.promoId).Return(nil)

				f.accountServiceClient.EXPECT().GetCompanyNameByCompanyID(gomock.Any(), a.companyId).Return("companyName", nil)

				f.redisDb.EXPECT().Set(gomock.Any(), a.promoId, gomock.Any(), gomock.Any()).Return(redis.NewStatusCmd(context.Background(), "OK"))

				f.redisDb.EXPECT().Get(gomock.Any(), gomock.Eq(a.promoId)).Return(redis.NewStringResult("", redis.Nil))

				f.redisDb.EXPECT().Del(gomock.Any(), a.promoId).Return(redis.NewIntCmd(context.Background(), 1))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			tt.prepare(&tt.fields, &tt.args)

			s := &Service{
				log:                  tt.fields.log,
				promoRepository:      tt.fields.promoRepository,
				promoCodeRepository:  tt.fields.promoCodeRepository,
				redisDb:              tt.fields.redisDb,
				accountServiceClient: tt.fields.accountServiceClient,
			}

			err := s.Delete(tt.args.ctx, tt.args.promoId, tt.args.companyId)

			if (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}

		})
	}
}
