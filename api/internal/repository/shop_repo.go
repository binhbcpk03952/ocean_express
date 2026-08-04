package repository

import (
	"context"
	"errors"
	"ocean-express-api/internal/domain"

	"gorm.io/gorm"
)

type shopRepository struct {
	db *gorm.DB
}

func NewShopRepository(db *gorm.DB) domain.ShopRepository {
	return &shopRepository{db: db}
}

func (r *shopRepository) GetByAPIKey(ctx context.Context, apiKey string) (*domain.Shop, error) {
	var shop domain.Shop
	err := r.db.WithContext(ctx).Where("api_key = ?", apiKey).First(&shop).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &shop, nil
}

func (r *shopRepository) GetByID(ctx context.Context, id string) (*domain.Shop, error) {
	var shop domain.Shop
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&shop).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &shop, nil
}

func (r *shopRepository) GetByEmail(ctx context.Context, email string) (*domain.Shop, error) {
	var shop domain.Shop
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&shop).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &shop, nil
}

func (r *shopRepository) GetByPhoneOrEmail(ctx context.Context, identifier string) (*domain.Shop, error) {
	var shop domain.Shop
	err := r.db.WithContext(ctx).Where("email = ? OR phone = ?", identifier, identifier).First(&shop).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &shop, nil
}

func (r *shopRepository) FindAll(ctx context.Context, status string) ([]*domain.Shop, error) {
	var shops []*domain.Shop
	query := r.db.WithContext(ctx)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	err := query.Order("created_at DESC").Find(&shops).Error
	return shops, err
}

func (r *shopRepository) Create(ctx context.Context, shop *domain.Shop) error {
	return r.db.WithContext(ctx).Create(shop).Error
}

func (r *shopRepository) Update(ctx context.Context, shop *domain.Shop) error {
	// Select tường minh các cột được phép ghi. Mọi luồng gọi Update đều load shop
	// đầy đủ qua GetByID trước khi sửa, nên api_key/password_hash mang giá trị hiện
	// hữu (không bị ghi rỗng). ReviewShop set api_key mới khi duyệt lần đầu.
	return r.db.WithContext(ctx).Model(shop).
		Select("name", "email", "password_hash", "webhook_url", "api_key", "location_id", "address_detail", "status", "is_active").
		Updates(shop).Error
}
