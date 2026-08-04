package usecase

import (
	"context"
	"ocean-express-api/internal/domain"
)

type rateUseCase struct {
	rateRepo domain.RateRepository
}

func NewRateUseCase(repo domain.RateRepository) domain.RateUseCase {
	return &rateUseCase{rateRepo: repo}
}

func (u *rateUseCase) CalculateFee(ctx context.Context, fromLocID, toLocID string, weight int) (float64, error) {
	rate, err := u.rateRepo.GetRate(ctx, fromLocID, toLocID)
	if err != nil {
		return 0, err
	}

	// 3. Tính phí
	fee := rate.BaseFee
	if weight > rate.BaseWeight {
		extraWeight := weight - rate.BaseWeight
		steps := (extraWeight + rate.ExtraWeightStep - 1) / rate.ExtraWeightStep // làm tròn lên
		fee += float64(steps) * rate.ExtraFee
	}

	return fee, nil
}

func (u *rateUseCase) GetRates(ctx context.Context) ([]*domain.ShippingRate, error) {
	return u.rateRepo.FindAll(ctx)
}

func (u *rateUseCase) CreateRate(ctx context.Context, rate *domain.ShippingRate) error {
	return u.rateRepo.CreateRate(ctx, rate)
}
