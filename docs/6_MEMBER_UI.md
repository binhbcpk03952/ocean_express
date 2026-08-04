# Giao diện Member (Driver & Hub Staff)

Tài liệu mô tả UI/UX cho 3 role vận hành ngoài admin: `first_mile_driver`, `hub_staff`, `last_mile_driver`. Web admin panel ([admin/](../admin/)) phục vụ `admin`; phần này đặc tả **giao diện member** — mobile-first vì các role này làm việc ngoài hiện trường.

Giao diện member **dùng chung codebase** với admin (cùng project [admin/](../admin/), Vue 3 + Tailwind), phân tách bằng route: vùng `/` cho admin, vùng `/m` cho member. Xem [router/index.js](../admin/src/router/index.js).

Tham chiếu: trạng thái & màu lấy từ [useStatus.js](../admin/src/composables/useStatus.js). Phân quyền theo [5_RBAC.md](./5_RBAC.md). Luồng trạng thái theo [4_STATE_MACHINE.md](./4_STATE_MACHINE.md).

## 1. Nguyên tắc thiết kế

- **Mobile-first**: layout một cột, nút bấm cao (≥ 44px), điều hướng bằng bottom tab bar. Xem [MemberLayout.vue](../admin/src/components/MemberLayout.vue).
- **Nhãn trạng thái tiếng Việt** dùng đúng `STATUS_MAP` và [StatusBadge.vue](../admin/src/components/ui/StatusBadge.vue) — đồng nhất với admin.
- **Chỉ hiện hành động hợp lệ**: nút chuyển trạng thái chỉ render khi (role + trạng thái hiện tại) cho phép. Nguồn ánh xạ: [useMemberActions.js](../admin/src/composables/useMemberActions.js), phản chiếu transition rules trong [order_usecase.go](../api/internal/usecase/order_usecase.go). Backend vẫn là nơi enforce thật.
- **Quét đơn = nhập tay tracking number** (MVP). Chưa dùng camera/barcode — ô nhập text có `autocapitalize=characters`, font mono. Có thể nâng cấp lên quét camera sau mà không đổi luồng.

## 2. Điều hướng & bố cục

Sau khi đăng nhập, `homeRoute()` trong [authStore.js](../admin/src/stores/authStore.js) điều hướng theo role:
- `admin` → `/` (admin panel).
- `first_mile_driver` / `hub_staff` / `last_mile_driver` → `/m` (MemberTasks).

[MemberLayout.vue](../admin/src/components/MemberLayout.vue): top bar (tên + role + đổi theme + đăng xuất) và bottom tab bar. Tab "Quét đơn" chỉ hiện với `hub_staff`.

**Đăng ký tài xế (self-register)**: tài xế tự đăng ký tại [RegisterShipperView.vue](../admin/src/views/register/RegisterShipperView.vue) (route công khai `/register/shipper`) — chọn role (`first_mile_driver`/`last_mile_driver`) và **tự chọn bưu cục** từ danh sách `GET /hubs` (công khai). Tài khoản tạo ở trạng thái `pending` (`is_active=false`), chưa đăng nhập được cho tới khi Admin duyệt ở [ApprovalsView.vue](../admin/src/views/ApprovalsView.vue). Chi tiết luồng duyệt: mục 6 của [5_RBAC.md](./5_RBAC.md).

| Route (name) | View | Ai truy cập |
|---|---|---|
| `/m` (`MemberTasks`) | [MemberTasksView.vue](../admin/src/views/member/MemberTasksView.vue) | 3 role member |
| `/m/orders/:id` (`MemberOrderDetail`) | [MemberOrderDetailView.vue](../admin/src/views/member/MemberOrderDetailView.vue) | 3 role member |
| `/m/scan` (`HubScan`) | [HubScanView.vue](../admin/src/views/member/HubScanView.vue) | chỉ `hub_staff` |

## 3. Màn hình Nhiệm vụ (`MemberTasks`)

Danh sách đơn dạng card, gọi `GET /orders` — backend đã lọc theo phạm vi role sẵn ([order_repo.go](../api/internal/repository/order_repo.go) `FindAll`):
- Driver (`first_mile_driver`/`last_mile_driver`): thấy đơn `current_driver_id = mình` hoặc chưa gán (`NULL`).
- `hub_staff`: thấy đơn có `current_hub_id = hub của mình`.

Mỗi card hiện mã vận đơn, `StatusBadge`, tên/SĐT người nhận, địa chỉ, COD. Bấm vào card → màn chi tiết. Có chip lọc nhanh theo các trạng thái đang tồn tại trong danh sách.

## 4. Màn hình Chi tiết đơn (`MemberOrderDetail`)

Gọi `GET /orders/:id` → hiện thông tin người nhận (SĐT bấm gọi được qua `tel:`), COD, khối lượng, và timeline hành trình từ `logs`.

Thanh hành động cố định đáy màn (trên tab bar) render các nút theo `actionsFor(role, status)`:

| Role | Trạng thái hiện tại | Nút | Transition |
|---|---|---|---|
| `first_mile_driver` | `ready_to_pick` | Xác nhận đã lấy hàng | → `picked_up` |
| `first_mile_driver` | `picked_up` | Báo hoàn / không lấy được | → `returned` |
| `last_mile_driver` | `hub_outbound` | Nhận đơn đi giao | → `delivering` |
| `last_mile_driver` | `delivering` | Giao thành công | → `delivered` |
| `last_mile_driver` | `delivering` | Giao thất bại / hoàn | → `returned` |

Mỗi nút gọi `PUT /orders/:id/status { status, note, latitude, longitude }`. Toạ độ lấy qua `navigator.geolocation` (nếu user từ chối/không hỗ trợ → gửi `0,0`, không chặn thao tác). Sau khi thành công, tải lại chi tiết để phản ánh trạng thái mới.

## 5. Màn hình Quét đơn (`HubScan`, chỉ `hub_staff`)

Vì đơn chưa từng inbound có `current_hub_id = NULL` nên **không xuất hiện** trong danh sách hub — hub staff cần tra cứu trực tiếp. Luồng:

1. Nhập mã vận đơn → `GET /orders/tracking/:tracking_number` (mọi role nội bộ tra được, xem [5_RBAC.md](./5_RBAC.md) mục 2).
2. Hiện thẻ đơn + `StatusBadge` + nút thao tác theo `actionsFor('hub_staff', status)`:

| Trạng thái hiện tại | Nút | Transition |
|---|---|---|
| `picked_up` | Nhập kho (Inbound) | → `hub_inbound` |
| `in_transit` | Nhập kho hub đích | → `hub_inbound` |
| `hub_inbound` | Xuất kho giao Last-mile | → `hub_outbound` |
| `hub_inbound` | Chuyển tiếp hub khác | → `in_transit` |

Nút gọi `PUT /orders/:id/status`. Backend gán `current_hub_id` theo hub của staff khi `hub_inbound` và chặn nếu sai luồng/sai quyền (`403`/`422`). Sau thao tác, tự tra cứu lại để cập nhật.

## 6. Xử lý lỗi

- **Không có nhiệm vụ**: empty state "Chưa có nhiệm vụ nào".
- **Tra cứu không thấy** (`HubScan`): hiện "Không tìm thấy vận đơn" kèm mã đã nhập.
- **Chuyển trạng thái bị từ chối** (`403`/`422`): toast đỏ lấy `error.message` từ backend, không đổi UI trạng thái đơn.

## 7. Ngoài phạm vi MVP (nâng cấp sau)

- Quét camera/barcode (hiện là nhập tay), quét hàng loạt (batch scan).
- Đính kèm ảnh/chữ ký khi giao thành công, chọn lý do hoàn có cấu trúc.
- Offline queue & đồng bộ, push notification (FCM — bảng `employee_devices` đã có sẵn trong schema).
- Tách thành PWA/app riêng nếu cần cài đặt lên thiết bị.
