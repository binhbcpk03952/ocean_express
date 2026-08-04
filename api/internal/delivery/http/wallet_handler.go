package http

import (
	"net/http"
	"ocean-express-api/internal/delivery/http/middleware"
	"ocean-express-api/internal/domain"

	"github.com/gin-gonic/gin"
)

type WalletHandler struct {
	walletUseCase domain.WalletUseCase
}

func NewWalletHandler(r *gin.Engine, walletUC domain.WalletUseCase) {
	handler := &WalletHandler{walletUseCase: walletUC}

	api := r.Group("/api/v1")
	{
		// Portal Shop (role 'shop'): xem ví + lịch sử đối soát của chính mình.
		shopPortal := api.Group("/shop")
		shopPortal.Use(middleware.AuthRequired(), middleware.RoleRequired(domain.RoleShop))
		{
			shopPortal.GET("/wallet", handler.GetMyWallet)
			shopPortal.GET("/settlements", handler.GetMySettlements)
		}

		// Quản trị (role 'admin'): chốt & chi trả đối soát cho shop.
		admin := api.Group("/settlements")
		admin.Use(middleware.AuthRequired(), middleware.RoleRequired("admin"))
		{
			admin.GET("", handler.ListSettlements)
			admin.POST("", handler.CreateSettlement)
			admin.PATCH("/:id/paid", handler.MarkPaid)
		}

		// Admin xem ví của một shop bất kỳ (phục vụ đối soát).
		adminWallet := api.Group("/shops")
		adminWallet.Use(middleware.AuthRequired(), middleware.RoleRequired("admin"))
		{
			adminWallet.GET("/:id/wallet", handler.GetShopWallet)
		}
	}
}

// GetMyWallet: shop xem số dư khả dụng + lịch sử bút toán của chính mình.
func (h *WalletHandler) GetMyWallet(c *gin.Context) {
	shopID, _ := c.Get("user_id")
	balance, txs, err := h.walletUseCase.GetWallet(c.Request.Context(), shopID.(string))
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{
		"available_balance": balance,
		"transactions":      txs,
	}})
}

// GetShopWallet: admin xem ví của một shop cụ thể (theo :id).
func (h *WalletHandler) GetShopWallet(c *gin.Context) {
	shopID := c.Param("id")
	balance, txs, err := h.walletUseCase.GetWallet(c.Request.Context(), shopID)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{
		"available_balance": balance,
		"transactions":      txs,
	}})
}

// GetMySettlements: shop xem lịch sử phiên chi trả của chính mình.
func (h *WalletHandler) GetMySettlements(c *gin.Context) {
	shopID, _ := c.Get("user_id")
	settlements, err := h.walletUseCase.ListSettlements(c.Request.Context(), shopID.(string))
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": settlements})
}

// ListSettlements: admin xem tất cả phiên chi trả; lọc theo shop_id nếu có query.
func (h *WalletHandler) ListSettlements(c *gin.Context) {
	shopID := c.Query("shop_id") // rỗng = tất cả
	settlements, err := h.walletUseCase.ListSettlements(c.Request.Context(), shopID)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": settlements})
}

type CreateSettlementRequest struct {
	ShopID string `json:"shop_id" binding:"required"`
	Note   string `json:"note"`
}

// CreateSettlement: admin chốt phiên chi trả — gom bút toán chưa đối soát của shop.
func (h *WalletHandler) CreateSettlement(c *gin.Context) {
	var req CreateSettlementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Thiếu shop_id"})
		return
	}

	settlement, err := h.walletUseCase.CreateSettlement(c.Request.Context(), req.ShopID, req.Note)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": settlement})
}

// MarkPaid: admin đánh dấu phiên đã chi tiền cho shop.
func (h *WalletHandler) MarkPaid(c *gin.Context) {
	id := c.Param("id")
	settlement, err := h.walletUseCase.MarkSettlementPaid(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": settlement})
}
