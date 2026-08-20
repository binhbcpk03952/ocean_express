package http

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"ocean-express-api/internal/delivery/http/middleware"
	"ocean-express-api/internal/domain"
	"ocean-express-api/pkg/pdf"
	"ocean-express-api/pkg/utils"

	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	orderUseCase domain.OrderUseCase
	rateUseCase  domain.RateUseCase
	shopRepo     domain.ShopRepository
}

func NewOrderHandler(r *gin.Engine, orderUC domain.OrderUseCase, rateUC domain.RateUseCase, shopRepo domain.ShopRepository) {
	handler := &OrderHandler{
		orderUseCase: orderUC,
		rateUseCase:  rateUC,
		shopRepo:     shopRepo,
	}

	api := r.Group("/api/v1")
	{
		// 1. API Quản lý Vận đơn (Nội bộ JWT hoặc Đối tác Shop qua X-API-Key)
		ordersGroup := api.Group("")
		ordersGroup.Use(middleware.CombinedAuth(shopRepo))
		{
			ordersGroup.POST("/rates/calculate", handler.CalculateRate)
			ordersGroup.POST("/orders", handler.CreateOrder)
			ordersGroup.GET("/orders", handler.GetOrders)
			ordersGroup.GET("/orders/:id", handler.GetOrder)
			ordersGroup.GET("/orders/:id/label", handler.GetOrderLabel)
			ordersGroup.GET("/orders/:id/pdf", handler.GetOrderLabel)
			ordersGroup.GET("/orders/:id/print", handler.GetOrderLabel)
			ordersGroup.GET("/orders/:id/print-label", handler.PrintLabelJSON)
			ordersGroup.POST("/orders/print-label", handler.PrintLabelJSON)
			ordersGroup.POST("/orders/labels/batch", handler.GetBatchOrderLabels)
			ordersGroup.GET("/tracking/:tracking_number", handler.GetOrderByTracking)
			ordersGroup.PUT("/orders/:id/status", handler.UpdateStatus)
			ordersGroup.POST("/orders/:id/assign", middleware.RoleRequired(string(domain.RoleAdmin), string(domain.RoleHubStaff)), handler.AssignOrder)
			ordersGroup.POST("/orders/submit-cod", middleware.RoleRequired("first_mile_driver", "last_mile_driver"), handler.SubmitCOD)
		}

		// 2. Portal Shop (JWT role 'shop'): tạo đơn + tính cước bằng phiên đăng nhập
		// thay vì API key. shopID lấy từ user_id trong token.
		shopPortalGroup := api.Group("/shop")
		shopPortalGroup.Use(middleware.AuthRequired(), middleware.RoleRequired(domain.RoleShop))
		{
			shopPortalGroup.POST("/orders", handler.CreateOrderPortal)
			shopPortalGroup.POST("/orders/import", handler.ImportOrders)
			shopPortalGroup.POST("/rates/calculate", handler.CalculateRate)
			shopPortalGroup.GET("/orders/:id/pdf", handler.GetOrderLabel)
			shopPortalGroup.GET("/orders/:id/label", handler.GetOrderLabel)
			shopPortalGroup.GET("/orders/:id/print-label", handler.PrintLabelJSON)
			shopPortalGroup.POST("/orders/labels/batch", handler.GetBatchOrderLabels)
		}
		
		// 3. API Public & In vận đơn (Tra cứu vận đơn & in tem công khai không chặn auth để iframe / popup in dễ dàng)
		publicGroup := api.Group("/public")
		{
			publicGroup.GET("/tracking/:tracking_number", handler.GetPublicTracking)
			publicGroup.GET("/tracking/:tracking_number/label", handler.GetOrderLabel)
			publicGroup.GET("/tracking/:tracking_number/pdf", handler.GetOrderLabel)
			publicGroup.GET("/orders/:id/label", handler.GetOrderLabel)
			publicGroup.GET("/orders/:id/pdf", handler.GetOrderLabel)
			publicGroup.GET("/orders/:id/print-label", handler.PrintLabelJSON)
			publicGroup.POST("/orders/print-label", handler.PrintLabelJSON)
			publicGroup.POST("/orders/labels/batch", handler.GetBatchOrderLabels)
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

type BatchLabelsRequest struct {
	OrderIDs []string `json:"order_ids" binding:"required"`
}

func (h *OrderHandler) GetOrderLabel(c *gin.Context) {
	id := c.Param("id")
	if id == "" || id == "label" || id == "pdf" {
		id = c.Param("tracking_number")
	}
	if id == "" {
		id = c.Query("tracking_number")
	}
	if id == "" {
		id = c.Query("order_code")
	}
	if id == "" {
		id = c.Query("order_id")
	}
	if id == "" {
		id = c.Query("id")
	}

	order, _, err := h.orderUseCase.GetOrderDetails(c.Request.Context(), id)
	if err != nil || order == nil {
		order, _, err = h.orderUseCase.GetOrderDetailsByTrackingNumber(c.Request.Context(), id)
	}
	if err != nil || order == nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Không tìm thấy vận đơn"})
		return
	}

	userRoleStr, _ := c.Get("role")
	userIdStr, _ := c.Get("user_id")
	if userRoleStr != nil && userRoleStr.(string) == "shop" {
		if userIdStr != nil && order.ShopID != userIdStr.(string) {
			c.JSON(http.StatusForbidden, gin.H{"success": false, "error": "Bạn không có quyền in vận đơn này"})
			return
		}
	}

	var shop *domain.Shop
	if order.ShopID != "" {
		shop, _ = h.shopRepo.GetByID(c.Request.Context(), order.ShopID)
	}

	pdfBytes, err := pdf.GenerateOrderLabelPDF(order, shop)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Lỗi tạo file PDF: " + err.Error()})
		return
	}

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "inline; filename=\"label_"+order.TrackingNumber+".pdf\"")
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}

func (h *OrderHandler) PrintLabelJSON(c *gin.Context) {
	id := c.Param("id")
	if id == "print-label" {
		id = ""
	}
	if id == "" {
		id = c.Param("tracking_number")
	}
	if id == "" {
		id = c.Query("order_id")
		if id == "" {
			id = c.Query("tracking_number")
		}
		if id == "" {
			id = c.Query("order_code")
		}
		if id == "" {
			id = c.Query("code")
		}
		if id == "" {
			id = c.Query("id")
		}
	}
	
	if id == "" {
		if bodyBytes, err := io.ReadAll(c.Request.Body); err == nil && len(bodyBytes) > 0 {
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			var rawMap map[string]interface{}
			if err := json.Unmarshal(bodyBytes, &rawMap); err == nil {
				for _, k := range []string{"tracking_number", "order_code", "order_id", "id", "code"} {
					if val, ok := rawMap[k]; ok {
						if strVal, ok := val.(string); ok && strVal != "" {
							id = strVal
							break
						}
					}
				}
				// Handle array of order_codes (e.g. GHN format: {"order_codes": ["OE-..."]})
				if id == "" {
					if codes, ok := rawMap["order_codes"].([]interface{}); ok && len(codes) > 0 {
						if firstCode, ok := codes[0].(string); ok {
							id = firstCode
						}
					}
				}
			}
		}
	}

	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "status": "error", "success": false, "message": "Thiếu mã đơn hàng hoặc mã vận đơn", "error": "Thiếu mã đơn hàng hoặc mã vận đơn"})
		return
	}

	order, _, err := h.orderUseCase.GetOrderDetails(c.Request.Context(), id)
	if err != nil || order == nil {
		order, _, err = h.orderUseCase.GetOrderDetailsByTrackingNumber(c.Request.Context(), id)
	}
	if err != nil || order == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "status": "error", "success": false, "message": "Không tìm thấy vận đơn", "error": "Không tìm thấy vận đơn"})
		return
	}

	st := utils.GetStatusInfo(order.Status)
	labelURL := fmt.Sprintf("https://api.oceanexpress.bcbdev.id.vn/api/v1/public/orders/%s/label", order.TrackingNumber)
	trackingURL := fmt.Sprintf("https://oceanexpress.bcbdev.id.vn/tracking?code=%s", order.TrackingNumber)

	c.JSON(http.StatusOK, gin.H{
		"code":            200,
		"status":          "success",
		"success":         true,
		"message":         "Lấy thông tin in vận đơn thành công",
		"label_url":       labelURL,
		"pdf_url":         labelURL,
		"tracking_url":    trackingURL,
		"data": gin.H{
			"order_id":           order.ID,
			"tracking_number":    order.TrackingNumber,
			"token":              "oe_" + order.TrackingNumber,
			"print_url":          labelURL,
			"status":             order.Status,
			"status_name":        st.Name,
			"status_label":       st.Label,
			"status_description": st.Description,
			"status_badge":       st,
			"label_url":          labelURL,
			"pdf_url":            labelURL,
			"tracking_url":       trackingURL,
		},
	})
}

func (h *OrderHandler) GetBatchOrderLabels(c *gin.Context) {
	var req BatchLabelsRequest
	if err := c.ShouldBindJSON(&req); err != nil || len(req.OrderIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Danh sách mã vận đơn không hợp lệ"})
		return
	}

	userRoleStr, _ := c.Get("role")
	userIdStr, _ := c.Get("user_id")

	var orders []*domain.ShippingOrder
	shopMap := make(map[string]*domain.Shop)

	for _, id := range req.OrderIDs {
		order, _, err := h.orderUseCase.GetOrderDetails(c.Request.Context(), id)
		if err == nil && order != nil {
			if userRoleStr != nil && userRoleStr.(string) == "shop" {
				if userIdStr != nil && order.ShopID != userIdStr.(string) {
					continue // bỏ qua đơn không thuộc quyền sở hữu của shop
				}
			}
			orders = append(orders, order)
			if order.ShopID != "" {
				if _, exists := shopMap[order.ShopID]; !exists {
					shop, _ := h.shopRepo.GetByID(c.Request.Context(), order.ShopID)
					if shop != nil {
						shopMap[order.ShopID] = shop
					}
				}
			}
		}
	}

	if len(orders) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Không tìm thấy vận đơn nào hợp lệ để in"})
		return
	}

	pdfBytes, err := pdf.GenerateBatchOrderLabelsPDF(orders, shopMap)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Lỗi tạo file PDF batch: " + err.Error()})
		return
	}

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "inline; filename=\"batch_labels.pdf\"")
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
	if err != nil || order == nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Không tìm thấy vận đơn"})
		return
	}

	userRoleStr, _ := c.Get("role")
	userIdStr, _ := c.Get("user_id")
	if userRoleStr != nil && userRoleStr.(string) == "shop" {
		if userIdStr != nil && order.ShopID != userIdStr.(string) {
			c.JSON(http.StatusForbidden, gin.H{"success": false, "error": "Bạn không có quyền truy cập vận đơn này"})
			return
		}
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
			"message": "Không tìm thấy vận đơn với mã này",
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
		"message": "Nộp COD thành công",
		"data": gin.H{
			"amount_submitted": total,
		},
	})
}

func (h *OrderHandler) GetPublicTracking(c *gin.Context) {
	trackingNumber := c.Param("tracking_number")
	
	order, logs, err := h.orderUseCase.GetOrderDetailsByTrackingNumber(c.Request.Context(), trackingNumber)
	if err != nil || order == nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Không tìm thấy thông tin vận đơn"})
		return
	}

	// Filter and format data for public tracking
	st := utils.GetStatusInfo(order.Status)
	publicData := gin.H{
		"id":                      order.ID,
		"tracking_number":         order.TrackingNumber,
		"status":                  order.Status,
		"status_name":             st.Name,
		"status_label":            st.Label,
		"status_description":      st.Description,
		"status_display":          st.Name,
		"status_badge":            st,
		"receiver_name":           order.ReceiverName,
		"receiver_phone":          order.ReceiverPhone,
		"weight":                  order.Weight,
		"length":                  order.Length,
		"width":                   order.Width,
		"height":                  order.Height,
		"cod_amount":              order.CodAmount,
		"shipping_fee":            order.ShippingFee,
		"estimated_delivery_time": order.EstimatedDeliveryTime,
		"sla_deadline":            order.SlaDeadline,
		"sla_breached":            order.SlaBreached,
		"delivery_attempts":       order.DeliveryAttempts,
		"failure_reason":          order.FailureReason,
		"created_at":              order.CreatedAt,
		"updated_at":              order.UpdatedAt,
		"tracking_logs":           logs,
		"sender_latitude":         order.SenderLatitude,
		"sender_longitude":        order.SenderLongitude,
		"receiver_latitude":       order.ReceiverLatitude,
		"receiver_longitude":      order.ReceiverLongitude,
		"sender_address_detail":   order.SenderAddressDetail,
		"receiver_address_detail": order.ReceiverAddressDetail,
		"sender_location_id":      order.SenderLocationID,
		"receiver_location_id":    order.ReceiverLocationID,
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
