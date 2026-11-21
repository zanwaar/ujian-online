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
: "${BACKUP_DIR:=$PROJECT_ROOT/backups}"

timestamp="$(date -u +%Y%m%dT%H%M%SZ)"
backup_file="${BACKUP_FILE:-$BACKUP_DIR/${DB_NAME}_${timestamp}.dump}"
mkdir -p "$(dirname "$backup_file")"

export PGPASSWORD="${DB_PASSWORD:-}"
export PGSSLMODE="$DB_SSLMODE"
pg_dump \
  --format=custom \
  --file="$backup_file" \
  --host="$DB_HOST" \
  --port="$DB_PORT" \
  --username="$DB_USER" \
  --dbname="$DB_NAME" \
  --no-owner \
  --no-privileges \
  --no-comments \
  --verbose >/dev/null

pg_restore --list "$backup_file" >/dev/null
chmod 600 "$backup_file"
printf '%s\n' "$backup_file"
