// Nguồn chân lý chung cho vòng đời trạng thái đơn hàng.
// Dùng bởi StatusBadge, Dashboard chart, filter bảng, timeline... để nhãn + màu
// luôn nhất quán ở mọi nơi.

// Thứ tự phản ánh tiến trình vận hành (state machine).
export const STATUS_ORDER = [
  'ready_to_pick',
  'picked_up',
  'hub_inbound',
  'in_transit',
  'hub_outbound',
  'delivering',
  'delivery_failed',
  'return_requested',
  'returning',
  'return_hub',
  'delivered',
  'returned',
];

export const STATUS_MAP = {
  ready_to_pick:    { label: 'Chờ lấy hàng',          fg: 'var(--st-ready-fg)',      bg: 'var(--st-ready-bg)' },
  picked_up:        { label: 'Đã lấy hàng',           fg: 'var(--st-picked-fg)',     bg: 'var(--st-picked-bg)' },
  hub_inbound:      { label: 'Đã nhập kho',           fg: 'var(--st-inbound-fg)',    bg: 'var(--st-inbound-bg)' },
  in_transit:       { label: 'Đang trung chuyển',      fg: 'var(--st-transit-fg)',    bg: 'var(--st-transit-bg)' },
  hub_outbound:     { label: 'Đã xuất kho giao',       fg: 'var(--st-outbound-fg)',   bg: 'var(--st-outbound-bg)' },
  delivering:       { label: 'Đang giao hàng',        fg: 'var(--st-delivering-fg)', bg: 'var(--st-delivering-bg)' },
  delivery_failed:  { label: 'Giao không thành công',  fg: '#ef4444',                 bg: '#fee2e2' },
  return_requested: { label: 'Yêu cầu chuyển hoàn',   fg: '#f97316',                 bg: '#ffedd5' },
  returning:        { label: 'Đang chuyển hoàn',      fg: '#b45309',                 bg: '#fef3c7' },
  return_hub:       { label: 'Đã nhập kho hoàn',      fg: 'var(--st-inbound-fg)',    bg: 'var(--st-inbound-bg)' },
  delivered:        { label: 'Giao thành công',       fg: 'var(--st-delivered-fg)',  bg: 'var(--st-delivered-bg)' },
  returned:         { label: 'Đã hoàn trả Shop',      fg: 'var(--st-returned-fg)',   bg: 'var(--st-returned-bg)' },
  cancelled:        { label: 'Đã hủy đơn',            fg: '#dc2626',                 bg: '#fee2e2' },
};

const FALLBACK = { label: '', fg: 'var(--text-meta)', bg: 'var(--bg-subtle)' };

// Alias để các view import theo tên quen thuộc.
export const STATUS_META = STATUS_MAP;

export function statusConfig(status) {
  return STATUS_MAP[status] || { ...FALLBACK, label: status };
}

export function statusLabel(status) {
  return (STATUS_MAP[status] || FALLBACK).label || status;
}

// Trả về màu foreground (dot/bar) của trạng thái — dùng cho chart, marker.
export function statusColor(status) {
  return (STATUS_MAP[status] || FALLBACK).fg;
}

export function useStatus() {
  return { STATUS_MAP, STATUS_META, STATUS_ORDER, statusConfig, statusLabel, statusColor };
}
