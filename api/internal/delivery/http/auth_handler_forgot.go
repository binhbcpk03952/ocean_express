package http

import (
	"net/http"
	"github.com/gin-gonic/gin"
)

type ForgotPasswordRequest struct {
	Identifier string `json:"identifier" binding:"required"`
	Type       string `json:"type" binding:"required"` // "employee" hoặc "shop"
}

func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Dữ liệu không hợp lệ"})
		return
	}
	
	ctx := c.Request.Context()
	var err error
	
	if req.Type == "employee" {
		err = h.authUseCase.ForgotPassword(ctx, req.Identifier)
	} else if req.Type == "shop" {
		err = h.shopAuthUseCase.ForgotPassword(ctx, req.Identifier)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Type không hợp lệ"})
		return
	}
	
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Hướng dẫn đặt lại mật khẩu đã được gửi đến số điện thoại hoặc email của bạn"})
}

type ResetPasswordRequest struct {
	Identifier  string `json:"identifier" binding:"required"`
	Type        string `json:"type" binding:"required"` // "employee" hoặc "shop"
	OTP         string `json:"otp" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Dữ liệu không hợp lệ"})
		return
	}
	
	ctx := c.Request.Context()
	var err error
	
	if req.Type == "employee" {
		err = h.authUseCase.ResetPassword(ctx, req.Identifier, req.OTP, req.NewPassword)
	} else if req.Type == "shop" {
		err = h.shopAuthUseCase.ResetPassword(ctx, req.Identifier, req.OTP, req.NewPassword)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Type không hợp lệ"})
		return
	}
	
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Đặt lại mật khẩu thành công"})
}
