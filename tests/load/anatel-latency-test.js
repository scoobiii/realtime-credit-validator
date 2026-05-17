import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  thresholds: {
    http_req_duration: ['p(95)<500'], // Anatel exige latência < 500ms
  },
  vus: 10,
  duration: '30s',
};

export default function () {
  const token = __ENV.TOKEN;
  const gatewayUrl = __ENV.GATEWAY_URL || 'http://localhost:8080';
  
  const payload = JSON.stringify({
    user_id: 'testuser',
    amount: 100,
    idempotency_key: `${__VU}-${__ITER}-${Date.now()}`,
    payment_method: 'pix'
  });
  
  const params = {
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json',
    },
  };
  
  const res = http.post(`${gatewayUrl}/v1/credit`, payload, params);
  
  check(res, {
    'status is 200': (r) => r.status === 200,
    'credit confirmed': (r) => r.json('status') === 'confirmed',
    'latency under 500ms': (r) => r.timings.duration < 500,
  });
  
  sleep(0.1);
}
