#!/bin/sh

set -e

echo "Waiting for postgres..."

until PGPASSWORD=$DATABASE_PASSWORD psql -h postgres -p 5432 -U postgres -d cshotel -c '\q'; do
  >&2 echo "Postgres is unavailable - sleeping"
  sleep 1
done

echo "Postgres is up - executing command"
echo "Running: $@"
exec "$@"