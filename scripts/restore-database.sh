#!/usr/bin/env bash

set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ENV_FILE="${ENV_FILE:-$PROJECT_ROOT/.env}"

if [[ -f "$ENV_FILE" ]]; then
  set -a
  source "$ENV_FILE"
  set +a
fi

: "${DB_HOST:=localhost}"
: "${DB_PORT:=5432}"
: "${DB_USER:=postgres}"
: "${DB_NAME:?DB_NAME is required}"
: "${DB_SSLMODE:=disable}"
: "${BACKUP_FILE:?BACKUP_FILE is required}"
: "${TARGET_DB_NAME:=$DB_NAME}"

if [[ "${CONFIRM_RESTORE:-}" != "YES" ]]; then
  printf '%s\n' 'Refusing restore: set CONFIRM_RESTORE=YES to continue.' >&2
  exit 1
fi

if [[ "$TARGET_DB_NAME" == "$DB_NAME" && "${ALLOW_DESTRUCTIVE_RESTORE:-}" != "YES" ]]; then
  printf '%s\n' 'Refusing restore into source database: set ALLOW_DESTRUCTIVE_RESTORE=YES or use TARGET_DB_NAME.' >&2
  exit 1
fi

if [[ ! -f "$BACKUP_FILE" ]]; then
  printf 'Backup file not found: %s\n' "$BACKUP_FILE" >&2
  exit 1
fi

export PGPASSWORD="${DB_PASSWORD:-}"
export PGSSLMODE="$DB_SSLMODE"
pg_restore \
  --clean \
  --if-exists \
  --exit-on-error \
  --no-owner \
  --no-privileges \
  --host="$DB_HOST" \
  --port="$DB_PORT" \
  --username="$DB_USER" \
  --dbname="$TARGET_DB_NAME" \
  "$BACKUP_FILE"

printf 'Restored %s into %s\n' "$BACKUP_FILE" "$TARGET_DB_NAME"
