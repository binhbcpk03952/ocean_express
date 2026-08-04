package domain

import (
	"context"
	"time"
)

// Location đại diện cho một đơn vị hành chính (locations table)
type Location struct {
	ID        string    `json:"id" gorm:"primaryKey;column:id"`
	Name      string    `json:"name" gorm:"column:name"`
	Type      string    `json:"type" gorm:"column:type"` // province, district, ward
	ParentID  *string   `json:"parent_id" gorm:"column:parent_id"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
}

func (Location) TableName() string {
	return "locations"
}

// LocationRepository quản lý tương tác CSDL cho Location
type LocationRepository interface {
	GetByID(ctx context.Context, id string) (*Location, error)
	FindAll(ctx context.Context, parentID *string, locType *string) ([]*Location, error)
	Create(ctx context.Context, loc *Location) error
}

// LocationUseCase xử lý logic nghiệp vụ cho Location
type LocationUseCase interface {
	GetLocations(ctx context.Context, parentID *string, locType *string) ([]*Location, error)
	CreateLocation(ctx context.Context, id, name, locType string, parentID *string) (*Location, error)
}
