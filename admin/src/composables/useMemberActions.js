// Ánh xạ (role, trạng thái hiện tại) -> các hành động chuyển trạng thái mà member
// được phép thực hiện trên giao diện. Đây là bản phản chiếu của transition + role rules
// trong backend (order_usecase.go). FE chỉ dùng để hiện đúng nút; backend vẫn là nơi
// enforce thật sự.

// Mỗi action: { to: <status đích>, label, variant }
const ACTIONS = {
  first_mile_driver: {
    ready_to_pick: [
      { to: 'picked_up', label: 'Xác nhận đã lấy hàng', variant: 'primary' },
      { to: 'returned', label: 'Không lấy được hàng / Báo hủy', variant: 'danger' }
    ],
    picked_up: [
      { to: 'hub_inbound', label: 'Bàn giao nhập kho Hub', variant: 'primary' },
      { to: 'returned', label: 'Báo sự cố / Trả hàng', variant: 'danger' }
    ],
    returning: [
      { to: 'return_hub', label: 'Bàn giao kho hoàn', variant: 'primary' }
    ]
  },
  hub_staff: {
    picked_up: [{ to: 'hub_inbound', label: 'Nhập kho (Inbound)', variant: 'primary' }],
    in_transit: [{ to: 'hub_inbound', label: 'Nhập kho từ Hub khác', variant: 'primary' }],
    hub_inbound: [
      { to: 'hub_outbound', label: 'Xuất kho giao Last-mile', variant: 'primary' },
      { to: 'in_transit', label: 'Chuyển tiếp sang Hub khác', variant: 'secondary' },
    ],
    returning: [
      { to: 'return_hub', label: 'Nhập kho hàng hoàn', variant: 'primary' }
    ],
    return_hub: [
      { to: 'returned', label: 'Xác nhận đã trả hàng cho Shop', variant: 'primary' }
    ]
  },
  last_mile_driver: {
    hub_outbound: [{ to: 'delivering', label: 'Nhận đơn đi giao', variant: 'primary' }],
    delivering: [
      { to: 'delivered', label: 'Giao thành công (Thu COD)', variant: 'primary' },
      { to: 'delivery_failed', label: 'Giao thất bại / Khách hẹn lại', variant: 'danger' },
      { to: 'hub_inbound', label: 'Giao lại về kho', variant: 'secondary' }
    ],
    delivery_failed: [
      { to: 'delivering', label: 'Tiến hành phát lại', variant: 'primary' },
      { to: 'return_requested', label: 'Yêu cầu chuyển hoàn', variant: 'danger' }
    ],
    return_requested: [
      { to: 'returning', label: 'Vận chuyển về Hub hoàn hàng', variant: 'primary' }
    ]
  },
};

export function actionsFor(role, status) {
  return ACTIONS[role]?.[status] || [];
}

