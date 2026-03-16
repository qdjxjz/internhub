#!/usr/bin/env bash
# 在项目根目录一键启动所有服务（需先 docker compose up -d 启动 PostgreSQL）

set -e
cd "$(dirname "$0")/.."
[ -f .env ] && source .env

PIDS=()
cleanup() {
  echo ""
  echo "正在停止所有服务..."
  for pid in "${PIDS[@]}"; do
    kill "$pid" 2>/dev/null || true
  done
  exit 0
}
trap cleanup SIGINT SIGTERM

echo "启动 api-gateway :8080 ..."
go run ./api-gateway/cmd & PIDS+=($!)
sleep 0.5
echo "启动 auth-service :8081 ..."
go run ./auth-service & PIDS+=($!)
sleep 0.5
echo "启动 user-service :8082 ..."
go run ./user-service/cmd & PIDS+=($!)
sleep 0.5
echo "启动 job-service :8083 ..."
go run ./job-service/cmd & PIDS+=($!)
sleep 0.5
echo "启动 apply-service :8084 ..."
go run ./apply-service/cmd & PIDS+=($!)
sleep 0.5
echo "启动 recommend-service :8085 ..."
go run ./recommend-service/cmd & PIDS+=($!)

echo ""
echo "全部服务已启动。按 Ctrl+C 停止所有服务。"
echo "网关: http://localhost:8080  健康检查: curl http://localhost:8080/health"
wait
