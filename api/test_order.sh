#!/bin/bash

echo "=== 1. INSERT DỮ LIỆU MẪU (MOCK DATA) VÀO DATABASE ==="
docker exec -i ocean_express_db psql -U root -d ocean_express_db -c "
-- Tạo 2 Locations (Tỉnh A, Tỉnh B)
INSERT INTO locations (id, name, type) VALUES ('LOC_A', 'Tỉnh A', 'province') ON CONFLICT DO NOTHING;
INSERT INTO locations (id, name, type) VALUES ('LOC_B', 'Tỉnh B', 'province') ON CONFLICT DO NOTHING;

-- Tạo 1 cấu hình ShippingRate cho tuyến A -> B
-- Base: 1000g = 30000đ. Mỗi 500g tiếp theo = 5000đ
INSERT INTO shipping_rates (from_location_id, to_location_id, base_weight, base_fee, extra_weight_step, extra_fee)
VALUES ('LOC_A', 'LOC_B', 1000, 30000, 500, 5000) ON CONFLICT DO NOTHING;

-- Tạo 1 Shop (BC Sport) nằm ở Tỉnh A
-- api_key = 'API_KEY_BC_SPORT'
INSERT INTO shops (id, name, webhook_url, api_key, location_id, address_detail) 
VALUES (
  '12345678-1234-1234-1234-123456789012', 
  'BC Sport', 
  'https://bcsport.example.com/webhook', 
  'API_KEY_BC_SPORT', 
  'LOC_A', 
  '123 Đường thể thao'
) ON CONFLICT DO NOTHING;
"

echo -e "\n=== 2. GỌI API TẠO ĐƠN HÀNG (POST /orders) ==="
# Test Case: Đơn nặng 1200g
# Theo công thức: weight = 1200g. 
# BaseWeight = 1000g => Dư 200g.
# ExtraWeightStep = 500g => 200g được làm tròn lên 1 step.
# Fee = BaseFee (30000) + 1 * ExtraFee (5000) = 35000đ
curl -s -X POST http://localhost:8080/api/v1/orders \
  -H "Content-Type: application/json" \
  -H "X-API-Key: API_KEY_BC_SPORT" \
  -d '{
    "receiver_name": "Nguyen Khach Hàng",
    "receiver_phone": "0911223344",
    "receiver_location_id": "LOC_B",
    "receiver_address_detail": "456 Đường nhận hàng",
    "weight": 1200,
    "cod_amount": 500000
  }' | jq . || echo -e "\n(Cài jq để format JSON, hoặc xem text phía trên)"
