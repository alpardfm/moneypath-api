#!/usr/bin/env bash

set -euo pipefail

cd "$(dirname "$0")/.."

if [ ! -f ".env.vps" ]; then
  echo "missing .env.vps in $(pwd)"
  exit 1
fi

if [ ! -f "docker-compose.vps.yml" ]; then
  echo "missing docker-compose.vps.yml in $(pwd)"
  exit 1
fi

sudo docker-compose --env-file .env.vps -f docker-compose.vps.yml up -d --build

for attempt in $(seq 1 20); do
  if curl -fsS http://127.0.0.1:18080/health >/dev/null 2>&1; then
    echo "deployment healthy"
    exit 0
  fi

  sleep 3
done

echo "deployment health check failed"
sudo docker-compose --env-file .env.vps -f docker-compose.vps.yml ps
sudo docker-compose --env-file .env.vps -f docker-compose.vps.yml logs api --tail=100
exit 1
