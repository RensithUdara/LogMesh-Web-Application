import http from "k6/http";
import { check, sleep } from "k6";

export const options = {
  vus: 20,
  duration: "30s"
};

export default function () {
  const level = __ITER % 10 === 0 ? "ERROR" : __ITER % 5 === 0 ? "WARN" : "INFO";
  const payload = JSON.stringify({
    service: "k6-load-test",
    environment: "development",
    level,
    message: `k6 log ${__ITER}`,
    host: `vu-${__VU}`,
    trace_id: `k6-${__VU}-${__ITER}`,
    metadata: {
      iteration: __ITER
    }
  });

  const response = http.post(`${__ENV.LOGMESH_API_URL || "http://localhost:8081"}/v1/logs`, payload, {
    headers: {
      "Content-Type": "application/json"
    }
  });

  check(response, {
    "accepted": (res) => res.status === 202
  });

  sleep(0.1);
}
