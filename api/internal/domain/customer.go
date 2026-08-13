package domain

import (
	"context"
	"time"
)

type Customer struct {
	ID            string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	ShopID        string    `gorm:"type:uuid;not null;index" json:"shop_id"`
	Name          string    `gorm:"type:varchar(255);not null" json:"name"`
	Phone         string    `gorm:"type:varchar(20);not null;index" json:"phone"`
	LocationID    string    `gorm:"type:varchar(50)" json:"location_id"`
	AddressDetail string    `gorm:"type:text" json:"address_detail"`
	Latitude      *float64  `gorm:"type:decimal(10,8)" json:"latitude"`
	Longitude     *float64  `gorm:"type:decimal(11,8)" json:"longitude"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type CustomerRepository interface {
	CreateOrUpdate(ctx context.Context, customer *Customer) error
	Search(ctx context.Context, shopID string, query string) ([]Customer, error)
	GetByPhone(ctx context.Context, shopID string, phone string) (*Customer, error)
}

type CustomerUseCase interface {
	SaveCustomer(ctx context.Context, shopID, name, phone, locationID, addressDetail string, lat, lng *float64) (*Customer, error)
	SearchCustomers(ctx context.Context, shopID, query string) ([]Customer, error)
}
