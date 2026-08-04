package http

import (
	"net/http"
	"ocean-express-api/internal/delivery/http/middleware"
	"ocean-express-api/internal/domain"

	"github.com/gin-gonic/gin"
)

type RateHandler struct {
	rateUseCase domain.RateUseCase
}

func NewRateHandler(r *gin.Engine, rateUC domain.RateUseCase) {
	handler := &RateHandler{
		rateUseCase: rateUC,
	}

	api := r.Group("/api/v1/rates")
	
	api.Use(middleware.AuthRequired())
	{
		api.GET("", handler.GetRates)
		
		// Chỉ admin được tạo bảng giá cước
		adminGroup := api.Group("")
		adminGroup.Use(middleware.RoleRequired("admin"))
		{
			adminGroup.POST("", handler.CreateRate)
		}
	}
}

func (h *RateHandler) GetRates(c *gin.Context) {
	rates, err := h.rateUseCase.GetRates(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": rates})
}

func (h *RateHandler) CreateRate(c *gin.Context) {
	var req domain.ShippingRate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Dữ liệu không hợp lệ"})
		return
	}

	// Xử lý null value an toàn
	if req.FromLocationID != nil && *req.FromLocationID == "" {
		req.FromLocationID = nil
	}
	if req.ToLocationID != nil && *req.ToLocationID == "" {
		req.ToLocationID = nil
	}

	err := h.rateUseCase.CreateRate(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "data": req})
}
