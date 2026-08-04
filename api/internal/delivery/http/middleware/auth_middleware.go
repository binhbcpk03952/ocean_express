package middleware

import (
	"net/http"
	"ocean-express-api/internal/domain"
	"ocean-express-api/pkg/utils"
	"strings"

	"github.com/gin-gonic/gin"
)

// SessionStore là session repo dùng chung cho AuthRequired. Được set một lần lúc
// khởi động (main.go). Nếu nil, middleware chỉ xác thực JWT stateless như cũ;
// nếu set, mỗi request bị kiểm tra jti còn hiệu lực trong Redis (hỗ trợ thu hồi).
var SessionStore domain.SessionRepository

// AuthRequired kiểm tra tính hợp lệ của JWT token
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "Vui lòng cung cấp token hợp lệ",
				},
			})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := utils.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "INVALID_TOKEN",
					"message": "Token không hợp lệ hoặc đã hết hạn",
				},
			})
			c.Abort()
			return
		}

		// Kiểm tra phiên còn hiệu lực trong Redis (nếu session store được bật).
		// Cho phép thu hồi token thật sự khi logout / khóa tài khoản.
		if SessionStore != nil {
			ok, serr := SessionStore.Exists(c.Request.Context(), claims.ID)
			if serr != nil || !ok {
				c.JSON(http.StatusUnauthorized, gin.H{
					"success": false,
					"error": gin.H{
						"code":    "SESSION_REVOKED",
						"message": "Phiên đăng nhập đã kết thúc, vui lòng đăng nhập lại",
					},
				})
				c.Abort()
				return
			}
		}

		// Lưu thông tin user vào context để handler sử dụng
		c.Set("user_id", claims.UserID)
		c.Set("role", claims.Role)
		c.Set("hub_id", claims.HubID)
		c.Set("jti", claims.ID)

		c.Next()
	}
}

// RoleRequired kiểm tra xem người dùng có Role yêu cầu hay không
func RoleRequired(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Chưa đăng nhập"})
			c.Abort()
			return
		}

		roleStr := userRole.(string)
		hasRole := false
		for _, r := range roles {
			if r == roleStr {
				hasRole = true
				break
			}
		}

		if !hasRole {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "FORBIDDEN",
					"message": "Bạn không có quyền thực hiện chức năng này",
				},
			})
			c.Abort()
			return
		}
		
		c.Next()
	}
}
