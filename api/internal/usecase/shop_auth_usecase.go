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

type shopAuthUseCase struct {
	shopRepo    domain.ShopRepository
	sessionRepo domain.SessionRepository
	emailSvc    domain.EmailService
}

// NewShopAuthUseCase khởi tạo use case đăng nhập portal cho Shop. Tái dùng cùng
// hạ tầng JWT + session Redis như employee, chỉ khác định danh (email) và role.
func NewShopAuthUseCase(repo domain.ShopRepository, sessionRepo domain.SessionRepository, emailSvc domain.EmailService) domain.ShopAuthUseCase {
	return &shopAuthUseCase{
		shopRepo:    repo,
		sessionRepo: sessionRepo,
		emailSvc:    emailSvc,
	}
}

func (u *shopAuthUseCase) Login(ctx context.Context, identifier, password string) (string, *domain.Shop, error) {
	shop, err := u.shopRepo.GetByPhoneOrEmail(ctx, identifier)
	if err != nil {
		return "", nil, errors.New("lỗi hệ thống khi truy xuất dữ liệu")
	}
	if shop == nil || shop.PasswordHash == "" {
		return "", nil, errors.New("thông tin đăng nhập không đúng")
	}

	if !utils.CheckPasswordHash(password, shop.PasswordHash) {
		return "", nil, errors.New("email hoặc mật khẩu không đúng")
	}

	// Cổng duyệt onboarding: chưa duyệt thì chưa vào portal được.
	switch shop.Status {
	case domain.StatusPending:
		return "", nil, errors.New("tài khoản đang chờ Admin duyệt")
	case domain.StatusRejected:
		return "", nil, errors.New("tài khoản đã bị từ chối")
	}

	if !shop.IsActive {
		return "", nil, errors.New("tài khoản đã bị khóa")
	}

	// Shop không thuộc hub nào -> hubID rỗng. Role cố định là 'shop'.
	token, jti, err := utils.GenerateToken(shop.ID, domain.RoleShop, "")
	if err != nil {
		return "", nil, errors.New("không thể tạo phiên đăng nhập")
	}

	if err := u.sessionRepo.Create(ctx, jti, shop.ID, utils.TokenTTL); err != nil {
		return "", nil, errors.New("không thể lưu phiên đăng nhập")
	}

	return token, shop, nil
}

func (u *shopAuthUseCase) ForgotPassword(ctx context.Context, identifier string) error {
	shop, err := u.shopRepo.GetByPhoneOrEmail(ctx, identifier)
	if err != nil {
		return errors.New("lỗi hệ thống")
	}
	if shop == nil || !shop.IsActive || shop.Status != domain.StatusApproved {
		return errors.New("tài khoản không tồn tại hoặc không hợp lệ")
	}

	otp := fmt.Sprintf("%06d", rand.Intn(1000000))
	
	if err := u.sessionRepo.SetOTP(ctx, identifier, otp, 5*time.Minute); err != nil {
		return errors.New("không thể tạo OTP")
	}
	
	if shop.Email != nil && *shop.Email != "" {
		_ = u.emailSvc.SendOTP(*shop.Email, otp, "Shop E-commerce")
	}

	return nil
}

func (u *shopAuthUseCase) ResetPassword(ctx context.Context, identifier, otp, newPassword string) error {
	savedOTP, err := u.sessionRepo.GetOTP(ctx, identifier)
	if err != nil {
		return errors.New("lỗi hệ thống")
	}
	if savedOTP == "" || savedOTP != otp {
		return errors.New("mã OTP không hợp lệ hoặc đã hết hạn")
	}

	shop, err := u.shopRepo.GetByPhoneOrEmail(ctx, identifier)
	if err != nil || shop == nil {
		return errors.New("tài khoản không tồn tại")
	}

	hash, err := utils.HashPassword(newPassword)
	if err != nil {
		return errors.New("lỗi xử lý mật khẩu")
	}

	shop.PasswordHash = hash
	if err := u.shopRepo.Update(ctx, shop); err != nil {
		return errors.New("lỗi cập nhật mật khẩu")
	}

	_ = u.sessionRepo.DeleteOTP(ctx, identifier)

	return nil
}
