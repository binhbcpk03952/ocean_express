package utils

// StatusInfo chứa thông tin chi tiết và định dạng chuẩn của trạng thái vận đơn
type StatusInfo struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Label       string `json:"label"`
	Description string `json:"description"`
	BadgeColor  string `json:"badge_color"`
	BadgeBg     string `json:"badge_bg"`
	BadgeText   string `json:"badge_text"`
}

var statusMap = map[string]StatusInfo{
	"ready_to_pick": {
		Code:        "ready_to_pick",
		Name:        "Chờ lấy hàng",
		Label:       "Chờ lấy hàng",
		Description: "Đơn hàng đã được tạo thành công trên hệ thống, đang chờ tài xế tiếp nhận và đến lấy hàng.",
		BadgeColor:  "amber",
		BadgeBg:     "#fef3c7",
		BadgeText:   "#d97706",
	},
	"picked_up": {
		Code:        "picked_up",
		Name:        "Đã lấy hàng",
		Label:       "Đã lấy hàng",
		Description: "Tài xế đã lấy hàng thành công từ người gửi và đang di chuyển về bưu cục trung chuyển.",
		BadgeColor:  "blue",
		BadgeBg:     "#dbeafe",
		BadgeText:   "#1d4ed8",
	},
	"hub_inbound": {
		Code:        "hub_inbound",
		Name:        "Đã nhập kho",
		Label:       "Đã nhập kho",
		Description: "Kiện hàng đã được quét nhập kho tại bưu cục trung chuyển Ocean Express.",
		BadgeColor:  "indigo",
		BadgeBg:     "#e0e7ff",
		BadgeText:   "#4338ca",
	},
	"in_transit": {
		Code:        "in_transit",
		Name:        "Đang trung chuyển",
		Label:       "Đang trung chuyển",
		Description: "Kiện hàng đang được vận chuyển giữa các trung tâm khai thác bưu cục.",
		BadgeColor:  "purple",
		BadgeBg:     "#f3e8ff",
		BadgeText:   "#7e22ce",
	},
	"hub_outbound": {
		Code:        "hub_outbound",
		Name:        "Đã xuất kho giao",
		Label:       "Đã xuất kho giao",
		Description: "Kiện hàng đã xuất khỏi bưu cục phát và được phân chia cho tài xế giao hàng chặng cuối.",
		BadgeColor:  "sky",
		BadgeBg:     "#e0f2fe",
		BadgeText:   "#0284c7",
	},
	"delivering": {
		Code:        "delivering",
		Name:        "Đang giao hàng",
		Label:       "Đang giao hàng",
		Description: "Tài xế đang trên đường giao hàng đến địa chỉ người nhận.",
		BadgeColor:  "cyan",
		BadgeBg:     "#cffafe",
		BadgeText:   "#0891b2",
	},
	"delivered": {
		Code:        "delivered",
		Name:        "Giao thành công",
		Label:       "Giao thành công",
		Description: "Đơn hàng đã được giao thành công tới người nhận.",
		BadgeColor:  "emerald",
		BadgeBg:     "#d1fae5",
		BadgeText:   "#059669",
	},
	"delivery_failed": {
		Code:        "delivery_failed",
		Name:        "Giao không thành công",
		Label:       "Giao thất bại",
		Description: "Giao hàng chưa thành công (Khách hẹn lại / Không liên lạc được / Sai địa chỉ).",
		BadgeColor:  "rose",
		BadgeBg:     "#ffe4e6",
		BadgeText:   "#e11d48",
	},
	"return_requested": {
		Code:        "return_requested",
		Name:        "Yêu cầu chuyển hoàn",
		Label:       "Yêu cầu hoàn",
		Description: "Đơn hàng đã được xác nhận hoàn trả về cho người gửi sau các lần giao không thành công.",
		BadgeColor:  "orange",
		BadgeBg:     "#ffedd5",
		BadgeText:   "#c2410c",
	},
	"returning": {
		Code:        "returning",
		Name:        "Đang chuyển hoàn",
		Label:       "Đang hoàn hàng",
		Description: "Kiện hàng đang trên lộ trình vận chuyển hoàn trả về kho gửi.",
		BadgeColor:  "amber",
		BadgeBg:     "#fef3c7",
		BadgeText:   "#b45309",
	},
	"return_hub": {
		Code:        "return_hub",
		Name:        "Đã nhập kho hoàn",
		Label:       "Kho hoàn trả",
		Description: "Kiện hàng đã về tới bưu cục trả hàng, sẵn sàng hoàn lại cho Shop.",
		BadgeColor:  "indigo",
		BadgeBg:     "#e0e7ff",
		BadgeText:   "#4338ca",
	},
	"returned": {
		Code:        "returned",
		Name:        "Đã hoàn trả Shop",
		Label:       "Đã hoàn hàng",
		Description: "Kiện hàng đã được hoàn trả thành công về tay người gửi (Shop).",
		BadgeColor:  "slate",
		BadgeBg:     "#f1f5f9",
		BadgeText:   "#475569",
	},
	"cancelled": {
		Code:        "cancelled",
		Name:        "Đã hủy",
		Label:       "Đã hủy",
		Description: "Đơn hàng đã được hủy bỏ.",
		BadgeColor:  "red",
		BadgeBg:     "#fee2e2",
		BadgeText:   "#dc2626",
	},
}

// GetStatusInfo trả về thông tin chi tiết tên tiếng Việt, mô tả và màu sắc của trạng thái
func GetStatusInfo(code string) StatusInfo {
	if info, ok := statusMap[code]; ok {
		return info
	}
	return StatusInfo{
		Code:        code,
		Name:        code,
		Label:       code,
		Description: code,
		BadgeColor:  "slate",
		BadgeBg:     "#f1f5f9",
		BadgeText:   "#475569",
	}
}
