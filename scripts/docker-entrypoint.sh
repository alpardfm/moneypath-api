#!/bin/sh

set -eu

if [ "${AUTO_MIGRATE:-false}" = "true" ]; then
  ATTEMPT=1
  MAX_ATTEMPTS=15

  while true; do
    if /usr/local/bin/moneypath-migrate -path "${MIGRATIONS_PATH:-file:///app/migrations}" up; then
      break
    fi

    if [ "${ATTEMPT}" -ge "${MAX_ATTEMPTS}" ]; then
      echo "migration failed after ${MAX_ATTEMPTS} attempts"
      exit 1
    fi

    echo "migration attempt ${ATTEMPT} failed, retrying..."
    ATTEMPT=$((ATTEMPT + 1))
    sleep 2
  done
fi

exec /usr/local/bin/moneypath-api
