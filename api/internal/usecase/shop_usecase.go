package usecase

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"ocean-express-api/internal/domain"
	"ocean-express-api/pkg/utils"
	"strings"
	"time"

	"github.com/google/uuid"
)

type shopUseCase struct {
	shopRepo     domain.ShopRepository
	sessionRepo  domain.SessionRepository
	emailService domain.EmailService
}

func NewShopUseCase(repo domain.ShopRepository, sessionRepo domain.SessionRepository, emailService domain.EmailService) domain.ShopUseCase {
	return &shopUseCase{
		shopRepo:     repo,
		sessionRepo:  sessionRepo,
		emailService: emailService,
	}
}

func (u *shopUseCase) GetShops(ctx context.Context, status string) ([]*domain.Shop, error) {
	return u.shopRepo.FindAll(ctx, status)
}

func (u *shopUseCase) GetByID(ctx context.Context, id string) (*domain.Shop, error) {
	shop, err := u.shopRepo.GetByID(ctx, id)
	if err != nil || shop == nil {
		return shop, err
	}
	// Tự động sinh API key nếu shop trước đó chưa có
	if shop.APIKey == "" {
		if rawKey, err := utils.GenerateAPIKey(); err == nil {
			shop.APIKey = rawKey
			_ = u.shopRepo.Update(ctx, shop)
		}
	}
	return shop, nil
}

// CreateShop tạo đối tác mới và sinh API key. API key thô chỉ được trả về đúng
// một lần tại đây (giá trị lưu trong struct bị ẩn khỏi JSON qua tag `json:"-"`).
func (u *shopUseCase) CreateShop(ctx context.Context, name, phone, webhookURL string, locationID *string, addressDetail string, latitude, longitude *float64) (*domain.Shop, string, error) {
	name = utils.FixMojibake(strings.TrimSpace(name))
	addressDetail = utils.FixMojibake(strings.TrimSpace(addressDetail))
	if name == "" || addressDetail == "" {
		return nil, "", errors.New("tên và địa chỉ chi tiết không được để trống")
	}

	apiKey, err := utils.GenerateAPIKey()
	if err != nil || apiKey == "" {
		apiKey = "oe_" + uuid.New().String()
	}

	var phonePtr *string
	if phone != "" {
		phonePtr = &phone
	}

	shop := &domain.Shop{
		Name:          name,
		Phone:         phonePtr,
		WebhookURL:    webhookURL,
		APIKey:        apiKey,
		LocationID:    locationID,
		AddressDetail: addressDetail,
		Latitude:      latitude,
		Longitude:     longitude,
		Status:        domain.StatusApproved, // Admin tạo trực tiếp -> duyệt ngay
		IsActive:      true,
	}

	if err := u.shopRepo.Create(ctx, shop); err != nil {
		return nil, "", err
	}

	return shop, apiKey, nil
}

// RegisterShop: shop tự đăng ký bằng email + mật khẩu. Tài khoản được kích hoạt ngay lập tức
// và tự động sinh sẵn API Key cho Shop để tích hợp nhanh chóng.
func (u *shopUseCase) RegisterShop(ctx context.Context, name, phone, email, password string, locationID *string, addressDetail string, latitude, longitude *float64) (*domain.Shop, error) {
	name = utils.FixMojibake(strings.TrimSpace(name))
	addressDetail = utils.FixMojibake(strings.TrimSpace(addressDetail))
	if name == "" || email == "" || password == "" || addressDetail == "" {
		return nil, errors.New("tên, email, mật khẩu và địa chỉ không được để trống")
	}

	existing, err := u.shopRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, errors.New("lỗi hệ thống khi kiểm tra email")
	}
	if existing != nil {
		return nil, errors.New("email đã được sử dụng")
	}

	hash, err := utils.HashPassword(password)
	if err != nil {
		return nil, errors.New("không thể mã hóa mật khẩu")
	}

	var phonePtr *string
	if phone != "" {
		phonePtr = &phone
	}

	rawKey, _ := utils.GenerateAPIKey()
	if rawKey == "" {
		rawKey = "oe_" + uuid.New().String()
	}

	shop := &domain.Shop{
		Name:          name,
		Phone:         phonePtr,
		Email:         &email,
		PasswordHash:  hash,
		APIKey:        rawKey,
		LocationID:    locationID,
		AddressDetail: addressDetail,
		Latitude:      latitude,
		Longitude:     longitude,
		Status:        domain.StatusApproved, // Tự động duyệt — không cần Admin xem xét
		IsActive:      true,
	}

	if err := u.shopRepo.Create(ctx, shop); err != nil {
		return nil, errors.New("lỗi khi tạo tài khoản đối tác")
	}

	return shop, nil
}

// ReviewShop: Admin duyệt hoặc từ chối đối tác đang chờ. Khi duyệt lần đầu (shop
// chưa có API key), sinh key mới và trả về đúng một lần.
func (u *shopUseCase) ReviewShop(ctx context.Context, id string, approve bool) (*domain.Shop, string, error) {
	shop, err := u.shopRepo.GetByID(ctx, id)
	if err != nil {
		return nil, "", errors.New("lỗi hệ thống khi truy xuất đối tác")
	}
	if shop == nil {
		return nil, "", errors.New("không tìm thấy đối tác")
	}

	if !approve {
		shop.Status = domain.StatusRejected
		shop.IsActive = false
		if err := u.shopRepo.Update(ctx, shop); err != nil {
			return nil, "", errors.New("lỗi khi cập nhật trạng thái duyệt")
		}
		return shop, "", nil
	}

	shop.Status = domain.StatusApproved
	shop.IsActive = true

	// Sinh API key nếu shop chưa có (duyệt lần đầu). API key thô chỉ trả về đây một lần.
	var rawKey string
	if shop.APIKey == "" {
		rawKey, err = utils.GenerateAPIKey()
		if err != nil {
			return nil, "", errors.New("không thể sinh API key")
		}
		shop.APIKey = rawKey
	}

	if err := u.shopRepo.Update(ctx, shop); err != nil {
		return nil, "", errors.New("lỗi khi duyệt đối tác")
	}

	// Gửi email thông báo cho đối tác
	if shop.Email != nil && u.emailService != nil {
		u.emailService.SendShopApprovedEmail(*shop.Email)
	}

	return shop, rawKey, nil
}

// UpdateShop cập nhật thông tin đối tác. Không đụng tới API key (chỉ sinh khi tạo mới).
// webhook_url là tùy chọn: shop có thể không có endpoint nhận webhook.
func (u *shopUseCase) UpdateShop(ctx context.Context, id, name, phone, webhookURL string, locationID *string, addressDetail string, latitude, longitude *float64) (*domain.Shop, error) {
	shop, err := u.shopRepo.GetByID(ctx, id)
	if err != nil || shop == nil {
		return nil, errors.New("không tìm thấy đối tác")
	}

	name = utils.FixMojibake(strings.TrimSpace(name))
	addressDetail = utils.FixMojibake(strings.TrimSpace(addressDetail))
	if name == "" || addressDetail == "" {
		return nil, errors.New("tên và địa chỉ chi tiết không được để trống")
	}

	var phonePtr *string
	if phone != "" {
		phonePtr = &phone
	} else if shop.Phone != nil {
		phonePtr = shop.Phone
	}

	shop.Name = name
	shop.Phone = phonePtr
	shop.WebhookURL = webhookURL
	shop.LocationID = locationID
	shop.AddressDetail = addressDetail
	shop.Latitude = latitude
	shop.Longitude = longitude

	if err := u.shopRepo.Update(ctx, shop); err != nil {
		return nil, err
	}

	return shop, nil
}

// RequestAPIKeyOTP sinh mã OTP 6 số, lưu vào Redis (5 phút), và gửi qua email cho shop.
// Shop phải dùng mã này cùng mật khẩu để tạo API key.
func (u *shopUseCase) RequestAPIKeyOTP(ctx context.Context, shopID string) error {
	shop, err := u.shopRepo.GetByID(ctx, shopID)
	if err != nil || shop == nil {
		return errors.New("không tìm thấy đối tác")
	}
	if shop.Email == nil || *shop.Email == "" {
		return errors.New("tài khoản chưa liên kết email")
	}

	otp := fmt.Sprintf("%06d", rand.New(rand.NewSource(time.Now().UnixNano())).Intn(1000000))
	otpKey := "shop_apikey_otp:" + shopID

	if err := u.sessionRepo.SetOTP(ctx, otpKey, otp, 5*time.Minute); err != nil {
		return errors.New("không thể tạo mã OTP")
	}

	if err := u.emailService.SendOTP(*shop.Email, otp, shop.Name); err != nil {
		// OTP đã lưu Redis, không cần rollback. Chỉ log lỗi gửi mail.
		return errors.New("không thể gửi email OTP, vui lòng thử lại")
	}

	return nil
}

// RegenerateAPIKey xác thực mật khẩu (và OTP nếu có), sau đó sinh API key mới.
func (u *shopUseCase) RegenerateAPIKey(ctx context.Context, id, password, otp string) (string, error) {
	shop, err := u.shopRepo.GetByID(ctx, id)
	if err != nil {
		return "", errors.New("lỗi hệ thống khi truy xuất đối tác")
	}
	if shop == nil {
		return "", errors.New("không tìm thấy đối tác")
	}

	// Bước 1: Xác thực mật khẩu
	if !utils.CheckPasswordHash(password, shop.PasswordHash) {
		return "", errors.New("mật khẩu không chính xác")
	}

	// Bước 2: Xác thực OTP nếu có gửi lên
	if otp != "" {
		otpKey := "shop_apikey_otp:" + id
		savedOTP, err := u.sessionRepo.GetOTP(ctx, otpKey)
		if err == nil && savedOTP != "" {
			if savedOTP != otp {
				return "", errors.New("mã OTP không hợp lệ hoặc đã hết hạn")
			}
			_ = u.sessionRepo.DeleteOTP(ctx, otpKey)
		}
	}

	// Bước 3: Sinh API key mới
	newKey, err := utils.GenerateAPIKey()
	if err != nil {
		return "", errors.New("không thể sinh API key mới")
	}

	shop.APIKey = newKey
	if err := u.shopRepo.Update(ctx, shop); err != nil {
		return "", errors.New("lỗi khi lưu API key mới")
	}

	return newKey, nil
}
