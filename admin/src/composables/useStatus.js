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
  ready_to_pick:    { label: 'Chờ lấy',          fg: 'var(--st-ready-fg)',      bg: 'var(--st-ready-bg)' },
  picked_up:        { label: 'Đã lấy',           fg: 'var(--st-picked-fg)',     bg: 'var(--st-picked-bg)' },
  hub_inbound:      { label: 'Nhập kho',         fg: 'var(--st-inbound-fg)',    bg: 'var(--st-inbound-bg)' },
  in_transit:       { label: 'Luân chuyển',      fg: 'var(--st-transit-fg)',    bg: 'var(--st-transit-bg)' },
  hub_outbound:     { label: 'Xuất kho',         fg: 'var(--st-outbound-fg)',   bg: 'var(--st-outbound-bg)' },
  delivering:       { label: 'Đang giao',        fg: 'var(--st-delivering-fg)', bg: 'var(--st-delivering-bg)' },
  delivery_failed:  { label: 'Giao thất bại',    fg: '#ef4444',                 bg: '#fee2e2' },
  return_requested: { label: 'Yêu cầu chuyển hoàn', fg: '#f97316',              bg: '#ffedd5' },
  returning:        { label: 'Đang hoàn về Hub', fg: 'var(--st-transit-fg)',    bg: 'var(--st-transit-bg)' },
  return_hub:       { label: 'Tại kho hàng hoàn', fg: 'var(--st-inbound-fg)',   bg: 'var(--st-inbound-bg)' },
  delivered:        { label: 'Thành công',       fg: 'var(--st-delivered-fg)',  bg: 'var(--st-delivered-bg)' },
  returned:         { label: 'Đã trả hàng cho Shop', fg: 'var(--st-returned-fg)', bg: 'var(--st-returned-bg)' },
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
