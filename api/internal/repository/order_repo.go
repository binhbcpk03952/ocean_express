package repository

import (
	"context"
	"ocean-express-api/internal/domain"

	"gorm.io/gorm"
)

type orderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) domain.OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) GetByID(ctx context.Context, id string) (*domain.ShippingOrder, error) {
	var order domain.ShippingOrder
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) FindAll(ctx context.Context, role, employeeID, hubID string, pageParams domain.PaginationParams) ([]*domain.ShippingOrder, int64, error) {
	var orders []*domain.ShippingOrder
	var total int64
	query := r.db.WithContext(ctx).Model(&domain.ShippingOrder{})

	if role == "first_mile_driver" {
		if hubID != "" {
			query = query.Where("current_driver_id = ? OR (current_driver_id IS NULL AND status = 'ready_to_pick' AND pickup_hub_id = ?)", employeeID, hubID)
		} else {
			query = query.Where("current_driver_id = ?", employeeID)
		}
	} else if role == "last_mile_driver" {
		if hubID != "" {
			query = query.Where("current_driver_id = ? OR (current_driver_id IS NULL AND status = 'hub_outbound' AND delivery_hub_id = ?)", employeeID, hubID)
		} else {
			query = query.Where("current_driver_id = ?", employeeID)
		}
	} else if role == "hub_staff" && hubID != "" {
		// Hub staff chỉ thấy đơn đang nằm tại bưu cục của mình
		query = query.Where("current_hub_id = ?", hubID)
	} else if role == "shop" {
		// Shop portal: chỉ thấy đơn của chính mình. employeeID mang shop.ID.
		query = query.Where("shop_id = ?", employeeID)
	}
	// Admin xem toàn bộ

	// Đếm tổng số bản ghi
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("created_at DESC").
		Offset(pageParams.GetOffset()).
		Limit(pageParams.GetLimit()).
		Find(&orders).Error
	return orders, total, err
}

func (r *orderRepository) GetByTrackingNumber(ctx context.Context, trackingNumber string) (*domain.ShippingOrder, error) {
	var order domain.ShippingOrder
	err := r.db.WithContext(ctx).Where("tracking_number = ?", trackingNumber).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) GetOrderLogs(ctx context.Context, orderID string) ([]*domain.TrackingLog, error) {
	var logs []*domain.TrackingLog
	err := r.db.WithContext(ctx).
		Table("tracking_logs").
		Select("tracking_logs.*, employees.name as employee_name").
		Joins("LEFT JOIN employees ON employees.id = tracking_logs.employee_id").
		Where("tracking_logs.order_id = ?", orderID).
		Order("tracking_logs.created_at ASC").
		Find(&logs).Error
	return logs, err
}

func (r *orderRepository) CreateOrderWithLog(ctx context.Context, order *domain.ShippingOrder, log *domain.TrackingLog) error {
	// Sử dụng Transaction để đảm bảo tính toàn vẹn dữ liệu
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(order).Error; err != nil {
			return err
		}
		
		log.OrderID = order.ID // Cập nhật OrderID sau khi insert (nếu ko truyền uuid trước)
		if err := tx.Create(log).Error; err != nil {
			return err
		}
		
		return nil
	})
}

func (r *orderRepository) UpdateStatus(ctx context.Context, order *domain.ShippingOrder, log *domain.TrackingLog) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Cập nhật trạng thái + tài xế đang ôm đơn (current_driver_id có thể được gán hoặc reset).
		// Dùng Select để GORM ghi cả giá trị NULL cho current_driver_id khi reset.
		if err := tx.Model(order).
			Select("status", "current_driver_id", "current_hub_id", "updated_at").
			Updates(order).Error; err != nil {
			return err
		}

		// Ghi log mới
		if err := tx.Create(log).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *orderRepository) SubmitCOD(ctx context.Context, driverID string) (float64, error) {
	var total float64
	
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Table("shipping_orders").
			Where("current_driver_id = ? AND status = ? AND cod_collected = ?", driverID, "delivered", false).
			Select("COALESCE(SUM(cod_amount), 0)").
			Scan(&total).Error
		if err != nil {
			return err
		}

		if total > 0 {
			err = tx.Table("shipping_orders").
				Where("current_driver_id = ? AND status = ? AND cod_collected = ?", driverID, "delivered", false).
				Updates(map[string]interface{}{
					"cod_collected": true,
					"cod_collected_at": gorm.Expr("NOW()"),
				}).Error
			if err != nil {
				return err
			}
		}
		
		return nil
	})
	
	return total, err
}
