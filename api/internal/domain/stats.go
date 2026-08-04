package domain

import "context"

// DashboardStats tổng hợp số liệu giám sát toàn cục cho Admin.
type DashboardStats struct {
	TotalOrders     int64            `json:"total_orders"`
	StatusCounts    map[string]int64 `json:"status_counts"`    // số đơn theo từng trạng thái
	DeliveredCount  int64            `json:"delivered_count"`  // đã giao thành công
	ReturnedCount   int64            `json:"returned_count"`   // đã hoàn
	InProgressCount int64            `json:"in_progress_count"` // đang trong luồng (chưa delivered/returned)
	SuccessRate     float64          `json:"success_rate"`     // delivered / (delivered + returned), phần trăm
	CongestedHubs   []HubLoad        `json:"congested_hubs"`   // các hub đang giữ nhiều đơn nhất
}

// HubLoad biểu diễn số đơn đang tồn tại một hub (phục vụ phát hiện ùn tắc).
type HubLoad struct {
	HubID   string `json:"hub_id"`
	HubName string `json:"hub_name"`
	Count   int64  `json:"count"`
}

// MemberStats thống kê hiệu suất cho tài xế/nhân viên kho.
type MemberStats struct {
	TotalDelivered  int64   `json:"total_delivered"` // Shipper: số đơn giao thành công
	TotalFailed     int64   `json:"total_failed"`    // Shipper: số đơn hoàn
	TotalCodHolding float64 `json:"total_cod_holding"` // Shipper: tổng COD đang giữ
	TotalScans      int64   `json:"total_scans"`     // HubStaff: số lượt quét mã (từ tracking_logs)
}

// ShopStats thống kê hiệu suất cho shop.
type ShopStats struct {
	TotalOrders    int64   `json:"total_orders"`
	TotalShippingFee float64 `json:"total_shipping_fee"`
	TotalCod       float64 `json:"total_cod"` // Tổng COD chưa đối soát
	ReturnRate     float64 `json:"return_rate"` // Tỷ lệ hoàn hàng
	StatusCounts   map[string]int64 `json:"status_counts"` // số đơn theo từng trạng thái
}

// StatsRepository truy vấn số liệu tổng hợp từ CSDL.
type StatsRepository interface {
	CountOrders(ctx context.Context) (int64, error)
	CountByStatus(ctx context.Context) (map[string]int64, error)
	CongestedHubs(ctx context.Context, limit int) ([]HubLoad, error)
	
	GetMemberStats(ctx context.Context, memberID string, role string) (*MemberStats, error)
	GetShopStats(ctx context.Context, shopID string) (*ShopStats, error)
}

// StatsUseCase tổng hợp số liệu cho dashboard.
type StatsUseCase interface {
	GetDashboard(ctx context.Context) (*DashboardStats, error)
	GetMemberStats(ctx context.Context, memberID string, role string) (*MemberStats, error)
	GetShopStats(ctx context.Context, shopID string) (*ShopStats, error)
}
