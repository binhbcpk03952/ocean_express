package usecase

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"ocean-express-api/internal/domain"
	"ocean-express-api/pkg/utils"
	"time"
)

type authUseCase struct {
	employeeRepo domain.EmployeeRepository
	sessionRepo  domain.SessionRepository
	emailSvc     domain.EmailService
}

// NewAuthUseCase khởi tạo AuthUseCase với các dependencies.
func NewAuthUseCase(repo domain.EmployeeRepository, sessionRepo domain.SessionRepository, emailSvc domain.EmailService) domain.AuthUseCase {
	return &authUseCase{
		employeeRepo: repo,
		sessionRepo:  sessionRepo,
		emailSvc:     emailSvc,
	}
}

func (u *authUseCase) Login(ctx context.Context, identifier, password string) (string, *domain.Employee, error) {
	emp, err := u.employeeRepo.GetByPhoneOrEmail(ctx, identifier)
	if err != nil {
		return "", nil, errors.New("lỗi hệ thống khi truy xuất dữ liệu")
	}

	if emp == nil {
		return "", nil, errors.New("tài khoản hoặc mật khẩu không đúng")
	}

	// Tài khoản tự đăng ký còn chờ duyệt / bị từ chối: chặn với message rõ ràng
	// thay vì "đã bị khóa" (is_active=false với pending là do chưa duyệt).
	if emp.Status == domain.StatusPending {
		return "", nil, errors.New("tài khoản đang chờ Admin duyệt")
	}
	if emp.Status == domain.StatusRejected {
		return "", nil, errors.New("tài khoản đã bị từ chối")
	}

	if !emp.IsActive {
		return "", nil, errors.New("tài khoản đã bị khóa")
	}

	if !utils.CheckPasswordHash(password, emp.PasswordHash) {
		return "", nil, errors.New("tài khoản hoặc mật khẩu không đúng")
	}

	var hubIDStr string
	if emp.HubID != nil {
		hubIDStr = *emp.HubID
	}

	token, jti, err := utils.GenerateToken(emp.ID, string(emp.Role), hubIDStr)
	if err != nil {
		return "", nil, errors.New("không thể tạo phiên đăng nhập")
	}

	// Ghi session vào Redis: token chỉ hợp lệ khi jti còn tồn tại.
	if err := u.sessionRepo.Create(ctx, jti, emp.ID, utils.TokenTTL); err != nil {
		return "", nil, errors.New("không thể lưu phiên đăng nhập")
	}

	return token, emp, nil
}

func (u *authUseCase) Logout(ctx context.Context, jti string) error {
	if jti == "" {
		return nil
	}
	return u.sessionRepo.Revoke(ctx, jti)
}

func (u *authUseCase) ForgotPassword(ctx context.Context, identifier string) error {
	emp, err := u.employeeRepo.GetByPhoneOrEmail(ctx, identifier)
	if err != nil {
		return errors.New("lỗi hệ thống")
	}
	if emp == nil || !emp.IsActive {
		return errors.New("tài khoản không tồn tại hoặc đã bị khóa")
	}

	// Generate 6-digit OTP
	otp := fmt.Sprintf("%06d", rand.Intn(1000000))
	
	if err := u.sessionRepo.SetOTP(ctx, identifier, otp, 5*time.Minute); err != nil {
		return errors.New("không thể tạo OTP")
	}
	
	// Send email if user has email
	if emp.Email != nil && *emp.Email != "" {
		_ = u.emailSvc.SendOTP(*emp.Email, otp, "Nhân viên")
	}

	return nil
}

func (u *authUseCase) ResetPassword(ctx context.Context, identifier, otp, newPassword string) error {
	savedOTP, err := u.sessionRepo.GetOTP(ctx, identifier)
	if err != nil {
		return errors.New("lỗi hệ thống")
	}
	if savedOTP == "" || savedOTP != otp {
		return errors.New("mã OTP không hợp lệ hoặc đã hết hạn")
	}

	emp, err := u.employeeRepo.GetByPhoneOrEmail(ctx, identifier)
	if err != nil || emp == nil {
		return errors.New("tài khoản không tồn tại")
	}

	hash, err := utils.HashPassword(newPassword)
	if err != nil {
		return errors.New("lỗi xử lý mật khẩu")
	}

	emp.PasswordHash = hash
	if err := u.employeeRepo.Update(ctx, emp); err != nil {
		return errors.New("lỗi cập nhật mật khẩu")
	}

	// Xóa OTP sau khi dùng thành công
	_ = u.sessionRepo.DeleteOTP(ctx, identifier)

	return nil
}
