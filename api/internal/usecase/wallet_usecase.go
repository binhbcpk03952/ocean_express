package usecase

import (
	"context"
	"fmt"
	"ocean-express-api/internal/domain"
)

type walletUseCase struct {
	walletRepo domain.WalletRepository
}

func NewWalletUseCase(repo domain.WalletRepository) domain.WalletUseCase {
	return &walletUseCase{walletRepo: repo}
}

// RecordCOD ghi 2 bút toán khi đơn giao thành công: +COD thu hộ và -cước.
// Idempotent theo đơn: nếu đơn đã có bút toán thì bỏ qua (tránh ghi trùng khi
// UpdateOrderStatus bị gọi lại). Chỉ ghi bút toán COD dương nếu cod_amount > 0.
func (u *walletUseCase) RecordCOD(ctx context.Context, order *domain.ShippingOrder) error {
	if order == nil {
		return fmt.Errorf("%w: đơn hàng rỗng", domain.ErrValidation)
	}

	has, err := u.walletRepo.HasOrderTransactions(ctx, order.ID)
	if err != nil {
		return err
	}
	if has {
		// Đã ghi nhận rồi -> không làm gì (idempotent).
		return nil
	}

	// Bút toán +COD thu hộ (chỉ ghi nếu có tiền thu hộ).
	if order.CodAmount > 0 {
		if err := u.walletRepo.AddTransaction(ctx, &domain.WalletTransaction{
			ShopID:  order.ShopID,
			OrderID: &order.ID,
			Type:    domain.WalletTxCodCredit,
			Amount:  order.CodAmount,
			Note:    fmt.Sprintf("Thu hộ COD đơn %s", order.TrackingNumber),
		}); err != nil {
			return err
		}
	}

	// Bút toán -cước vận chuyển (amount âm).
	if order.ShippingFee > 0 {
		if err := u.walletRepo.AddTransaction(ctx, &domain.WalletTransaction{
			ShopID:  order.ShopID,
			OrderID: &order.ID,
			Type:    domain.WalletTxFeeDebit,
			Amount:  -order.ShippingFee,
			Note:    fmt.Sprintf("Cước vận chuyển đơn %s", order.TrackingNumber),
		}); err != nil {
			return err
		}
	}

	return nil
}

func (u *walletUseCase) RecordReturnFee(ctx context.Context, order *domain.ShippingOrder) error {
	if order == nil {
		return fmt.Errorf("%w: đơn hàng rỗng", domain.ErrValidation)
	}

	has, err := u.walletRepo.HasOrderReturnTransaction(ctx, order.ID)
	if err != nil {
		return err
	}
	if has {
		return nil
	}

	if order.ReturnFee > 0 {
		if err := u.walletRepo.AddTransaction(ctx, &domain.WalletTransaction{
			ShopID:  order.ShopID,
			OrderID: &order.ID,
			Type:    domain.WalletTxReturnFee,
			Amount:  -order.ReturnFee,
			Note:    fmt.Sprintf("Phí hoàn hàng đơn %s", order.TrackingNumber),
		}); err != nil {
			return err
		}
	}

	return nil
}

func (u *walletUseCase) GetWallet(ctx context.Context, shopID string) (float64, []*domain.WalletTransaction, error) {
	balance, err := u.walletRepo.AvailableBalance(ctx, shopID)
	if err != nil {
		return 0, nil, err
	}
	txs, err := u.walletRepo.ListTransactions(ctx, shopID)
	if err != nil {
		return 0, nil, err
	}
	return balance, txs, nil
}

func (u *walletUseCase) CreateSettlement(ctx context.Context, shopID, note string) (*domain.Settlement, error) {
	if shopID == "" {
		return nil, fmt.Errorf("%w: thiếu shop_id", domain.ErrValidation)
	}
	return u.walletRepo.CreateSettlement(ctx, shopID, note)
}

func (u *walletUseCase) MarkSettlementPaid(ctx context.Context, settlementID string) (*domain.Settlement, error) {
	return u.walletRepo.MarkPaid(ctx, settlementID)
}

func (u *walletUseCase) ListSettlements(ctx context.Context, shopID string) ([]*domain.Settlement, error) {
	return u.walletRepo.ListSettlements(ctx, shopID)
}
