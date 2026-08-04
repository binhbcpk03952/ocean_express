# Load & Smoke Testing — Ocean Express API

Bộ kiểm thử tải/khói cho API, phục vụ Phase 4 (Kiểm thử & Tối ưu). Hai công cụ,
tách biệt hoàn toàn với code nghiệp vụ — chỉ gọi API qua HTTP.

## 1. `smoke.sh` — smoke test nhanh (không cần cài gì thêm)

Chỉ cần `bash` + `curl` (có sẵn trên máy dev). Kiểm tra các endpoint đọc chính
còn sống và đo độ trễ thô. Dùng để xác nhận nhanh sau khi deploy/rebuild.

```bash
# Server chạy ở localhost:8080 (mặc định)
bash api/test/load/smoke.sh

# Trỏ tới host khác + tài khoản khác
BASE_URL=http://localhost:8080 PHONE=0900000000 PASSWORD=admin123 \
  bash api/test/load/smoke.sh
```

Thoát code `0` nếu tất cả pass, `1` nếu có bước fail (dùng được trong CI).

## 2. `k6_smoke.js` — load test bằng k6 (ramping VUs + thresholds)

Cần cài [k6](https://k6.io/docs/get-started/installation/). Kịch bản: login lấy
token một lần, sau đó vòng lặp gọi các endpoint đọc dưới tải tăng dần.

```bash
# Cài k6 (một lần)
#   macOS:   brew install k6
#   Windows: winget install k6 --source winget
#   Linux:   xem trang cài đặt chính thức

# Chạy với cấu hình mặc định
k6 run api/test/load/k6_smoke.js

# Trỏ host / tài khoản qua biến môi trường
k6 run -e BASE_URL=http://localhost:8080 -e PHONE=0900000000 -e PASSWORD=admin123 \
  api/test/load/k6_smoke.js
```

### Thresholds (ngưỡng đạt/rớt)
- `http_req_duration p(95) < 500ms` — 95% request nhanh hơn 500ms.
- `http_req_failed rate < 1%` — dưới 1% request lỗi.

k6 thoát code khác `0` nếu vi phạm threshold → tích hợp CI được.

## Ghi chú
- Đây là read-load (đọc). Không tạo/sửa dữ liệu để tránh làm bẩn DB. Muốn test
  ghi (tạo đơn qua API key shop) thì bổ sung scenario riêng, tách VU pool.
- Con số tham chiếu trên máy dev (server rảnh): GET /orders ~6ms trung bình.
  Dùng làm mốc so sánh khi tối ưu hoặc phát hiện hồi quy hiệu năng.
