package usecase

import (
	"context"
	"errors"
	"fmt"
	"ocean-express-api/internal/domain"
	"ocean-express-api/pkg/geocoding"
	"ocean-express-api/pkg/utils"
	"strings"
	"time"

	"github.com/google/uuid"
)

type orderUseCase struct {
	orderRepo         domain.OrderRepository
	rateUC            domain.RateUseCase
	shopRepo          domain.ShopRepository
	hubRepo           domain.HubRepository
	locRepo           domain.LocationRepository
	geocoder          geocoding.Geocoder
	webhookSvc        domain.WebhookService
	walletUC          domain.WalletUseCase
	webhookDispatcher domain.WebhookDispatcher
	notifUC           domain.NotificationUseCase
	auditUC           domain.AuditUseCase
}

func NewOrderUseCase(orderRepo domain.OrderRepository, rateUC domain.RateUseCase, shopRepo domain.ShopRepository, hubRepo domain.HubRepository, locRepo domain.LocationRepository, geocoder geocoding.Geocoder, webhookSvc domain.WebhookService, walletUC domain.WalletUseCase, auditUC domain.AuditUseCase, webhookDispatcher domain.WebhookDispatcher, notifUC domain.NotificationUseCase) domain.OrderUseCase {
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
		notifUC:           notifUC,
	}
}

func (u *orderUseCase) CreateOrder(ctx context.Context, shopID, receiverName, receiverPhone, receiverLocID, receiverAddress string, weight, length, width, height int, codAmount float64, senderLat, senderLng, receiverLat, receiverLng *float64) (*domain.ShippingOrder, error) {
	shop, err := u.shopRepo.GetByID(ctx, shopID)
	if err != nil || shop == nil {
		return nil, errors.New("không tìm thấy thông tin đối tác")
	}

	if shop.LocationID == nil {
		return nil, errors.New("đối tác chưa cấu hình khu vực gửi hàng mặc định")
	}

	chargeableWeight := weight
	if length > 0 && width > 0 && height > 0 {
		volWeight := (length * width * height) / 5
		if volWeight > chargeableWeight {
			chargeableWeight = volWeight
		}
	}

	fee, err := u.rateUC.CalculateFee(ctx, *shop.LocationID, receiverLocID, chargeableWeight)
	if err != nil {
		return nil, fmt.Errorf("lỗi tính phí vận chuyển: %v", err)
	}

	trackingNumber := fmt.Sprintf("OE-%d", time.Now().UnixNano()/1000)
	eta := time.Now().AddDate(0, 0, 3)
	sla := time.Now().Add(48 * time.Hour)
	orderID := uuid.New().String()

	order := &domain.ShippingOrder{
		ID:                    orderID,
		TrackingNumber:        trackingNumber,
		ShopID:                shop.ID,
		SenderPhone:           shop.Phone,
		SenderLocationID:      shop.LocationID,
		SenderAddressDetail:   shop.AddressDetail,
		ReceiverName:          receiverName,
		ReceiverPhone:         receiverPhone,
		ReceiverLocationID:    &receiverLocID,
		ReceiverAddressDetail: receiverAddress,
		Weight:                weight,
		Length:                length,
		Width:                 width,
		Height:                height,
		ShippingFee:           fee,
		CodAmount:             codAmount,
		EstimatedDeliveryTime: &eta,
		SlaDeadline:           &sla,
		Status:                "ready_to_pick",
		StatusName:            utils.GetStatusInfo("ready_to_pick").Name,
		StatusLabel:           utils.GetStatusInfo("ready_to_pick").Label,
		StatusDescription:     utils.GetStatusInfo("ready_to_pick").Description,
	}

	// 1. Geocoding Sender Address
	if senderLat != nil && senderLng != nil {
		order.SenderLatitude = senderLat
		order.SenderLongitude = senderLng
		if hub, err := u.hubRepo.FindNearestHub(ctx, *senderLat, *senderLng); err == nil {
			order.PickupHubID = &hub.ID
		}
	} else {
		senderGeoAddr := u.buildFullAddress(ctx, shop.AddressDetail, shop.LocationID)
		if senderGeoAddr != "" {
			if senderCoords, err := u.geocoder.GetCoordinates(senderGeoAddr); err == nil {
				order.SenderLatitude = &senderCoords.Latitude
				order.SenderLongitude = &senderCoords.Longitude
				
				// Find nearest Pickup Hub
				if hub, err := u.hubRepo.FindNearestHub(ctx, senderCoords.Latitude, senderCoords.Longitude); err == nil {
					order.PickupHubID = &hub.ID
				}
			}
		}
	}
	
	// Fallback nếu không có PickupHubID
	if order.PickupHubID == nil && shop.LocationID != nil {
		if hubID := u.findHubByLocationHierarchy(ctx, *shop.LocationID); hubID != nil {
			order.PickupHubID = hubID
		}
	}

	// 2. Geocoding Receiver Address
	if receiverLat != nil && receiverLng != nil {
		order.ReceiverLatitude = receiverLat
		order.ReceiverLongitude = receiverLng
		if hub, err := u.hubRepo.FindNearestHub(ctx, *receiverLat, *receiverLng); err == nil {
			order.DeliveryHubID = &hub.ID
		}
	} else {
		var receiverLocIDPtr *string
		if receiverLocID != "" {
			receiverLocIDPtr = &receiverLocID
		}
		receiverGeoAddr := u.buildFullAddress(ctx, receiverAddress, receiverLocIDPtr)
		if receiverGeoAddr != "" {
			if receiverCoords, err := u.geocoder.GetCoordinates(receiverGeoAddr); err == nil {
				order.ReceiverLatitude = &receiverCoords.Latitude
				order.ReceiverLongitude = &receiverCoords.Longitude
				
				// Find nearest Delivery Hub
				if hub, err := u.hubRepo.FindNearestHub(ctx, receiverCoords.Latitude, receiverCoords.Longitude); err == nil {
					order.DeliveryHubID = &hub.ID
				}
			}
		}
	}
	// Fallback nếu không có DeliveryHubID
	if order.DeliveryHubID == nil && receiverLocID != "" {
		if hubID := u.findHubByLocationHierarchy(ctx, receiverLocID); hubID != nil {
			order.DeliveryHubID = hubID
		}
	}

	logEntry := &domain.TrackingLog{
		OrderID: order.ID,
		Status:  "ready_to_pick",
		Note:    "Đơn hàng được tạo bởi Shop",
	}

	err = u.orderRepo.CreateOrderWithLog(ctx, order, logEntry)
	if err != nil {
		return nil, errors.New("lỗi khi lưu vận đơn vào hệ thống")
	}

	// Trigger webhook qua dispatcher
	if u.webhookDispatcher != nil {
		u.webhookDispatcher.Dispatch(domain.WebhookJob{
			ShopID:         shop.ID,
			WebhookURL:     shop.WebhookURL,
			TrackingNumber: order.TrackingNumber,
			Status:         order.Status,
			Note:           logEntry.Note,
		})
	}

	// Ghi nhận Audit Log cho hành động tạo đơn
	if u.auditUC != nil {
		u.auditUC.LogAction(ctx, shopID, "create_order", "shipping_orders", order.ID, "{}", `{"status": "ready_to_pick", "tracking_number": "`+order.TrackingNumber+`"}`)
	}

	return order, nil
}

func (u *orderUseCase) UpdateOrderStatus(ctx context.Context, orderID, status, note, failureReason, employeeID, employeeRole, employeeHubID string, lat, lng float64) error {
	order, err := u.orderRepo.GetByID(ctx, orderID)
	if err != nil || order == nil {
		return fmt.Errorf("%w: không tìm thấy vận đơn", domain.ErrNotFound)
	}

	// 1. Phân quyền Shipper: chỉ được thao tác trên đơn của mình
	if employeeRole == "first_mile_driver" || employeeRole == "last_mile_driver" {
		if order.CurrentDriverID != nil && *order.CurrentDriverID != employeeID {
			return fmt.Errorf("%w: đơn hàng này đang do người khác phụ trách", domain.ErrForbidden)
		}
		// Tự động gán nếu chưa có (khi lấy hàng / giao hàng)
		if order.CurrentDriverID == nil {
			order.CurrentDriverID = &employeeID
		}
	}

	// 2. State Machine Validation
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
	}

	allowed, ok := validTransitions[order.Status]
	if !ok {
		return fmt.Errorf("%w: trạng thái hiện tại '%s' không hợp lệ", domain.ErrInvalidTransition, order.Status)
	}

	isValidTransition := false
	for _, allowedStatus := range allowed {
		if status == allowedStatus {
			isValidTransition = true
			break
		}
	}

	if !isValidTransition {
		return fmt.Errorf("%w: không thể chuyển từ %s sang %s", domain.ErrInvalidTransition, order.Status, status)
	}

	// 3. Phân quyền theo hành động: mỗi trạng thái đích chỉ cho phép một số role thực hiện.
	// Admin được bỏ qua mọi ràng buộc role.
	if employeeRole != "admin" {
		allowedRoles := map[string][]string{
			"picked_up":        {"first_mile_driver"},
			"hub_inbound":      {"hub_staff", "first_mile_driver", "last_mile_driver"},
			"hub_outbound":     {"hub_staff"},
			"in_transit":       {"hub_staff"},
			"delivering":       {"last_mile_driver"},
			"delivery_failed":  {"last_mile_driver"},
			"return_requested": {"last_mile_driver", "admin"},
			"returning":        {"last_mile_driver", "first_mile_driver", "hub_staff"},
			"return_hub":       {"hub_staff"},
			"delivered":        {"last_mile_driver"},
			"returned":         {"first_mile_driver", "last_mile_driver", "hub_staff"},
		}

		if roles, exists := allowedRoles[status]; exists {
			permitted := false
			for _, r := range roles {
				if employeeRole == r {
					permitted = true
					break
				}
			}
			if !permitted {
				return fmt.Errorf("%w: role %s không được phép chuyển trạng thái sang %s", domain.ErrForbidden, employeeRole, status)
			}
		}
	}

	oldStatus := order.Status

	if status == "delivery_failed" {
		order.DeliveryAttempts++
		if failureReason != "" {
			order.FailureReason = &failureReason
			note = fmt.Sprintf("Lý do: %s. %s", failureReason, note)
		}
		if order.DeliveryAttempts >= 3 {
			status = "return_requested"
			note = "Giao thất bại 3 lần, tự động chuyển hoàn. " + note
		}
	}

	// Kiểm tra SLA Breach
	if order.SlaDeadline != nil && time.Now().After(*order.SlaDeadline) && !order.SlaBreached {
		// Chưa hoàn thành giao hàng hoặc trả hàng thì vi phạm
		if status != "delivered" && status != "returned" {
			order.SlaBreached = true
		}
	}

	// Tính phí hoàn hàng khi đơn chuyển sang returned
	if status == "returned" && oldStatus != "returned" {
		order.ReturnFee = order.ShippingFee * 0.5
	}

	order.Status = status

	// Khi hàng nằm trong kho, không tài xế nào đang ôm đơn. Reset để last-mile driver
	// hoặc hub staff có thể quét tiếp.
	if status == "hub_received" || status == "return_hub" || status == "returned" {
		order.CurrentDriverID = nil
	}
	// có thể tự nhận đơn ở bước delivering (nếu không sẽ vẫn bị gán cho first-mile driver cũ).
	if status == "hub_inbound" || status == "hub_outbound" {
		order.CurrentDriverID = nil
	}

	// Ghi nhận đơn đang nằm ở bưu cục nào (theo hub của staff quét nhập kho),
	// phục vụ lọc đơn theo hub cho Hub Staff.
	if status == "hub_inbound" && employeeHubID != "" {
		order.CurrentHubID = &employeeHubID
	}

	var empIDPtr *string
	if employeeID != "" {
		empIDPtr = &employeeID
	}

	logEntry := &domain.TrackingLog{
		OrderID:    order.ID,
		Status:     status,
		Note:       note,
		EmployeeID: empIDPtr,
		Latitude:   lat,
		Longitude:  lng,
	}

	err = u.orderRepo.UpdateStatus(ctx, order, logEntry)
	if err != nil {
		return fmt.Errorf("lỗi cập nhật trạng thái: %v", err)
	}

	// Khi giao thành công: ghi bút toán COD vào ví shop (+COD thu hộ, -cước).
	// Idempotent theo đơn trong wallet usecase. Không chặn luồng nếu ghi lỗi —
	// chỉ log, để việc cập nhật trạng thái vẫn thành công.
	if status == "delivered" && u.walletUC != nil {
		if err := u.walletUC.RecordCOD(ctx, order); err != nil {
			fmt.Printf("cảnh báo: không ghi được bút toán COD cho đơn %s: %v\n", order.ID, err)
		}
	}

	// Khi trả hàng thành công: ghi bút toán phí trả hàng vào ví shop (-cước hoàn hàng).
	if status == "returned" && u.walletUC != nil {
		if err := u.walletUC.RecordReturnFee(ctx, order); err != nil {
			fmt.Printf("cảnh báo: không ghi được bút toán Return Fee cho đơn %s: %v\n", order.ID, err)
		}
	}

	// Ghi nhận Audit Log
	if u.auditUC != nil {
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
		if u.notifUC != nil {
			title := fmt.Sprintf("Đơn hàng %s thay đổi trạng thái", order.TrackingNumber)
			message := fmt.Sprintf("Trạng thái mới: %s. %s", order.Status, note)
			_ = u.notifUC.CreateNotification(ctx, shop.ID, title, message, "order_update")
		}
	}

	return nil
}

func (u *orderUseCase) GetOrders(ctx context.Context, role, employeeID, hubID string, pageParams domain.PaginationParams) (*domain.PaginatedResponse, error) {
	orders, total, err := u.orderRepo.FindAll(ctx, role, employeeID, hubID, pageParams)
	if err != nil {
		return nil, errors.New("không thể lấy danh sách vận đơn")
	}

	return &domain.PaginatedResponse{
		Data: orders,
		Meta: domain.PaginationMeta{
			Page:       pageParams.Page,
			Limit:      pageParams.Limit,
			TotalItems: total,
			TotalPages: domain.CalculateTotalPages(total, pageParams.GetLimit()),
		},
	}, nil
}

func (u *orderUseCase) GetOrderDetails(ctx context.Context, id string) (*domain.ShippingOrder, []*domain.TrackingLog, error) {
	order, err := u.orderRepo.GetByID(ctx, id)
	if err != nil || order == nil {
		// Fallback: nếu id truyền vào là tracking_number (dạng OE-...)
		var errTracking error
		order, errTracking = u.orderRepo.GetByTrackingNumber(ctx, id)
		if errTracking != nil || order == nil {
			if err != nil {
				return nil, nil, err
			}
			return nil, nil, errTracking
		}
	}
	logs, err := u.orderRepo.GetOrderLogs(ctx, order.ID)
	if err != nil {
		return nil, nil, err
	}
	return order, logs, nil
}

func (u *orderUseCase) GetOrderDetailsByTrackingNumber(ctx context.Context, trackingNumber string) (*domain.ShippingOrder, []*domain.TrackingLog, error) {
	order, err := u.orderRepo.GetByTrackingNumber(ctx, trackingNumber)
	if err != nil {
		return nil, nil, err
	}
	logs, err := u.orderRepo.GetOrderLogs(ctx, order.ID)
	if err != nil {
		return nil, nil, err
	}
	return order, logs, nil
}

func (u *orderUseCase) SubmitCOD(ctx context.Context, driverID string) (float64, error) {
	return u.orderRepo.SubmitCOD(ctx, driverID)
}

func (u *orderUseCase) AssignOrder(ctx context.Context, orderID, shipperID, assignerID, role string) error {
	if role != "admin" && role != "hub_staff" {
		return fmt.Errorf("%w: chỉ admin hoặc điều phối viên mới được phép gán đơn", domain.ErrForbidden)
	}

	order, err := u.orderRepo.GetByID(ctx, orderID)
	if err != nil || order == nil {
		return fmt.Errorf("%w: không tìm thấy vận đơn", domain.ErrNotFound)
	}

	// Verify shipper exists
	// Ideally we would fetch the employee here, but for now we just assign
	order.CurrentDriverID = &shipperID

	logEntry := &domain.TrackingLog{
		OrderID:    order.ID,
		Status:     order.Status, // status doesn't change
		Note:       "Đơn hàng được phân công cho nhân viên giao nhận",
		EmployeeID: &assignerID,
	}

	err = u.orderRepo.UpdateStatus(ctx, order, logEntry)
	if err != nil {
		return fmt.Errorf("lỗi gán đơn hàng: %v", err)
	}

	if u.auditUC != nil {
		u.auditUC.LogAction(ctx, assignerID, "assign_order", "shipping_orders", order.ID, "{}", `{"assigned_to": "`+shipperID+`"}`)
	}

	return nil
}

// buildFullAddress ghép địa chỉ chi tiết với tên đơn vị hành chính (xã/huyện/tỉnh) để tăng độ chính xác geocoding.
// Ví dụ: "Hà Huy Tập" + locationID của Xã tại Đắk Lắk → "Hà Huy Tập, Buôn Ma Thuột, Đắk Lắk"
func (u *orderUseCase) buildFullAddress(ctx context.Context, addressDetail string, locationID *string) string {
	if addressDetail == "" && locationID == nil {
		return ""
	}

	var parts []string
	if addressDetail != "" {
		parts = append(parts, addressDetail)
	}

	if locationID != nil {
		// Dùng đệ qui để lấy toàn bộ chuỗi phân cấp: ward → district → province
		currentID := *locationID
		for i := 0; i < 3; i++ { // Tối đa 3 cấp: xã, huyện, tỉnh
			loc, err := u.locRepo.GetByID(ctx, currentID)
			if err != nil || loc == nil {
				break
			}
			parts = append(parts, loc.Name)
			if loc.ParentID == nil {
				break
			}
			currentID = *loc.ParentID
		}
	}

	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, ", ")
}

func (u *orderUseCase) findHubByLocationHierarchy(ctx context.Context, startLocID string) *string {
	currentLocID := &startLocID
	for currentLocID != nil {
		hubs, err := u.hubRepo.FindAll(ctx, currentLocID)
		if err == nil && len(hubs) > 0 {
			return &hubs[0].ID
		}
		
		// Move up the hierarchy
		loc, err := u.locRepo.GetByID(ctx, *currentLocID)
		if err != nil || loc == nil || loc.ParentID == nil {
			break
		}
		currentLocID = loc.ParentID
	}
	return nil
}
