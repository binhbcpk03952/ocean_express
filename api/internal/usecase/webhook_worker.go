package usecase

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"ocean-express-api/internal/domain"
	"ocean-express-api/pkg/utils"
	"time"
)

type webhookDispatcher struct {
	jobQueue chan domain.WebhookJob
	client   *http.Client
}

func NewWebhookDispatcher(queueSize int) domain.WebhookDispatcher {
	return &webhookDispatcher{
		jobQueue: make(chan domain.WebhookJob, queueSize),
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (w *webhookDispatcher) Dispatch(job domain.WebhookJob) {
	if job.WebhookURL == "" {
		return
	}
	w.jobQueue <- job
}

func (w *webhookDispatcher) Start(workers int) {
	for i := 0; i < workers; i++ {
		go w.worker()
	}
}

func (w *webhookDispatcher) Stop() {
	close(w.jobQueue)
}

func (w *webhookDispatcher) worker() {
	for job := range w.jobQueue {
		w.processJob(job)
	}
}

func (w *webhookDispatcher) processJob(job domain.WebhookJob) {
	st := utils.GetStatusInfo(job.Status)
	payload := domain.WebhookPayload{
		TrackingNumber:    job.TrackingNumber,
		Status:            job.Status,
		StatusName:        st.Name,
		StatusLabel:       st.Label,
		StatusDescription: st.Description,
		Note:              job.Note,
		Timestamp:         time.Now(),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Webhook encode error: %v", err)
		return
	}

	maxRetries := 3
	for attempt := 1; attempt <= maxRetries; attempt++ {
		resp, err := w.client.Post(job.WebhookURL, "application/json", bytes.NewBuffer(body))
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				log.Printf("Webhook delivered for tracking %s to %s", job.TrackingNumber, job.WebhookURL)
				return // Success
			}
			log.Printf("Webhook failed with status %d (attempt %d/%d)", resp.StatusCode, attempt, maxRetries)
		} else {
			log.Printf("Webhook request error: %v (attempt %d/%d)", err, attempt, maxRetries)
		}

		// Exponential backoff
		time.Sleep(time.Duration(attempt) * time.Second)
	}
}
