package repository

import (
	"context"
	"ocean-express-api/internal/domain"

	"gorm.io/gorm"
)

type hubRepository struct {
	db *gorm.DB
}

func NewHubRepository(db *gorm.DB) domain.HubRepository {
	return &hubRepository{db: db}
}

func (r *hubRepository) FindAll(ctx context.Context, locationID *string) ([]*domain.Hub, error) {
	query := r.db.WithContext(ctx)
	if locationID != nil {
		query = query.Where("location_id = ?", *locationID)
	}

	var hubs []*domain.Hub
	err := query.Order("name ASC").Find(&hubs).Error
	return hubs, err
}

func (r *hubRepository) Create(ctx context.Context, hub *domain.Hub) error {
	return r.db.WithContext(ctx).Create(hub).Error
}

func (r *hubRepository) FindNearestHub(ctx context.Context, lat, lng float64) (*domain.Hub, error) {
	var hub domain.Hub
	// Haversine formula in Postgres to calculate distance in km
	// 6371 is the Earth's radius in km
	query := `
		SELECT *, 
			(6371 * acos(cos(radians(?)) * cos(radians(latitude)) * 
			cos(radians(longitude) - radians(?)) + 
			sin(radians(?)) * sin(radians(latitude)))) AS distance 
		FROM hubs 
		WHERE latitude IS NOT NULL AND longitude IS NOT NULL
		ORDER BY distance ASC 
		LIMIT 1
	`
	err := r.db.WithContext(ctx).Raw(query, lat, lng, lat).Scan(&hub).Error
	if err != nil {
		return nil, err
	}
	if hub.ID == "" {
		return nil, gorm.ErrRecordNotFound
	}
	return &hub, nil
}
