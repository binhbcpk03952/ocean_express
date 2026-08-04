package middleware

import (
	"net/http"
	"ocean-express-api/internal/domain"
	"strings"

	"github.com/gin-gonic/gin"
)

// ShopAPIKeyAuth kiểm tra xác thực cho API của Đối tác (Shop)
func ShopAPIKeyAuth(shopRepo domain.ShopRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "MISSING_API_KEY",
					"message": "Vui lòng cung cấp X-API-Key trong Header",
				},
			})
			c.Abort()
			return
		}

		apiKey = strings.TrimSpace(apiKey)
		shop, err := shopRepo.GetByAPIKey(c.Request.Context(), apiKey)
		
		if err != nil || shop == nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "INVALID_API_KEY",
					"message": "API Key không hợp lệ",
				},
			})
			c.Abort()
			return
		}

		// Lưu shop_id vào context
		c.Set("shop_id", shop.ID)
		
		c.Next()
	}
}
