package http

import (
	"net/http"
	"ocean-express-api/internal/delivery/http/middleware"
	"ocean-express-api/internal/domain"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authUseCase     domain.AuthUseCase
	shopAuthUseCase domain.ShopAuthUseCase
}

func NewAuthHandler(r *gin.Engine, authUC domain.AuthUseCase, shopAuthUC domain.ShopAuthUseCase) {
	handler := &AuthHandler{
		authUseCase:     authUC,
		shopAuthUseCase: shopAuthUC,
	}

	api := r.Group("/api/v1")
	{
		api.POST("/auth/login", handler.Login)          // nhân viên/tài xế: phone/email + password
		api.POST("/auth/shop/login", handler.ShopLogin) // đối tác shop: email/phone + password
		
		api.POST("/auth/forgot-password", handler.ForgotPassword)
		api.POST("/auth/reset-password", handler.ResetPassword)

		authed := api.Group("/auth")
		authed.Use(middleware.AuthRequired())
		{
			authed.POST("/logout", handler.Logout)
		}
	}
}

type LoginRequest struct {
	Identifier string `json:"identifier" binding:"required"`
	Password   string `json:"password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "Dữ liệu không hợp lệ",
			},
		})
		return
	}

	token, emp, err := h.authUseCase.Login(c.Request.Context(), req.Identifier, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "UNAUTHORIZED",
				"message": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"token": token,
			"employee": gin.H{
				"id":     emp.ID,
				"name":   emp.Name,
				"role":   emp.Role,
				"hub_id": emp.HubID,
			},
		},
	})
}

type ShopLoginRequest struct {
	Identifier string `json:"identifier" binding:"required"`
	Password   string `json:"password" binding:"required"`
}

// ShopLogin đăng nhập portal cho đối tác Shop bằng email + mật khẩu. Trả về JWT
// role 'shop' dùng chung cơ chế session như nhân viên.
func (h *AuthHandler) ShopLogin(c *gin.Context) {
	var req ShopLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "Dữ liệu không hợp lệ",
			},
		})
		return
	}

	token, shop, err := h.shopAuthUseCase.Login(c.Request.Context(), req.Identifier, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "UNAUTHORIZED",
				"message": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"token": token,
			"shop": gin.H{
				"id":    shop.ID,
				"name":  shop.Name,
				"email": shop.Email,
				"role":  domain.RoleShop,
			},
		},
	})
}

// Logout thu hồi phiên hiện tại. jti được middleware AuthRequired trích từ token
// và lưu vào context.
func (h *AuthHandler) Logout(c *gin.Context) {
	jtiVal, _ := c.Get("jti")
	var jti string
	if jtiVal != nil {
		jti = jtiVal.(string)
	}

	if err := h.authUseCase.Logout(c.Request.Context(), jti); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "LOGOUT_FAILED",
				"message": "Không thể đăng xuất",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Đã đăng xuất"})
}
