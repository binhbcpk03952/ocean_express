package usecase

import (
	"context"
	"errors"
	"fmt"
	"ocean-express-api/internal/domain"
	"ocean-express-api/pkg/geocoding"
	"strings"
	"time"

	"github.com/google/uuid"
)

type orderUseCase struct {
	orderRepo  domain.OrderRepository
	rateUC     domain.RateUseCase
	shopRepo   domain.ShopRepository
	hubRepo    domain.HubRepository
	locRepo    domain.LocationRepository
	geocoder   geocoding.Geocoder
	webhookSvc domain.WebhookService
	walletUC   domain.WalletUseCase // nil-able: test cũ truyền nil, không ghi ví
}

// NewOrderUseCase khởi tạo order use case. walletUC có thể nil trong test không cần
// vòng tiền; production truyền thật để ghi bút toán COD khi đơn delivered.
func NewOrderUseCase(orderRepo domain.OrderRepository, rateUC domain.RateUseCase, shopRepo domain.ShopRepository, hubRepo domain.HubRepository, locRepo domain.LocationRepository, geocoder geocoding.Geocoder, webhookSvc domain.WebhookService, walletUC domain.WalletUseCase) domain.OrderUseCase {
	return &orderUseCase{
		orderRepo:  orderRepo,
		rateUC:     rateUC,
		shopRepo:   shopRepo,
		hubRepo:    hubRepo,
		locRepo:    locRepo,
		geocoder:   geocoder,
		webhookSvc: webhookSvc,
		walletUC:   walletUC,
	}
}

func (u *orderUseCase) CreateOrder(ctx context.Context, shopID, receiverName, receiverPhone, receiverLocID, receiverAddress string, weight int, codAmount float64, senderLat, senderLng, receiverLat, receiverLng *float64) (*domain.ShippingOrder, error) {
	shop, err := u.shopRepo.GetByID(ctx, shopID)
	if err != nil || shop == nil {
		return nil, errors.New("không tìm thấy thông tin đối tác")
	}

	if shop.LocationID == nil {
		return nil, errors.New("đối tác chưa cấu hình khu vực gửi hàng mặc định")
	}

	fee, err := u.rateUC.CalculateFee(ctx, *shop.LocationID, receiverLocID, weight)
	if err != nil {
		return nil, fmt.Errorf("lỗi tính phí vận chuyển: %v", err)
	}

	trackingNumber := fmt.Sprintf("OE-%d", time.Now().UnixNano()/1000)
	eta := time.Now().AddDate(0, 0, 3)
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
		ShippingFee:           fee,
		CodAmount:             codAmount,
		EstimatedDeliveryTime: &eta,
		Status:                "ready_to_pick",
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

	// Trigger webhook bất đồng bộ
	u.webhookSvc.SendOrderStatus(shop.WebhookURL, domain.WebhookPayload{
		TrackingNumber: order.TrackingNumber,
		Status:         order.Status,
		Note:           logEntry.Note,
		Timestamp:      time.Now(),
	})

	return order, nil
}

func (u *orderUseCase) UpdateOrderStatus(ctx context.Context, orderID, status, note, employeeID, employeeRole, employeeHubID string, lat, lng float64) error {
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
		"ready_to_pick": {"picked_up", "returned"},
		"picked_up":     {"hub_inbound", "returned"},
		"hub_inbound":   {"in_transit", "hub_outbound"},
		"in_transit":    {"hub_inbound"},
		"hub_outbound":  {"delivering"},
		"delivering":    {"delivered", "returned", "hub_inbound"},
		"delivered":     {},
		"returned":      {},
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
			"picked_up":    {"first_mile_driver"},
			"hub_inbound":  {"hub_staff"},
			"hub_outbound": {"hub_staff"},
			"in_transit":   {"hub_staff"},
			"delivering":   {"last_mile_driver"},
			"delivered":    {"last_mile_driver"},
			// returned: mọi role vận hành (driver/hub_staff) đều có thể ghi nhận hoàn hàng
			"returned": {"first_mile_driver", "last_mile_driver", "hub_staff"},
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

	order.Status = status

	// Khi hàng nằm trong kho, không tài xế nào đang ôm đơn. Reset để last-mile driver
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

	shop, err := u.shopRepo.GetByID(ctx, order.ShopID)
	if err == nil && shop != nil {
		u.webhookSvc.SendOrderStatus(shop.WebhookURL, domain.WebhookPayload{
			TrackingNumber: order.TrackingNumber,
			Status:         order.Status,
			Note:           note,
			Timestamp:      time.Now(),
		})
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
	if err != nil {
		return nil, nil, err
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
