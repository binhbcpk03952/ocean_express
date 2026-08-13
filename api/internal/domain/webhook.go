package domain

type WebhookJob struct {
	ShopID         string
	WebhookURL     string
	TrackingNumber string
	Status         string
	Note           string
}

// WebhookDispatcher interface
type WebhookDispatcher interface {
	Dispatch(job WebhookJob)
	Start(workers int)
	Stop()
}
