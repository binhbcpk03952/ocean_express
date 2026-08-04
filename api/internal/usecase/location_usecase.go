package usecase

import (
	"context"
	"errors"
	"ocean-express-api/internal/domain"
)

type locationUseCase struct {
	locationRepo domain.LocationRepository
}

func NewLocationUseCase(repo domain.LocationRepository) domain.LocationUseCase {
	return &locationUseCase{locationRepo: repo}
}

func (u *locationUseCase) GetLocations(ctx context.Context, parentID *string, locType *string) ([]*domain.Location, error) {
	return u.locationRepo.FindAll(ctx, parentID, locType)
}

func (u *locationUseCase) CreateLocation(ctx context.Context, id, name, locType string, parentID *string) (*domain.Location, error) {
	if id == "" || name == "" || locType == "" {
		return nil, errors.New("id, name và type không được để trống")
	}

	loc := &domain.Location{
		ID:       id,
		Name:     name,
		Type:     locType,
		ParentID: parentID,
	}

	if err := u.locationRepo.Create(ctx, loc); err != nil {
		return nil, err
	}
	return loc, nil
}
