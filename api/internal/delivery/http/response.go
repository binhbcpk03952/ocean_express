package http

import (
	"ocean-express-api/internal/domain"

	"github.com/gin-gonic/gin"
)

// respondError trả về lỗi theo cấu trúc JSON chuẩn, tự map sentinel error trong
// domain sang đúng HTTP status + error.code. Lỗi không phân loại được -> 500.
func respondError(c *gin.Context, err error) {
	status := domain.HTTPStatusForError(err)
	code := domain.ErrorCode(err)
	if code == "" {
		code = "INTERNAL_ERROR"
	}
	c.JSON(status, gin.H{
		"success": false,
		"error": gin.H{
			"code":    code,
			"message": err.Error(),
		},
	})
}
