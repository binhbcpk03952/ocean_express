#!/usr/bin/env bash
# Smoke + micro-load test cho Ocean Express API, chỉ dùng bash + curl.
# Dùng khi chưa cài k6 — verify nhanh rằng các endpoint đọc chính còn sống và
# đo sơ bộ độ trễ dưới tải nhẹ tuần tự.
#
# Cách chạy:
#   bash api/test/load/smoke.sh
#   BASE_URL=http://localhost:8080 ADMIN_PHONE=0900000000 ADMIN_PASS=admin123 REQUESTS=30 bash api/test/load/smoke.sh
#
# Thoát mã 0 nếu mọi kiểm tra pass, 1 nếu có bất kỳ lỗi nào (dùng được trong CI).

set -uo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"
API="${BASE_URL}/api/v1"
ADMIN_PHONE="${ADMIN_PHONE:-0900000000}"
ADMIN_PASS="${ADMIN_PASS:-admin123}"
REQUESTS="${REQUESTS:-20}"

pass=0
fail=0

# check_status <label> <expected_code> <actual_code>
check_status() {
  local label="$1" expected="$2" actual="$3"
  if [ "$actual" = "$expected" ]; then
    echo "  PASS  $label (HTTP $actual)"
    pass=$((pass + 1))
  else
    echo "  FAIL  $label (mong đợi $expected, nhận $actual)"
    fail=$((fail + 1))
  fi
}

echo "== Ocean Express smoke/load — target: $BASE_URL =="

# 1. Health
code=$(curl -s -o /dev/null -w '%{http_code}' "$BASE_URL/ping")
check_status "GET /ping" "200" "$code"

# 2. Login lấy token
login_body=$(curl -s -X POST "$API/auth/login" \
  -H 'Content-Type: application/json' \
  -d "{\"phone\":\"$ADMIN_PHONE\",\"password\":\"$ADMIN_PASS\"}")

# Trích token không cần jq (grep JSON field).
TOKEN=$(printf '%s' "$login_body" | grep -o '"token":"[^"]*"' | head -1 | sed 's/"token":"//;s/"$//')
if [ -n "$TOKEN" ]; then
  echo "  PASS  POST /auth/login (có token)"
  pass=$((pass + 1))
else
  echo "  FAIL  POST /auth/login — không lấy được token. Response: $login_body"
  fail=$((fail + 1))
  echo ""
  echo "Không có token, bỏ qua các endpoint cần auth. pass=$pass fail=$fail"
  exit 1
fi

AUTH=(-H "Authorization: Bearer $TOKEN")

# 3. Các endpoint đọc cần auth
for path in "/orders" "/rates" "/stats/dashboard" "/locations"; do
  code=$(curl -s -o /dev/null -w '%{http_code}' "${AUTH[@]}" "$API$path")
  check_status "GET $path" "200" "$code"
done

# 4. Micro-load: bắn REQUESTS lần GET /orders tuần tự, đo tổng thời gian.
echo ""
echo "== Micro-load: $REQUESTS × GET /orders =="
total_ms=0
slowest_ms=0
load_fail=0
for _ in $(seq 1 "$REQUESTS"); do
  t=$(curl -s -o /dev/null -w '%{time_total}' "${AUTH[@]}" "$API/orders")
  ms=$(awk "BEGIN{printf \"%d\", $t * 1000}")
  total_ms=$((total_ms + ms))
  if [ "$ms" -gt "$slowest_ms" ]; then slowest_ms=$ms; fi
done
avg_ms=$((total_ms / REQUESTS))
echo "  Trung bình: ${avg_ms}ms | Chậm nhất: ${slowest_ms}ms | Tổng: ${total_ms}ms"

# Ngưỡng mềm: trung bình dưới 500ms coi là đạt cho tải nhẹ local.
if [ "$avg_ms" -lt 500 ]; then
  echo "  PASS  độ trễ trung bình dưới 500ms"
  pass=$((pass + 1))
else
  echo "  WARN  độ trễ trung bình ${avg_ms}ms ≥ 500ms (chấp nhận tùy môi trường)"
fi

echo ""
echo "== Kết quả: $pass pass, $fail fail =="
[ "$fail" -eq 0 ] && exit 0 || exit 1
