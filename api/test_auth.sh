#!/bin/bash

echo "=== 1. TẠO TÀI KHOẢN ADMIN VÀO DATABASE ==="
# Mật khẩu "123456" đã được hash bằng bcrypt
# Chú ý: Cổng DB có thể là 5432, username/password mặc định trong docker-compose.yml
docker exec -i ocean_express_db psql -U root -d ocean_express_db -c "
INSERT INTO employees (id, name, phone, password_hash, role, is_active) 
VALUES (
  '11111111-1111-1111-1111-111111111111', 
  'System Admin', 
  '0987654321', 
  '\$2a\$14\$vC.8lR9/i6/C/z3M9e0kueR5S3I1qK6n1iO.B5qN6V6/O1nI4O7v.', 
  'admin', 
  true
) ON CONFLICT DO NOTHING;
"

echo -e "\n=== 2. GỌI API LOGIN ==="
LOGIN_RES=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "0987654321",
    "password": "123456"
  }')

echo "Response từ API Login: $LOGIN_RES"

# Trích xuất Token từ JSON Response bằng grep/sed
TOKEN=$(echo $LOGIN_RES | grep -o '"token":"[^"]*' | sed 's/"token":"//')

if [ -z "$TOKEN" ]; then
  echo -e "\n[LỖI] Không lấy được Token. Vui lòng kiểm tra lại quá trình Login."
  exit 1
fi

echo -e "\n=== 3. GỌI API ĐƯỢC BẢO VỆ ==="
# Route /api/v1/me yêu cầu token để trả về thông tin User
curl -s -X GET http://localhost:8080/api/v1/me \
  -H "Authorization: Bearer $TOKEN" | jq . || echo "\n(Nếu không có jq, xem kết quả dạng text phía trên)"
