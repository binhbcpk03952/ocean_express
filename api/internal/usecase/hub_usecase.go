package usecase

import (
	"context"
	"errors"
	"ocean-express-api/internal/domain"
)

type hubUseCase struct {
	hubRepo      domain.HubRepository
	locationRepo domain.LocationRepository
}

func NewHubUseCase(hubRepo domain.HubRepository, locRepo domain.LocationRepository) domain.HubUseCase {
	return &hubUseCase{hubRepo: hubRepo, locationRepo: locRepo}
}

func (u *hubUseCase) GetHubs(ctx context.Context, locationID *string) ([]*domain.Hub, error) {
	return u.hubRepo.FindAll(ctx, locationID)
}

func (u *hubUseCase) CreateHub(ctx context.Context, name string, hubType string, locationID *string, addressDetail string, lat *float64, lng *float64) (*domain.Hub, error) {
	hub := &domain.Hub{
		Name:          name,
		Type:          hubType,
		LocationID:    locationID,
		AddressDetail: addressDetail,
		Latitude:      lat,
		Longitude:     lng,
	}
	if name == "" || addressDetail == "" {
		return nil, errors.New("tên và địa chỉ chi tiết không được để trống")
	}

	// Validate location exists if provided
	if locationID != nil {
		locs, err := u.locationRepo.FindAll(ctx, nil, nil)
		if err != nil {
			return nil, err
		}
		
		found := false
		for _, loc := range locs {
			if loc.ID == *locationID {
				found = true
				break
			}
		}
		if !found {
			return nil, errors.New("khu vực không hợp lệ")
		}
	}



	if err := u.hubRepo.Create(ctx, hub); err != nil {
		return nil, err
	}
	return hub, nil
}
