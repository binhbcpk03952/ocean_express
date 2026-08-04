package notification

import (
	"context"
	"log"
	"ocean-express-api/internal/domain"
)

// stubService là bản hiện thực log-only của NotificationService.
// Nó tra danh sách thiết bị của nhân viên rồi ghi log thay vì gọi FCM thật —
// để tầng nghiệp vụ chạy được offline. Khi cắm Firebase, chỉ cần thay thân
// goroutine bên dưới bằng lời gọi FCM SDK, không đổi interface hay tầng gọi.
type stubService struct {
	deviceRepo domain.DeviceRepository
}

// NewStubService tạo NotificationService log-only.
func NewStubService(deviceRepo domain.DeviceRepository) domain.NotificationService {
	return &stubService{deviceRepo: deviceRepo}
}

// NotifyEmployee gửi thông báo tới toàn bộ thiết bị của một nhân viên.
// Fire-and-forget qua goroutine để không block request gốc (giống webhook).
func (s *stubService) NotifyEmployee(employeeID string, msg domain.PushMessage) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[Push] panic khi gửi thông báo cho %s: %v", employeeID, r)
			}
		}()

		if employeeID == "" {
			return
		}

		devices, err := s.deviceRepo.FindByEmployee(context.Background(), employeeID)
		if err != nil {
			log.Printf("[Push] không lấy được thiết bị của %s: %v", employeeID, err)
			return
		}
		if len(devices) == 0 {
			log.Printf("[Push] nhân viên %s chưa đăng ký thiết bị nào, bỏ qua thông báo %q", employeeID, msg.Title)
			return
		}

		for _, d := range devices {
			// STUB: thay dòng log này bằng lời gọi Firebase FCM (token = d.FCMToken).
			log.Printf("[Push→%s] token=%s title=%q body=%q data=%v",
				d.DeviceName, maskToken(d.FCMToken), msg.Title, msg.Body, msg.Data)
		}
	}()
}

// maskToken ẩn bớt token trong log để không lộ giá trị đầy đủ.
func maskToken(t string) string {
	if len(t) <= 8 {
		return "***"
	}
	return t[:4] + "..." + t[len(t)-4:]
}
