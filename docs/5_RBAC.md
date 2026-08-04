# Phân quyền (RBAC - Role Based Access Control)

Tài liệu này đặc tả chi tiết cơ chế phân quyền, bổ sung cho phần mô tả khái niệm ở [1_BRD.md](./1_BRD.md) và ràng buộc chuyển trạng thái ở [4_STATE_MACHINE.md](./4_STATE_MACHINE.md).

## 1. Các Vai trò

| Role (enum) | Tên gọi | Phạm vi làm việc | Nền tảng chính |
|---|---|---|---|
| `admin` | Quản trị viên | Toàn hệ thống | Web (Admin panel) — vùng `/` |
| `first_mile_driver` | Tài xế lấy hàng | Đơn được gán cho mình | Web member (mobile-first) — vùng `/m` |
| `hub_staff` | Nhân viên kho | Đơn vật lý tại Hub của mình (`hub_id`) | Web member (mobile-first) — vùng `/m` |
| `last_mile_driver` | Tài xế giao hàng | Đơn được gán cho mình | Web member (mobile-first) — vùng `/m` |
| `shop` | Đối tác gửi hàng | Đơn của chính shop (`shop_id`) | Web shop portal — vùng `/shop` |

`shop` **không phải** employee — là tài khoản portal của đối tác, đăng nhập bằng **email + mật khẩu** (`POST /auth/shop/login`), dùng chung hạ tầng JWT + session Redis nhưng lưu ở bảng `shops`. Xem [shop_auth_usecase.go](../api/internal/usecase/shop_auth_usecase.go).

Nguyên tắc nền tảng:
- **Data scoping (giới hạn dữ liệu)**: Driver chỉ thấy đơn `current_driver_id = employee.id` **hoặc đơn chưa có tài xế** (`current_driver_id IS NULL`) — để tài xế có thể tự nhận đơn mới trong khu vực. Hub Staff chỉ thấy đơn đang nằm tại `hub_id` của mình (`current_hub_id`). Shop chỉ thấy đơn `shop_id = shop.id` của mình. Xem `FindAll` trong [order_repo.go](../api/internal/repository/order_repo.go).
- **Action scoping (giới hạn hành động)**: mỗi role chỉ được thực hiện các bước chuyển trạng thái hợp lệ (xem mục 3). Enforce trong `UpdateOrderStatus` ([order_usecase.go](../api/internal/usecase/order_usecase.go)).
- **Deny by default**: nếu một role không được liệt kê rõ ràng cho một hành động thì mặc định bị từ chối.

## 2. Ma trận quyền theo Endpoint (API x Role)

Ký hiệu: ✅ được phép · ❌ bị từ chối · 🔒 được phép nhưng bị giới hạn phạm vi dữ liệu (data scope) · 🌐 công khai (không cần đăng nhập). Tất cả path đều có tiền tố `/api/v1`.

| Endpoint | Method | admin | shop | first_mile_driver | hub_staff | last_mile_driver |
|---|---|:---:|:---:|:---:|:---:|:---:|
| `/auth/login` | POST | 🌐 | 🌐 | 🌐 | 🌐 | 🌐 |
| `/auth/shop/login` | POST | ❌ | 🌐 | ❌ | ❌ | ❌ |
| `/shops/register` | POST | 🌐 | 🌐 | 🌐 | 🌐 | 🌐 |
| `/employees/register` | POST | 🌐 | 🌐 | 🌐 | 🌐 | 🌐 |
| `/orders` | POST (tạo đơn, API key) | — | shop API key | ❌ | ❌ | ❌ |
| `/shop/orders` | POST (tạo đơn, portal JWT) | ❌ | ✅ | ❌ | ❌ | ❌ |
| `/shop/rates/calculate` | POST | ❌ | ✅ | ❌ | ❌ | ❌ |
| `/orders` | GET (danh sách) | ✅ | 🔒 đơn của shop | 🔒 đơn của mình + chưa gán | 🔒 đơn tại hub | 🔒 đơn của mình + chưa gán |
| `/orders/:id` | GET (chi tiết) | ✅ | ❌ | ✅ | ✅ | ✅ |
| `/tracking/:tracking_number` | GET (tra cứu) | ✅ | ❌ | ✅ | ✅ | ✅ |
| `/orders/:id/status` | PUT | ✅ | ❌ | 🔒 theo state | 🔒 theo state | 🔒 theo state |
| `/shops/me` | GET (portal) | ❌ | ✅ | ❌ | ❌ | ❌ |
| `/employees` | GET/POST | ✅ | ❌ | ❌ | ❌ | ❌ |
| `/employees/:id/review` | PATCH (duyệt) | ✅ | ❌ | ❌ | ❌ | ❌ |
| `/shops` | GET/POST | ✅ | ❌ | ❌ | ❌ | ❌ |
| `/shops/:id/review` | PATCH (duyệt) | ✅ | ❌ | ❌ | ❌ | ❌ |
| `/hubs` | GET | 🌐 | 🌐 | 🌐 | 🌐 | 🌐 |
| `/hubs` | POST | ✅ | ❌ | ❌ | ❌ | ❌ |
| `/locations` | GET | 🌐 | 🌐 | 🌐 | 🌐 | 🌐 |
| `/locations` | POST | ✅ | ❌ | ❌ | ❌ | ❌ |
| `/rates` | GET | ✅ | ❌ | ✅ | ✅ | ✅ |
| `/rates` | POST | ✅ | ❌ | ❌ | ❌ | ❌ |
| `/stats/dashboard` | GET | ✅ | ❌ | ❌ | ❌ | ❌ |

> Ghi chú:
> - **Hai kênh tạo đơn cho Shop**: `POST /orders` (xác thực bằng **API key** qua `ShopAPIKeyAuth` — tích hợp máy-với-máy) và `POST /shop/orders` (xác thực bằng **JWT role `shop`** — dùng trong shop portal). Cả hai cùng gọi `CreateOrder`, chỉ khác nguồn lấy `shop_id` (context API key vs `user_id` trong token).
> - `GET /hubs` và `GET /locations` mở **công khai** để trang đăng ký tự phục vụ (shipper chọn hub, shop chọn khu vực) load được khi chưa đăng nhập.
> - `GET /orders/:id` và `GET /tracking/:tracking_number` hiện chỉ chặn ở mức đăng nhập (`AuthRequired`), chưa lọc data-scope ở tầng chi tiết. Đây là chủ ý để Hub Staff tra được đơn **chưa** nằm trong hub mình. Việc đổi trạng thái vẫn bị `UpdateOrderStatus` chặn theo role + scope.
> - `GET /orders` cho role `shop` lọc theo `shop_id` (nhánh mới trong `FindAll`) — shop chỉ thấy đơn của chính mình.
> - Điều phối gán tài xế thủ công (kiểu `/orders/:id/assign`) **chưa có** trong MVP — hiện đơn tự gán cho tài xế đầu tiên thao tác. Ngoài phạm vi lần này.

## 3. Ma trận quyền theo Chuyển trạng thái (Transition x Role)

Áp dụng cho `PUT /orders/:id/status`. Ngoài đúng role, còn phải thỏa **điều kiện phạm vi** ở cột cuối. Bảng này phản chiếu `validTransitions` + `allowedRoles` trong [order_usecase.go](../api/internal/usecase/order_usecase.go). `admin` bỏ qua ràng buộc role (nhưng vẫn tuân theo luồng trạng thái).

| Chuyển trạng thái | Role được phép | Điều kiện phạm vi bắt buộc |
|---|---|---|
| `ready_to_pick` → `picked_up` | `first_mile_driver` | Đơn chưa gán, hoặc gán cho chính tài xế (`current_driver_id`) |
| `ready_to_pick` → `returned` | mọi role vận hành | — |
| `picked_up` → `hub_inbound` | `hub_staff` | Ghi nhận đơn vào `hub_id` của staff (set `current_hub_id`) |
| `picked_up` → `returned` | mọi role vận hành | — |
| `hub_inbound` → `hub_outbound` | `hub_staff` | Đơn đang tại hub của staff |
| `hub_inbound` → `in_transit` | `hub_staff` | Chuyển tiếp sang hub khác |
| `in_transit` → `hub_inbound` | `hub_staff` | Nhập vào hub đích của staff |
| `hub_outbound` → `delivering` | `last_mile_driver` | Đơn chưa gán, hoặc gán cho chính tài xế |
| `delivering` → `delivered` | `last_mile_driver` | Đơn được gán cho chính tài xế |
| `delivering` → `returned` | mọi role vận hành | Đơn được gán cho chính tài xế (nếu là driver) |
| `delivering` → `hub_inbound` | `hub_staff` | Giao thất bại, nhập hoàn về kho |

> "mọi role vận hành" = `first_mile_driver`, `last_mile_driver`, `hub_staff` (đích `returned` không giới hạn role cụ thể trong `allowedRoles`).

Ràng buộc bổ sung (đồng bộ với [4_STATE_MACHINE.md](./4_STATE_MACHINE.md)):
- Nhảy trạng thái không có trong `validTransitions` (ví dụ `ready_to_pick` → `delivered`) → trả lỗi (HTTP 500 với message `không thể chuyển từ X sang Y` ở implementation hiện tại — xem mục 5 về việc chuẩn hóa mã lỗi).
- Khi vào kho (`hub_inbound`/`hub_outbound`) hệ thống **reset** `current_driver_id` để last-mile driver có thể tự nhận đơn ở bước `delivering`.

## 4. Cơ chế thực thi

### 4.1 Backend (Golang / Gin)
Thực thi phân quyền bằng 2 lớp middleware trong [middleware/auth_middleware.go](../api/internal/delivery/http/middleware/auth_middleware.go), đặt trước Handler:

1. **`AuthRequired()`**: đọc header `Authorization: Bearer <jwt>`, validate token, gắn `user_id`, `role`, `hub_id` vào `gin.Context`. Thiếu/không hợp lệ → `401` (`UNAUTHORIZED` / `INVALID_TOKEN`).
2. **`RoleRequired(roles ...string)`**: so khớp `role` trong context. Không khớp → `403 FORBIDDEN`.

Data scoping và transition scoping được kiểm tra **trong UseCase** (`order_usecase.go`), không nhét vào middleware, vì cần truy vấn dữ liệu đơn để so `current_driver_id` / `current_hub_id` với người đang thao tác.

```go
// Route thật trong order_handler.go — nhóm nội bộ chỉ cần AuthRequired,
// việc chặn theo role/scope nằm trong UpdateOrderStatus.
internalGroup := api.Group("/")
internalGroup.Use(middleware.AuthRequired())
{
    internalGroup.GET("/orders", handler.GetOrders)                       // scope filter trong usecase
    internalGroup.GET("/tracking/:tracking_number", handler.GetOrderByTracking)
    internalGroup.GET("/orders/:id", handler.GetOrder)
    internalGroup.PUT("/orders/:id/status", handler.UpdateStatus)         // scope + transition check trong usecase
}

// Route chỉ dành cho admin dùng RoleRequired (employee/shop/stats/rate-create...):
api.Use(middleware.AuthRequired(), middleware.RoleRequired("admin"))
```

```go
// Scope + transition check thật trong UpdateOrderStatus (order_usecase.go, rút gọn):
// 1. Driver chỉ thao tác đơn của mình (tự gán nếu đơn chưa có ai giữ).
if role == "first_mile_driver" || role == "last_mile_driver" {
    if order.CurrentDriverID != nil && *order.CurrentDriverID != employeeID {
        return errors.New("bạn không có quyền thao tác trên đơn hàng này")
    }
}
// 2. Kiểm tra transition hợp lệ theo validTransitions[order.Status].
// 3. Kiểm tra role được phép cho trạng thái đích (admin bỏ qua bước này).
```

### 4.2 Frontend (Vue 3)
Đã triển khai trong [router/index.js](../admin/src/router/index.js) và [authStore.js](../admin/src/stores/authStore.js):

- **`authStore.hasRole(...roles)`**: kiểm tra role hiện tại; **`homeRoute()`**: trả route mặc định theo role (admin → `Dashboard`, shop → `ShopDashboard`, còn lại → `MemberTasks`).
- Router tách 3 vùng gắn `meta.roles`: cây `/` (AdminLayout) chỉ `admin`; cây `/m` (MemberLayout) cho 3 role vận hành, riêng `/m/scan` giới hạn `hub_staff`; cây `/shop` (ShopLayout) chỉ role `shop`.
- Các route công khai (`meta.requiresAuth === false`): `/login`, `/shop/login`, `/register/shop`, `/register/shipper` — nếu đã đăng nhập mà truy cập thì bị đẩy về `homeRoute()`.
- Navigation guard gộp `meta.roles` của cả record cha + con trên đường match; sai role → đẩy về `homeRoute()` của chính user (không có trang 403 riêng, giữ UX đơn giản).
- Ẩn/hiện menu theo role: sidebar admin chỉ hiện nhóm "Quản trị" (Duyệt tài khoản/Shops/Nhân sự) với admin; tab "Quét đơn" chỉ hiện với `hub_staff`.

```js
// Guard thật (rút gọn) trong beforeEach:
const required = to.matched.map((r) => r.meta?.roles).filter(Boolean);
if (required.some((roles) => !authStore.hasRole(...roles))) {
  return authStore.homeRoute();
}
```

> Lưu ý bảo mật: check ở FE chỉ để cải thiện UX (ẩn nút/menu, điều hướng). **Nguồn chân lý phân quyền luôn là backend** — mọi endpoint phải tự kiểm tra, không tin vào FE.

## 5. Cấu trúc lỗi phân quyền

Response lỗi theo cấu trúc chuẩn ở [.agent/backend.md](../.agent/backend.md):

```json
{
  "success": false,
  "error": {
    "code": "FORBIDDEN",
    "message": "Bạn không có quyền thực hiện chức năng này"
  }
}
```

### 5.1 Lỗi từ middleware (đã chuẩn hóa)
Lớp `AuthRequired` / `RoleRequired` ([auth_middleware.go](../api/internal/delivery/http/middleware/auth_middleware.go)) trả về đúng HTTP status + code:

| Code | HTTP | Ý nghĩa |
|---|---|---|
| `UNAUTHORIZED` | 401 | Thiếu header `Authorization: Bearer ...` |
| `INVALID_TOKEN` | 401 | Token sai chữ ký hoặc hết hạn |
| `FORBIDDEN` | 403 | Role không nằm trong danh sách `RoleRequired` cho phép |

### 5.2 Lỗi từ UseCase (đã chuẩn hóa)
Các kiểm tra **data-scope** (không sở hữu đơn / sai hub) và **transition** (sai luồng, sai role theo `allowedRoles`) trong `UpdateOrderStatus` ([order_usecase.go](../api/internal/usecase/order_usecase.go)) nay bọc các **sentinel error** định nghĩa ở [domain/errors.go](../api/internal/domain/errors.go) qua `fmt.Errorf("%w: ...", domain.ErrXxx)`. Handler dùng helper `respondError` ([response.go](../api/internal/delivery/http/response.go)) để map sentinel → đúng HTTP status + `error.code`:

| Sentinel (domain) | Code | HTTP | Dùng cho |
|---|---|---|---|
| `ErrForbidden` | `FORBIDDEN` | 403 | Sai scope: "đơn hàng này đang do người khác phụ trách", sai role cho transition |
| `ErrInvalidTransition` | `INVALID_TRANSITION` | 422 | Sai luồng state machine: "không thể chuyển từ X sang Y" |
| `ErrNotFound` | `NOT_FOUND` | 404 | Không tìm thấy đơn |
| `ErrValidation` | `VALIDATION_ERROR` | 400 | Dữ liệu đầu vào không hợp lệ |
| (không khớp sentinel) | `INTERNAL_ERROR` | 500 | Lỗi hệ thống không phân loại |

`HTTPStatusForError` và `ErrorCode` (cùng file `errors.go`) dùng `errors.Is` để so khớp, nên message chi tiết vẫn giữ được (bọc bằng `%w`) mà vẫn phân loại đúng. Có test bảo vệ trong [state_machine_test.go](../api/internal/usecase/state_machine_test.go) (`TestUpdateStatus_InvalidTransitionWrapsSentinel`, `TestUpdateStatus_WrongRoleWrapsForbidden`, `TestUpdateStatus_DriverOwnershipWrapsForbidden`).

> Message tiếng Việt vẫn trả nguyên trong `error.message` cho người dùng đọc; `error.code` để client phân biệt nguyên nhân máy-đọc-được.

## 6. Onboarding tự phục vụ (Self-register + duyệt)

Ngoài luồng Admin tạo tài khoản trực tiếp, hệ thống cho phép Shop và tài xế **tự đăng ký** rồi chờ Admin duyệt.

### 6.1 Hai cột trạng thái tách bạch
- **`status`** (`pending` / `approved` / `rejected`): cổng onboarding. Tài khoản Admin tạo trực tiếp mặc định `approved`; tài khoản tự đăng ký khởi tạo `pending`.
- **`is_active`** (bool): công tắc khóa/mở của Admin **sau khi** đã duyệt. Đăng nhập yêu cầu `status = approved` **và** `is_active = true`.

Áp dụng cho cả `employees` và `shops` (xem hằng `StatusPending/Approved/Rejected` trong [employee.go](../api/internal/domain/employee.go)).

### 6.2 Đăng ký (công khai, không cần auth)
| Endpoint | Method | Ghi chú |
|---|---|---|
| `/employees/register` | POST | Chỉ chấp nhận role `first_mile_driver` / `last_mile_driver`; **bắt buộc chọn `hub_id`**. Tạo `pending` + `is_active=false`. |
| `/shops/register` | POST | Đăng ký bằng `email` + `password`. Tạo `pending`; **chưa sinh API key** (key chỉ sinh khi Admin duyệt). |

### 6.3 Đăng nhập
| Đối tượng | Endpoint | Định danh |
|---|---|---|
| Nhân viên/tài xế | `/auth/login` | `phone` + password |
| Shop (portal) | `/auth/shop/login` | `email` + password → JWT role `shop` |

Cả hai đều chặn đăng nhập khi `status != approved` (message rõ: "đang chờ duyệt" / "đã bị từ chối") hoặc `is_active=false`.

### 6.4 Duyệt (chỉ admin)
| Endpoint | Method | Hành vi |
|---|---|---|
| `/employees/:id/review` | PATCH `{approve}` | `approve=true` → `approved` + `is_active=true`; `false` → `rejected`. |
| `/shops/:id/review` | PATCH `{approve}` | Như trên. Khi duyệt lần đầu (shop chưa có key), **sinh API key và trả về đúng một lần** trong response. |

### 6.5 Quyền của role `shop`
Role `shop` chỉ truy cập được portal của chính mình (data-scope theo `shop_id` lấy từ JWT):
- `GET /shops/me` — thông tin tài khoản.
- `POST /shop/orders` — tạo đơn qua phiên đăng nhập (thay cho API key máy-máy).
- `POST /shop/rates/calculate` — tính cước thử.
- `GET /orders` — tự lọc về đơn của shop mình (nhánh `role=shop` trong [order_repo.go](../api/internal/repository/order_repo.go)).
