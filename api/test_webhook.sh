#!/bin/bash

echo "=== HƯỚNG DẪN TEST WEBHOOK VÀ GPS ==="
echo ""
echo "1. Cấu hình Webhook giả lập:"
echo "   - Bước 1: Truy cập trang https://webhook.site/"
echo "   - Bước 2: Copy 'Your unique URL' (Ví dụ: https://webhook.site/xxx-yyy-zzz)"
echo "   - Bước 3: Update dòng webhook_url của bảng shops trong database thành URL vừa copy."
echo "     (Có thể dùng lệnh SQL: UPDATE shops SET webhook_url = 'https://webhook.site/xxx' WHERE id = 'ID_CỦA_SHOP';)"
echo ""
echo "2. Lấy Order ID để test:"
echo "   - Bạn có thể lấy 1 ID đơn hàng bằng cách gọi API GET /api/v1/orders"
echo "   - (Giả sử ID đơn hàng là: 99999999-9999-9999-9999-999999999999)"
echo ""
echo "3. Gọi API cập nhật trạng thái kèm GPS (Mô phỏng Shipper):"
echo "   Chạy lệnh cURL sau (thay thế JWT_TOKEN và ORDER_ID cho phù hợp):"
echo ""

cat << 'EOF'
curl -X PUT http://localhost:8080/api/v1/orders/ORDER_ID/status \
  -H "Authorization: Bearer JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "status": "delivering",
    "note": "Shipper đang giao hàng",
    "latitude": 10.762622,
    "longitude": 106.660172
  }'
EOF

echo ""
echo "4. Kiểm tra kết quả:"
echo "   - Backend sẽ trả về 200 OK."
echo "   - Ngay lập tức, hãy nhìn sang màn hình webhook.site, bạn sẽ thấy 1 HTTP POST Request bay tới."
echo "   - Trong body của Webhook sẽ chứa JSON: { tracking_number, status, note, timestamp }."
echo "   - Mở DB bảng tracking_logs, bạn sẽ thấy record mới có latitude = 10.762622 và longitude = 106.660172."
