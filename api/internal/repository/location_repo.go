package repository

import (
	"context"
	"ocean-express-api/internal/domain"

	"gorm.io/gorm"
)

type locationRepository struct {
	db *gorm.DB
}

func NewLocationRepository(db *gorm.DB) domain.LocationRepository {
	return &locationRepository{db: db}
}

func (r *locationRepository) GetByID(ctx context.Context, id string) (*domain.Location, error) {
	var loc domain.Location
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&loc).Error
	if err != nil {
		return nil, err
	}
	return &loc, nil
}

func (r *locationRepository) FindAll(ctx context.Context, parentID *string, locType *string) ([]*domain.Location, error) {
	query := r.db.WithContext(ctx)
	if parentID != nil {
		query = query.Where("parent_id = ?", *parentID)
	}
	if locType != nil {
		query = query.Where("type = ?", *locType)
	}

	var locs []*domain.Location
	err := query.Order("name ASC").Find(&locs).Error
	return locs, err
}

func (r *locationRepository) Create(ctx context.Context, loc *domain.Location) error {
	return r.db.WithContext(ctx).Create(loc).Error
}
