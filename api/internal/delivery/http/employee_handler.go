package http

import (
	"net/http"
	"ocean-express-api/internal/delivery/http/middleware"
	"ocean-express-api/internal/domain"

	"github.com/gin-gonic/gin"
)

type EmployeeHandler struct {
	employeeUseCase domain.EmployeeUseCase
}

func NewEmployeeHandler(r *gin.Engine, empUC domain.EmployeeUseCase) {
	handler := &EmployeeHandler{
		employeeUseCase: empUC,
	}

	api := r.Group("/api/v1/employees")
	{
		// Public: shipper tự đăng ký (chờ Admin duyệt).
		api.POST("/register", handler.Register)

		// Quản trị (role 'admin').
		admin := api.Group("")
		admin.Use(middleware.AuthRequired(), middleware.RoleRequired("admin"))
		{
			admin.GET("", handler.GetEmployees)
			admin.POST("", handler.CreateEmployee)
			admin.PUT("/:id", handler.UpdateEmployee)
			admin.PATCH("/:id/active", handler.SetActive)
			admin.PATCH("/:id/review", handler.Review)
		}
	}
}

func (h *EmployeeHandler) GetEmployees(c *gin.Context) {
	var hubID *string
	if hub := c.Query("hub_id"); hub != "" {
		hubID = &hub
	}

	var status *string
	if s := c.Query("status"); s != "" {
		status = &s
	}

	var pageParams domain.PaginationParams
	if err := c.ShouldBindQuery(&pageParams); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Tham số phân trang không hợp lệ"})
		return
	}

	paginatedResp, err := h.employeeUseCase.GetEmployees(c.Request.Context(), hubID, status, pageParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true, 
		"data": paginatedResp.Data,
		"meta": paginatedResp.Meta,
	})
}

type RegisterEmployeeRequest struct {
	Name     string  `json:"name" binding:"required"`
	Phone    string  `json:"phone" binding:"required"`
	Email    string  `json:"email"`
	Password string  `json:"password" binding:"required"`
	Role     string  `json:"role" binding:"required"`
	HubID    *string `json:"hub_id" binding:"required"`
}

// Register là shipper tự đăng ký (chờ Admin duyệt). Public, không cần auth.
func (h *EmployeeHandler) Register(c *gin.Context) {
	var req RegisterEmployeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	emp, err := h.employeeUseCase.RegisterEmployee(c.Request.Context(), req.Name, req.Phone, req.Email, req.Password, req.Role, req.HubID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "data": emp, "message": "Đăng ký thành công, tài khoản đang chờ Admin duyệt"})
}

type ReviewEmployeeRequest struct {
	Approve *bool `json:"approve" binding:"required"`
}

// Review duyệt/từ chối một tài khoản shipper đang chờ.
func (h *EmployeeHandler) Review(c *gin.Context) {
	id := c.Param("id")

	var req ReviewEmployeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Thiếu trường approve"})
		return
	}

	emp, err := h.employeeUseCase.ReviewEmployee(c.Request.Context(), id, *req.Approve)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": emp})
}

type CreateEmployeeRequest struct {
	Name     string  `json:"name" binding:"required"`
	Phone    string  `json:"phone" binding:"required"`
	Email    string  `json:"email"`
	Password string  `json:"password" binding:"required"`
	Role     string  `json:"role" binding:"required"`
	HubID    *string `json:"hub_id"`
}

func (h *EmployeeHandler) CreateEmployee(c *gin.Context) {
	var req CreateEmployeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Dữ liệu không hợp lệ"})
		return
	}

	if req.HubID != nil && *req.HubID == "" {
		req.HubID = nil
	}

	emp, err := h.employeeUseCase.CreateEmployee(c.Request.Context(), req.Name, req.Phone, req.Email, req.Password, req.Role, req.HubID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "data": emp})
}

type UpdateEmployeeRequest struct {
	Name     string  `json:"name" binding:"required"`
	Phone    string  `json:"phone" binding:"required"`
	Email    string  `json:"email"`
	Password string  `json:"password"` // rỗng = giữ nguyên mật khẩu cũ
	Role     string  `json:"role" binding:"required"`
	HubID    *string `json:"hub_id"`
}

func (h *EmployeeHandler) UpdateEmployee(c *gin.Context) {
	id := c.Param("id")

	var req UpdateEmployeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Dữ liệu không hợp lệ"})
		return
	}

	if req.HubID != nil && *req.HubID == "" {
		req.HubID = nil
	}

	emp, err := h.employeeUseCase.UpdateEmployee(c.Request.Context(), id, req.Name, req.Phone, req.Email, req.Password, req.Role, req.HubID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": emp})
}

type SetActiveRequest struct {
	IsActive *bool `json:"is_active" binding:"required"`
}

func (h *EmployeeHandler) SetActive(c *gin.Context) {
	id := c.Param("id")

	var req SetActiveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Thiếu trường is_active"})
		return
	}

	emp, err := h.employeeUseCase.SetActive(c.Request.Context(), id, *req.IsActive)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": emp})
}
