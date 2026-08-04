package usecase_test

import (
	"context"
	"testing"

	"ocean-express-api/internal/domain"
	"ocean-express-api/internal/usecase"
)

// MockRateRepo implements domain.RateRepository for testing
type MockRateRepo struct {
	MockRate *domain.ShippingRate
	MockErr  error
}

func (m *MockRateRepo) GetRate(ctx context.Context, fromLocID, toLocID string) (*domain.ShippingRate, error) {
	return m.MockRate, m.MockErr
}
func (m *MockRateRepo) FindAll(ctx context.Context) ([]*domain.ShippingRate, error) {
	return nil, nil
}
func (m *MockRateRepo) CreateRate(ctx context.Context, rate *domain.ShippingRate) error {
	return nil
}

func TestCalculateFee(t *testing.T) {
	mockRepo := &MockRateRepo{
		MockRate: &domain.ShippingRate{
			BaseWeight:      1000,
			BaseFee:         30000,
			ExtraWeightStep: 500,
			ExtraFee:        5000,
		},
		MockErr: nil,
	}

	rateUC := usecase.NewRateUseCase(mockRepo)
	ctx := context.Background()

	tests := []struct {
		name        string
		weight      int
		expectedFee float64
	}{
		{"Weight below base", 500, 30000},
		{"Weight equal base", 1000, 30000},
		{"Weight slightly above base", 1100, 35000},
		{"Weight exactly one step above", 1500, 35000},
		{"Weight exactly two steps above", 2000, 40000},
		{"Weight three steps above", 2001, 45000},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fee, err := rateUC.CalculateFee(ctx, "A", "B", tc.weight)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if fee != tc.expectedFee {
				t.Errorf("expected fee %v, got %v", tc.expectedFee, fee)
			}
		})
	}
}
