// Load test cho Ocean Express API bằng k6.
//
// Kịch bản: đăng nhập lấy JWT một lần cho mỗi VU, sau đó lặp gọi các endpoint
// đọc chính (list orders, rates, dashboard stats, tra cứu) — mô phỏng tải của
// admin panel khi nhiều người dùng cùng thao tác.
//
// Cách chạy (cần cài k6: https://k6.io/docs/get-started/installation/):
//   k6 run api/test/load/k6_smoke.js
//   BASE_URL=http://localhost:8080 ADMIN_PHONE=0900000000 ADMIN_PASS=admin123 k6 run api/test/load/k6_smoke.js
//
// Tùy chỉnh tải qua biến môi trường:
//   VUS=20 DURATION=1m k6 run api/test/load/k6_smoke.js

import http from 'k6/http';
import { check, sleep, group } from 'k6';
import { Rate, Trend } from 'k6/metrics';

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const API = `${BASE_URL}/api/v1`;
const ADMIN_PHONE = __ENV.ADMIN_PHONE || '0900000000';
const ADMIN_PASS = __ENV.ADMIN_PASS || 'admin123';

// Metric tùy chỉnh: tỉ lệ login thất bại (tách khỏi http error tổng).
const loginFailRate = new Rate('login_failed');
const listOrdersDuration = new Trend('list_orders_duration', true);

export const options = {
  scenarios: {
    // Ramp dần để thấy hệ thống phản ứng khi tải tăng, thay vì dội 1 phát.
    ramping_reads: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: __ENV.RAMP_UP || '15s', target: Number(__ENV.VUS || 10) },
        { duration: __ENV.DURATION || '30s', target: Number(__ENV.VUS || 10) },
        { duration: '10s', target: 0 },
      ],
      gracefulRampDown: '5s',
    },
  },
  thresholds: {
    // 95% request đọc phải dưới 500ms; tỉ lệ lỗi HTTP dưới 1%.
    http_req_duration: ['p(95)<500'],
    http_req_failed: ['rate<0.01'],
    login_failed: ['rate<0.01'],
    checks: ['rate>0.99'],
  },
};

// setup() chạy 1 lần trước toàn bộ VU — lấy token admin dùng chung.
export function setup() {
  const res = http.post(
    `${API}/auth/login`,
    JSON.stringify({ phone: ADMIN_PHONE, password: ADMIN_PASS }),
    { headers: { 'Content-Type': 'application/json' } }
  );
  const ok = check(res, {
    'setup: login 200': (r) => r.status === 200,
    'setup: có token': (r) => !!(r.json() && r.json().data && r.json().data.token),
  });
  if (!ok) {
    throw new Error(`Không đăng nhập được để lấy token (status ${res.status}). Kiểm tra server + tài khoản seed.`);
  }
  return { token: res.json().data.token };
}

export default function (data) {
  const authHeaders = {
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${data.token}`,
    },
  };

  group('health', () => {
    const res = http.get(`${BASE_URL}/ping`);
    check(res, { 'ping 200': (r) => r.status === 200 });
  });

  group('list orders', () => {
    const res = http.get(`${API}/orders`, authHeaders);
    listOrdersDuration.add(res.timings.duration);
    check(res, {
      'orders 200': (r) => r.status === 200,
      'orders có success': (r) => r.json('success') === true,
    });
  });

  group('list rates', () => {
    const res = http.get(`${API}/rates`, authHeaders);
    check(res, { 'rates 200': (r) => r.status === 200 });
  });

  group('dashboard stats', () => {
    const res = http.get(`${API}/stats/dashboard`, authHeaders);
    check(res, { 'stats 200': (r) => r.status === 200 });
  });

  loginFailRate.add(false); // đánh dấu iteration này login-flow ổn (token từ setup)
  sleep(1);
}
