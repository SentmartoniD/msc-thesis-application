import http from 'k6/http';
import { check } from 'k6';
import { SharedArray } from 'k6/data';

const codes = new SharedArray('codes', () => JSON.parse(open('./codes.json')));
const BASE = __ENV.BASE_URL;
const API  = __ENV.API_URL;

export const options = {
  scenarios: {
    redirect: {
      executor: 'constant-arrival-rate',
      exec: 'redirect',          // ← which function this scenario calls
      rate: 2000, timeUnit: '1s', duration: '5m',
      preAllocatedVUs: 500, maxVUs: 4000,
    },
    shorten: {
      executor: 'constant-arrival-rate',
      exec: 'shorten',
      rate: 20, timeUnit: '1s', duration: '5m',
      preAllocatedVUs: 20, maxVUs: 200,
    },
    stats: {
      executor: 'constant-arrival-rate',
      exec: 'stats',
      rate: 5, timeUnit: '1s', duration: '5m',
      preAllocatedVUs: 10, maxVUs: 50,
    },
  },

  thresholds: {
    'http_req_duration{scenario:redirect}': ['p(95)<200'],
    'http_req_duration{scenario:shorten}':  ['p(95)<500'],
    'http_req_failed{scenario:redirect}':   ['rate<0.001'],
    'http_req_failed{scenario:shorten}':    ['rate<0.001'],
  },
};

export function redirect() {
  const code = codes[Math.floor(Math.random() * codes.length)];
  const res = http.get(`${BASE}/${code}`, { redirects: 0 });
  check(res, { 'is 302': (r) => r.status === 302 });
}

export function shorten() {
  const res = http.post(`${API}/api/v1/links`,
    JSON.stringify({ url: `https://example.com/p/${Date.now()}` }),
    { headers: { 'Content-Type': 'application/json' } });
  check(res, { 'is 201': (r) => r.status === 201 });
}

export function stats() {
  const code = codes[Math.floor(Math.random() * codes.length)];
  const res = http.get(`${API}/api/v1/analytics/${code}`);
  check(res, { 'is 200': (r) => r.status === 200 });
}
