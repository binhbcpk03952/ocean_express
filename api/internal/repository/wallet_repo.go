package repository

import (
	"context"
	"ocean-express-api/internal/domain"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type walletRepository struct {
	db *gorm.DB
}

func NewWalletRepository(db *gorm.DB) domain.WalletRepository {
	return &walletRepository{db: db}
}

func (r *walletRepository) AddTransaction(ctx context.Context, tx *domain.WalletTransaction) error {
	return r.db.WithContext(ctx).Create(tx).Error
}

func (r *walletRepository) ListTransactions(ctx context.Context, shopID string) ([]*domain.WalletTransaction, error) {
	var txs []*domain.WalletTransaction
	err := r.db.WithContext(ctx).
		Where("shop_id = ?", shopID).
		Order("created_at DESC").
		Find(&txs).Error
	return txs, err
}

func (r *walletRepository) AvailableBalance(ctx context.Context, shopID string) (float64, error) {
	var balance float64
	// COALESCE để tổng rỗng trả 0 thay vì NULL.
	err := r.db.WithContext(ctx).
		Model(&domain.WalletTransaction{}).
		Where("shop_id = ? AND settlement_id IS NULL", shopID).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&balance).Error
	return balance, err
}

func (r *walletRepository) CreateSettlement(ctx context.Context, shopID, note string) (*domain.Settlement, error) {
	var settlement *domain.Settlement

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Khóa các bút toán chưa đối soát của shop để tránh race khi 2 admin cùng chốt.
		var txs []*domain.WalletTransaction
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("shop_id = ? AND settlement_id IS NULL", shopID).
			Find(&txs).Error; err != nil {
			return err
		}

		if len(txs) == 0 {
			return domain.ErrValidation
		}

		var total float64
		for _, t := range txs {
			total += t.Amount
		}

		s := &domain.Settlement{
			ShopID:      shopID,
			TotalAmount: total,
			Status:      domain.SettlementPending,
			Note:        note,
		}
		if err := tx.Create(s).Error; err != nil {
			return err
		}

		// Gán settlement_id cho toàn bộ bút toán vừa gom.
		if err := tx.Model(&domain.WalletTransaction{}).
			Where("shop_id = ? AND settlement_id IS NULL", shopID).
			Update("settlement_id", s.ID).Error; err != nil {
			return err
		}

		settlement = s
		return nil
	})

	return settlement, err
}

func (r *walletRepository) MarkPaid(ctx context.Context, settlementID string) (*domain.Settlement, error) {
	var s domain.Settlement
	if err := r.db.WithContext(ctx).Where("id = ?", settlementID).First(&s).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	if err := r.db.WithContext(ctx).
		Model(&s).
		Updates(map[string]interface{}{
			"status":  domain.SettlementPaid,
			"paid_at": gorm.Expr("CURRENT_TIMESTAMP"),
		}).Error; err != nil {
		return nil, err
	}

	// Đọc lại để trả paid_at đã set.
	if err := r.db.WithContext(ctx).Where("id = ?", settlementID).First(&s).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *walletRepository) HasOrderTransactions(ctx context.Context, orderID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.WalletTransaction{}).
		Where("order_id = ?", orderID).
		Count(&count).Error
	return count > 0, err
}

func (r *walletRepository) HasOrderReturnTransaction(ctx context.Context, orderID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.WalletTransaction{}).
		Where("order_id = ? AND type = ?", orderID, domain.WalletTxReturnFee).
		Count(&count).Error
	return count > 0, err
}

func (r *walletRepository) ListSettlements(ctx context.Context, shopID string) ([]*domain.Settlement, error) {
	var settlements []*domain.Settlement
	query := r.db.WithContext(ctx)
	if shopID != "" {
		query = query.Where("shop_id = ?", shopID)
	}
	err := query.Order("created_at DESC").Find(&settlements).Error
	return settlements, err
}
