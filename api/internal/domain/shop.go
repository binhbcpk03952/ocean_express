package domain

import (
	"context"
	"time"
)

// RoleShop là role JWT dành cho tài khoản portal của đối tác Shop. Tách khỏi
// EmployeeRole vì shop không phải nhân sự nội bộ, nhưng dùng chung cơ chế JWT.
const RoleShop = "shop"

// Shop đại diện cho đối tác E-commerce
type Shop struct {
	ID            string    `json:"id" gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	Name          string    `json:"name" gorm:"column:name"`
	// Credential đăng nhập portal (self-register). Nullable với shop cũ do Admin
	// tạo trực tiếp — các shop đó chỉ dùng api_key, chưa có tài khoản portal.
	Email         *string   `json:"email" gorm:"column:email"`
	Phone         *string   `json:"phone" gorm:"column:phone"`
	PasswordHash  string    `json:"-" gorm:"column:password_hash"`
	WebhookURL    string    `json:"webhook_url" gorm:"column:webhook_url"`
	APIKey        string    `json:"-" gorm:"column:api_key"`
	LocationID    *string   `json:"location_id" gorm:"column:location_id"`
	AddressDetail string    `json:"address_detail" gorm:"column:address_detail"`
	Latitude      *float64  `json:"latitude" gorm:"column:latitude"`
	Longitude     *float64  `json:"longitude" gorm:"column:longitude"`
	// Status: cổng duyệt onboarding (pending/approved/rejected). IsActive: admin
	// khóa/mở sau khi đã duyệt. Xem hằng StatusPending/... ở employee.go.
	Status        string    `json:"status" gorm:"column:status;default:approved"`
	IsActive      bool      `json:"is_active" gorm:"column:is_active;default:true"`
	CreatedAt     time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
}

func (Shop) TableName() string {
	return "shops"
}

// ShopRepository định nghĩa interface tương tác CSDL cho Shop
type ShopRepository interface {
	GetByAPIKey(ctx context.Context, apiKey string) (*Shop, error)
	GetByID(ctx context.Context, id string) (*Shop, error)
	GetByEmail(ctx context.Context, email string) (*Shop, error)
	GetByPhoneOrEmail(ctx context.Context, identifier string) (*Shop, error)
	// FindAll trả về danh sách shop; status rỗng = mọi trạng thái, ngược lại lọc theo status.
	FindAll(ctx context.Context, status string) ([]*Shop, error)
	Create(ctx context.Context, shop *Shop) error
	Update(ctx context.Context, shop *Shop) error
}

// ShopUseCase xử lý logic nghiệp vụ cho Shop (quản trị bởi Admin + self-service portal)
type ShopUseCase interface {
	// GetShops liệt kê đối tác; status rỗng = tất cả, hoặc lọc pending/approved/rejected.
	GetShops(ctx context.Context, status string) ([]*Shop, error)
	CreateShop(ctx context.Context, name, phone, webhookURL string, locationID *string, addressDetail string, latitude, longitude *float64) (*Shop, string, error)
	// UpdateShop cập nhật thông tin đối tác (tên, webhook_url, khu vực, địa chỉ). Không đổi API key.
	UpdateShop(ctx context.Context, id, name, phone, webhookURL string, locationID *string, addressDetail string, latitude, longitude *float64) (*Shop, error)
	// RegisterShop: shop tự đăng ký bằng email/mật khẩu, tạo tài khoản ở trạng thái pending (chờ Admin duyệt).
	RegisterShop(ctx context.Context, name, phone, email, password string, locationID *string, addressDetail string, latitude, longitude *float64) (*Shop, error)
	// ReviewShop: Admin duyệt (approved) hoặc từ chối (rejected). Khi duyệt lần đầu sẽ sinh API key và trả về (chỉ một lần).
	ReviewShop(ctx context.Context, id string, approve bool) (*Shop, string, error)
	// GetByID phục vụ endpoint portal "me".
	GetByID(ctx context.Context, id string) (*Shop, error)
	// RegenerateAPIKey xác thực OTP + mật khẩu rồi sinh API key mới cho shop.
	RegenerateAPIKey(ctx context.Context, id, password, otp string) (string, error)
	// RequestAPIKeyOTP gửi mã OTP qua email để shop xác thực khi yêu cầu API key.
	RequestAPIKeyOTP(ctx context.Context, shopID string) error
}

// ShopAuthUseCase xử lý đăng nhập portal cho Shop (email/phone + mật khẩu).
type ShopAuthUseCase interface {
	Login(ctx context.Context, identifier, password string) (string, *Shop, error)
	ForgotPassword(ctx context.Context, identifier string) error
	ResetPassword(ctx context.Context, identifier, otp, newPassword string) error
}

// WebhookPayload cấu trúc dữ liệu gửi cho Shop
type WebhookPayload struct {
	TrackingNumber    string    `json:"tracking_number"`
	Status            string    `json:"status"`
	StatusName        string    `json:"status_name"`
	StatusLabel       string    `json:"status_label"`
	StatusDescription string    `json:"status_description"`
	Note              string    `json:"note"`
	Timestamp         time.Time `json:"timestamp"`
}

// WebhookService định nghĩa chức năng bắn webhook
type WebhookService interface {
	SendOrderStatus(url string, payload WebhookPayload)
}
