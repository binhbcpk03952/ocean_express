package usecase

import (
	"context"
	"ocean-express-api/internal/domain"
)

type statsUseCase struct {
	statsRepo domain.StatsRepository
}

func NewStatsUseCase(repo domain.StatsRepository) domain.StatsUseCase {
	return &statsUseCase{statsRepo: repo}
}

func (u *statsUseCase) GetDashboard(ctx context.Context) (*domain.DashboardStats, error) {
	total, err := u.statsRepo.CountOrders(ctx)
	if err != nil {
		return nil, err
	}

	statusCounts, err := u.statsRepo.CountByStatus(ctx)
	if err != nil {
		return nil, err
	}

	congested, err := u.statsRepo.CongestedHubs(ctx, 5)
	if err != nil {
		return nil, err
	}

	delivered := statusCounts["delivered"]
	returned := statusCounts["returned"]
	inProgress := total - delivered - returned

	// Tỷ lệ giao thành công tính trên các đơn đã kết thúc luồng (delivered + returned).
	var successRate float64
	if finished := delivered + returned; finished > 0 {
		successRate = float64(delivered) / float64(finished) * 100
	}

	return &domain.DashboardStats{
		TotalOrders:     total,
		StatusCounts:    statusCounts,
		DeliveredCount:  delivered,
		ReturnedCount:   returned,
		InProgressCount: inProgress,
		SuccessRate:     successRate,
		CongestedHubs:   congested,
	}, nil
}

func (u *statsUseCase) GetMemberStats(ctx context.Context, memberID string, role string) (*domain.MemberStats, error) {
	return u.statsRepo.GetMemberStats(ctx, memberID, role)
}

func (u *statsUseCase) GetShopStats(ctx context.Context, shopID string) (*domain.ShopStats, error) {
	return u.statsRepo.GetShopStats(ctx, shopID)
}
