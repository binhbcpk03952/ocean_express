package notification

import (
	"fmt"
	"log"
	"net/smtp"
	"ocean-express-api/internal/domain"
)

type smtpEmailService struct {
	host string
	port string
	user string
	pass string
}

// NewSmtpEmailService creates a real SMTP EmailService.
func NewSmtpEmailService(host, port, user, pass string) domain.EmailService {
	return &smtpEmailService{
		host: host,
		port: port,
		user: user,
		pass: pass,
	}
}

func (s *smtpEmailService) SendShopApprovedEmail(email string) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[Email] panic khi gửi email cho %s: %v", email, r)
			}
		}()

		if email == "" {
			return
		}

		subject := "Subject: Tài khoản của bạn đã được duyệt\n"
		mime := "MIME-version: 1.0;\nContent-Type: text/plain; charset=\"UTF-8\";\n\n"
		body := "Chúc mừng! Tài khoản đối tác của bạn tại Ocean Express đã được duyệt. Vui lòng đăng nhập vào hệ thống để lấy API Key của bạn."
		msg := []byte(subject + mime + body)

		auth := smtp.PlainAuth("", s.user, s.pass, s.host)
		addr := fmt.Sprintf("%s:%s", s.host, s.port)

		err := smtp.SendMail(addr, auth, s.user, []string{email}, msg)
		if err != nil {
			log.Printf("[Email] Lỗi gửi email đến %s qua SMTP %s: %v", email, addr, err)
			return
		}

		log.Printf("[Email→%s] Đã gửi thành công (SMTP)", email)
	}()
}

func (s *smtpEmailService) SendOTP(toEmail, otp, role string) error {
	if toEmail == "" {
		return fmt.Errorf("địa chỉ email trống")
	}

	subject := "Subject: Mã OTP Khôi Phục Mật Khẩu - Ocean Express\n"
	mime := "MIME-version: 1.0;\nContent-Type: text/plain; charset=\"UTF-8\";\n\n"
	body := fmt.Sprintf("Xin chào,\n\nBạn đã yêu cầu khôi phục mật khẩu tài khoản %s.\nMã OTP của bạn là: %s\n\nMã này có hiệu lực trong 5 phút. Nếu bạn không yêu cầu, vui lòng bỏ qua email này.\n\nTrân trọng,\nĐội ngũ Ocean Express", role, otp)
	msg := []byte(subject + mime + body)

	auth := smtp.PlainAuth("", s.user, s.pass, s.host)
	addr := fmt.Sprintf("%s:%s", s.host, s.port)

	err := smtp.SendMail(addr, auth, s.user, []string{toEmail}, msg)
	if err != nil {
		log.Printf("[Email] Lỗi gửi OTP đến %s qua SMTP %s: %v", toEmail, addr, err)
		return err
	}

	log.Printf("[Email→%s] Đã gửi OTP thành công (SMTP)", toEmail)
	return nil
}
