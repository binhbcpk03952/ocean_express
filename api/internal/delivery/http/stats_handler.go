package http

import (
	"net/http"
	"ocean-express-api/internal/delivery/http/middleware"
	"ocean-express-api/internal/domain"

	"github.com/gin-gonic/gin"
)

type StatsHandler struct {
	statsUseCase domain.StatsUseCase
}

func NewStatsHandler(r *gin.Engine, statsUC domain.StatsUseCase) {
	handler := &StatsHandler{
		statsUseCase: statsUC,
	}

	api := r.Group("/api/v1/stats")
	api.Use(middleware.AuthRequired())
	{
		// Admin only
		adminStats := api.Group("")
		adminStats.Use(middleware.RoleRequired("admin"))
		{
			adminStats.GET("/dashboard", handler.GetDashboard)
			adminStats.GET("/shop/:id", handler.GetShopStats)
		}
		
		// For members (Shipper, Hub staff)
		api.GET("/member/me", handler.GetMemberStats)
		
		// Shop can see their own stats via their specific endpoint in the future
		shopStats := api.Group("/shop")
		shopStats.Use(middleware.RoleRequired("shop"))
		{
			shopStats.GET("/me", handler.GetShopStatsForMe)
		}
	}
}

func (h *StatsHandler) GetDashboard(c *gin.Context) {
	stats, err := h.statsUseCase.GetDashboard(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": stats})
}

func (h *StatsHandler) GetMemberStats(c *gin.Context) {
	userIDStr, _ := c.Get("user_id")
	if userIDStr == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Unauthorized"})
		return
	}
	userID := userIDStr.(string)
	
	roleStr, _ := c.Get("role")
	role := ""
	if roleStr != nil {
		role = roleStr.(string)
	}
	
	stats, err := h.statsUseCase.GetMemberStats(c.Request.Context(), userID, role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": stats})
}

func (h *StatsHandler) GetShopStats(c *gin.Context) {
	shopID := c.Param("id")
	stats, err := h.statsUseCase.GetShopStats(c.Request.Context(), shopID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": stats})
}

func (h *StatsHandler) GetShopStatsForMe(c *gin.Context) {
	shopIDStr, _ := c.Get("user_id")
	if shopIDStr == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Unauthorized"})
		return
	}
	shopID := shopIDStr.(string)

	stats, err := h.statsUseCase.GetShopStats(c.Request.Context(), shopID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": stats})
}
