package domain

import (
	"context"
	"time"
)

// Loại bút toán ví shop.
const (
	WalletTxCodCredit = "cod_credit" // +COD thu hộ khi đơn delivered
	WalletTxFeeDebit  = "fee_debit"  // -cước vận chuyển
)

// Trạng thái phiên chi trả.
const (
	SettlementPending = "pending" // đã chốt, chờ chi tiền
	SettlementPaid    = "paid"    // đã chi cho shop
)

// WalletTransaction là một bút toán có dấu trong sổ cái ví của shop.
// amount > 0 cộng vào số dư shop (COD thu hộ), amount < 0 trừ (cước).
// SettlementID = nil nghĩa là chưa đối soát (còn nằm trong số dư khả dụng).
type WalletTransaction struct {
	ID           string    `json:"id" gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	ShopID       string    `json:"shop_id" gorm:"column:shop_id"`
	OrderID      *string   `json:"order_id" gorm:"column:order_id"`
	Type         string    `json:"type" gorm:"column:type"`
	Amount       float64   `json:"amount" gorm:"column:amount"`
	SettlementID *string   `json:"settlement_id" gorm:"column:settlement_id"`
	Note         string    `json:"note" gorm:"column:note"`
	CreatedAt    time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
}

func (WalletTransaction) TableName() string {
	return "wallet_transactions"
}

// Settlement là một phiên chi trả gom nhiều bút toán ví chưa đối soát của shop.
type Settlement struct {
	ID          string     `json:"id" gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	ShopID      string     `json:"shop_id" gorm:"column:shop_id"`
	TotalAmount float64    `json:"total_amount" gorm:"column:total_amount"`
	Status      string     `json:"status" gorm:"column:status;default:pending"`
	Note        string     `json:"note" gorm:"column:note"`
	CreatedAt   time.Time  `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	PaidAt      *time.Time `json:"paid_at" gorm:"column:paid_at"`
}

func (Settlement) TableName() string {
	return "settlements"
}

// WalletRepository quản lý sổ cái ví và phiên đối soát.
type WalletRepository interface {
	// AddTransaction ghi một bút toán mới vào ví shop.
	AddTransaction(ctx context.Context, tx *WalletTransaction) error
	// HasOrderTransactions cho biết đơn đã có bút toán ví chưa (phục vụ idempotency
	// khi ghi nhận COD lúc delivered — tránh ghi trùng nếu gọi lại).
	HasOrderTransactions(ctx context.Context, orderID string) (bool, error)
	// ListTransactions liệt kê bút toán của shop (mới nhất trước).
	ListTransactions(ctx context.Context, shopID string) ([]*WalletTransaction, error)
	// AvailableBalance tính tổng bút toán CHƯA đối soát (settlement_id IS NULL) của shop.
	AvailableBalance(ctx context.Context, shopID string) (float64, error)
	// CreateSettlement gom toàn bộ bút toán chưa đối soát của shop thành một phiên chi trả,
	// gán settlement_id cho chúng và ghi total_amount. Thực hiện trong transaction.
	// Trả về phiên vừa tạo; nếu không có bút toán nào chưa đối soát trả ErrValidation.
	CreateSettlement(ctx context.Context, shopID, note string) (*Settlement, error)
	// MarkPaid đánh dấu phiên đã chi tiền.
	MarkPaid(ctx context.Context, settlementID string) (*Settlement, error)
	// ListSettlements liệt kê phiên chi trả; shopID rỗng = tất cả (cho admin).
	ListSettlements(ctx context.Context, shopID string) ([]*Settlement, error)
}

// WalletUseCase xử lý logic ví shop và đối soát.
type WalletUseCase interface {
	// RecordCOD ghi 2 bút toán khi đơn giao thành công: +COD và -cước.
	// Idempotent theo order: nếu đơn đã ghi nhận rồi thì bỏ qua.
	RecordCOD(ctx context.Context, order *ShippingOrder) error
	// GetWallet trả về số dư khả dụng + lịch sử bút toán của shop.
	GetWallet(ctx context.Context, shopID string) (float64, []*WalletTransaction, error)
	// CreateSettlement (admin) chốt phiên chi trả cho shop.
	CreateSettlement(ctx context.Context, shopID, note string) (*Settlement, error)
	// MarkSettlementPaid (admin) đánh dấu đã chi tiền.
	MarkSettlementPaid(ctx context.Context, settlementID string) (*Settlement, error)
	// ListSettlements liệt kê phiên chi trả (admin: tất cả; shop: của mình).
	ListSettlements(ctx context.Context, shopID string) ([]*Settlement, error)
}
