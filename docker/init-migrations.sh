#!/bin/sh
set -e

echo "Waiting for PostgreSQL to be ready..."
until pg_isready -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER"; do
  sleep 1
done

echo "PostgreSQL is ready. Running migrations..."
export PGPASSWORD="${DB_PASSWORD}"

# Execute SQL files in migrations directory
for file in /root/migrations/*.up.sql; do
  if [ -f "$file" ]; then
    echo "Running migration: $file"
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f "$file"
  fi
done

echo "Migrations completed."

