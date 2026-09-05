import http from 'k6/http';
import { check } from 'k6';
import { SharedArray } from 'k6/data';

const codes = new SharedArray('codes', () =>
  JSON.parse(open('./codes.json').replace(/^\uFEFF/, ''))
);

const BASE_REDIRECT  = __ENV.BASE_REDIRECT;
const BASE_SHORTEN   = __ENV.BASE_SHORTEN;
const BASE_ANALYTICS = __ENV.BASE_ANALYTICS;

// one step = 10s ramp + 90s hold
function step(rate) {
  return [
    { duration: '10s', target: rate },
    { duration: '90s', target: rate },
  ];
}

export const options = {
  scenarios: {
    staircase: {
      executor: 'ramping-arrival-rate',
      startRate: 1000,
      timeUnit: '1s',
      preAllocatedVUs: 2000,
      maxVUs: 8000,
      stages: [
        { duration: '2m', target: 1000 },        // warm-up
        ...step(2000),
        ...step(3000),
        ...step(4000),
        ...step(5000),
        ...step(6000),
        ...step(7000),
        ...step(8000),
      ],
    },
  },
  thresholds: {
    'http_req_duration{op:redirect}': ['p(95)<200'],
    'http_req_failed': ['rate<0.001'],
    // stop early once it is clearly broken, so weak configs finish fast
    'http_req_duration{op:redirect}': [
      { threshold: 'p(95)<1000', abortOnFail: true, delayAbortEval: '30s' },
    ],
  },
};

export default function () {
  const r = Math.random();
  if (r < 0.990)      redirect();
  else if (r < 0.999) createLink();
  else                analytics();
}

function redirect() {
  const code = codes[Math.floor(Math.random() * codes.length)];
  const res = http.get(`${BASE_REDIRECT}/${code}`, {
    redirects: 0,
    tags: { op: 'redirect' },
  });
  check(res, { 'is 302': (r) => r.status === 302 });
}

function createLink() {
  const res = http.post(
    `${BASE_SHORTEN}/api/v1/links`,
    JSON.stringify({ url: 'https://example.com/page' }),
    { headers: { 'Content-Type': 'application/json' }, tags: { op: 'create' } }
  );
  check(res, { 'is 201': (r) => r.status === 201 });
}

function analytics() {
  const code = codes[Math.floor(Math.random() * codes.length)];
  const res = http.get(`${BASE_ANALYTICS}/api/v1/analytics/${code}`, {
    tags: { op: 'analytics' },
  });
  check(res, { 'is 200': (r) => r.status === 200 });
}