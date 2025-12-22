#!/bin/sh
set -e

echo "Starting crawler application..."
export PGPASSWORD="${DB_PASSWORD}"

# Wait for PostgreSQL to be ready
echo "Waiting for PostgreSQL..."
until pg_isready -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" 2>/dev/null; do
  echo "PostgreSQL is unavailable - sleeping"
  sleep 2
done

echo "PostgreSQL is up!"

# Run migrations if migration files exist
if [ -d "/root/migrations" ] && [ "$(ls -A /root/migrations/*.up.sql 2>/dev/null)" ]; then
  echo "Running database migrations..."
  for file in /root/migrations/*.up.sql; do
    if [ -f "$file" ]; then
      echo "Executing migration: $(basename "$file")"
      psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f "$file" || true
    fi
  done
  echo "Migrations completed."
fi

# Run the application
exec ./crawler

