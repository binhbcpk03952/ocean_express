package usecase_test

import (
	"context"
	"errors"

	"ocean-express-api/internal/domain"
	"ocean-express-api/pkg/geocoding"
)

// Các mock bổ sung cho những dependency được thêm vào NewOrderUseCase
// sau đợt refactor (hubRepo, locRepo, geocoder, walletUC).
// Mặc định chúng hành xử "trung tính": không lỗi, không side-effect,
// để các test hiện có tập trung vào luồng đơn hàng và state machine.

type MockHubRepo struct {
	NearestHub *domain.Hub
}

func (m *MockHubRepo) FindAll(ctx context.Context, locationID *string) ([]*domain.Hub, error) {
	if m.NearestHub != nil {
		return []*domain.Hub{m.NearestHub}, nil
	}
	return nil, nil
}

func (m *MockHubRepo) FindNearestHub(ctx context.Context, lat, lng float64) (*domain.Hub, error) {
	if m.NearestHub != nil {
		return m.NearestHub, nil
	}
	return nil, errors.New("không tìm thấy hub phù hợp")
}

func (m *MockHubRepo) Create(ctx context.Context, hub *domain.Hub) error { return nil }

type MockLocationRepo struct {
	MockLocation *domain.Location
}

func (m *MockLocationRepo) GetByID(ctx context.Context, id string) (*domain.Location, error) {
	if m.MockLocation != nil {
		return m.MockLocation, nil
	}
	return &domain.Location{ID: id, Name: "Địa điểm test"}, nil
}

func (m *MockLocationRepo) FindAll(ctx context.Context, parentID *string, locType *string) ([]*domain.Location, error) {
	return nil, nil
}

func (m *MockLocationRepo) Create(ctx context.Context, loc *domain.Location) error { return nil }

// MockGeocoder trả về toạ độ cố định, không gọi ra Nominatim thật khi chạy test.
type MockGeocoder struct {
	Coords *geocoding.Coordinates
	Err    error
}

func (m *MockGeocoder) GetCoordinates(address string) (*geocoding.Coordinates, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	if m.Coords != nil {
		return m.Coords, nil
	}
	return &geocoding.Coordinates{Latitude: 21.0278, Longitude: 105.8342}, nil
}

// MockWalletUseCase ghi nhận việc RecordCOD có được gọi hay không,
// phục vụ kiểm tra luồng giao hàng thành công.
type MockWalletUseCase struct {
	RecordedOrders []*domain.ShippingOrder
	RecordErr      error
}

func (m *MockWalletUseCase) RecordCOD(ctx context.Context, order *domain.ShippingOrder) error {
	if m.RecordErr != nil {
		return m.RecordErr
	}
	m.RecordedOrders = append(m.RecordedOrders, order)
	return nil
}

func (m *MockWalletUseCase) GetWallet(ctx context.Context, shopID string) (float64, []*domain.WalletTransaction, error) {
	return 0, nil, nil
}

func (m *MockWalletUseCase) CreateSettlement(ctx context.Context, shopID, note string) (*domain.Settlement, error) {
	return nil, nil
}

func (m *MockWalletUseCase) MarkSettlementPaid(ctx context.Context, settlementID string) (*domain.Settlement, error) {
	return nil, nil
}

func (m *MockWalletUseCase) ListSettlements(ctx context.Context, shopID string) ([]*domain.Settlement, error) {
	return nil, nil
}
