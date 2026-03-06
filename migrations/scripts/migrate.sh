#!/usr/bin/env sh
set -e

if [ -z "$DATABASE_URL" ]; then
  echo "DATABASE_URL is not set"
  exit 1
fi

if ! command -v migrate >/dev/null 2>&1; then
  echo "migrate command is required"
  exit 1
fi

if ! command -v psql >/dev/null 2>&1; then
  echo "psql command is required"
  exit 1
fi

DB_SCHEMA="${DB_SCHEMA:-ums}"
case "$DB_SCHEMA" in
  *[!A-Za-z0-9_]*|"")
    echo "DB_SCHEMA must match [A-Za-z0-9_]+"
    exit 1
    ;;
esac

ACTION=${1:-up}
STEPS=$2

MIGRATIONS_DIR="."

echo "Running migration"
echo "Dir: $MIGRATIONS_DIR"
echo "Schema: $DB_SCHEMA"

case "$DATABASE_URL" in
  *"search_path="*)
    DATABASE_URL_MIGRATE="$DATABASE_URL"
    ;;
  *\?*)
    DATABASE_URL_MIGRATE="${DATABASE_URL}&search_path=${DB_SCHEMA}%2Cpublic"
    ;;
  *)
    DATABASE_URL_MIGRATE="${DATABASE_URL}?search_path=${DB_SCHEMA}%2Cpublic"
    ;;
esac

psql -X "$DATABASE_URL" -v ON_ERROR_STOP=1 -c "CREATE SCHEMA IF NOT EXISTS \"$DB_SCHEMA\";" >/dev/null

case "$ACTION" in
  up)
    migrate -path "$MIGRATIONS_DIR" -database "$DATABASE_URL_MIGRATE" up
    ;;
  down)
    migrate -path "$MIGRATIONS_DIR" -database "$DATABASE_URL_MIGRATE" down
    ;;
  step)
    if [ -z "$STEPS" ]; then
      echo "Steps required"
      exit 1
    fi
    migrate -path "$MIGRATIONS_DIR" -database "$DATABASE_URL_MIGRATE" up "$STEPS"
    ;;
  force)
    if [ -z "$STEPS" ]; then
      echo "Need version"
      exit 1
    fi
    migrate -path "$MIGRATIONS_DIR" -database "$DATABASE_URL_MIGRATE" force "$STEPS"
    ;;
  version)
    migrate -path "$MIGRATIONS_DIR" -database "$DATABASE_URL_MIGRATE" version
    ;;
  *)
    echo "Unknown action"
    exit 1
    ;;
esac

echo "Done"
