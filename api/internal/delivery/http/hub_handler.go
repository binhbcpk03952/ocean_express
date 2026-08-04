package http

import (
	"net/http"
	"ocean-express-api/internal/delivery/http/middleware"
	"ocean-express-api/internal/domain"

	"github.com/gin-gonic/gin"
)

type HubHandler struct {
	hubUseCase domain.HubUseCase
}

func NewHubHandler(r *gin.Engine, hubUC domain.HubUseCase) {
	handler := &HubHandler{
		hubUseCase: hubUC,
	}

	api := r.Group("/api/v1/hubs")

	// GET công khai: shipper cần xem danh sách bưu cục để chọn khi tự đăng ký
	// (chưa có token). Giống pattern GET /locations.
	api.GET("", handler.GetHubs)

	// Chỉ Admin được tạo Hub mới
	adminGroup := api.Group("")
	adminGroup.Use(middleware.AuthRequired(), middleware.RoleRequired("admin"))
	{
		adminGroup.POST("", handler.CreateHub)
	}
}

func (h *HubHandler) GetHubs(c *gin.Context) {
	var locationID *string
	if l := c.Query("location_id"); l != "" {
		locationID = &l
	}

	hubs, err := h.hubUseCase.GetHubs(c.Request.Context(), locationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": hubs})
}

type CreateHubRequest struct {
	Name          string   `json:"name" binding:"required"`
	Type          string   `json:"type" binding:"required"`
	LocationID    *string  `json:"location_id"`
	AddressDetail string   `json:"address_detail" binding:"required"`
	Latitude      *float64 `json:"latitude"`
	Longitude     *float64 `json:"longitude"`
}

func (h *HubHandler) CreateHub(c *gin.Context) {
	var req CreateHubRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Dữ liệu không hợp lệ"})
		return
	}

	hub, err := h.hubUseCase.CreateHub(c.Request.Context(), req.Name, req.Type, req.LocationID, req.AddressDetail, req.Latitude, req.Longitude)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "data": hub})
}
