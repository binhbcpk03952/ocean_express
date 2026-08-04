package repository

import (
	"context"
	"ocean-express-api/internal/domain"

	"gorm.io/gorm"
)

type statsRepository struct {
	db *gorm.DB
}

func NewStatsRepository(db *gorm.DB) domain.StatsRepository {
	return &statsRepository{db: db}
}

func (r *statsRepository) CountOrders(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.ShippingOrder{}).Count(&count).Error
	return count, err
}

func (r *statsRepository) CountByStatus(ctx context.Context) (map[string]int64, error) {
	type row struct {
		Status string
		Count  int64
	}
	var rows []row
	err := r.db.WithContext(ctx).
		Model(&domain.ShippingOrder{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	result := make(map[string]int64, len(rows))
	for _, r := range rows {
		result[r.Status] = r.Count
	}
	return result, nil
}

// CongestedHubs trả về các hub đang giữ nhiều đơn nhất (đơn chưa giao/hoàn mà đang gán tại hub).
func (r *statsRepository) CongestedHubs(ctx context.Context, limit int) ([]domain.HubLoad, error) {
	var hubs []domain.HubLoad
	err := r.db.WithContext(ctx).
		Table("shipping_orders AS o").
		Select("o.current_hub_id AS hub_id, h.name AS hub_name, COUNT(*) AS count").
		Joins("JOIN hubs h ON h.id = o.current_hub_id").
		Where("o.current_hub_id IS NOT NULL AND o.status NOT IN ?", []string{"delivered", "returned"}).
		Group("o.current_hub_id, h.name").
		Order("count DESC").
		Limit(limit).
		Scan(&hubs).Error
	return hubs, err
}

func (r *statsRepository) GetMemberStats(ctx context.Context, memberID string, role string) (*domain.MemberStats, error) {
	stats := &domain.MemberStats{}
	
	if role == string(domain.RoleHubStaff) {
		// Đếm số lượt quét mã trong bảng tracking_logs
		var count int64
		err := r.db.WithContext(ctx).Table("tracking_logs").
			Where("employee_id = ?", memberID).
			Count(&count).Error
		if err != nil {
			return nil, err
		}
		stats.TotalScans = count
	} else {
		// Shipper (first_mile_driver, last_mile_driver)
		var delivered, failed int64
		r.db.WithContext(ctx).Table("shipping_orders").
			Where("current_driver_id = ? AND status = ?", memberID, "delivered").
			Count(&delivered)
		
		r.db.WithContext(ctx).Table("shipping_orders").
			Where("current_driver_id = ? AND status = ?", memberID, "returned").
			Count(&failed)
			
		var codHolding float64
		r.db.WithContext(ctx).Table("shipping_orders").
			Where("current_driver_id = ? AND status = ? AND cod_collected = false", memberID, "delivered").
			Select("COALESCE(SUM(cod_amount), 0)").
			Scan(&codHolding)
			
		stats.TotalDelivered = delivered
		stats.TotalFailed = failed
		stats.TotalCodHolding = codHolding
	}
	
	return stats, nil
}

func (r *statsRepository) GetShopStats(ctx context.Context, shopID string) (*domain.ShopStats, error) {
	stats := &domain.ShopStats{}
	
	var totalOrders int64
	r.db.WithContext(ctx).Table("shipping_orders").
		Where("shop_id = ?", shopID).
		Count(&totalOrders)
		
	var shippingFee float64
	r.db.WithContext(ctx).Table("shipping_orders").
		Where("shop_id = ?", shopID).
		Select("COALESCE(SUM(shipping_fee), 0)").
		Scan(&shippingFee)
		
	var returnedCount int64
	r.db.WithContext(ctx).Table("shipping_orders").
		Where("shop_id = ? AND status = ?", shopID, "returned").
		Count(&returnedCount)
		
	if totalOrders > 0 {
		stats.ReturnRate = float64(returnedCount) / float64(totalOrders) * 100
	}
	
	rows, err := r.db.WithContext(ctx).Table("shipping_orders").
		Select("status, COUNT(id) as count").
		Where("shop_id = ?", shopID).
		Group("status").
		Rows()
	
	statusCounts := make(map[string]int64)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var status string
			var count int64
			if err := rows.Scan(&status, &count); err == nil {
				statusCounts[status] = count
			}
		}
	}
	
	stats.StatusCounts = statusCounts
	stats.TotalOrders = totalOrders
	stats.TotalShippingFee = shippingFee
	// TotalCod chưa đối soát sẽ phức tạp hơn nếu dựa vào wallet, ta có thể lấy tổng COD của các đơn delivered mà shop chưa rút.
	// Hiện tại tính tổng COD của các đơn được giao thành công nhưng chưa chuyển vào wallet. Hoặc tính bằng wallet_repo.
	
	return stats, nil
}
