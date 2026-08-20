package http

import (
	"net/http"
	"ocean-express-api/internal/delivery/http/middleware"
	"ocean-express-api/internal/domain"

	"github.com/gin-gonic/gin"
)

type ShopHandler struct {
	shopUseCase domain.ShopUseCase
}

func NewShopHandler(r *gin.Engine, shopUC domain.ShopUseCase) {
	handler := &ShopHandler{
		shopUseCase: shopUC,
	}

	api := r.Group("/api/v1/shops")
	{
		// Public: shop tự đăng ký (chờ Admin duyệt).
		api.POST("/register", handler.Register)

		// Portal của Shop (role 'shop'): xem thông tin tài khoản của chính mình.
		shopPortal := api.Group("")
		shopPortal.Use(middleware.AuthRequired(), middleware.RoleRequired(domain.RoleShop))
		{
		shopPortal.GET("/me", handler.GetMe)
			shopPortal.PUT("/me", handler.UpdateMe)
			shopPortal.POST("/me/api-key/request-otp", handler.RequestAPIKeyOTP)
			shopPortal.POST("/me/api-key", handler.RegenerateAPIKey)
		}

		// Quản trị (role 'admin').
		admin := api.Group("")
		admin.Use(middleware.AuthRequired(), middleware.RoleRequired("admin"))
		{
			admin.GET("", handler.GetShops)
			admin.POST("", handler.CreateShop)
			admin.PUT("/:id", handler.UpdateShop)
			admin.PATCH("/:id/review", handler.Review)
		}
	}
}

func (h *ShopHandler) GetShops(c *gin.Context) {
	status := c.Query("status") // rỗng = tất cả
	shops, err := h.shopUseCase.GetShops(c.Request.Context(), status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": shops})
}

type RegisterShopRequest struct {
	Name          string  `json:"name" binding:"required"`
	Phone         string  `json:"phone" binding:"required"`
	Email         string  `json:"email" binding:"required,email"`
	Password      string  `json:"password" binding:"required"`
	LocationID    *string  `json:"location_id"`
	AddressDetail string   `json:"address_detail" binding:"required"`
	Latitude      *float64 `json:"latitude"`
	Longitude     *float64 `json:"longitude"`
}

func (h *ShopHandler) Register(c *gin.Context) {
	var req RegisterShopRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	if req.LocationID != nil && *req.LocationID == "" {
		req.LocationID = nil
	}

	shop, err := h.shopUseCase.RegisterShop(c.Request.Context(), req.Name, req.Phone, req.Email, req.Password, req.LocationID, req.AddressDetail, req.Latitude, req.Longitude)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "data": shop, "message": "Đăng ký thành công! Bạn có thể đăng nhập ngay bây giờ."})
}

func (h *ShopHandler) GetMe(c *gin.Context) {
	shopID, _ := c.Get("user_id")
	shop, err := h.shopUseCase.GetByID(c.Request.Context(), shopID.(string))
	if err != nil || shop == nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Không tìm thấy thông tin đối tác"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true, 
		"data": gin.H{
			"id":             shop.ID,
			"name":           shop.Name,
			"email":          shop.Email,
			"webhook_url":    shop.WebhookURL,
			"location_id":    shop.LocationID,
			"address_detail": shop.AddressDetail,
			"status":         shop.Status,
			"is_active":      shop.IsActive,
			"created_at":     shop.CreatedAt,
			"api_key":        shop.APIKey, // Trả về API Key để shop có thể xem sau khi được duyệt
		},
	})
}

type UpdateMeRequest struct {
	Name          string  `json:"name" binding:"required"`
	Phone         string  `json:"phone"`
	WebhookURL    string  `json:"webhook_url"` // optional: shop có thể không có endpoint webhook
	LocationID    *string  `json:"location_id"`
	AddressDetail string   `json:"address_detail" binding:"required"`
	Latitude      *float64 `json:"latitude"`
	Longitude     *float64 `json:"longitude"`
}

// UpdateMe cho shop tự cập nhật thông tin của chính mình trong portal (khu vực gửi
// hàng, địa chỉ, webhook). shopID lấy từ JWT — không cho sửa shop khác.
func (h *ShopHandler) UpdateMe(c *gin.Context) {
	shopID, _ := c.Get("user_id")

	var req UpdateMeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Dữ liệu không hợp lệ"})
		return
	}

	if req.LocationID != nil && *req.LocationID == "" {
		req.LocationID = nil
	}

	shop, err := h.shopUseCase.UpdateShop(c.Request.Context(), shopID.(string), req.Name, req.Phone, req.WebhookURL, req.LocationID, req.AddressDetail, req.Latitude, req.Longitude)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": shop})
}

type ReviewShopRequest struct {
	Approve *bool `json:"approve" binding:"required"`
}

func (h *ShopHandler) Review(c *gin.Context) {
	id := c.Param("id")

	var req ReviewShopRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Thiếu trường approve"})
		return
	}

	shop, apiKey, err := h.shopUseCase.ReviewShop(c.Request.Context(), id, *req.Approve)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	// api_key chỉ khác rỗng khi duyệt lần đầu (sinh key mới) — trả về đúng một lần.
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{
		"shop":    shop,
		"api_key": apiKey,
	}})
}

type CreateShopRequest struct {
	Name          string  `json:"name" binding:"required"`
	Phone         string  `json:"phone"`
	WebhookURL    string  `json:"webhook_url" binding:"required"`
	LocationID    *string  `json:"location_id"`
	AddressDetail string   `json:"address_detail" binding:"required"`
	Latitude      *float64 `json:"latitude"`
	Longitude     *float64 `json:"longitude"`
}

func (h *ShopHandler) CreateShop(c *gin.Context) {
	var req CreateShopRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Dữ liệu không hợp lệ"})
		return
	}

	if req.LocationID != nil && *req.LocationID == "" {
		req.LocationID = nil
	}

	shop, apiKey, err := h.shopUseCase.CreateShop(c.Request.Context(), req.Name, req.Phone, req.WebhookURL, req.LocationID, req.AddressDetail, req.Latitude, req.Longitude)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	// API key thô chỉ trả về đúng một lần khi tạo, sau đó không thể xem lại.
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": gin.H{
		"shop":    shop,
		"api_key": apiKey,
	}})
}

type UpdateShopRequest struct {
	Name          string  `json:"name" binding:"required"`
	Phone         string  `json:"phone"`
	WebhookURL    string  `json:"webhook_url" binding:"required"`
	LocationID    *string  `json:"location_id"`
	AddressDetail string   `json:"address_detail" binding:"required"`
	Latitude      *float64 `json:"latitude"`
	Longitude     *float64 `json:"longitude"`
}

func (h *ShopHandler) UpdateShop(c *gin.Context) {
	id := c.Param("id")

	var req UpdateShopRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Dữ liệu không hợp lệ"})
		return
	}

	if req.LocationID != nil && *req.LocationID == "" {
		req.LocationID = nil
	}

	shop, err := h.shopUseCase.UpdateShop(c.Request.Context(), id, req.Name, req.Phone, req.WebhookURL, req.LocationID, req.AddressDetail, req.Latitude, req.Longitude)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": shop})
}

type RegenerateAPIKeyRequest struct {
	Password string `json:"password" binding:"required"`
	OTP      string `json:"otp" binding:"required"`
}

func (h *ShopHandler) RequestAPIKeyOTP(c *gin.Context) {
	shopID, _ := c.Get("user_id")
	if err := h.shopUseCase.RequestAPIKeyOTP(c.Request.Context(), shopID.(string)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Mã OTP đã được gửi đến email của bạn. Hiệu lực trong 5 phút."})
}

func (h *ShopHandler) RegenerateAPIKey(c *gin.Context) {
	shopID, _ := c.Get("user_id")

	var req RegenerateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Thiếu mật khẩu hoặc mã OTP"})
		return
	}

	newKey, err := h.shopUseCase.RegenerateAPIKey(c.Request.Context(), shopID.(string), req.Password, req.OTP)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"api_key": newKey,
		},
		"message": "Đã tạo lại API Key thành công",
	})
}
