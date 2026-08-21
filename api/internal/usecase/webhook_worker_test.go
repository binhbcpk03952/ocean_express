package usecase_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"ocean-express-api/internal/domain"
	"ocean-express-api/internal/usecase"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestWebhookPayloadFormat(t *testing.T) {
	var receivedPayload domain.WebhookPayload
	var receivedHeaders http.Header
	receivedChan := make(chan struct{}, 1)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeaders = r.Header.Clone()
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &receivedPayload)
		w.WriteHeader(http.StatusOK)
		receivedChan <- struct{}{}
	}))
	defer server.Close()

	dispatcher := usecase.NewWebhookDispatcher(10)
	dispatcher.Start(2)
	defer dispatcher.Stop()

	fixedTime := time.Date(2026, 8, 20, 12, 4, 0, 0, time.UTC)
	job := domain.WebhookJob{
		EventID:        "evt-uuid-123",
		ShopID:         "shop-test",
		WebhookURL:     server.URL,
		TrackingNumber: "BCS99999",
		Status:         "ready_to_pick",
		Note:           "Đơn mới",
		SequenceID:     1,
		Timestamp:      fixedTime,
	}

	dispatcher.Dispatch(job)

	select {
	case <-receivedChan:
	case <-time.After(3 * time.Second):
		t.Fatal("Timeout waiting for webhook")
	}

	// 1. Kiểm tra Sequence ID
	if receivedPayload.SequenceID != 1 {
		t.Errorf("Expected SequenceID=1, got %d", receivedPayload.SequenceID)
	}

	// 2. Kiểm tra Timestamp theo chuẩn ISO-8601 UTC
	expectedISO := "2026-08-20T12:04:00Z"
	if receivedPayload.Timestamp != expectedISO {
		t.Errorf("Expected Timestamp=%s, got %s", expectedISO, receivedPayload.Timestamp)
	}
	if receivedPayload.CreatedAt != expectedISO {
		t.Errorf("Expected CreatedAt=%s, got %s", expectedISO, receivedPayload.CreatedAt)
	}

	// 3. Kiểm tra Epoch timestamp (milliseconds)
	if receivedPayload.TimestampEpoch != fixedTime.UnixMilli() {
		t.Errorf("Expected TimestampEpoch=%d, got %d", fixedTime.UnixMilli(), receivedPayload.TimestampEpoch)
	}

	// 4. Kiểm tra Event ID & Tracking Number
	if receivedPayload.EventID != "evt-uuid-123" {
		t.Errorf("Expected EventID=evt-uuid-123, got %s", receivedPayload.EventID)
	}
	if receivedPayload.TrackingNumber != "BCS99999" {
		t.Errorf("Expected TrackingNumber=BCS99999, got %s", receivedPayload.TrackingNumber)
	}

	// 5. Kiểm tra Headers
	if receivedHeaders.Get("Content-Type") != "application/json" {
		t.Errorf("Expected Content-Type=application/json, got %s", receivedHeaders.Get("Content-Type"))
	}
	if receivedHeaders.Get("X-Tracking-Number") != "BCS99999" {
		t.Errorf("Expected X-Tracking-Number=BCS99999, got %s", receivedHeaders.Get("X-Tracking-Number"))
	}
	if receivedHeaders.Get("X-Sequence-ID") != "1" {
		t.Errorf("Expected X-Sequence-ID=1, got %s", receivedHeaders.Get("X-Sequence-ID"))
	}
}

func TestWebhookAckOn200NoRetry(t *testing.T) {
	var callCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&callCount, 1)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}))
	defer server.Close()

	dispatcher := usecase.NewWebhookDispatcher(10)
	dispatcher.Start(2)
	defer dispatcher.Stop()

	dispatcher.Dispatch(domain.WebhookJob{
		WebhookURL:     server.URL,
		TrackingNumber: "BCS11111",
		Status:         "ready_to_pick",
		SequenceID:     1,
	})

	time.Sleep(300 * time.Millisecond)

	count := atomic.LoadInt32(&callCount)
	if count != 1 {
		t.Errorf("Expected exactly 1 call on HTTP 200, got %d calls (unnecessary retry)", count)
	}
}

func TestWebhookFIFOInOrderDelivery(t *testing.T) {
	var mu sync.Mutex
	var receivedSequences []int64
	doneChan := make(chan struct{}, 1)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload domain.WebhookPayload
		_ = json.NewDecoder(r.Body).Decode(&payload)

		mu.Lock()
		receivedSequences = append(receivedSequences, payload.SequenceID)
		if len(receivedSequences) == 3 {
			doneChan <- struct{}{}
		}
		mu.Unlock()

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	dispatcher := usecase.NewWebhookDispatcher(10)
	dispatcher.Start(3)
	defer dispatcher.Stop()

	trackingNumber := "BCS-FIFO-TEST"

	// Dispatch 3 bước liên tiếp của cùng 1 đơn hàng
	dispatcher.Dispatch(domain.WebhookJob{
		WebhookURL:     server.URL,
		TrackingNumber: trackingNumber,
		Status:         "ready_to_pick",
		SequenceID:     1,
	})
	dispatcher.Dispatch(domain.WebhookJob{
		WebhookURL:     server.URL,
		TrackingNumber: trackingNumber,
		Status:         "picked_up",
		SequenceID:     2,
	})
	dispatcher.Dispatch(domain.WebhookJob{
		WebhookURL:     server.URL,
		TrackingNumber: trackingNumber,
		Status:         "delivering",
		SequenceID:     3,
	})

	select {
	case <-doneChan:
	case <-time.After(3 * time.Second):
		t.Fatal("Timeout waiting for all 3 FIFO webhooks")
	}

	mu.Lock()
	defer mu.Unlock()

	if len(receivedSequences) != 3 {
		t.Fatalf("Expected 3 received webhooks, got %d", len(receivedSequences))
	}
	for i, expectedSeq := range []int64{1, 2, 3} {
		if receivedSequences[i] != expectedSeq {
			t.Errorf("Step %d: Expected sequence %d, got %d (Out of order delivery!)", i, expectedSeq, receivedSequences[i])
		}
	}
}
