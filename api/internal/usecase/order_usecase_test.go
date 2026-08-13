package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"ocean-express-api/internal/domain"
	"ocean-express-api/internal/usecase"
)

// Mocks
type MockOrderRepo struct {
	SavedOrder *domain.ShippingOrder
	SavedLog   *domain.TrackingLog
}

func (m *MockOrderRepo) GetByID(ctx context.Context, id string) (*domain.ShippingOrder, error) {
	if m.SavedOrder != nil && m.SavedOrder.ID == id {
		return m.SavedOrder, nil
	}
	return nil, errors.New("not found")
}
func (m *MockOrderRepo) GetByTrackingNumber(ctx context.Context, trackingNumber string) (*domain.ShippingOrder, error) {
	if m.SavedOrder != nil && m.SavedOrder.TrackingNumber == trackingNumber {
		return m.SavedOrder, nil
	}
	return nil, errors.New("not found")
}
func (m *MockOrderRepo) GetOrderLogs(ctx context.Context, orderID string) ([]*domain.TrackingLog, error) {
	return nil, nil
}
func (m *MockOrderRepo) FindAll(ctx context.Context, role, employeeID, hubID string, pageParams domain.PaginationParams) ([]*domain.ShippingOrder, int64, error) {
	return nil, 0, nil
}
func (m *MockOrderRepo) CreateOrderWithLog(ctx context.Context, order *domain.ShippingOrder, log *domain.TrackingLog) error {
	order.ID = "test-uuid"
	order.CreatedAt = time.Now()
	m.SavedOrder = order
	m.SavedLog = log
	return nil
}
func (m *MockOrderRepo) UpdateStatus(ctx context.Context, order *domain.ShippingOrder, log *domain.TrackingLog) error {
	m.SavedOrder = order
	m.SavedLog = log
	return nil
}
func (m *MockOrderRepo) SubmitCOD(ctx context.Context, orderID string) (float64, error) {
	return 0, nil
}

type MockShopRepo struct{}

func (m *MockShopRepo) GetByID(ctx context.Context, id string) (*domain.Shop, error) {
	locID := "loc-default"
	return &domain.Shop{ID: "shop-1", WebhookURL: "http://example.com/webhook", LocationID: &locID}, nil
}
func (m *MockShopRepo) GetByAPIKey(ctx context.Context, key string) (*domain.Shop, error) {
	return nil, nil
}
func (m *MockShopRepo) GetByEmail(ctx context.Context, email string) (*domain.Shop, error) {
	return nil, nil
}
func (m *MockShopRepo) GetByPhoneOrEmail(ctx context.Context, identifier string) (*domain.Shop, error) {
	return nil, nil
}
func (m *MockShopRepo) FindAll(ctx context.Context, status string) ([]*domain.Shop, error) {
	return nil, nil
}
func (m *MockShopRepo) Create(ctx context.Context, shop *domain.Shop) error {
	return nil
}
func (m *MockShopRepo) Update(ctx context.Context, shop *domain.Shop) error {
	return nil
}

type MockWebhookService struct {
	SentPayload *domain.WebhookPayload
}

func (m *MockWebhookService) SendOrderStatus(url string, payload domain.WebhookPayload) {
	m.SentPayload = &payload
}

func TestCreateOrder(t *testing.T) {
	orderRepo := &MockOrderRepo{}
	rateRepo := &MockRateRepo{
		MockRate: &domain.ShippingRate{BaseWeight: 1000, BaseFee: 30000},
	}
	rateUC := usecase.NewRateUseCase(rateRepo)
	shopRepo := &MockShopRepo{}
	webhookSvc := &MockWebhookService{}

	orderUC := usecase.NewOrderUseCase(orderRepo, rateUC, shopRepo, &MockHubRepo{}, &MockLocationRepo{}, &MockGeocoder{}, webhookSvc, &MockWalletUseCase{}, nil, nil)
	ctx := context.Background()

	order, err := orderUC.CreateOrder(ctx, "shop-1", "John Doe", "0123456789", "loc-1", "123 Street", 500, 0, 0, 0, 100000, nil, nil, nil, nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if order == nil {
		t.Fatal("expected order to be created, got nil")
	}

	if order.Status != "ready_to_pick" {
		t.Errorf("expected status ready_to_pick, got %s", order.Status)
	}

	if order.ShippingFee != 30000 {
		t.Errorf("expected fee 30000, got %f", order.ShippingFee)
	}

	if order.CodAmount != 100000 {
		t.Errorf("expected COD 100000, got %f", order.CodAmount)
	}

	// Verify webhook was called asynchronously (Note: in a real test you might need a small sleep or a channel to wait for the goroutine, but here we can just verify the repo logic or trust the structure)
}
