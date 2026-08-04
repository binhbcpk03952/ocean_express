package repository

import (
	"context"
	"errors"
	"ocean-express-api/internal/domain"

	"gorm.io/gorm"
)

type employeeRepository struct {
	db *gorm.DB
}

// NewEmployeeRepository khởi tạo EmployeeRepository với GORM DB
func NewEmployeeRepository(db *gorm.DB) domain.EmployeeRepository {
	return &employeeRepository{db: db}
}

func (r *employeeRepository) GetByPhoneOrEmail(ctx context.Context, identifier string) (*domain.Employee, error) {
	var emp domain.Employee
	err := r.db.WithContext(ctx).Where("phone = ? OR email = ?", identifier, identifier).First(&emp).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Trả về nil khi không tìm thấy thay vì lỗi hệ thống
		}
		return nil, err
	}
	return &emp, nil
}

func (r *employeeRepository) GetByPhone(ctx context.Context, phone string) (*domain.Employee, error) {
	var emp domain.Employee
	err := r.db.WithContext(ctx).Where("phone = ?", phone).First(&emp).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Trả về nil khi không tìm thấy thay vì lỗi hệ thống
		}
		return nil, err
	}
	return &emp, nil
}

func (r *employeeRepository) GetByID(ctx context.Context, id string) (*domain.Employee, error) {
	var emp domain.Employee
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&emp).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &emp, nil
}

func (r *employeeRepository) FindAll(ctx context.Context, hubID *string, status *string, pageParams domain.PaginationParams) ([]*domain.Employee, int64, error) {
	var emps []*domain.Employee
	var total int64
	query := r.db.WithContext(ctx).Model(&domain.Employee{})

	if hubID != nil {
		query = query.Where("hub_id = ?", *hubID)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("created_at DESC").
		Offset(pageParams.GetOffset()).
		Limit(pageParams.GetLimit()).
		Find(&emps).Error
	return emps, total, err
}

func (r *employeeRepository) Create(ctx context.Context, emp *domain.Employee) error {
	return r.db.WithContext(ctx).Create(emp).Error
}

func (r *employeeRepository) Update(ctx context.Context, emp *domain.Employee) error {
	// Select tường minh các cột được phép cập nhật để tránh ghi đè ngoài ý muốn
	// (vd password_hash khi đổi mật khẩu, is_active khi khóa/mở tài khoản).
	return r.db.WithContext(ctx).Model(emp).
		Select("name", "phone", "password_hash", "role", "hub_id", "status", "is_active").
		Updates(emp).Error
}
