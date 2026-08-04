package repository

import (
	"context"
	"errors"
	"ocean-express-api/internal/domain"

	"gorm.io/gorm"
)

type rateRepository struct {
	db *gorm.DB
}

func NewRateRepository(db *gorm.DB) domain.RateRepository {
	return &rateRepository{db: db}
}

func (r *rateRepository) GetRate(ctx context.Context, fromLocID, toLocID string) (*domain.ShippingRate, error) {
	var rate domain.ShippingRate
	err := r.db.WithContext(ctx).
		Where("(? = from_location_id OR ? LIKE from_location_id || '-%' OR from_location_id IS NULL) AND (? = to_location_id OR ? LIKE to_location_id || '-%' OR to_location_id IS NULL)", fromLocID, fromLocID, toLocID, toLocID).
		Order("COALESCE(LENGTH(from_location_id), 0) + COALESCE(LENGTH(to_location_id), 0) DESC").
		First(&rate).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("không tìm thấy bảng giá phù hợp")
		}
		return nil, err
	}
	return &rate, nil
}

func (r *rateRepository) FindAll(ctx context.Context) ([]*domain.ShippingRate, error) {
	var rates []*domain.ShippingRate
	err := r.db.WithContext(ctx).Order("id DESC").Find(&rates).Error
	return rates, err
}

func (r *rateRepository) CreateRate(ctx context.Context, rate *domain.ShippingRate) error {
	return r.db.WithContext(ctx).Create(rate).Error
}
