package domain

import (
	"context"
	"time"
)

// Hub đại diện cho một bưu cục / kho hàng (hubs table)
type Hub struct {
	ID            string    `json:"id" gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	Name          string    `json:"name" gorm:"column:name"`
	Type          string    `json:"type" gorm:"column:type"`
	LocationID    *string   `json:"location_id" gorm:"column:location_id"`
	AddressDetail string    `json:"address_detail" gorm:"column:address_detail"`
	Latitude      *float64  `json:"latitude" gorm:"column:latitude"`
	Longitude     *float64  `json:"longitude" gorm:"column:longitude"`
	CreatedAt     time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
}

func (Hub) TableName() string {
	return "hubs"
}

// HubRepository quản lý tương tác CSDL cho Hub
type HubRepository interface {
	FindAll(ctx context.Context, locationID *string) ([]*Hub, error)
	FindNearestHub(ctx context.Context, lat, lng float64) (*Hub, error)
	Create(ctx context.Context, hub *Hub) error
}

// HubUseCase xử lý logic nghiệp vụ cho Hub
type HubUseCase interface {
	GetHubs(ctx context.Context, locationID *string) ([]*Hub, error)
	CreateHub(ctx context.Context, name string, hubType string, locationID *string, addressDetail string, lat *float64, lng *float64) (*Hub, error)
}
