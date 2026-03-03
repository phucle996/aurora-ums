#!/usr/bin/env sh
set -e

if [ -z "$DATABASE_URL" ]; then
  echo "DATABASE_URL is not set"
  exit 1
fi

ACTION=${1:-up}
STEPS=$2

MIGRATIONS_DIR="."

echo "Running migration"
echo "Dir: $MIGRATIONS_DIR"

case "$ACTION" in
  up)
    migrate -path "$MIGRATIONS_DIR" -database "$DATABASE_URL" up
    ;;
  down)
    migrate -path "$MIGRATIONS_DIR" -database "$DATABASE_URL" down
    ;;
  step)
    if [ -z "$STEPS" ]; then
      echo "Steps required"
      exit 1
    fi
    migrate -path "$MIGRATIONS_DIR" -database "$DATABASE_URL" up "$STEPS"
    ;;
  force)
    if [ -z "$STEPS" ]; then
      echo "Need version"
      exit 1
    fi
    migrate -path "$MIGRATIONS_DIR" -database "$DATABASE_URL" force "$STEPS"
    ;;
  version)
    migrate -path "$MIGRATIONS_DIR" -database "$DATABASE_URL" version
    ;;
  *)
    echo "Unknown action"
    exit 1
    ;;
esac

echo "Done"
