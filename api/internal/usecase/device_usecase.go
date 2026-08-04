package usecase

import (
	"context"
	"errors"
	"ocean-express-api/internal/domain"
)

type deviceUseCase struct {
	deviceRepo domain.DeviceRepository
}

func NewDeviceUseCase(repo domain.DeviceRepository) domain.DeviceUseCase {
	return &deviceUseCase{deviceRepo: repo}
}

// Register đăng ký (hoặc refresh) thiết bị hiện tại cho nhân viên đang đăng nhập.
func (u *deviceUseCase) Register(ctx context.Context, employeeID, deviceName, fcmToken string) (*domain.EmployeeDevice, error) {
	if employeeID == "" {
		return nil, errors.New("thiếu thông tin nhân viên")
	}
	if fcmToken == "" {
		return nil, errors.New("fcm_token không được để trống")
	}

	device := &domain.EmployeeDevice{
		EmployeeID: employeeID,
		DeviceName: deviceName,
		FCMToken:   fcmToken,
	}

	if err := u.deviceRepo.Upsert(ctx, device); err != nil {
		return nil, errors.New("không thể đăng ký thiết bị")
	}

	return device, nil
}

// Unregister gỡ thiết bị của nhân viên (logout thiết bị / gỡ app).
func (u *deviceUseCase) Unregister(ctx context.Context, employeeID, fcmToken string) error {
	if employeeID == "" || fcmToken == "" {
		return errors.New("thiếu thông tin thiết bị")
	}
	return u.deviceRepo.DeleteByToken(ctx, employeeID, fcmToken)
}
