import http from 'k6/http';
import { check } from 'k6';
import { SharedArray } from 'k6/data';

// .replace strips the UTF-8 BOM that PowerShell writes at the start of files.
const codes = new SharedArray('codes', () =>
  JSON.parse(open('./codes.json').replace(/^\uFEFF/, ''))
);

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8000';

export const options = {
  scenarios: {
    redirect: {
      executor: 'constant-arrival-rate',
      rate: 200,
      timeUnit: '1s',
      duration: '60s',
      preAllocatedVUs: 50,
      maxVUs: 500,
    },
  },
  thresholds: {
    http_req_duration: ['p(95)<200'],
    http_req_failed:   ['rate<0.001'],
  },
};

export default function () {
  const code = codes[Math.floor(Math.random() * codes.length)];
  const res = http.get(`${BASE_URL}/${code}`, { redirects: 0 });
  check(res, { 'is 302': (r) => r.status === 302 });
}
