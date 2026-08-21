package domain

import "time"

type WebhookJob struct {
	EventID        string
	ShopID         string
	WebhookURL     string
	TrackingNumber string
	Status         string
	Note           string
	SequenceID     int64
	Timestamp      time.Time
}

// WebhookDispatcher interface
type WebhookDispatcher interface {
	Dispatch(job WebhookJob)
	Start(workers int)
	Stop()
}
