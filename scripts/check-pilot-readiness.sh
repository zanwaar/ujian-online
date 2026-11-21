#!/usr/bin/env bash

set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ENV_FILE="${ENV_FILE:-$PROJECT_ROOT/.env}"
BASE_URL="${BASE_URL:-http://127.0.0.1:8080}"
LOG_FILE="${LOG_FILE:-$PROJECT_ROOT/logs/server.log}"
REQUIRE_LOGS="${REQUIRE_LOGS:-true}"
REQUIRE_CONFIG="${REQUIRE_CONFIG:-true}"
REQUIRE_BACKUP="${REQUIRE_BACKUP:-false}"
EXPECTED_ENVIRONMENT="${EXPECTED_ENVIRONMENT:-}"
ALLOW_DEVELOPMENT_ENV="${ALLOW_DEVELOPMENT_ENV:-}"

command -v curl >/dev/null || { printf '%s\n' 'curl is required.' >&2; exit 1; }
command -v jq >/dev/null || { printf '%s\n' 'jq is required.' >&2; exit 1; }

if [[ ! -x "$PROJECT_ROOT/scripts/check-observability.sh" ]]; then
  printf '%s\n' 'check-observability.sh is missing or not executable.' >&2
  exit 1
fi

if [[ "$REQUIRE_CONFIG" == "true" ]]; then
  if [[ ! -f "$ENV_FILE" ]]; then
    printf 'Environment file not found: %s\n' "$ENV_FILE" >&2
    exit 1
  fi

  set -a
  source "$ENV_FILE"
  set +a

  : "${DB_NAME:?DB_NAME is required in ENV_FILE}"
  : "${JWT_SECRET:?JWT_SECRET is required in ENV_FILE}"
  : "${CORS_ALLOWED_ORIGINS:?CORS_ALLOWED_ORIGINS is required in ENV_FILE}"
  ATTEMPT_EXPIRY_INTERVAL_SECONDS="${ATTEMPT_EXPIRY_INTERVAL_SECONDS:-5}"

  if [[ "$DB_NAME" == "mydb" || "$DB_NAME" == "your_database_name" ]]; then
    printf 'DB_NAME still uses a placeholder: %s\n' "$DB_NAME" >&2
    exit 1
  fi
  jwt_placeholder='your_super_secret_jwt_key_change_this_in_production_12345'
  if [[ "$JWT_SECRET" == "$jwt_placeholder" || ${#JWT_SECRET} -lt 32 ]]; then
    printf '%s\n' 'JWT_SECRET is empty, a placeholder, or shorter than 32 characters.' >&2
    exit 1
  fi
  if [[ "$CORS_ALLOWED_ORIGINS" == '*' ]]; then
    printf '%s\n' 'CORS_ALLOWED_ORIGINS must not be wildcard for a pilot.' >&2
    exit 1
  fi
  if ! [[ "$ATTEMPT_EXPIRY_INTERVAL_SECONDS" =~ ^[1-9][0-9]*$ ]]; then
    printf 'ATTEMPT_EXPIRY_INTERVAL_SECONDS must be a positive integer: %s\n' "$ATTEMPT_EXPIRY_INTERVAL_SECONDS" >&2
    exit 1
  fi
  if [[ -n "$EXPECTED_ENVIRONMENT" && "${ENVIRONMENT:-}" != "$EXPECTED_ENVIRONMENT" ]]; then
    printf 'ENVIRONMENT=%s, expected %s.\n' "${ENVIRONMENT:-}" "$EXPECTED_ENVIRONMENT" >&2
    exit 1
  fi
  if [[ "${ENVIRONMENT:-}" == "development" && "$ALLOW_DEVELOPMENT_ENV" != "YES" ]]; then
    printf '%s\n' 'Development environment is blocked; set ALLOW_DEVELOPMENT_ENV=YES only for local rehearsal.' >&2
    exit 1
  fi
fi

BASE_URL="$BASE_URL" LOG_FILE="$LOG_FILE" REQUIRE_LOGS="$REQUIRE_LOGS" \
  "$PROJECT_ROOT/scripts/check-observability.sh"

if [[ "$REQUIRE_BACKUP" == "true" ]]; then
  : "${BACKUP_FILE:?BACKUP_FILE is required when REQUIRE_BACKUP=true}"
  if [[ ! -f "$BACKUP_FILE" ]]; then
    printf 'Backup file not found: %s\n' "$BACKUP_FILE" >&2
    exit 1
  fi
  command -v pg_restore >/dev/null || { printf '%s\n' 'pg_restore is required to validate BACKUP_FILE.' >&2; exit 1; }
  pg_restore --list "$BACKUP_FILE" >/dev/null
  printf 'Backup OK: %s\n' "$BACKUP_FILE"
fi

printf 'Pilot readiness passed: %s\n' "$BASE_URL"
