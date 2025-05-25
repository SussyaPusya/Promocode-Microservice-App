package promo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gitlab.com/pisya-dev/promo-code-service/internal/domain/dto/promo"
	"gitlab.com/pisya-dev/promo-code-service/internal/domain/dto/target"
	promoenum "gitlab.com/pisya-dev/promo-code-service/internal/domain/enum/promo"
	domainerrors "gitlab.com/pisya-dev/promo-code-service/internal/domain/errors"
	"gitlab.com/pisya-dev/promo-code-service/internal/pkg/functional"
	"gitlab.com/pisya-dev/promo-code-service/internal/storage/model"
	promoStorage "gitlab.com/pisya-dev/promo-code-service/internal/storage/promo"
	"gitlab.com/pisya-dev/promo-code-service/internal/storage/promo_code"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrPermissionDenied    = errors.New("permission denied")
	ErrNotFound            = errors.New("not found")
	ErrFraudDetected       = errors.New("fraud detected")
	ErrNoActivations       = errors.New("no activations")
	ErrCompanyDoesNotExist = errors.New("company does not exist")
)

type Service struct {
	log                  *zap.Logger
	promoRepository      promoRepository
	promoCodeRepository  promoCodeRepository
	redisDb              redisDb
	accountServiceClient accountServiceClient
}

func New(
	log *zap.Logger,
	promoRepository promoRepository,
	promoCodeRepository promoCodeRepository,
	redisDb redisDb,
	accountServiceClient accountServiceClient,
) *Service {
	return &Service{
		log:                  log,
		promoRepository:      promoRepository,
		promoCodeRepository:  promoCodeRepository,
		redisDb:              redisDb,
		accountServiceClient: accountServiceClient,
	}
}

func (s *Service) Create(ctx context.Context, promoDto *promo.CreatePromoDTO) (id string, err error) {
	const op = "service.promo.Create"

	err = promoDto.Validate()
	if err != nil {
		return "", err
	}

	promoModel := &model.Promo{
		Id:               uuid.New().String(),
		CompanyId:        promoDto.CompanyId,
		Description:      promoDto.Description,
		ImageUrl:         promoDto.ImageUrl,
		ActiveFrom:       promoDto.ActiveFrom,
		ActiveUntil:      promoDto.ActiveUntil,
		CreatedAt:        time.Now(),
		Mode:             promoDto.Mode,
		TargetAgeFrom:    int(promoDto.Target.AgeFrom),
		TargetAgeUntil:   int(promoDto.Target.AgeUntil),
		TargetCountry:    promoDto.Target.Country,
		TargetCategories: promoDto.Target.Categories,
		PromoCommon:      promoDto.PromoCommon,
		PromoUnique:      promoDto.PromoUnique,
		MaxCount:         promoDto.MaxCount,
	}

	id, err = s.promoRepository.Create(ctx, promoModel)

	if err != nil {
		s.log.Error("Failed to save promo to db", zap.Error(err))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	if promoDto.Mode == promo.COMMON {
		promoCodeModel := &model.PromoCode{
			Id:       uuid.New().String(),
			PromoId:  id,
			Code:     promoDto.PromoCommon,
			MaxCount: promoDto.MaxCount,
		}

		_, err = s.promoCodeRepository.Create(ctx, promoCodeModel)
		if err != nil {
			s.log.Error("Failed to save promo code to db", zap.Error(err))
			return "", fmt.Errorf("%s: %w", op, err)
		}
	} else if promoDto.Mode == promo.UNIQUE {
		for _, code := range promoDto.PromoUnique {
			promoCodeModel := &model.PromoCode{
				Id:       uuid.New().String(),
				PromoId:  id,
				Code:     code,
				MaxCount: 1,
			}

			_, err = s.promoCodeRepository.Create(ctx, promoCodeModel)
			if err != nil {
				s.log.Error("Failed to save unique promo code to db", zap.Error(err))
			}
		}
	} else {
		s.log.Error("Invalid promo mode", zap.String("mode", string(promoDto.Mode)))
		return "", domainerrors.ValidationError{Field: "mode", Message: "Invalid promo mode"}
	}

	return id, nil
}

func (s *Service) List(ctx context.Context, companyId string, countries []string, limit int, offset int, sortBy promoenum.SortBy) (promoDTOs []promo.DTO, err error) {

	promoModels, err := s.promoRepository.List(ctx, companyId, countries, offset, sortBy, limit)

	if err != nil {
		return nil, fmt.Errorf("promoRepository.List: %w", err)
	}

	promoDTOs = make([]promo.DTO, len(promoModels))

	for idx, promoModel := range promoModels {

		companyName, err := s.accountServiceClient.GetCompanyNameByCompanyID(ctx, promoModel.CompanyId)
		if err != nil {
			if status.Code(err) == codes.NotFound {
				s.log.Error("company does not exist", zap.String("company_id", promoModel.CompanyId))
				return nil, ErrCompanyDoesNotExist
			}
			return nil, fmt.Errorf("accountServiceClient.GetCompanyNameByCompanyID: %w", err)
		}

		promoDTOs[idx] = promo.DTO{
			PromoId:     promoModel.Id,
			CompanyId:   promoModel.CompanyId,
			CompanyName: companyName,
			Mode:        promoModel.Mode,
			Description: promoModel.Description,
			ImageURL:    promoModel.ImageUrl,
			Codes: functional.Map(promoModel.Codes, func(code promoStorage.CodeDTO) promo.Code {
				return promo.Code{
					Code:        code.Code,
					MaxCount:    code.MaxCount,
					Activations: code.Activations,
				}
			}),
			Target: &target.DTO{
				AgeFrom:    int64(promoModel.TargetAgeFrom),
				AgeUntil:   int64(promoModel.TargetAgeUntil),
				Country:    promoModel.TargetCountry,
				Categories: promoModel.TargetCategories,
			},
			ActiveFrom:  promoModel.ActiveFrom,
			ActiveUntil: promoModel.ActiveUntil,
		}
	}

	return promoDTOs, nil
}

func (s *Service) Count(ctx context.Context, companyId string, countries []string) (count int, err error) {
	count, err = s.promoRepository.Count(ctx, companyId, countries)
	if err != nil {
		return 0, fmt.Errorf("promoRepository.Count: %w", err)
	}
	return count, nil
}

func (s *Service) GetById(ctx context.Context, promoId string, companyId string) (promoDTO *promo.DTO, err error) {

	promoModel := new(promoStorage.PromoDetails)

	cashedPromoModel := s.redisDb.Get(ctx, promoId).Val()

	if cashedPromoModel != "" {
		err = json.Unmarshal([]byte(cashedPromoModel), promoModel)
		if err != nil {
			return nil, fmt.Errorf("json.Unmarshal: %w", err)
		}
	} else {
		promoModel, err = s.promoRepository.GetById(ctx, promoId)
		if err != nil {
			return nil, fmt.Errorf("promoRepository.GetById: %w", err)
		}

		if promoModel == nil {
			return nil, ErrNotFound
		}

		promoModelJSON, err := json.Marshal(promoModel)
		if err != nil {
			return nil, fmt.Errorf("json.Marshal: %w", err)
		}

		err = s.redisDb.Set(ctx, promoId, string(promoModelJSON), 0).Err()
		if err != nil {
			s.log.Warn("s.redisDb.Set: Failed to save promo to redis", zap.Error(err))
		}
	}

	if promoModel.CompanyId != companyId {
		return nil, ErrPermissionDenied
	}

	companyName, err := s.accountServiceClient.GetCompanyNameByCompanyID(ctx, promoModel.CompanyId)
	if err != nil {
		return nil, fmt.Errorf("accountServiceClient.GetCompanyNameByCompanyID: %w", err)
	}

	promoDTO = &promo.DTO{
		PromoId:     promoModel.Id,
		CompanyId:   promoModel.CompanyId,
		CompanyName: companyName,
		Mode:        promoModel.Mode,
		Description: promoModel.Description,
		ImageURL:    promoModel.ImageUrl,
		Codes: functional.Map(promoModel.Codes, func(code promoStorage.CodeDTO) promo.Code {
			return promo.Code{
				Code:        code.Code,
				MaxCount:    code.MaxCount,
				Activations: code.Activations,
			}
		}),
		Target: &target.DTO{
			AgeFrom:    int64(promoModel.TargetAgeFrom),
			AgeUntil:   int64(promoModel.TargetAgeUntil),
			Country:    promoModel.TargetCountry,
			Categories: promoModel.TargetCategories,
		},
		ActiveFrom:  promoModel.ActiveFrom,
		ActiveUntil: promoModel.ActiveUntil,
	}

	return promoDTO, nil
}

func (s *Service) Update(ctx context.Context,
	promoId string,
	companyId string,
	description string,
	imageUrl string,
	targetAgeFrom int64,
	targetAgeUntil int64,
	targetCountry string,
	targetCategories []string,
	activeFrom time.Time,
	activeUntil time.Time) error {

	err := validateUpdate(
		description,
		imageUrl,
		targetAgeFrom,
		targetAgeUntil,
		targetCountry,
		targetCategories,
		activeFrom,
		activeUntil,
	)
	if err != nil {
		return err
	}

	promoDTO, err := s.GetById(ctx, promoId, companyId)

	if err != nil {
		return fmt.Errorf("s.GetById: %w", err)
	}

	if promoDTO == nil {
		return ErrNotFound
	}

	if promoDTO.CompanyId != companyId {
		return ErrPermissionDenied
	}

	err = s.promoRepository.Update(ctx, promoId, description, imageUrl, targetAgeFrom, targetAgeUntil, targetCountry, targetCategories, activeFrom, activeUntil)

	defer func() {
		err = s.redisDb.Del(ctx, promoId).Err()
		if err != nil {
			s.log.Warn("s.redisDb.Del: Failed to delete promo from redis", zap.Error(err))
		}
	}()

	if err != nil {
		return fmt.Errorf("promoRepository.Update: %w", err)
	}

	return nil

}

func validateUpdate(description string,
	imageUrl string,
	targetAgeFrom int64,
	targetAgeUntil int64,
	targetCountry string,
	targetCategories []string,
	activeFrom time.Time,
	activeUntil time.Time,
) error {
	validate := validator.New()

	err := validate.Var(description, "required")
	if err != nil {
		return domainerrors.ValidationError{
			Field:   "description",
			Message: err.Error(),
		}

	}

	return nil
}

func (s *Service) Delete(ctx context.Context, promoId string, companyId string) error {
	promoDTO, err := s.GetById(ctx, promoId, companyId)

	if err != nil {
		return fmt.Errorf("promoRepository.GetById: %w", err)
	}

	if promoDTO == nil {
		return ErrNotFound
	}

	if promoDTO.CompanyId != companyId {
		return ErrPermissionDenied
	}

	defer func() {
		err = s.redisDb.Del(ctx, promoId).Err()
		if err != nil {
			s.log.Warn("s.redisDb.Del: Failed to delete promo from redis", zap.Error(err))
		}
	}()

	err = s.promoRepository.Delete(ctx, promoId)

	if err != nil {
		return fmt.Errorf("promoRepository.Delete: %w", err)
	}

	return nil
}

func (s *Service) Activate(ctx context.Context, promoId string) (code string, err error) {

	code, err = s.promoCodeRepository.Activate(ctx, promoId)

	if antifraud() {
		return "", ErrFraudDetected
	}

	if err != nil {

		if errors.Is(err, promo_code.ErrNoActivations) {
			return "", ErrNoActivations
		}

		return "", fmt.Errorf("promoCodeRepository.Activate: %w", err)
	}

	if code == "" {
		return "", ErrNotFound
	}

	return code, nil
}

// В будущем логика антифрода
func antifraud() bool {
	if rand.Intn(100) < 25 {
		return true
	}
	return false
}
