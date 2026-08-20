package domain

import (
	"context"
	"time"
)

type Notification struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	Type      string    `json:"type"`
	IsRead    bool      `json:"is_read"`
	CreatedAt time.Time `json:"created_at"`
}

type NotificationRepository interface {
	Create(ctx context.Context, notif *Notification) error
	FindByUserID(ctx context.Context, userID string, pageParams PaginationParams) ([]Notification, int, error)
	MarkAsRead(ctx context.Context, id, userID string) error
	MarkAllAsRead(ctx context.Context, userID string) error
	CountUnread(ctx context.Context, userID string) (int, error)
}

type NotificationUseCase interface {
	CreateNotification(ctx context.Context, userID, title, message, notifType string) error
	GetUserNotifications(ctx context.Context, userID string, pageParams PaginationParams) (*PaginatedResponse, error)
	MarkAsRead(ctx context.Context, id, userID string) error
	MarkAllAsRead(ctx context.Context, userID string) error
	GetUnreadCount(ctx context.Context, userID string) (int, error)
}
