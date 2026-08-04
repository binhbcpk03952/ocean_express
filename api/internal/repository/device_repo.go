package repository

import (
	"context"
	"ocean-express-api/internal/domain"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type deviceRepository struct {
	db *gorm.DB
}

func NewDeviceRepository(db *gorm.DB) domain.DeviceRepository {
	return &deviceRepository{db: db}
}

// Upsert đăng ký/cập nhật thiết bị theo fcm_token. Nếu token đã tồn tại thì cập
// nhật employee_id + device_name + last_active thay vì tạo bản ghi trùng.
func (r *deviceRepository) Upsert(ctx context.Context, device *domain.EmployeeDevice) error {
	device.LastActive = time.Now()
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "fcm_token"}},
		DoUpdates: clause.AssignmentColumns([]string{"employee_id", "device_name", "last_active"}),
	}).Create(device).Error
}

func (r *deviceRepository) FindByEmployee(ctx context.Context, employeeID string) ([]*domain.EmployeeDevice, error) {
	var devices []*domain.EmployeeDevice
	err := r.db.WithContext(ctx).
		Where("employee_id = ?", employeeID).
		Order("last_active DESC").
		Find(&devices).Error
	return devices, err
}

func (r *deviceRepository) DeleteByToken(ctx context.Context, employeeID, fcmToken string) error {
	return r.db.WithContext(ctx).
		Where("employee_id = ? AND fcm_token = ?", employeeID, fcmToken).
		Delete(&domain.EmployeeDevice{}).Error
}
