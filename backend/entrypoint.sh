#!/bin/sh
set -euo pipefail

# 环境变量
DB_URL="${DATABASE_URL:-}"
DB_HOST="${POSTGRES_HOST:-postgres}"
DB_PORT="${POSTGRES_PORT:-5432}"

if [ -z "$DB_URL" ]; then
  echo "ERROR: DATABASE_URL 未设置，例如 postgres://user:password@postgres:5432/guardian?sslmode=disable" >&2
  exit 1
fi

echo "[entrypoint] 等待 PostgreSQL $DB_HOST:$DB_PORT 就绪..."
for i in $(seq 1 60); do
  if nc -z "$DB_HOST" "$DB_PORT" >/dev/null 2>&1; then
    echo "[entrypoint] PostgreSQL 已就绪"
    break
  fi
  sleep 1
done

echo "[entrypoint] 执行数据库迁移..."
/app/migrate -path /app/migrations -database "$DB_URL" up

echo "[entrypoint] 启动后端服务..."
exec /app/guardian


