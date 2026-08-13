package usecase

import (
	"context"
	"ocean-express-api/internal/domain"
)

type customerUseCase struct {
	customerRepo domain.CustomerRepository
}

func NewCustomerUseCase(repo domain.CustomerRepository) domain.CustomerUseCase {
	return &customerUseCase{customerRepo: repo}
}

func (uc *customerUseCase) SaveCustomer(ctx context.Context, shopID, name, phone, locationID, addressDetail string, lat, lng *float64) (*domain.Customer, error) {
	c := &domain.Customer{
		ShopID:        shopID,
		Name:          name,
		Phone:         phone,
		LocationID:    locationID,
		AddressDetail: addressDetail,
		Latitude:      lat,
		Longitude:     lng,
	}
	err := uc.customerRepo.CreateOrUpdate(ctx, c)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (uc *customerUseCase) SearchCustomers(ctx context.Context, shopID, query string) ([]domain.Customer, error) {
	return uc.customerRepo.Search(ctx, shopID, query)
}
