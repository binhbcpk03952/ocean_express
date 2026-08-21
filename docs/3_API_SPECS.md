# Đặc tả API (API Specifications)

Base URL: `/api/v1`

## 1. Đăng nhập (Login)
**Endpoint**: `POST /auth/login`
- **Request Body**:
  ```json
  {
    "phone": "0987654321",
    "password": "hashed_password"
  }
  ```
- **Response**:
  ```json
  {
    "success": true,
    "data": {
      "token": "jwt_token_string",
      "employee": {
        "id": "uuid",
        "name": "Nguyen Van A",
        "role": "first_mile_driver",
        "hub_id": "uuid | null"
      }
    }
  }
  ```
- **Ghi chú**: `role` là một trong `admin`, `first_mile_driver`, `hub_staff`, `last_mile_driver`. `hub_id` chỉ có giá trị với `hub_staff` và các tài xế gắn bưu cục.

## 2. Tạo Đơn hàng (Create Order)
**Endpoint**: `POST /orders`
**Auth**: Shop API Key (header `X-API-Key`) — đây là API dành cho đối tác Shop, không dùng JWT nhân viên.
- **Request Body**:
  ```json
  {
    "receiver_name": "Tran Thi B",
    "receiver_phone": "0123456789",
    "receiver_location_id": "VN-HCM-Q1",
    "receiver_address_detail": "123 Le Loi, Q1, HCM",
    "weight": 500,
    "cod_amount": 100000
  }
  ```
- **Ghi chú**: `shop_id`, địa chỉ gửi và phí ship được suy ra từ API key của Shop; client không gửi `shop_id`. `weight` tính bằng gram.
- **Response**: trả về đối tượng đơn hàng đầy đủ (`id`, `tracking_number`, `status = ready_to_pick`, `shipping_fee`, `estimated_delivery_time`, ...).

## 3. Cập nhật Trạng thái (Update Status)
**Endpoint**: `PUT /orders/:id/status`  (`:id` là UUID của đơn)
**Auth**: JWT nhân viên (Role + scope được kiểm tra trong UseCase — xem [5_RBAC.md](./5_RBAC.md)).
- **Request Body**:
  ```json
  {
    "status": "picked_up",
    "note": "Hàng dễ vỡ",
    "latitude": 10.776,
    "longitude": 106.700
  }
  ```
- **Ghi chú**: `latitude`/`longitude` là tùy chọn (client mobile lấy từ GPS thiết bị); nếu không có, gửi 0. Mỗi lần cập nhật hợp lệ sẽ ghi thêm một bản ghi vào `tracking_logs`.
- **Response**:
  ```json
  {
    "success": true,
    "message": "Cập nhật trạng thái thành công"
  }
  ```

## 4. Lấy chi tiết đơn hàng theo ID (Get Order)
**Endpoint**: `GET /orders/:id`  (`:id` là UUID)
**Auth**: JWT nhân viên.
- **Response**:
  ```json
  {
    "success": true,
    "data": {
      "order": {
        "id": "uuid",
        "tracking_number": "OE-1712345678",
        "status": "picked_up",
        "receiver_name": "Tran Thi B",
        "receiver_phone": "0123456789",
        "receiver_address_detail": "123 Le Loi, Q1, HCM",
        "cod_amount": 100000,
        "shipping_fee": 30000,
        "weight": 500
      },
      "logs": [
        { "status": "ready_to_pick", "created_at": "2023-10-01T10:00:00Z" },
        { "status": "picked_up", "note": "Hàng dễ vỡ", "latitude": 10.776, "longitude": 106.700, "created_at": "2023-10-01T14:30:00Z" }
      ]
    }
  }
  ```

## 5. Tra cứu đơn theo Mã vận đơn (Lookup by Tracking Number)
**Endpoint**: `GET /tracking/:tracking_number`
**Auth**: JWT nhân viên (mọi role nội bộ đều tra được — đặc biệt Hub Staff dùng để xử lý đơn chưa nằm trong danh sách hub của mình).
- **Ghi chú**: dùng path riêng `/tracking/...` thay vì `/orders/tracking/...` do ràng buộc router của Gin (static segment không đứng cùng vị trí với param `:id`).
- **Response**: cùng cấu trúc `{ order, logs }` như mục 4.

## 6. Danh sách đơn hàng (List Orders)
**Endpoint**: `GET /orders`
**Auth**: JWT (nhân viên hoặc shop). Kết quả **tự động giới hạn phạm vi** theo role:
- `admin`: toàn bộ đơn.
- `first_mile_driver` / `last_mile_driver`: đơn đang gán cho mình hoặc chưa có tài xế (`current_driver_id IS NULL`).
- `hub_staff`: đơn đang nằm tại hub của mình (`current_hub_id = hub_id`).
- `shop`: đơn do chính shop tạo (`shop_id = <shop trong token>`).
- **Response**: `{ "success": true, "data": [ <order>, ... ] }`.

---

# Onboarding tự phục vụ & Portal Shop

Hệ thống hỗ trợ đăng ký tự phục vụ cho **Shop** (đối tác gửi hàng) và **Shipper** (tài xế). Tài khoản tự đăng ký ở trạng thái `pending`, chờ Admin duyệt mới dùng được. Xem cơ chế phân quyền chi tiết ở [5_RBAC.md](./5_RBAC.md).

Vòng đời trạng thái tài khoản (`status`): `pending` → `approved` / `rejected`. Tách bạch với `is_active` (Admin khóa/mở sau khi đã duyệt).

## 7. Shop tự đăng ký (Register Shop)
**Endpoint**: `POST /shops/register`
**Auth**: Không (công khai).
- **Request Body**:
  ```json
  {
    "name": "BC Sport",
    "email": "shop@example.com",
    "password": "matkhau",
    "location_id": "VN-HCM-Q1",
    "address_detail": "123 Le Loi, Q1, HCM"
  }
  ```
- **Ghi chú**: tài khoản tạo ở trạng thái `pending`, chưa có API key (API key sinh khi Admin duyệt). `webhook_url` set sau trong portal.
- **Response**: `{ "success": true, "data": <shop>, "message": "Đăng ký thành công, tài khoản đang chờ Admin duyệt" }`.

## 8. Đăng nhập Portal Shop (Shop Login)
**Endpoint**: `POST /auth/shop/login`
**Auth**: Không (công khai).
- **Request Body**: `{ "email": "shop@example.com", "password": "matkhau" }`
- **Response**:
  ```json
  {
    "success": true,
    "data": {
      "token": "jwt_token_string",
      "shop": { "id": "uuid", "name": "BC Sport", "email": "shop@example.com", "role": "shop" }
    }
  }
  ```
- **Ghi chú**: JWT mang `role = shop`. Tài khoản `pending`/`rejected`/bị khóa sẽ bị từ chối với thông báo tương ứng.

## 9. Portal Shop — Thông tin tài khoản (Me)
**Endpoint**: `GET /shops/me`
**Auth**: JWT role `shop`.
- **Response**: `{ "success": true, "data": <shop> }` (không bao gồm `password_hash`, `api_key`).

## 10. Portal Shop — Tạo đơn qua phiên đăng nhập
**Endpoint**: `POST /shop/orders`
**Auth**: JWT role `shop`.
- **Request Body**: giống mục 2, nhưng `shop_id` suy ra từ token thay vì API key.
- **Ghi chú**: tương đương `POST /orders` (API key) nhưng dành cho shop thao tác trực tiếp trên portal.

## 11. Portal Shop — Tính cước thử
**Endpoint**: `POST /shop/rates/calculate`
**Auth**: JWT role `shop`.
- **Request Body**: `{ "receiver_location_id": "VN-HCM-Q1", "weight": 500 }`
- **Response**: `{ "success": true, "data": { "fee": 30000 } }`.

## 12. Shipper tự đăng ký (Register Shipper)
**Endpoint**: `POST /employees/register`
**Auth**: Không (công khai).
- **Request Body**:
  ```json
  {
    "name": "Nguyen Van A",
    "phone": "0987654321",
    "password": "matkhau",
    "role": "first_mile_driver",
    "hub_id": "uuid"
  }
  ```
- **Ghi chú**: chỉ chấp nhận `role` là `first_mile_driver` hoặc `last_mile_driver` (admin/hub_staff do Admin tạo trực tiếp). `hub_id` bắt buộc (shipper tự chọn bưu cục). Tài khoản tạo ở trạng thái `pending`, `is_active = false` cho tới khi Admin duyệt.
- **Response**: `{ "success": true, "data": <employee>, "message": "..." }`.

## 13. Admin duyệt/từ chối Shop
**Endpoint**: `PATCH /shops/:id/review`
**Auth**: JWT role `admin`.
- **Request Body**: `{ "approve": true }` (true = duyệt, false = từ chối).
- **Ghi chú**: khi duyệt lần đầu (shop chưa có API key), hệ thống sinh API key và trả về **một lần duy nhất** trong `api_key`.
- **Response**: `{ "success": true, "data": { "shop": <shop>, "api_key": "oe_... | \"\"" } }`.

## 14. Admin duyệt/từ chối Shipper
**Endpoint**: `PATCH /employees/:id/review`
**Auth**: JWT role `admin`.
- **Request Body**: `{ "approve": true }`.
- **Ghi chú**: duyệt → `status=approved` + `is_active=true`; từ chối → `status=rejected`.
- **Response**: `{ "success": true, "data": <employee> }`.

## 15. Danh sách tài khoản theo trạng thái (Admin)
- `GET /shops?status=pending` — lọc đối tác theo trạng thái duyệt (rỗng = tất cả).
- `GET /employees?status=pending` — lọc nhân sự theo trạng thái duyệt. Vẫn hỗ trợ `?hub_id=` như cũ.

---

## 16. Webhook Đối tác (Partner Webhooks)

Khi đơn hàng có thay đổi trạng thái trong vòng đời vận chuyển, Ocean Express sẽ chủ động bắn HTTP POST request tới `webhook_url` do Shop cấu hình.

### Header HTTP:
- `Content-Type`: `application/json`
- `User-Agent`: `OceanExpress-Webhook/1.0`
- `X-Event-ID`: `<uuid>` (Khóa Idempotency định danh duy nhất của sự kiện)
- `X-Tracking-Number`: `<mã_vận_đơn>` (VD: `BCS12345`)
- `X-Sequence-ID`: `<số_thứ_tự_bước>` (VD: `1`, `2`, `3`)

### Webhook JSON Payload Schema:
```json
{
  "event_id": "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d",
  "tracking_number": "BCS12345",
  "status": "ready_to_pick",
  "status_name": "Chờ lấy hàng",
  "status_label": "Chờ lấy",
  "status_description": "Đơn hàng mới tạo, đang chờ phân công tài xế lấy hàng",
  "note": "Đơn hàng được tạo bởi Shop",
  "sequence_id": 1,
  "timestamp": "2026-08-20T12:04:00Z",
  "created_at": "2026-08-20T12:04:00Z",
  "timestamp_epoch": 1787227440000
}
```

### Quy tắc Ack & Retry:
1. **Thành công (Ack)**: Khi máy chủ của Shop phản hồi mã **HTTP 2xx (200, 201, 204)** trong vòng **10 giây**, Ocean Express ghi nhận thành công và kết thúc luồng (không retry).
2. **Thất bại & Thử lại (Retry)**: Khi máy chủ của Shop phản hồi mã lỗi **4xx / 5xx** hoặc bị **timeout / lỗi kết nối mạng quá 10s**, Ocean Express sẽ tự động retry tối đa **3 lần** kèm exponential backoff (2s, 4s).
3. **Bảo đảm thứ tự (FIFO & Partitioning)**: Các webhook của cùng một `tracking_number` luôn được xử lý tuần tự theo đúng tiến trình thời gian thực. Shop có thể dựa vào `sequence_id` hoặc `timestamp_epoch` để loại trừ các bản ghi out-of-order nếu mạng phía đối tác có độ trễ.

