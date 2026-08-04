package http

import (
	"net/http"
	"ocean-express-api/internal/delivery/http/middleware"
	"ocean-express-api/internal/domain"

	"github.com/gin-gonic/gin"
)

type LocationHandler struct {
	locationUseCase domain.LocationUseCase
}

func NewLocationHandler(r *gin.Engine, locUC domain.LocationUseCase) {
	handler := &LocationHandler{
		locationUseCase: locUC,
	}

	api := r.Group("/api/v1/locations")
	// Lấy danh sách (Public hoặc yêu cầu đăng nhập tùy thiết kế, ở đây tạm để mọi người xem được)
	api.GET("", handler.GetLocations)

	// Admin mới được tạo mới
	api.Use(middleware.AuthRequired(), middleware.RoleRequired("admin"))
	{
		api.POST("", handler.CreateLocation)
	}
}

func (h *LocationHandler) GetLocations(c *gin.Context) {
	var parentID *string
	var locType *string

	if p := c.Query("parent_id"); p != "" {
		parentID = &p
	}
	if t := c.Query("type"); t != "" {
		locType = &t
	}

	locs, err := h.locationUseCase.GetLocations(c.Request.Context(), parentID, locType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": locs})
}

type CreateLocationRequest struct {
	ID       string  `json:"id" binding:"required"`
	Name     string  `json:"name" binding:"required"`
	Type     string  `json:"type" binding:"required"`
	ParentID *string `json:"parent_id"`
}

func (h *LocationHandler) CreateLocation(c *gin.Context) {
	var req CreateLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Dữ liệu không hợp lệ"})
		return
	}

	loc, err := h.locationUseCase.CreateLocation(c.Request.Context(), req.ID, req.Name, req.Type, req.ParentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "data": loc})
}
