import re

with open("api/internal/usecase/order_usecase.go", "r", encoding="utf-8") as f:
    content = f.read()

# 1. Update struct and NewOrderUseCase
content = re.sub(
    r"type orderUseCase struct {[\s\S]*?}",
    """type orderUseCase struct {
	orderRepo         domain.OrderRepository
	rateUC            domain.RateUseCase
	shopRepo          domain.ShopRepository
	hubRepo           domain.HubRepository
	locRepo           domain.LocationRepository
	geocoder          geocoding.Geocoder
	webhookSvc        domain.WebhookService
	walletUC          domain.WalletUseCase
	auditUC           domain.AuditUseCase
	webhookDispatcher domain.WebhookDispatcher
}

func NewOrderUseCase(orderRepo domain.OrderRepository, rateUC domain.RateUseCase, shopRepo domain.ShopRepository, hubRepo domain.HubRepository, locRepo domain.LocationRepository, geocoder geocoding.Geocoder, webhookSvc domain.WebhookService, walletUC domain.WalletUseCase, auditUC domain.AuditUseCase, webhookDispatcher domain.WebhookDispatcher) domain.OrderUseCase {
	return &orderUseCase{
		orderRepo:         orderRepo,
		rateUC:            rateUC,
		shopRepo:          shopRepo,
		hubRepo:           hubRepo,
		locRepo:           locRepo,
		geocoder:          geocoder,
		webhookSvc:        webhookSvc,
		walletUC:          walletUC,
		auditUC:           auditUC,
		webhookDispatcher: webhookDispatcher,
	}
}

func (u *orderUseCase) CreateOrder""",
    content,
    flags=re.MULTILINE
)

# 2. SLA in CreateOrder
content = content.replace(
    'eta := time.Now().AddDate(0, 0, 3)',
    'eta := time.Now().AddDate(0, 0, 3)\n\tsla := time.Now().Add(48 * time.Hour)'
)
content = content.replace(
    'Status:                "ready_to_pick",',
    'Status:                "ready_to_pick",\n\t\tSlaDeadline:           &sla,'
)

# 3. Webhook in CreateOrder
content = re.sub(
    r"u\.webhookSvc\.SendOrderStatus\(shop\.WebhookURL, domain\.WebhookPayload{[\s\S]*?}\)",
    """if u.webhookDispatcher != nil {
		u.webhookDispatcher.Dispatch(domain.WebhookJob{
			ShopID:         shop.ID,
			WebhookURL:     shop.WebhookURL,
			TrackingNumber: order.TrackingNumber,
			Status:         order.Status,
			Note:           logEntry.Note,
		})
	}""",
    content
)

# 4. State transitions
content = content.replace(
    '''	// 2. State Machine Validation
	validTransitions := map[string][]string{
		"ready_to_pick": {"picked_up", "returned"},
		"picked_up":     {"hub_inbound", "returned"},
		"hub_inbound":   {"in_transit", "hub_outbound"},
		"in_transit":    {"hub_inbound"},
		"hub_outbound":  {"delivering"},
		"delivering":    {"delivered", "returned", "hub_inbound"},
		"delivered":     {},
		"returned":      {},
	}''',
    '''	// 2. State Machine Validation
	validTransitions := map[string][]string{
		"ready_to_pick":    {"picked_up", "returned"},
		"picked_up":        {"hub_inbound", "returned"},
		"hub_inbound":      {"in_transit", "hub_outbound"},
		"in_transit":       {"hub_inbound"},
		"hub_outbound":     {"delivering"},
		"delivering":       {"delivered", "delivery_failed", "returned", "hub_inbound"},
		"delivery_failed":  {"delivering", "return_requested"},
		"return_requested": {"returning"},
		"returning":        {"return_hub"},
		"return_hub":       {"returned"},
		"delivered":        {},
		"returned":         {},
	}'''
)

# 5. Roles
content = content.replace(
    '''		allowedRoles := map[string][]string{
			"picked_up":    {"first_mile_driver"},
			"hub_inbound":  {"hub_staff"},
			"hub_outbound": {"hub_staff"},
			"in_transit":   {"hub_staff"},
			"delivering":   {"last_mile_driver"},
			"delivered":    {"last_mile_driver"},
			// returned: mọi role vận hành (driver/hub_staff) đều có thể ghi nhận hoàn hàng
			"returned": {"first_mile_driver", "last_mile_driver", "hub_staff"},
		}''',
    '''		allowedRoles := map[string][]string{
			"picked_up":        {"first_mile_driver"},
			"hub_inbound":      {"hub_staff"},
			"hub_outbound":     {"hub_staff"},
			"in_transit":       {"hub_staff"},
			"delivering":       {"last_mile_driver"},
			"delivery_failed":  {"last_mile_driver"},
			"return_requested": {"last_mile_driver", "admin"},
			"returning":        {"last_mile_driver", "first_mile_driver"},
			"return_hub":       {"hub_staff"},
			"delivered":        {"last_mile_driver"},
			"returned":         {"first_mile_driver", "last_mile_driver", "hub_staff"},
		}'''
)

# 6. Delivery Failed logic
content = content.replace(
    'order.Status = status',
    '''oldStatus := order.Status

	if status == "delivery_failed" {
		order.DeliveryAttempts++
		if failureReason != "" {
			order.FailureReason = &failureReason
		}
		if order.DeliveryAttempts >= 3 {
			status = "return_requested"
			note = "Giao thất bại 3 lần, tự động chuyển hoàn."
		}
	}

	order.Status = status'''
)

# 7. Audit Log and Webhook
content = re.sub(
    r"shop, err := u\.shopRepo\.GetByID\(ctx, order\.ShopID\)[\s\S]*?return nil\n}",
    """if u.auditUC != nil {
		u.auditUC.LogAction(ctx, employeeID, "update_status", "shipping_orders", order.ID, `{"status": "`+oldStatus+`"}`, `{"status": "`+status+`"}`)
	}

	shop, err := u.shopRepo.GetByID(ctx, order.ShopID)
	if err == nil && shop != nil {
		if u.webhookDispatcher != nil {
			u.webhookDispatcher.Dispatch(domain.WebhookJob{
				ShopID:         shop.ID,
				WebhookURL:     shop.WebhookURL,
				TrackingNumber: order.TrackingNumber,
				Status:         order.Status,
				Note:           note,
			})
		}
	}

	return nil
}""",
    content
)

with open("api/internal/usecase/order_usecase.go", "w", encoding="utf-8") as f:
    f.write(content)
