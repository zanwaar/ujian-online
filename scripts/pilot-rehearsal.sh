#!/usr/bin/env bash

set -euo pipefail

BASE_URL="${BASE_URL:-http://127.0.0.1:8080/api/v1}"
ASSIGNMENT_ID="${ASSIGNMENT_ID:?ASSIGNMENT_ID is required}"
CONFIRM_PILOT="${CONFIRM_PILOT:-}"
CANDIDATE_TOKEN="${CANDIDATE_TOKEN:-}"
CANDIDATE_EXAM_NUMBER="${CANDIDATE_EXAM_NUMBER:-}"
CANDIDATE_PASSWORD="${CANDIDATE_PASSWORD:-}"
REQUEST_TIMEOUT="${REQUEST_TIMEOUT:-15}"
TEMP_DIR="$(mktemp -d)"
RESPONSE_FILE="$TEMP_DIR/response.json"
trap 'rm -rf "$TEMP_DIR"' EXIT

if [[ "$CONFIRM_PILOT" != "YES" ]]; then
  printf '%s\n' 'Refusing pilot rehearsal: set CONFIRM_PILOT=YES to continue.' >&2
  exit 1
fi

if [[ -z "$CANDIDATE_TOKEN" && ( -z "$CANDIDATE_EXAM_NUMBER" || -z "$CANDIDATE_PASSWORD" ) ]]; then
  printf '%s\n' 'Set CANDIDATE_TOKEN or both CANDIDATE_EXAM_NUMBER and CANDIDATE_PASSWORD.' >&2
  exit 1
fi

request() {
  local method="$1"
  local path="$2"
  local expected_status="$3"
  local body="${4:-}"
  local http_status
  local -a curl_args=(
    --silent
    --show-error
    --connect-timeout "$REQUEST_TIMEOUT"
    --max-time "$REQUEST_TIMEOUT"
    --output "$RESPONSE_FILE"
    --write-out '%{http_code}'
    --request "$method"
    "$BASE_URL$path"
    --header 'Accept: application/json'
  )

  if [[ -n "$CANDIDATE_TOKEN" ]]; then
    curl_args+=(--header "Authorization: Bearer $CANDIDATE_TOKEN")
  fi
  if [[ -n "$body" ]]; then
    curl_args+=(--header 'Content-Type: application/json' --data "$body")
  fi

  http_status="$(curl "${curl_args[@]}")"
  if [[ "$http_status" != "$expected_status" ]]; then
    printf 'Unexpected HTTP status for %s %s: got %s, expected %s\n' "$method" "$path" "$http_status" "$expected_status" >&2
    cat "$RESPONSE_FILE" >&2
    exit 1
  fi
  jq -e '.success == true' "$RESPONSE_FILE" >/dev/null
}

if [[ -z "$CANDIDATE_TOKEN" ]]; then
  request POST /auth/candidate/login 200 "$(jq -nc --arg exam_number "$CANDIDATE_EXAM_NUMBER" --arg password "$CANDIDATE_PASSWORD" '{exam_number: $exam_number, password: $password}')"
  CANDIDATE_TOKEN="$(jq -r '.data.token // empty' "$RESPONSE_FILE")"
  if [[ -z "$CANDIDATE_TOKEN" ]]; then
    printf '%s\n' 'Candidate login succeeded but returned no token.' >&2
    exit 1
  fi
fi

request GET /candidate/exams 200
schedule="$(jq -c --arg assignment_id "$ASSIGNMENT_ID" '.data[] | select(.assignment_id == $assignment_id)' "$RESPONSE_FILE")"
if [[ -z "$schedule" ]]; then
  printf 'Assignment not found in candidate schedule: %s\n' "$ASSIGNMENT_ID" >&2
  exit 1
fi

exam_id="$(jq -r '.exam_id' <<<"$schedule")"
attempt_id="$(jq -r '.attempt_id // empty' <<<"$schedule")"
attempt_status="$(jq -r '.attempt_status // empty' <<<"$schedule")"
if [[ "$attempt_status" == "submitted" || "$attempt_status" == "expired" ]]; then
  printf 'Assignment %s already has a completed attempt; use a fresh assignment.\n' "$ASSIGNMENT_ID" >&2
  exit 1
fi

if [[ "$attempt_status" != "in_progress" || -z "$attempt_id" ]]; then
  encoded_assignment_id="$(jq -nr --arg value "$ASSIGNMENT_ID" '$value | @uri')"
  request POST "/candidate/exams/$exam_id/attempts?assignment_id=$encoded_assignment_id" 200
  attempt_id="$(jq -r '.data.id // empty' "$RESPONSE_FILE")"
fi
if [[ -z "$attempt_id" ]]; then
  printf '%s\n' 'Start/resume response returned no attempt ID.' >&2
  exit 1
fi

request GET "/candidate/attempts/$attempt_id" 200
resumed_attempt_id="$(jq -r '.data.id // empty' "$RESPONSE_FILE")"
resumed_status="$(jq -r '.data.status // empty' "$RESPONSE_FILE")"
if [[ "$resumed_attempt_id" != "$attempt_id" || "$resumed_status" != "in_progress" ]]; then
  printf 'Resume check failed: attempt=%s status=%s\n' "$resumed_attempt_id" "$resumed_status" >&2
  exit 1
fi

question_id="$(jq -r '.data.questions[0].id // empty' "$RESPONSE_FILE")"
option_key="$(jq -r '.data.questions[0].options[0].option_key // empty' "$RESPONSE_FILE")"
if [[ -z "$question_id" || -z "$option_key" ]]; then
  printf '%s\n' 'Resume response returned no question or option.' >&2
  exit 1
fi

request PUT "/candidate/attempts/$attempt_id/answers/$question_id" 200 "$(jq -nc --arg option_key "$option_key" '{selected_option_key: $option_key}')"
request GET "/candidate/attempts/$attempt_id" 200
jq -e --arg option_key "$option_key" '.data.questions[0].selected_option_key == $option_key' "$RESPONSE_FILE" >/dev/null

request POST "/candidate/attempts/$attempt_id/submit" 200
jq -e '.data.status == "submitted"' "$RESPONSE_FILE" >/dev/null
request POST "/candidate/attempts/$attempt_id/submit" 200
jq -e '.data.status == "submitted"' "$RESPONSE_FILE" >/dev/null

request GET "/candidate/attempts/$attempt_id/result" 200
jq -e '.data.status == "submitted"' "$RESPONSE_FILE" >/dev/null

show_result="$(jq -r '.data.show_result' "$RESPONSE_FILE")"
printf 'Pilot rehearsal passed: assignment=%s attempt=%s show_result=%s\n' "$ASSIGNMENT_ID" "$attempt_id" "$show_result"
