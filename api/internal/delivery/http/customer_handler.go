package http

import (
	"net/http"
	"ocean-express-api/internal/delivery/http/middleware"
	"ocean-express-api/internal/domain"

	"github.com/gin-gonic/gin"
)

type CustomerHandler struct {
	customerUC domain.CustomerUseCase
}

func NewCustomerHandler(r *gin.Engine, customerUC domain.CustomerUseCase) {
	handler := &CustomerHandler{customerUC: customerUC}

	api := r.Group("/api/v1/shop/customers")
	api.Use(middleware.AuthRequired(), middleware.RoleRequired(domain.RoleShop))
	{
		api.GET("", handler.SearchCustomers)
		api.POST("", handler.SaveCustomer)
	}
}

func (h *CustomerHandler) SearchCustomers(c *gin.Context) {
	shopID, _ := c.Get("user_id")
	query := c.Query("q")

	customers, err := h.customerUC.SearchCustomers(c.Request.Context(), shopID.(string), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": customers})
}

type SaveCustomerRequest struct {
	Name          string   `json:"name" binding:"required"`
	Phone         string   `json:"phone" binding:"required"`
	LocationID    string   `json:"location_id"`
	AddressDetail string   `json:"address_detail"`
	Latitude      *float64 `json:"latitude"`
	Longitude     *float64 `json:"longitude"`
}

func (h *CustomerHandler) SaveCustomer(c *gin.Context) {
	shopID, _ := c.Get("user_id")

	var req SaveCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	cust, err := h.customerUC.SaveCustomer(
		c.Request.Context(),
		shopID.(string),
		req.Name,
		req.Phone,
		req.LocationID,
		req.AddressDetail,
		req.Latitude,
		req.Longitude,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": cust})
}
