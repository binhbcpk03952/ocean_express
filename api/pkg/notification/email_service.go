package notification

import (
	"log"
	"ocean-express-api/internal/domain"
)

type stubEmailService struct{}

// NewStubEmailService creates a log-only EmailService for testing.
func NewStubEmailService() domain.EmailService {
	return &stubEmailService{}
}

// SendShopApprovedEmail simulates sending an approval email to the shop.
func (s *stubEmailService) SendShopApprovedEmail(email string) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[Email] panic khi gửi email cho %s: %v", email, r)
			}
		}()

		if email == "" {
			return
		}

		// STUB: Thay dòng log này bằng logic gửi email thật (ví dụ SendGrid, SMTP).
		log.Printf("[Email→%s] Chủ đề: \"Tài khoản của bạn đã được duyệt\". Nội dung: \"Chúc mừng! Tài khoản đối tác của bạn đã được duyệt. Vui lòng đăng nhập vào hệ thống để lấy API Key của bạn.\"", email)
	}()
}

func (s *stubEmailService) SendOTP(toEmail, otp, role string) error {
	log.Printf("[Email→%s] [STUB] Chủ đề: \"Mã OTP Khôi Phục Mật Khẩu\". Mã OTP của %s là: %s", toEmail, role, otp)
	return nil
}
