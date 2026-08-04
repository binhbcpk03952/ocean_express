package domain

import (
	"context"
	"time"
)

// SessionRepository quản lý phiên đăng nhập (JTI) trong Redis, phục vụ thu hồi
// token thật sự (logout / khóa tài khoản) thay vì chỉ dựa vào JWT stateless.
type SessionRepository interface {
	// Create lưu một phiên: jti -> userID với TTL bằng thời gian sống của token.
	Create(ctx context.Context, jti, userID string, ttl time.Duration) error
	// Exists kiểm tra phiên còn hiệu lực (chưa bị thu hồi và chưa hết hạn).
	Exists(ctx context.Context, jti string) (bool, error)
	// Revoke thu hồi một phiên cụ thể (logout thiết bị hiện tại).
	Revoke(ctx context.Context, jti string) error
	
	// SetOTP lưu mã OTP với identifier.
	SetOTP(ctx context.Context, identifier, otp string, ttl time.Duration) error
	// GetOTP lấy mã OTP.
	GetOTP(ctx context.Context, identifier string) (string, error)
	// DeleteOTP xóa mã OTP.
	DeleteOTP(ctx context.Context, identifier string) error
}
