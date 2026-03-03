#!/usr/bin/env sh

set -e

if [ -z "$1" ]; then
  echo "❌ Migration name is required"
  echo "👉 Usage: ./create_migration.sh add_users_table"
  exit 1
fi

MIGRATION_NAME=$1
MIGRATIONS_DIR="."

echo "🚀 Creating migration: $MIGRATION_NAME"

migrate create \
  -ext sql \
  -dir "$MIGRATIONS_DIR" \
  -seq \
  "$MIGRATION_NAME"

echo "✅ Migration created in $MIGRATIONS_DIR"
