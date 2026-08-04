package domain

import (
	"context"
	"time"
)

type EmployeeRole string

const (
	RoleAdmin           EmployeeRole = "admin"
	RoleFirstMileDriver EmployeeRole = "first_mile_driver"
	RoleHubStaff        EmployeeRole = "hub_staff"
	RoleTransitDriver   EmployeeRole = "transit_driver"
	RoleLastMileDriver  EmployeeRole = "last_mile_driver"
)

// Trạng thái duyệt tài khoản (dùng chung cho employee tự đăng ký và shop).
// Tách bạch với is_active: status là cổng onboarding (admin duyệt một lần),
// is_active để admin khóa/mở tài khoản sau khi đã duyệt.
const (
	StatusPending  = "pending"
	StatusApproved = "approved"
	StatusRejected = "rejected"
)

// Employee đại diện cho nhân viên / tài xế trong hệ thống
type Employee struct {
	ID           string       `json:"id" gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	Name         string       `json:"name" gorm:"column:name"`
	Phone        string       `json:"phone" gorm:"column:phone"`
	Email        *string      `json:"email" gorm:"column:email"`
	PasswordHash string       `json:"-" gorm:"column:password_hash"`
	Role         EmployeeRole `json:"role" gorm:"column:role"`
	HubID        *string      `json:"hub_id" gorm:"column:hub_id"`
	Status       string       `json:"status" gorm:"column:status;default:approved"`
	IsActive     bool         `json:"is_active" gorm:"column:is_active;default:true"`
	CreatedAt    time.Time    `json:"created_at" gorm:"column:created_at;autoCreateTime"`
}

func (Employee) TableName() string {
	return "employees"
}

// EmployeeDevice quản lý phiên đăng nhập đa thiết bị và FCM Token
type EmployeeDevice struct {
	ID         string    `json:"id" gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	EmployeeID string    `json:"employee_id" gorm:"column:employee_id"`
	DeviceName string    `json:"device_name" gorm:"column:device_name"`
	FCMToken   string    `json:"fcm_token" gorm:"column:fcm_token"`
	LastActive time.Time `json:"last_active" gorm:"column:last_active;autoUpdateTime"`
}

func (EmployeeDevice) TableName() string {
	return "employee_devices"
}

// EmployeeRepository định nghĩa interface tương tác CSDL cho Employee
type EmployeeRepository interface {
	GetByPhoneOrEmail(ctx context.Context, identifier string) (*Employee, error)
	GetByPhone(ctx context.Context, phone string) (*Employee, error)
	GetByID(ctx context.Context, id string) (*Employee, error)
	// FindAll lọc theo hub và/hoặc status (nil = không lọc theo tiêu chí đó).
	FindAll(ctx context.Context, hubID *string, status *string, pageParams PaginationParams) ([]*Employee, int64, error)
	Create(ctx context.Context, emp *Employee) error
	Update(ctx context.Context, emp *Employee) error
}

// AuthUseCase định nghĩa interface cho logic xác thực
type AuthUseCase interface {
	// Login xác thực bằng số điện thoại hoặc email + mật khẩu, trả về JWT token và thông tin nhân viên.
	Login(ctx context.Context, identifier, password string) (string, *Employee, error)
	// Logout thu hồi phiên hiện tại (theo jti trích từ token).
	Logout(ctx context.Context, jti string) error
	ForgotPassword(ctx context.Context, identifier string) error
	ResetPassword(ctx context.Context, identifier, otp, newPassword string) error
}

// EmployeeUseCase xử lý logic quản lý nhân sự (quản trị bởi Admin)
type EmployeeUseCase interface {
	// GetEmployees lọc theo hub và/hoặc status (nil = không lọc theo tiêu chí đó).
	GetEmployees(ctx context.Context, hubID *string, status *string, pageParams PaginationParams) (*PaginatedResponse, error)
	// CreateEmployee do Admin tạo trực tiếp — tài khoản được duyệt ngay (status=approved).
	CreateEmployee(ctx context.Context, name, phone, email, password, role string, hubID *string) (*Employee, error)
	// RegisterEmployee là shipper tự đăng ký (first/last-mile driver). Tài khoản ở trạng thái
	// chờ duyệt (status=pending, is_active=false) cho tới khi Admin duyệt. Tự chọn hub.
	RegisterEmployee(ctx context.Context, name, phone, email, password, role string, hubID *string) (*Employee, error)
	// UpdateEmployee cập nhật thông tin nhân sự. password rỗng nghĩa là giữ nguyên mật khẩu cũ.
	UpdateEmployee(ctx context.Context, id, name, phone, email, password, role string, hubID *string) (*Employee, error)
	// SetActive khóa/mở tài khoản nhân sự.
	SetActive(ctx context.Context, id string, active bool) (*Employee, error)
	// ReviewEmployee duyệt/từ chối một tài khoản đang chờ (approve=true -> approved + is_active,
	// approve=false -> rejected).
	ReviewEmployee(ctx context.Context, id string, approve bool) (*Employee, error)
}
