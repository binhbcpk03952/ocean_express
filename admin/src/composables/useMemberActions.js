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
      { to: 'returned', label: 'Báo sự cố / Trả hàng', variant: 'danger' }
    ],
  },
  hub_staff: {
    picked_up: [{ to: 'hub_inbound', label: 'Nhập kho (Inbound)', variant: 'primary' }],
    in_transit: [{ to: 'hub_inbound', label: 'Nhập kho từ Hub khác', variant: 'primary' }],
    hub_inbound: [
      { to: 'hub_outbound', label: 'Xuất kho giao Last-mile', variant: 'primary' },
      { to: 'in_transit', label: 'Chuyển tiếp sang Hub khác', variant: 'secondary' },
    ],
  },
  last_mile_driver: {
    hub_outbound: [{ to: 'delivering', label: 'Nhận đơn đi giao', variant: 'primary' }],
    delivering: [
      { to: 'delivered', label: 'Giao thành công', variant: 'primary' },
      { to: 'returned', label: 'Giao thất bại / Hoàn hàng', variant: 'danger' },
      { to: 'hub_inbound', label: 'Giao lại về kho', variant: 'secondary' }
    ],
  },
};

export function actionsFor(role, status) {
  return ACTIONS[role]?.[status] || [];
}
