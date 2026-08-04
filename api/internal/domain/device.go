package domain

import "context"

// DeviceRepository quản lý thiết bị (FCM token) của nhân viên/tài xế trong CSDL.
type DeviceRepository interface {
	// Upsert đăng ký hoặc cập nhật một thiết bị theo fcm_token (idempotent:
	// cùng token đăng nhập lại chỉ refresh, không tạo bản ghi trùng).
	Upsert(ctx context.Context, device *EmployeeDevice) error
	// FindByEmployee trả về danh sách thiết bị đang hoạt động của một nhân viên.
	FindByEmployee(ctx context.Context, employeeID string) ([]*EmployeeDevice, error)
	// DeleteByToken gỡ đăng ký một thiết bị (logout thiết bị / gỡ app).
	DeleteByToken(ctx context.Context, employeeID, fcmToken string) error
}

// DeviceUseCase xử lý logic đăng ký/gỡ thiết bị của nhân viên đang đăng nhập.
type DeviceUseCase interface {
	// Register đăng ký thiết bị hiện tại cho nhân viên (từ token trong context auth).
	Register(ctx context.Context, employeeID, deviceName, fcmToken string) (*EmployeeDevice, error)
	// Unregister gỡ thiết bị của nhân viên.
	Unregister(ctx context.Context, employeeID, fcmToken string) error
}

// PushMessage là nội dung một thông báo đẩy tới thiết bị.
type PushMessage struct {
	Title string            `json:"title"`
	Body  string            `json:"body"`
	Data  map[string]string `json:"data,omitempty"` // payload phụ (vd order_id, status) để app điều hướng
}

// NotificationService gửi push notification tới các thiết bị của một nhân viên.
// Hiện tại có bản stub (log-only); sau này thay bằng Firebase FCM thật mà không
// đổi tầng gọi (usecase chỉ phụ thuộc interface này).
type NotificationService interface {
	// NotifyEmployee gửi thông báo tới toàn bộ thiết bị của một nhân viên (bất đồng bộ).
	NotifyEmployee(employeeID string, msg PushMessage)
}
