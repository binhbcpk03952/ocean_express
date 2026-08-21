package usecase

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"ocean-express-api/internal/domain"
	"ocean-express-api/pkg/utils"
	"strconv"
	"sync"
	"time"
)

type webhookDispatcher struct {
	mu           sync.RWMutex
	workerQueues []chan domain.WebhookJob
	queueSize    int
	workers      int
	client       *http.Client
	started      bool
	stopped      bool
}

func hashKey(s string) uint32 {
	var h uint32 = 2166136261
	for i := 0; i < len(s); i++ {
		h *= 16777619
		h ^= uint32(s[i])
	}
	return h
}

// NewWebhookDispatcher khởi tạo Webhook Dispatcher
func NewWebhookDispatcher(queueSize int) domain.WebhookDispatcher {
	if queueSize <= 0 {
		queueSize = 100
	}
	return &webhookDispatcher{
		queueSize: queueSize,
		client: &http.Client{
			Timeout: 10 * time.Second, // Chuẩn hóa timeout 10 giây
		},
	}
}

// Start khởi chạy worker pools với Partitioned Queues theo hash của TrackingNumber
func (w *webhookDispatcher) Start(workers int) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.started {
		return
	}
	if workers <= 0 {
		workers = 5
	}
	w.workers = workers
	w.workerQueues = make([]chan domain.WebhookJob, workers)
	perQueueSize := w.queueSize / workers
	if perQueueSize < 10 {
		perQueueSize = 10
	}

	for i := 0; i < workers; i++ {
		w.workerQueues[i] = make(chan domain.WebhookJob, perQueueSize)
		go w.worker(w.workerQueues[i])
	}
	w.started = true
}

// Stop dừng toàn bộ hàng đợi webhook
func (w *webhookDispatcher) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.stopped || !w.started {
		return
	}
	w.stopped = true
	for _, q := range w.workerQueues {
		close(q)
	}
}

// Dispatch phân phối job vào worker channel tương ứng theo TrackingNumber để đảm bảo thứ tự FIFO
func (w *webhookDispatcher) Dispatch(job domain.WebhookJob) {
	if job.WebhookURL == "" {
		return
	}

	w.mu.RLock()
	defer w.mu.RUnlock()

	if w.stopped || len(w.workerQueues) == 0 {
		// Nếu chưa Start hoặc đã dừng, fallback gửi async trực tiếp
		go w.processJob(job)
		return
	}

	// Consistent Partitioning: Tất cả webhook của cùng một TrackingNumber luôn vào cùng 1 worker
	idx := int(hashKey(job.TrackingNumber) % uint32(len(w.workerQueues)))
	select {
	case w.workerQueues[idx] <- job:
	default:
		// Tránh block hệ thống nếu hàng đợi riêng bị đầy
		go w.processJob(job)
	}
}

func (w *webhookDispatcher) worker(q <-chan domain.WebhookJob) {
	for job := range q {
		w.processJob(job)
	}
}

func (w *webhookDispatcher) processJob(job domain.WebhookJob) {
	st := utils.GetStatusInfo(job.Status)

	eventTime := job.Timestamp
	if eventTime.IsZero() {
		eventTime = time.Now().UTC()
	} else {
		eventTime = eventTime.UTC()
	}
	isoTime := eventTime.Format(time.RFC3339) // Chuẩn ISO-8601 UTC: e.g. "2026-08-20T12:04:00Z"

	payload := domain.WebhookPayload{
		EventID:           job.EventID,
		TrackingNumber:    job.TrackingNumber,
		Status:            job.Status,
		StatusName:        st.Name,
		StatusLabel:       st.Label,
		StatusDescription: st.Description,
		Note:              job.Note,
		SequenceID:        job.SequenceID,
		Timestamp:         isoTime,
		CreatedAt:         isoTime,
		TimestampEpoch:    eventTime.UnixMilli(),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		log.Printf("[Webhook] Marshal JSON lỗi đơn %s: %v", job.TrackingNumber, err)
		return
	}

	maxRetries := 3
	for attempt := 1; attempt <= maxRetries; attempt++ {
		req, err := http.NewRequest("POST", job.WebhookURL, bytes.NewBuffer(body))
		if err != nil {
			log.Printf("[Webhook] Lỗi tạo Request đơn %s: %v", job.TrackingNumber, err)
			return
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "OceanExpress-Webhook/1.0")
		if job.EventID != "" {
			req.Header.Set("X-Event-ID", job.EventID)
		}
		req.Header.Set("X-Tracking-Number", job.TrackingNumber)
		if job.SequenceID > 0 {
			req.Header.Set("X-Sequence-ID", strconv.FormatInt(job.SequenceID, 10))
		}

		resp, err := w.client.Do(req)
		if err == nil {
			// Drain body và close để tái sử dụng TCP connection an toàn
			_, _ = io.Copy(io.Discard, resp.Body)
			resp.Body.Close()

			// Thành công: Shop phản hồi HTTP 2xx (200, 201, 204...) -> dừng retry ngay
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				log.Printf("[Webhook] Thành công: %s (Đơn %s, Seq %d, Status %s) - HTTP %d",
					job.WebhookURL, job.TrackingNumber, job.SequenceID, job.Status, resp.StatusCode)
				return
			}

			// Thất bại: Shop phản hồi mã lỗi 4xx / 5xx
			log.Printf("[Webhook] Thất bại: %s (Đơn %s, Seq %d) - HTTP %d (Lần thử %d/%d)",
				job.WebhookURL, job.TrackingNumber, job.SequenceID, resp.StatusCode, attempt, maxRetries)
		} else {
			// Lỗi mạng hoặc quá thời gian timeout 10s
			log.Printf("[Webhook] Lỗi kết nối/timeout tới %s (Đơn %s, Seq %d): %v (Lần thử %d/%d)",
				job.WebhookURL, job.TrackingNumber, job.SequenceID, err, attempt, maxRetries)
		}

		// Áp dụng backoff trước khi thử lại
		if attempt < maxRetries {
			time.Sleep(time.Duration(attempt*2) * time.Second)
		}
	}
}
