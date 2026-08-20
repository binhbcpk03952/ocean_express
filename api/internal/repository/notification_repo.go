package repository

import (
	"context"
	"fmt"
	"ocean-express-api/internal/domain"

	"gorm.io/gorm"
)

type notificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) domain.NotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) Create(ctx context.Context, notif *domain.Notification) error {
	return r.db.WithContext(ctx).Create(notif).Error
}

func (r *notificationRepository) FindByUserID(ctx context.Context, userID string, pageParams domain.PaginationParams) ([]domain.Notification, int, error) {
	var total int64
	if err := r.db.WithContext(ctx).Model(&domain.Notification{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var notifs []domain.Notification
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Limit(pageParams.GetLimit()).Offset(pageParams.GetOffset()).Find(&notifs).Error; err != nil {
		return nil, 0, err
	}
	return notifs, int(total), nil
}

func (r *notificationRepository) MarkAsRead(ctx context.Context, id, userID string) error {
	res := r.db.WithContext(ctx).Model(&domain.Notification{}).Where("id = ? AND user_id = ?", id, userID).Update("is_read", true)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("không tìm thấy thông báo hoặc không có quyền")
	}
	return nil
}

func (r *notificationRepository) MarkAllAsRead(ctx context.Context, userID string) error {
	return r.db.WithContext(ctx).Model(&domain.Notification{}).Where("user_id = ? AND is_read = false", userID).Update("is_read", true).Error
}

func (r *notificationRepository) CountUnread(ctx context.Context, userID string) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.Notification{}).Where("user_id = ? AND is_read = false", userID).Count(&count).Error
	return int(count), err
}
