package repository

import (
	"context"
	"ocean-express-api/internal/domain"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type customerRepository struct {
	db *gorm.DB
}

func NewCustomerRepository(db *gorm.DB) domain.CustomerRepository {
	return &customerRepository{db: db}
}

func (r *customerRepository) CreateOrUpdate(ctx context.Context, c *domain.Customer) error {
	var existing domain.Customer
	err := r.db.WithContext(ctx).Where("shop_id = ? AND phone = ?", c.ShopID, c.Phone).First(&existing).Error
	if err == nil {
		c.ID = existing.ID
		return r.db.WithContext(ctx).Model(&existing).Updates(c).Error
	}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: false}).Create(c).Error
}

func (r *customerRepository) Search(ctx context.Context, shopID string, query string) ([]domain.Customer, error) {
	var customers []domain.Customer
	q := r.db.WithContext(ctx).Where("shop_id = ?", shopID)
	if query != "" {
		likeQ := "%" + query + "%"
		q = q.Where("name LIKE ? OR phone LIKE ?", likeQ, likeQ)
	}
	err := q.Order("updated_at desc").Limit(20).Find(&customers).Error
	return customers, err
}

func (r *customerRepository) GetByPhone(ctx context.Context, shopID string, phone string) (*domain.Customer, error) {
	var customer domain.Customer
	err := r.db.WithContext(ctx).Where("shop_id = ? AND phone = ?", shopID, phone).First(&customer).Error
	if err != nil {
		return nil, err
	}
	return &customer, nil
}
