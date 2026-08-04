package geocoding

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Coordinates represents a geographical point
type Coordinates struct {
	Latitude  float64
	Longitude float64
}

// Geocoder defines the interface for geocoding services
type Geocoder interface {
	GetCoordinates(address string) (*Coordinates, error)
}

type nominatimGeocoder struct {
	client *http.Client
}

// NewNominatimGeocoder creates a new OpenStreetMap Nominatim geocoder
func NewNominatimGeocoder() Geocoder {
	return &nominatimGeocoder{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (g *nominatimGeocoder) GetCoordinates(address string) (*Coordinates, error) {
	if address == "" {
		return nil, fmt.Errorf("địa chỉ trống")
	}

	// Tách chuỗi theo dấu phẩy để thực hiện cơ chế Fallback (Fallback Mechanism)
	parts := strings.Split(address, ",")

	for i := 0; i < len(parts); i++ {
		// Tạo chuỗi fallback bằng cách ghép các phần tử từ i đến cuối
		var fallbackParts []string
		for _, p := range parts[i:] {
			trimmed := strings.TrimSpace(p)
			if trimmed != "" {
				fallbackParts = append(fallbackParts, trimmed)
			}
		}

		if len(fallbackParts) == 0 {
			continue
		}

		// Luôn thêm "Việt Nam" vào cuối để tăng độ chính xác
		searchQuery := strings.Join(fallbackParts, ", ") + ", Việt Nam"

		coords, err := g.fetchNominatim(searchQuery)
		// Nếu thành công và có tọa độ, trả về ngay
		if err == nil && coords != nil {
			return coords, nil
		}
		// Nếu lỗi (không tìm thấy), vòng lặp sẽ tiếp tục cắt bỏ phần tử đầu tiên (địa chỉ chi tiết)
		// và thử tìm kiếm bằng cấp độ hành chính rộng hơn (Phường/Xã -> Quận/Huyện -> Tỉnh/Thành)
	}

	return nil, fmt.Errorf("không tìm thấy tọa độ cho địa chỉ: %s", address)
}

func (g *nominatimGeocoder) fetchNominatim(searchQuery string) (*Coordinates, error) {
	encodedAddress := url.QueryEscape(searchQuery)
	// countrycodes=vn: Giới hạn kết quả trong lãnh thổ Việt Nam, tránh nhầm địa danh trùng tên
	// viewbox: Bounding box bao trọn Việt Nam để ưu tiên kết quả trong nước
	// bounded=1: Chỉ trả về kết quả trong viewbox
	reqURL := fmt.Sprintf(
		"https://nominatim.openstreetmap.org/search?q=%s&format=json&limit=1&countrycodes=vn&viewbox=102.14,8.18,109.46,23.39&bounded=1",
		encodedAddress,
	)

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, err
	}
	// Bắt buộc khai báo User-Agent hợp lệ theo chính sách của Nominatim
	req.Header.Set("User-Agent", "OceanExpress/1.0 (contact@oceanexpress.com)")

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("lỗi từ Nominatim API, status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var results []struct {
		Lat string `json:"lat"`
		Lon string `json:"lon"`
	}

	if err := json.Unmarshal(body, &results); err != nil {
		return nil, fmt.Errorf("không thể parse JSON từ Nominatim: %v", err)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("không có kết quả")
	}

	lat, err := strconv.ParseFloat(results[0].Lat, 64)
	if err != nil {
		return nil, err
	}
	lon, err := strconv.ParseFloat(results[0].Lon, 64)
	if err != nil {
		return nil, err
	}

	return &Coordinates{
		Latitude:  lat,
		Longitude: lon,
	}, nil
}
