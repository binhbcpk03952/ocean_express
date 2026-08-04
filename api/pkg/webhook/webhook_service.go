package webhook

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"ocean-express-api/internal/domain"
	"time"
)

type webhookService struct {
	client *http.Client
}

func NewWebhookService() domain.WebhookService {
	return &webhookService{
		client: &http.Client{
			Timeout: 5 * time.Second, // Timeout 5s để không treo goroutine
		},
	}
}

// SendOrderStatus bắn payload đi bằng Goroutine
func (s *webhookService) SendOrderStatus(url string, payload domain.WebhookPayload) {
	// Bắn Fire-and-Forget
	go func() {
		if url == "" {
			return
		}

		jsonData, err := json.Marshal(payload)
		if err != nil {
			log.Printf("[Webhook] Marshal lỗi: %v", err)
			return
		}

		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
		if err != nil {
			log.Printf("[Webhook] Tạo Request lỗi: %v", err)
			return
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := s.client.Do(req)
		if err != nil {
			log.Printf("[Webhook] Bắn tới %s lỗi: %v", url, err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			log.Printf("[Webhook] Thành công: %s (Trạng thái: %s)", url, payload.Status)
		} else {
			log.Printf("[Webhook] %s phản hồi mã lỗi %d", url, resp.StatusCode)
		}
	}()
}
