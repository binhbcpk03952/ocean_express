package usecase

import (
	"context"
	"ocean-express-api/internal/domain"
	"time"

	"github.com/google/uuid"
)

type notificationUseCase struct {
	notifRepo domain.NotificationRepository
}

func NewNotificationUseCase(notifRepo domain.NotificationRepository) domain.NotificationUseCase {
	return &notificationUseCase{notifRepo: notifRepo}
}

func (u *notificationUseCase) CreateNotification(ctx context.Context, userID, title, message, notifType string) error {
	notif := &domain.Notification{
		ID:        uuid.New().String(),
		UserID:    userID,
		Title:     title,
		Message:   message,
		Type:      notifType,
		IsRead:    false,
		CreatedAt: time.Now(),
	}
	return u.notifRepo.Create(ctx, notif)
}

func (u *notificationUseCase) GetUserNotifications(ctx context.Context, userID string, pageParams domain.PaginationParams) (*domain.PaginatedResponse, error) {
	notifs, total, err := u.notifRepo.FindByUserID(ctx, userID, pageParams)
	if err != nil {
		return nil, err
	}
	return &domain.PaginatedResponse{
		Data: notifs,
		Meta: domain.PaginationMeta{
			Page:       pageParams.Page,
			Limit:      pageParams.Limit,
			TotalItems: int64(total),
			TotalPages: domain.CalculateTotalPages(int64(total), pageParams.GetLimit()),
		},
	}, nil
}

func (u *notificationUseCase) MarkAsRead(ctx context.Context, id, userID string) error {
	return u.notifRepo.MarkAsRead(ctx, id, userID)
}

func (u *notificationUseCase) MarkAllAsRead(ctx context.Context, userID string) error {
	return u.notifRepo.MarkAllAsRead(ctx, userID)
}

func (u *notificationUseCase) GetUnreadCount(ctx context.Context, userID string) (int, error) {
	return u.notifRepo.CountUnread(ctx, userID)
}
