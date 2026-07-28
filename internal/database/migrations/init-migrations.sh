#!/bin/bash
# This script runs only .up.sql migration files in order.
# PostgreSQL docker-entrypoint-initdb.d would run ALL .sql files alphabetically,
# including .down.sql files which should not run during initialization.

set -e

MIGRATIONS_DIR="/migrations"

# Find all .up.sql files, sort them, and execute
for migration in $(find "$MIGRATIONS_DIR" -name "*.up.sql" | sort); do
    echo "Running migration: $(basename "$migration")"
    psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" -f "$migration"
done

echo "All migrations completed successfully."
