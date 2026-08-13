package http

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"ocean-express-api/internal/delivery/http/middleware"
	"ocean-express-api/internal/domain"
	"ocean-express-api/pkg/pdf"

	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	orderUseCase domain.OrderUseCase
	rateUseCase  domain.RateUseCase
}

func NewOrderHandler(r *gin.Engine, orderUC domain.OrderUseCase, rateUC domain.RateUseCase, shopRepo domain.ShopRepository) {
	handler := &OrderHandler{
		orderUseCase: orderUC,
		rateUseCase:  rateUC,
	}

	api := r.Group("/api/v1")
	{
		// 1. API DÃ nh cho Äá»‘i tÃ¡c (Shop)
		shopGroup := api.Group("")
		shopGroup.Use(middleware.ShopAPIKeyAuth(shopRepo))
		{
			shopGroup.POST("/rates/calculate", handler.CalculateRate)
			shopGroup.POST("/orders", handler.CreateOrder)
		}

		// 2. Portal Shop (JWT role 'shop'): táº¡o Ä‘Æ¡n + tÃ­nh cÆ°á»›c báº±ng phiÃªn Ä‘Äƒng nháº­p
		// thay vÃ¬ API key. shopID láº¥y tá»« user_id trong token.
		shopPortalGroup := api.Group("/shop")
		shopPortalGroup.Use(middleware.AuthRequired(), middleware.RoleRequired(domain.RoleShop))
		{
			shopPortalGroup.POST("/orders", handler.CreateOrderPortal)
			shopPortalGroup.POST("/orders/import", handler.ImportOrders)
			shopPortalGroup.POST("/rates/calculate", handler.CalculateRate)
			shopPortalGroup.GET("/orders/:id/pdf", handler.GetOrderLabel)
			shopPortalGroup.GET("/orders/:id/label", handler.GetOrderLabel)
		}

		// 3. API DÃ nh cho Ná»™i bá»™ (NhÃ¢n viÃªn / Shipper)
		internalGroup := api.Group("")
		internalGroup.Use(middleware.AuthRequired())
		{
			internalGroup.GET("/orders", handler.GetOrders)
			// Tra cá»©u theo mÃ£ váº­n Ä‘Æ¡n: má» i role ná»™i bá»™ Ä‘á» u tra Ä‘Æ°á»£c (Ä‘áº·c biá»‡t Hub Staff
			// dÃ¹ng Ä‘á»ƒ quÃ©t Ä‘Æ¡n chÆ°a náº±m trong danh sÃ¡ch hub cá»§a mÃ¬nh). DÃ¹ng path riÃªng
			// /tracking/:tracking_number thay vÃ¬ lá»“ng dÆ°á»›i /orders/... vÃ¬ Gin khÃ´ng cho
			// static segment ("tracking") Ä‘á»©ng cÃ¹ng vá»‹ trÃ­ vá»›i param (:id).
			internalGroup.GET("/tracking/:tracking_number", handler.GetOrderByTracking)
			internalGroup.GET("/orders/:id", handler.GetOrder)
			internalGroup.GET("/orders/:id/label", handler.GetOrderLabel)
			internalGroup.GET("/orders/:id/pdf", handler.GetOrderLabel)
			internalGroup.PUT("/orders/:id/status", handler.UpdateStatus)
			internalGroup.POST("/orders/:id/assign", middleware.RoleRequired(domain.RoleAdmin, domain.RoleHubStaff), handler.AssignOrder)
			internalGroup.POST("/orders/submit-cod", middleware.RoleRequired("first_mile_driver", "last_mile_driver"), handler.SubmitCOD)
		}
		
		// 4. API Public (Tra cứu vận đơn cho khách hàng)
		publicGroup := api.Group("/public")
		{
			publicGroup.GET("/tracking/:tracking_number", handler.GetPublicTracking)
		}
	}
}

type CalculateRateRequest struct {
	ReceiverLocationID string `json:"receiver_location_id" binding:"required"`
	Weight             int    `json:"weight" binding:"required"`
	Length             int    `json:"length"`
	Width              int    `json:"width"`
	Height             int    `json:"height"`
}

func (h *OrderHandler) CalculateRate(c *gin.Context) {
	shopID, _ := c.Get("shop_id")
	
	var req CalculateRateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Dữ liệu không hợp lệ"})
		return
	}

	chargeableWeight := req.Weight
	if req.Length > 0 && req.Width > 0 && req.Height > 0 {
		volWeight := (req.Length * req.Width * req.Height) / 5
		if volWeight > chargeableWeight {
			chargeableWeight = volWeight
		}
	}

	fee, err := h.rateUseCase.CalculateFee(c.Request.Context(), "", req.ReceiverLocationID, chargeableWeight)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"fee": fee, "shop_id": shopID, "chargeable_weight": chargeableWeight}})
}

type CreateOrderRequest struct {
	ReceiverName          string   `json:"receiver_name" binding:"required"`
	ReceiverPhone         string   `json:"receiver_phone" binding:"required"`
	ReceiverLocationID    string   `json:"receiver_location_id" binding:"required"`
	ReceiverAddressDetail string   `json:"receiver_address_detail" binding:"required"`
	Weight                int      `json:"weight" binding:"required"`
	Length                int      `json:"length"`
	Width                 int      `json:"width"`
	Height                int      `json:"height"`
	CodAmount             float64  `json:"cod_amount"`
	SenderLatitude        *float64 `json:"sender_latitude"`
	SenderLongitude       *float64 `json:"sender_longitude"`
	ReceiverLatitude      *float64 `json:"receiver_latitude"`
	ReceiverLongitude     *float64 `json:"receiver_longitude"`
}

func (h *OrderHandler) CreateOrder(c *gin.Context) {
	shopIDStr, _ := c.Get("shop_id")
	shopID := shopIDStr.(string)

	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Dữ liệu không hợp lệ"})
		return
	}

	order, err := h.orderUseCase.CreateOrder(
		c.Request.Context(), 
		shopID, 
		req.ReceiverName, 
		req.ReceiverPhone, 
		req.ReceiverLocationID, 
		req.ReceiverAddressDetail, 
		req.Weight, 
		req.Length,
		req.Width,
		req.Height,
		req.CodAmount,
		req.SenderLatitude,
		req.SenderLongitude,
		req.ReceiverLatitude,
		req.ReceiverLongitude,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": order})
}

// CreateOrderPortal là biến thể của CreateOrder cho portal Shop: shopID lấy từ
// phiên đăng nhập (JWT user_id) thay vì API key. Tái dùng cùng usecase.
func (h *OrderHandler) CreateOrderPortal(c *gin.Context) {
	shopIDStr, _ := c.Get("user_id")
	shopID, _ := shopIDStr.(string)

	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Dữ liệu không hợp lệ"})
		return
	}

	order, err := h.orderUseCase.CreateOrder(
		c.Request.Context(),
		shopID,
		req.ReceiverName,
		req.ReceiverPhone,
		req.ReceiverLocationID,
		req.ReceiverAddressDetail,
		req.Weight,
		req.Length,
		req.Width,
		req.Height,
		req.CodAmount,
		req.SenderLatitude,
		req.SenderLongitude,
		req.ReceiverLatitude,
		req.ReceiverLongitude,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": order})
}

type UpdateStatusRequest struct {
	Status    string  `json:"status" binding:"required"`
	Note      string  `json:"note"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

func (h *OrderHandler) UpdateStatus(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Status        string  `json:"status" binding:"required"`
		Note          string  `json:"note"`
		FailureReason string  `json:"failure_reason"`
		Latitude      float64 `json:"latitude"`
		Longitude     float64 `json:"longitude"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userIdStr, _ := c.Get("user_id")
	userId := ""
	if userIdStr != nil {
		userId = userIdStr.(string)
	}

	userRoleStr, _ := c.Get("role")
	var userRole string
	if userRoleStr != nil {
		userRole = userRoleStr.(string)
	}

	userHubId, ok := c.Get("hub_id")
	var hubId string
	if ok && userHubId != nil {
		hubId = userHubId.(string)
	}

	err := h.orderUseCase.UpdateOrderStatus(c.Request.Context(), id, req.Status, req.Note, req.FailureReason, userId, string(userRole), hubId, req.Latitude, req.Longitude)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order status updated"})
}

type AssignOrderRequest struct {
	ShipperID string `json:"shipper_id" binding:"required"`
}

func (h *OrderHandler) AssignOrder(c *gin.Context) {
	id := c.Param("id")
	var req AssignOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userIdStr, _ := c.Get("user_id")
	userId := ""
	if userIdStr != nil {
		userId = userIdStr.(string)
	}

	userRoleStr, _ := c.Get("role")
	var userRole string
	if userRoleStr != nil {
		userRole = userRoleStr.(string)
	}

	err := h.orderUseCase.AssignOrder(c.Request.Context(), id, req.ShipperID, userId, userRole)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order assigned successfully"})
}

func (h *OrderHandler) GetOrderLabel(c *gin.Context) {
	id := c.Param("id")
	order, _, err := h.orderUseCase.GetOrderDetails(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	pdfBytes, err := pdf.GenerateOrderLabelPDF(order)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate PDF"})
		return
	}

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "inline; filename=\"label_"+order.TrackingNumber+".pdf\"")
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}

func (h *OrderHandler) GetOrders(c *gin.Context) {
	empIDStr, _ := c.Get("user_id")
	var empID string
	if empIDStr != nil {
		empID = empIDStr.(string)
	}

	roleStr, _ := c.Get("role")
	var empRole string
	if roleStr != nil {
		empRole = roleStr.(string)
	}

	hubIDStr, _ := c.Get("hub_id")
	var empHubID string
	if hubIDStr != nil {
		empHubID = hubIDStr.(string)
	}

	var pageParams domain.PaginationParams
	if err := c.ShouldBindQuery(&pageParams); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Tham số phân trang không hợp lệ"})
		return
	}

	paginatedResp, err := h.orderUseCase.GetOrders(c.Request.Context(), empRole, empID, empHubID, pageParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	// Return the structure directly
	c.JSON(http.StatusOK, gin.H{
		"success": true, 
		"data": paginatedResp.Data,
		"meta": paginatedResp.Meta,
	})
}

func (h *OrderHandler) GetOrder(c *gin.Context) {
	orderID := c.Param("id")
	order, logs, err := h.orderUseCase.GetOrderDetails(c.Request.Context(), orderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{
		"order": order,
		"logs":  logs,
	}})
}

func (h *OrderHandler) GetOrderByTracking(c *gin.Context) {
	trackingNumber := c.Param("tracking_number")
	order, logs, err := h.orderUseCase.GetOrderDetailsByTrackingNumber(c.Request.Context(), trackingNumber)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": gin.H{
			"code":    "NOT_FOUND",
			"message": "KhÃ´ng tÃ¬m tháº¥y váº­n Ä‘Æ¡n vá»›i mÃ£ nÃ y",
		}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{
		"order": order,
		"logs":  logs,
	}})
}
func (h *OrderHandler) SubmitCOD(c *gin.Context) {
	userId := c.MustGet("user_id").(string)
	
	total, err := h.orderUseCase.SubmitCOD(c.Request.Context(), userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "N?p COD thành công",
		"data": gin.H{
			"amount_submitted": total,
		},
	})
}

func (h *OrderHandler) GetPublicTracking(c *gin.Context) {
	trackingNumber := c.Param("tracking_number")
	
	order, logs, err := h.orderUseCase.GetOrderDetailsByTrackingNumber(c.Request.Context(), trackingNumber)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Khong tim thay don hang"})
		return
	}

	// Filter out sensitive data for public tracking
	publicData := gin.H{
		"tracking_number":         order.TrackingNumber,
		"status":                  order.Status,
		"receiver_name":           order.ReceiverName, // could be masked like Nguyen V***
		"receiver_phone":          order.ReceiverPhone, // could be masked like 090****123
		"created_at":              order.CreatedAt,
		"tracking_logs":           logs,
		"sender_latitude":         order.SenderLatitude,
		"sender_longitude":        order.SenderLongitude,
		"receiver_latitude":       order.ReceiverLatitude,
		"receiver_longitude":      order.ReceiverLongitude,
		"sender_address_detail":   order.SenderAddressDetail,
		"receiver_address_detail": order.ReceiverAddressDetail,
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": publicData})
}

// ImportOrders handles CSV file upload and batch order creation
func (h *OrderHandler) ImportOrders(c *gin.Context) {
	shopIDStr, _ := c.Get("user_id")
	shopID, _ := shopIDStr.(string)

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Vui lòng chọn file CSV"})
		return
	}

	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Không thể đọc file"})
		return
	}
	defer f.Close()

	importReader := csv.NewReader(f)
	records, err := importReader.ReadAll()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "File CSV không hợp lệ"})
		return
	}

	if len(records) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "File CSV không có dữ liệu"})
		return
	}

	successCount := 0
	var errors []string

	for i, row := range records[1:] { // Bỏ qua header
		if len(row) < 5 {
			continue
		}
		
		receiverName := row[0]
		receiverPhone := row[1]
		receiverLocID := row[2]
		receiverAddress := row[3]
		weightStr := row[4]
		
		weight := 1000 // default
		fmt.Sscanf(weightStr, "%d", &weight)

		_, err := h.orderUseCase.CreateOrder(
			c.Request.Context(),
			shopID,
			receiverName,
			receiverPhone,
			receiverLocID,
			receiverAddress,
			weight,
			0, 0, 0, 0, // Kích thước mặc định
			nil, nil, nil, nil, // Không có tọa độ GPS
		)

		if err != nil {
			errors = append(errors, fmt.Sprintf("Dòng %d: %v", i+2, err))
		} else {
			successCount++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"success_count": successCount,
			"errors":        errors,
		},
	})
}
