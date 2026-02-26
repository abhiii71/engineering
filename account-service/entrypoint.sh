#!/bin/bash
set -e

# PostgreSQL bin path (Ubuntu puts it under /usr/lib/postgresql/<ver>/bin)
PGBIN="${PGBIN:-$(ls -d /usr/lib/postgresql/*/bin 2>/dev/null | head -1)}"
PGDATA="${PGDATA:-/app/pgdata}"
export DATABASE_URL="${DATABASE_URL:-postgres://account:secret@localhost:5432/accountdb?sslmode=disable}"

# 1. Init PostgreSQL data dir if first run
if [ ! -f "$PGDATA/PG_VERSION" ]; then
  mkdir -p "$PGDATA"
  chown postgres:postgres "$PGDATA"
  su postgres -c "$PGBIN/initdb -D $PGDATA"
  echo "listen_addresses = 'localhost'" >> "$PGDATA/postgresql.conf"
fi

# 2. Start PostgreSQL
su postgres -c "$PGBIN/pg_ctl -D $PGDATA -l $PGDATA/logfile start"

# 3. Wait until Postgres is accepting connections
until su postgres -c "$PGBIN/pg_isready -d postgres" 2>/dev/null; do sleep 1; done

# 4. Create user and database if not exist
su postgres -c "psql -d postgres -tAc \"SELECT 1 FROM pg_roles WHERE rolname='account'\"" | grep -q 1 || \
  su postgres -c "psql -d postgres -c \"CREATE USER account WITH PASSWORD 'secret' CREATEDB;\""
su postgres -c "psql -d postgres -tAc \"SELECT 1 FROM pg_database WHERE datname='accountdb'\"" | grep -q 1 || \
  su postgres -c "psql -d postgres -c \"CREATE DATABASE accountdb OWNER account;\""

# 5. Run migration
export PGPASSWORD=secret
psql -h localhost -U account -d accountdb -f /app/migrations/000001_create_accounts_table.up.sql 2>/dev/null || true
psql -h localhost -U account -d accountdb -f /app/migrations/000002_create_transactions_and_activity_log.up.sql 2>/dev/null || true

# 6. Run the account service (foreground)
exec /app/account-service
