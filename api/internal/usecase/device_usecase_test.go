package usecase_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"ocean-express-api/internal/domain"
	"ocean-express-api/internal/usecase"
	"ocean-express-api/pkg/notification"
)

// mockDeviceRepo là bản ghi nhớ (in-memory) của DeviceRepository cho test.
type mockDeviceRepo struct {
	mu      sync.Mutex
	devices []*domain.EmployeeDevice
	failAll bool
}

func (m *mockDeviceRepo) Upsert(ctx context.Context, device *domain.EmployeeDevice) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.failAll {
		return errors.New("db lỗi")
	}
	// Idempotent theo fcm_token: nếu token đã có thì cập nhật, không thêm trùng.
	for _, d := range m.devices {
		if d.FCMToken == device.FCMToken {
			d.EmployeeID = device.EmployeeID
			d.DeviceName = device.DeviceName
			return nil
		}
	}
	m.devices = append(m.devices, device)
	return nil
}

func (m *mockDeviceRepo) FindByEmployee(ctx context.Context, employeeID string) ([]*domain.EmployeeDevice, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []*domain.EmployeeDevice
	for _, d := range m.devices {
		if d.EmployeeID == employeeID {
			out = append(out, d)
		}
	}
	return out, nil
}

func (m *mockDeviceRepo) DeleteByToken(ctx context.Context, employeeID, fcmToken string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	kept := m.devices[:0]
	for _, d := range m.devices {
		if d.EmployeeID == employeeID && d.FCMToken == fcmToken {
			continue
		}
		kept = append(kept, d)
	}
	m.devices = kept
	return nil
}

func TestRegisterDevice_Persists(t *testing.T) {
	repo := &mockDeviceRepo{}
	uc := usecase.NewDeviceUseCase(repo)

	dev, err := uc.Register(context.Background(), "emp-1", "iPhone 15", "tok-abc")
	if err != nil {
		t.Fatalf("không mong đợi lỗi: %v", err)
	}
	if dev.EmployeeID != "emp-1" || dev.FCMToken != "tok-abc" {
		t.Fatalf("thiết bị lưu sai: %+v", dev)
	}
	if len(repo.devices) != 1 {
		t.Fatalf("mong đợi 1 thiết bị, có %d", len(repo.devices))
	}
}

func TestRegisterDevice_IdempotentOnSameToken(t *testing.T) {
	repo := &mockDeviceRepo{}
	uc := usecase.NewDeviceUseCase(repo)

	if _, err := uc.Register(context.Background(), "emp-1", "iPhone 15", "tok-abc"); err != nil {
		t.Fatalf("lần 1 lỗi: %v", err)
	}
	// Cùng token đăng ký lại (vd mở app lại) — không được tạo bản ghi trùng.
	if _, err := uc.Register(context.Background(), "emp-1", "iPhone 15 Pro", "tok-abc"); err != nil {
		t.Fatalf("lần 2 lỗi: %v", err)
	}
	if len(repo.devices) != 1 {
		t.Fatalf("mong đợi vẫn 1 thiết bị (idempotent), có %d", len(repo.devices))
	}
	if repo.devices[0].DeviceName != "iPhone 15 Pro" {
		t.Errorf("mong đợi device_name được refresh, có %q", repo.devices[0].DeviceName)
	}
}

func TestRegisterDevice_RejectsEmptyToken(t *testing.T) {
	uc := usecase.NewDeviceUseCase(&mockDeviceRepo{})
	if _, err := uc.Register(context.Background(), "emp-1", "iPhone", ""); err == nil {
		t.Fatal("mong đợi lỗi khi fcm_token rỗng")
	}
}

func TestUnregisterDevice_Removes(t *testing.T) {
	repo := &mockDeviceRepo{}
	uc := usecase.NewDeviceUseCase(repo)

	_, _ = uc.Register(context.Background(), "emp-1", "iPhone", "tok-abc")
	if err := uc.Unregister(context.Background(), "emp-1", "tok-abc"); err != nil {
		t.Fatalf("unregister lỗi: %v", err)
	}
	if len(repo.devices) != 0 {
		t.Fatalf("mong đợi thiết bị bị gỡ, còn %d", len(repo.devices))
	}
}

// TestNotificationStub_DoesNotPanic đảm bảo stub log-only chạy an toàn (kể cả khi
// nhân viên chưa có thiết bị) và không làm sập tiến trình gọi.
func TestNotificationStub_DoesNotPanic(t *testing.T) {
	repo := &mockDeviceRepo{}
	_ = repo.Upsert(context.Background(), &domain.EmployeeDevice{EmployeeID: "emp-1", FCMToken: "tok-abc", DeviceName: "iPhone"})

	svc := notification.NewStubService(repo)
	// Fire-and-forget qua goroutine — chỉ cần không panic ở tầng gọi.
	svc.NotifyEmployee("emp-1", domain.PushMessage{Title: "Đơn mới", Body: "Bạn có đơn cần lấy"})
	svc.NotifyEmployee("emp-khong-co-thiet-bi", domain.PushMessage{Title: "x", Body: "y"})
}
