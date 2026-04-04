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

compose_cmd() {
  if sudo docker compose version >/dev/null 2>&1; then
    echo "sudo docker compose"
    return
  fi

  if sudo docker-compose version >/dev/null 2>&1; then
    echo "sudo docker-compose"
    return
  fi

  echo "docker compose is not installed"
  exit 1
}

COMPOSE="$(compose_cmd)"

# docker-compose v1 on Ubuntu often crashes with KeyError: 'ContainerConfig'
# during recreate flows. Removing the API container first avoids that buggy path
# while keeping the postgres volume intact.
sudo docker ps -aq --filter "name=moneypath-api" | xargs -r sudo docker rm -f
sudo docker ps -aq --filter "name=_moneypath-api" | xargs -r sudo docker rm -f

$COMPOSE --env-file .env.vps -f docker-compose.vps.yml up -d --build

for attempt in $(seq 1 20); do
  if curl -fsS http://127.0.0.1:18080/health >/dev/null 2>&1; then
    echo "deployment healthy"
    exit 0
  fi

  sleep 3
done

echo "deployment health check failed"
$COMPOSE --env-file .env.vps -f docker-compose.vps.yml ps
$COMPOSE --env-file .env.vps -f docker-compose.vps.yml logs api --tail=100
exit 1
