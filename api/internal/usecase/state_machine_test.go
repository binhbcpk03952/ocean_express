package usecase_test

import (
	"context"
	"errors"
	"testing"

	"ocean-express-api/internal/domain"
	"ocean-express-api/internal/usecase"
)

// seedOrderRepo tạo MockOrderRepo với một đơn hàng ở trạng thái cho trước.
func seedOrderRepo(status string, driverID *string) *MockOrderRepo {
	return &MockOrderRepo{
		SavedOrder: &domain.ShippingOrder{
			ID:              "order-1",
			TrackingNumber:  "OE-TEST",
			ShopID:          "shop-1",
			Status:          status,
			CurrentDriverID: driverID,
		},
	}
}

func newOrderUC(orderRepo *MockOrderRepo) domain.OrderUseCase {
	rateUC := usecase.NewRateUseCase(&MockRateRepo{MockRate: &domain.ShippingRate{BaseWeight: 1000, BaseFee: 30000}})
	return usecase.NewOrderUseCase(orderRepo, rateUC, &MockShopRepo{}, &MockWebhookService{}, nil)
}

func TestUpdateStatus_ValidTransitions(t *testing.T) {
	tests := []struct {
		name     string
		from     string
		to       string
		role     string
		wantErr  bool
	}{
		{"first-mile lấy hàng", "ready_to_pick", "picked_up", "first_mile_driver", false},
		{"hub nhập kho", "picked_up", "hub_inbound", "hub_staff", false},
		{"hub xuất kho", "hub_inbound", "hub_outbound", "hub_staff", false},
		{"last-mile giao hàng", "hub_outbound", "delivering", "last_mile_driver", false},
		{"giao thành công", "delivering", "delivered", "last_mile_driver", false},
		{"admin bỏ qua ràng buộc role", "ready_to_pick", "picked_up", "admin", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := seedOrderRepo(tc.from, nil)
			uc := newOrderUC(repo)
			err := uc.UpdateOrderStatus(context.Background(), "order-1", tc.to, "", "emp-1", tc.role, "hub-1", 0, 0)
			if tc.wantErr && err == nil {
				t.Fatalf("mong đợi lỗi nhưng không có")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("không mong đợi lỗi, nhận: %v", err)
			}
			if !tc.wantErr && repo.SavedOrder.Status != tc.to {
				t.Errorf("mong đợi status %s, nhận %s", tc.to, repo.SavedOrder.Status)
			}
		})
	}
}

func TestUpdateStatus_InvalidTransitionIsRejected(t *testing.T) {
	// Không được nhảy thẳng từ ready_to_pick sang delivered (bỏ qua hub).
	repo := seedOrderRepo("ready_to_pick", nil)
	uc := newOrderUC(repo)
	err := uc.UpdateOrderStatus(context.Background(), "order-1", "delivered", "", "emp-1", "admin", "", 0, 0)
	if err == nil {
		t.Fatal("mong đợi lỗi khi bỏ qua hub, nhưng không có")
	}
}

func TestUpdateStatus_RoleEnforcement(t *testing.T) {
	// last_mile_driver không được thực hiện thao tác lấy hàng (picked_up).
	repo := seedOrderRepo("ready_to_pick", nil)
	uc := newOrderUC(repo)
	err := uc.UpdateOrderStatus(context.Background(), "order-1", "picked_up", "", "emp-1", "last_mile_driver", "", 0, 0)
	if err == nil {
		t.Fatal("mong đợi last_mile_driver bị chặn khỏi picked_up, nhưng không có lỗi")
	}
}

func TestUpdateStatus_DriverOwnershipGuard(t *testing.T) {
	// Đơn đang được gán cho driver khác -> driver hiện tại không được thao tác.
	otherDriver := "driver-other"
	repo := seedOrderRepo("delivering", &otherDriver)
	uc := newOrderUC(repo)
	err := uc.UpdateOrderStatus(context.Background(), "order-1", "delivered", "", "driver-me", "last_mile_driver", "", 0, 0)
	if err == nil {
		t.Fatal("mong đợi driver bị chặn khỏi đơn của người khác, nhưng không có lỗi")
	}
}

func TestUpdateStatus_HubInboundResetsDriverAndSetsHub(t *testing.T) {
	firstMile := "driver-first"
	repo := seedOrderRepo("picked_up", &firstMile)
	uc := newOrderUC(repo)
	err := uc.UpdateOrderStatus(context.Background(), "order-1", "hub_inbound", "", "staff-1", "hub_staff", "hub-42", 0, 0)
	if err != nil {
		t.Fatalf("không mong đợi lỗi, nhận: %v", err)
	}
	if repo.SavedOrder.CurrentDriverID != nil {
		t.Errorf("mong đợi current_driver_id được reset về nil khi nhập kho")
	}
	if repo.SavedOrder.CurrentHubID == nil || *repo.SavedOrder.CurrentHubID != "hub-42" {
		t.Errorf("mong đợi current_hub_id = hub-42 khi nhập kho")
	}
}

func TestGetOrderDetailsByTrackingNumber(t *testing.T) {
	repo := seedOrderRepo("hub_inbound", nil)
	repo.SavedOrder.TrackingNumber = "OE-123456"
	uc := newOrderUC(repo)

	order, _, err := uc.GetOrderDetailsByTrackingNumber(context.Background(), "OE-123456")
	if err != nil {
		t.Fatalf("không mong đợi lỗi, nhận: %v", err)
	}
	if order == nil || order.ID != "order-1" {
		t.Fatalf("mong đợi tìm thấy đơn order-1 theo tracking number")
	}

	if _, _, err := uc.GetOrderDetailsByTrackingNumber(context.Background(), "OE-KHONG-TON-TAI"); err == nil {
		t.Fatal("mong đợi lỗi khi tra cứu mã vận đơn không tồn tại")
	}
}

// Các test dưới đây bảo vệ việc phân loại lỗi (sentinel) để handler map đúng HTTP
// status (403/422). Nếu ai đó đổi lại thành errors.New thô, test sẽ bắt được.

func TestUpdateStatus_InvalidTransitionWrapsSentinel(t *testing.T) {
	// Nhảy sai luồng (ready_to_pick -> delivered) phải bọc ErrInvalidTransition -> 422.
	repo := seedOrderRepo("ready_to_pick", nil)
	uc := newOrderUC(repo)
	err := uc.UpdateOrderStatus(context.Background(), "order-1", "delivered", "", "emp-1", "admin", "", 0, 0)
	if err == nil {
		t.Fatal("mong đợi lỗi khi nhảy sai luồng trạng thái")
	}
	if !errors.Is(err, domain.ErrInvalidTransition) {
		t.Fatalf("mong đợi lỗi bọc ErrInvalidTransition (map 422), nhận: %v", err)
	}
	if domain.HTTPStatusForError(err) != 422 {
		t.Fatalf("mong đợi HTTP 422 cho lỗi transition, nhận: %d", domain.HTTPStatusForError(err))
	}
}

func TestUpdateStatus_WrongRoleWrapsForbidden(t *testing.T) {
	// last_mile_driver không được thực hiện picked_up -> bọc ErrForbidden -> 403.
	repo := seedOrderRepo("ready_to_pick", nil)
	uc := newOrderUC(repo)
	err := uc.UpdateOrderStatus(context.Background(), "order-1", "picked_up", "", "emp-1", "last_mile_driver", "", 0, 0)
	if err == nil {
		t.Fatal("mong đợi lỗi khi sai role")
	}
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("mong đợi lỗi bọc ErrForbidden (map 403), nhận: %v", err)
	}
	if domain.HTTPStatusForError(err) != 403 {
		t.Fatalf("mong đợi HTTP 403 cho lỗi sai role, nhận: %d", domain.HTTPStatusForError(err))
	}
}

func TestUpdateStatus_DriverOwnershipWrapsForbidden(t *testing.T) {
	// Đơn đang do driver khác giữ -> driver hiện tại bị chặn với ErrForbidden -> 403.
	otherDriver := "driver-other"
	repo := seedOrderRepo("delivering", &otherDriver)
	uc := newOrderUC(repo)
	err := uc.UpdateOrderStatus(context.Background(), "order-1", "delivered", "", "driver-me", "last_mile_driver", "", 0, 0)
	if err == nil {
		t.Fatal("mong đợi lỗi khi thao tác đơn của người khác")
	}
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("mong đợi lỗi bọc ErrForbidden (map 403), nhận: %v", err)
	}
}
