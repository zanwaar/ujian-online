#!/usr/bin/env bash

set -euo pipefail

BASE_URL="${BASE_URL:-http://127.0.0.1:8080}"
LOG_FILE="${LOG_FILE:-logs/server.log}"
REQUIRE_LOGS="${REQUIRE_LOGS:-true}"
TEMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TEMP_DIR"' EXIT

check_endpoint() {
  local path="$1"
  local expected_status="$2"
  local expected_body_status="$3"
  local body_file="$TEMP_DIR/body"
  local header_file="$TEMP_DIR/headers"
  local http_status

  http_status="$(curl -sS -o "$body_file" -D "$header_file" -w '%{http_code}' "$BASE_URL$path")"
  if [[ "$http_status" != "$expected_status" ]]; then
    printf '%s returned HTTP %s, expected %s\n' "$path" "$http_status" "$expected_status" >&2
    cat "$body_file" >&2
    exit 1
  fi
  jq -e --arg expected "$expected_body_status" '.success == true and .data.status == $expected and (.data.request_id | type == "string" and length > 0)' "$body_file" >/dev/null
  if ! awk 'tolower($1) == "x-request-id:" {print $2}' "$header_file" | tr -d '\r' | grep -q '[^[:space:]]'; then
    printf '%s did not return X-Request-ID\n' "$path" >&2
    exit 1
  fi
  printf '%s OK\n' "$path"
}

check_endpoint "/health" 200 "ok"
check_endpoint "/ready" 200 "ready"

if [[ "$REQUIRE_LOGS" == "true" ]]; then
  if [[ ! -f "$LOG_FILE" ]]; then
    printf 'Log file not found: %s\n' "$LOG_FILE" >&2
    exit 1
  fi
  request_log="$(rg '"msg":"http_request"' "$LOG_FILE" | tail -1 || true)"
  if [[ -z "$request_log" ]]; then
    printf 'No structured http_request log found in %s\n' "$LOG_FILE" >&2
    exit 1
  fi
  printf '%s\n' "$request_log" | jq -e '(.request_id | type == "string" and length > 0) and (.path | type == "string") and (.status | type == "number") and (.latency_ms | type == "number")' >/dev/null
  printf '%s OK\n' "$LOG_FILE"
fi
