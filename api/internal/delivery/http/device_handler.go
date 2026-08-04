package http

import (
	"net/http"
	"ocean-express-api/internal/delivery/http/middleware"
	"ocean-express-api/internal/domain"

	"github.com/gin-gonic/gin"
)

type DeviceHandler struct {
	deviceUseCase domain.DeviceUseCase
}

func NewDeviceHandler(r *gin.Engine, deviceUC domain.DeviceUseCase) {
	handler := &DeviceHandler{
		deviceUseCase: deviceUC,
	}

	// Mọi role nội bộ (đặc biệt tài xế) đều đăng ký được thiết bị của chính mình.
	api := r.Group("/api/v1/devices")
	api.Use(middleware.AuthRequired())
	{
		api.POST("", handler.Register)
		api.DELETE("", handler.Unregister)
	}
}

// currentUserID lấy id nhân viên đang đăng nhập từ context (do AuthRequired set).
func currentUserID(c *gin.Context) string {
	v, _ := c.Get("user_id")
	if v == nil {
		return ""
	}
	id, _ := v.(string)
	return id
}

type RegisterDeviceRequest struct {
	DeviceName string `json:"device_name"`
	FCMToken   string `json:"fcm_token" binding:"required"`
}

func (h *DeviceHandler) Register(c *gin.Context) {
	var req RegisterDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Dữ liệu thiết bị không hợp lệ"})
		return
	}

	device, err := h.deviceUseCase.Register(c.Request.Context(), currentUserID(c), req.DeviceName, req.FCMToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "data": device})
}

type UnregisterDeviceRequest struct {
	FCMToken string `json:"fcm_token" binding:"required"`
}

func (h *DeviceHandler) Unregister(c *gin.Context) {
	var req UnregisterDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Dữ liệu thiết bị không hợp lệ"})
		return
	}

	if err := h.deviceUseCase.Unregister(c.Request.Context(), currentUserID(c), req.FCMToken); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Đã gỡ thiết bị"})
}
