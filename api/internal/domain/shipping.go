package domain

import (
	"context"
	"time"
)

// ShippingRate cấu hình tính tiền cước
type ShippingRate struct {
	ID              int       `json:"id" gorm:"primaryKey;autoIncrement"`
	FromLocationID  *string   `json:"from_location_id" gorm:"column:from_location_id"`
	ToLocationID    *string   `json:"to_location_id" gorm:"column:to_location_id"`
	BaseWeight      int       `json:"base_weight" gorm:"column:base_weight;default:1000"`
	BaseFee         float64   `json:"base_fee" gorm:"column:base_fee"`
	ExtraWeightStep int       `json:"extra_weight_step" gorm:"column:extra_weight_step"`
	ExtraFee        float64   `json:"extra_fee" gorm:"column:extra_fee"`
}

func (ShippingRate) TableName() string {
	return "shipping_rates"
}

// ShippingOrder đại diện cho đơn hàng
type ShippingOrder struct {
	ID                    string     `json:"id" gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	TrackingNumber        string     `json:"tracking_number" gorm:"column:tracking_number;unique"`
	ShopID                string     `json:"shop_id" gorm:"column:shop_id"`
	SenderPhone           *string    `json:"sender_phone" gorm:"column:sender_phone"`
	SenderLocationID      *string    `json:"sender_location_id" gorm:"column:sender_location_id"`
	SenderAddressDetail   string     `json:"sender_address_detail" gorm:"column:sender_address_detail"`
	SenderLatitude        *float64   `json:"sender_latitude" gorm:"column:sender_latitude"`
	SenderLongitude       *float64   `json:"sender_longitude" gorm:"column:sender_longitude"`
	ReceiverName          string     `json:"receiver_name" gorm:"column:receiver_name"`
	ReceiverPhone         string     `json:"receiver_phone" gorm:"column:receiver_phone"`
	ReceiverLocationID    *string    `json:"receiver_location_id" gorm:"column:receiver_location_id"`
	ReceiverAddressDetail string     `json:"receiver_address_detail" gorm:"column:receiver_address_detail"`
	ReceiverLatitude      *float64   `json:"receiver_latitude" gorm:"column:receiver_latitude"`
	ReceiverLongitude     *float64   `json:"receiver_longitude" gorm:"column:receiver_longitude"`
	Weight                int        `json:"weight" gorm:"column:weight"`
	Length                int        `json:"length" gorm:"column:length;default:0"`
	Width                 int        `json:"width" gorm:"column:width;default:0"`
	Height                int        `json:"height" gorm:"column:height;default:0"`
	ShippingFee           float64    `json:"shipping_fee" gorm:"column:shipping_fee"`
	CodAmount             float64    `json:"cod_amount" gorm:"column:cod_amount;default:0"`
	EstimatedDeliveryTime *time.Time `json:"estimated_delivery_time" gorm:"column:estimated_delivery_time"`
	Status                string     `json:"status" gorm:"column:status;default:ready_to_pick"`
	CurrentDriverID       *string    `json:"current_driver_id" gorm:"column:current_driver_id"`
	CurrentHubID          *string    `json:"current_hub_id" gorm:"column:current_hub_id"`
	PickupHubID           *string    `json:"pickup_hub_id" gorm:"column:pickup_hub_id"`
	DeliveryHubID         *string    `json:"delivery_hub_id" gorm:"column:delivery_hub_id"`
	CodCollected          bool       `json:"cod_collected" gorm:"column:cod_collected;default:false"`
	CodCollectedAt        *time.Time `json:"cod_collected_at" gorm:"column:cod_collected_at"`
	DeliveryAttempts      int        `json:"delivery_attempts" gorm:"column:delivery_attempts;default:0"`
	FailureReason         *string    `json:"failure_reason" gorm:"column:failure_reason"`
	SlaDeadline           *time.Time `json:"sla_deadline" gorm:"column:sla_deadline"`
	SlaBreached           bool       `json:"sla_breached" gorm:"column:sla_breached;default:false"`
	ReturnFee             float64    `json:"return_fee" gorm:"column:return_fee;default:0"`
	IsReturnConfirmed     bool       `json:"is_return_confirmed" gorm:"column:is_return_confirmed;default:false"`
	CreatedAt             time.Time  `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt             time.Time  `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

func (ShippingOrder) TableName() string {
	return "shipping_orders"
}

// TrackingLog lưu log hành trình
type TrackingLog struct {
	ID         string    `json:"id" gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	OrderID    string    `json:"order_id" gorm:"column:order_id"`
	Status     string    `json:"status" gorm:"column:status"`
	Note       string    `json:"note" gorm:"column:note"`
	EmployeeID   *string   `json:"employee_id" gorm:"column:employee_id"`
	EmployeeName *string   `json:"employee_name,omitempty" gorm:"-"`
	Latitude     float64   `json:"latitude" gorm:"column:latitude"`
	Longitude    float64   `json:"longitude" gorm:"column:longitude"`
	CreatedAt    time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
}

func (TrackingLog) TableName() string {
	return "tracking_logs"
}

// RateRepository quản lý dữ liệu bảng giá cước
type RateRepository interface {
	GetRate(ctx context.Context, fromLocID, toLocID string) (*ShippingRate, error)
	FindAll(ctx context.Context) ([]*ShippingRate, error)
	CreateRate(ctx context.Context, rate *ShippingRate) error
}

// RateUseCase xử lý logic tính cước phí
type RateUseCase interface {
	CalculateFee(ctx context.Context, fromLocID, toLocID string, weight int) (float64, error)
	GetRates(ctx context.Context) ([]*ShippingRate, error)
	CreateRate(ctx context.Context, rate *ShippingRate) error
}

// OrderRepository quản lý vận đơn và log hành trình
type OrderRepository interface {
	GetByID(ctx context.Context, id string) (*ShippingOrder, error)
	GetByTrackingNumber(ctx context.Context, trackingNumber string) (*ShippingOrder, error)
	FindAll(ctx context.Context, role, employeeID, hubID string, pageParams PaginationParams) ([]*ShippingOrder, int64, error)
	GetOrderLogs(ctx context.Context, orderID string) ([]*TrackingLog, error)
	CreateOrderWithLog(ctx context.Context, order *ShippingOrder, log *TrackingLog) error
	UpdateStatus(ctx context.Context, order *ShippingOrder, log *TrackingLog) error
	SubmitCOD(ctx context.Context, driverID string) (float64, error)
}

// OrderUseCase xử lý logic tạo vận đơn
type OrderUseCase interface {
	CreateOrder(ctx context.Context, shopID, receiverName, receiverPhone, receiverLocID, receiverAddress string, weight, length, width, height int, codAmount float64, senderLat, senderLng, receiverLat, receiverLng *float64) (*ShippingOrder, error)
	UpdateOrderStatus(ctx context.Context, orderID, status, note, failureReason, employeeID, employeeRole, employeeHubID string, lat, lng float64) error
	GetOrders(ctx context.Context, role, employeeID, hubID string, pageParams PaginationParams) (*PaginatedResponse, error)
	GetOrderDetails(ctx context.Context, id string) (*ShippingOrder, []*TrackingLog, error)
	GetOrderDetailsByTrackingNumber(ctx context.Context, trackingNumber string) (*ShippingOrder, []*TrackingLog, error)
	SubmitCOD(ctx context.Context, driverID string) (float64, error)
	AssignOrder(ctx context.Context, orderID, shipperID, assignerID, role string) error
}
