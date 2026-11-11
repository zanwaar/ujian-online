import http from "k6/http"
import { check } from "k6"

const baseURL = __ENV.BASE_URL || "http://127.0.0.1:8080"
const testMode = __ENV.TEST || "login"
const examNumber = __ENV.EXAM_NUMBER || "DEMO-002"
const password = __ENV.PASSWORD || "password"
const attemptID = __ENV.ATTEMPT_ID
const questionID = __ENV.QUESTION_ID

export const options = {
  scenarios: {
    api: {
      executor: "constant-arrival-rate",
      rate: Number(__ENV.RATE || 4),
      timeUnit: "1s",
      duration: __ENV.DURATION || "10s",
      preAllocatedVUs: Number(__ENV.VUS || 10),
      maxVUs: Number(__ENV.MAX_VUS || 20),
    },
  },
  thresholds: {
    http_req_failed: ["rate<0.01"],
    http_req_duration: ["p(95)<500"],
  },
}

export default function () {
  if (testMode === "login") {
    const response = http.post(
      `${baseURL}/api/v1/auth/candidate/login`,
      JSON.stringify({ exam_number: examNumber, password }),
      { headers: { "Content-Type": "application/json" } },
    )
    check(response, {
      "login returns 200": (result) => result.status === 200,
      "login returns token": (result) => result.body.includes('"token"'),
    })
    return
  }

  if (testMode !== "autosave" || !attemptID || !questionID || !__ENV.TOKEN) {
    throw new Error("autosave requires TOKEN, ATTEMPT_ID, and QUESTION_ID")
  }

  const response = http.put(
    `${baseURL}/api/v1/candidate/attempts/${attemptID}/answers/${questionID}`,
    JSON.stringify({ selected_option_key: __ENV.OPTION_KEY || "A" }),
    {
      headers: {
        Authorization: `Bearer ${__ENV.TOKEN}`,
        "Content-Type": "application/json",
      },
    },
  )
  check(response, {
    "autosave returns 200": (result) => result.status === 200,
  })
}
