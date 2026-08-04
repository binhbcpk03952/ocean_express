package http

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"ocean-express-api/internal/delivery/http/middleware"

	"github.com/gin-gonic/gin"
)

type RoutesHandler struct{}

func NewRoutesHandler(r *gin.Engine) {
	handler := &RoutesHandler{}

	api := r.Group("/api/v1/routes")
	api.Use(middleware.AuthRequired())
	{
		api.POST("/optimize", handler.OptimizeRoute)
	}
}

// Request struct cho API optimize
type OptimizeRequest struct {
	Locations []struct {
		ID  string  `json:"id"`
		Lat float64 `json:"lat"`
		Lng float64 `json:"lng"`
	} `json:"locations"`
}

func (h *RoutesHandler) OptimizeRoute(c *gin.Context) {
	var req OptimizeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	if len(req.Locations) < 2 {
		// Tối thiểu 2 điểm để tối ưu
		c.JSON(http.StatusOK, gin.H{"success": true, "data": req.Locations})
		return
	}

	// Xây dựng URL cho OSRM Trip API (TSP)
	// Định dạng: {lng},{lat};{lng},{lat}
	coords := ""
	for i, loc := range req.Locations {
		if i > 0 {
			coords += ";"
		}
		coords += fmt.Sprintf("%f,%f", loc.Lng, loc.Lat)
	}

	// OSRM chạy local ở cổng 5000 (cần phải config nếu chạy prod)
	// Trip API: /trip/v1/driving/{coordinates}?roundtrip=false&source=first
	osrmURL := fmt.Sprintf("http://localhost:5000/trip/v1/driving/%s?roundtrip=false&source=first", coords)

	resp, err := http.Get(osrmURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Lỗi kết nối OSRM: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Lỗi đọc OSRM response"})
		return
	}

	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "OSRM trả về lỗi: " + string(body)})
		return
	}

	var osrmRes map[string]interface{}
	if err := json.Unmarshal(body, &osrmRes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Lỗi parse OSRM JSON"})
		return
	}

	waypoints, ok := osrmRes["waypoints"].([]interface{})
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Không có waypoints từ OSRM"})
		return
	}

	// Sắp xếp lại locations dựa trên waypoint_index
	optimized := make([]interface{}, len(req.Locations))
	for i, wpInter := range waypoints {
		wp := wpInter.(map[string]interface{})
		idx := int(wp["waypoint_index"].(float64))
		if idx >= 0 && idx < len(optimized) {
			optimized[idx] = req.Locations[i]
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    optimized,
	})
}
