package domain

// EmailService defines the interface for sending emails.
type EmailService interface {
	// SendShopApprovedEmail sends an email to the shop owner after their account is approved.
	SendShopApprovedEmail(email string)
	// SendOTP sends an OTP code for password recovery
	SendOTP(toEmail, otp, role string) error
}
