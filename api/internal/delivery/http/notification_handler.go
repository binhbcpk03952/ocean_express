package http

import (
	"net/http"
	"ocean-express-api/internal/delivery/http/middleware"
	"ocean-express-api/internal/domain"
	"strconv"

	"github.com/gin-gonic/gin"
)

type NotificationHandler struct {
	notifUseCase domain.NotificationUseCase
}

func NewNotificationHandler(r *gin.Engine, notifUC domain.NotificationUseCase) {
	handler := &NotificationHandler{
		notifUseCase: notifUC,
	}

	api := r.Group("/api/v1/notifications")
	api.Use(middleware.AuthRequired())
	{
		api.GET("", handler.GetNotifications)
		api.GET("/unread-count", handler.GetUnreadCount)
		api.PUT("/:id/read", handler.MarkAsRead)
		api.PUT("/read-all", handler.MarkAllAsRead)
	}
}

func (h *NotificationHandler) GetNotifications(c *gin.Context) {
	userID := c.GetString("user_id") // set by AuthRequired middleware

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	pageParams := domain.PaginationParams{Page: page, Limit: limit}

	res, err := h.notifUseCase.GetUserNotifications(c.Request.Context(), userID, pageParams)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    res.Data,
		"meta":    res.Meta,
	})
}

func (h *NotificationHandler) GetUnreadCount(c *gin.Context) {
	userID := c.GetString("user_id")

	count, err := h.notifUseCase.GetUnreadCount(c.Request.Context(), userID)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"count": count}})
}

func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	userID := c.GetString("user_id")
	id := c.Param("id")

	if err := h.notifUseCase.MarkAsRead(c.Request.Context(), id, userID); err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Đã đánh dấu đọc thông báo"})
}

func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	userID := c.GetString("user_id")

	if err := h.notifUseCase.MarkAllAsRead(c.Request.Context(), userID); err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Đã đánh dấu đọc tất cả"})
}
