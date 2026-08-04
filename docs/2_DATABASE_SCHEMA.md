# Thiết kế Cơ sở Dữ liệu (Database Schema)

## Sơ đồ Bảng

### Bảng `employees` (Nhân viên/Tài xế)
| Column | Type | Description |
|---|---|---|
| id | UUID | Primary Key |
| name | VARCHAR(255) | Tên nhân viên |
| phone | VARCHAR(20) | Số điện thoại đăng nhập |
| role | VARCHAR(50) | Enum: `first_mile`, `hub_staff`, `last_mile`, `admin` |
| hub_id | UUID | ID của kho trực thuộc (Nullable) |
| created_at | TIMESTAMP | Thời gian tạo |

### Bảng `shops` (Khách hàng/Người gửi)
| Column | Type | Description |
|---|---|---|
| id | UUID | Primary Key |
| name | VARCHAR(255) | Tên cửa hàng |
| address | VARCHAR(255) | Địa chỉ lấy hàng |
| created_at | TIMESTAMP | Thời gian tạo |

### Bảng `shipping_orders` (Đơn hàng)
| Column | Type | Description |
|---|---|---|
| id | UUID | Primary Key |
| tracking_number| VARCHAR(50) | Mã vận đơn (Unique) |
| shop_id | UUID | Foreign Key -> `shops.id` |
| receiver_name | VARCHAR(255) | Tên người nhận |
| receiver_phone| VARCHAR(20) | Số điện thoại nhận |
| receiver_address| VARCHAR(255)| Địa chỉ nhận |
| status | VARCHAR(50) | Trạng thái đơn (Dùng Varchar kết hợp validation ở tầng App thay vì Enum DB cứng để dễ mở rộng) |
| current_driver_id| UUID | Driver đang giữ đơn (Nullable) |
| created_at | TIMESTAMP | Thời gian tạo |
| updated_at | TIMESTAMP | Thời gian cập nhật cuối |

### Bảng `tracking_logs` (Lịch sử hành trình)
| Column | Type | Description |
|---|---|---|
| id | UUID | Primary Key |
| order_id | UUID | Foreign Key -> `shipping_orders.id` |
| status | VARCHAR(50) | Trạng thái tại thời điểm log |
| location | VARCHAR(255) | Vị trí hiện tại (Tên Hub hoặc Tọa độ) |
| employee_id | UUID | Nhân viên/Driver thực hiện (Nullable) |
| created_at | TIMESTAMP | Thời điểm xảy ra |

## Giải thích thiết kế Enum/Varchar cho Status
Chúng ta sẽ sử dụng kiểu dữ liệu `VARCHAR` cho cột `status` thay vì native `ENUM` của PostgreSQL. 
**Lý do:**
1. Trạng thái giao hàng trong Logistics thay đổi và mở rộng rất thường xuyên (Thêm trạng thái Delay, Damage, Return_to_hub, v.v.). Việc ALTER TYPE của Enum trong Postgres rườm rà.
2. Dễ dàng validate ở tầng Application (Go struct tags hoặc code logic) mà không cần lo lỗi DB type mismatch.
