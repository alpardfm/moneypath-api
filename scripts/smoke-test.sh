#!/bin/sh

set -eu

BASE_URL="${BASE_URL:-http://localhost:18080}"
SUFFIX="$(date +%s)"
EMAIL="phase9-${SUFFIX}@example.com"
USERNAME="phase9_${SUFFIX}"
PASSWORD="password123"
FULL_NAME="Phase Nine"
TODAY="$(date +%F)"

require_command() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "required command not found: $1"
    exit 1
  fi
}

extract_json_value() {
  KEY="$1"
  sed -n "s/.*\"${KEY}\":\"\\([^\"]*\\)\".*/\\1/p" | head -n 1
}

assert_contains() {
  BODY="$1"
  EXPECTED="$2"
  if ! printf '%s' "$BODY" | grep -q "$EXPECTED"; then
    echo "expected response to contain: $EXPECTED"
    echo "$BODY"
    exit 1
  fi
}

request() {
  METHOD="$1"
  PATHNAME="$2"
  BODY="${3:-}"
  TOKEN="${4:-}"

  TMP_BODY="$(mktemp)"
  TMP_STATUS="$(mktemp)"

  if [ -n "$BODY" ]; then
    if [ -n "$TOKEN" ]; then
      curl -sS -X "$METHOD" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $TOKEN" \
        "$BASE_URL$PATHNAME" \
        -d "$BODY" \
        -o "$TMP_BODY" \
        -w "%{http_code}" > "$TMP_STATUS"
    else
      curl -sS -X "$METHOD" \
        -H "Content-Type: application/json" \
        "$BASE_URL$PATHNAME" \
        -d "$BODY" \
        -o "$TMP_BODY" \
        -w "%{http_code}" > "$TMP_STATUS"
    fi
  else
    if [ -n "$TOKEN" ]; then
      curl -sS -X "$METHOD" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $TOKEN" \
        "$BASE_URL$PATHNAME" \
        -o "$TMP_BODY" \
        -w "%{http_code}" > "$TMP_STATUS"
    else
      curl -sS -X "$METHOD" \
        -H "Content-Type: application/json" \
        "$BASE_URL$PATHNAME" \
        -o "$TMP_BODY" \
        -w "%{http_code}" > "$TMP_STATUS"
    fi
  fi

  STATUS="$(cat "$TMP_STATUS")"
  RESPONSE_BODY="$(cat "$TMP_BODY")"
  rm -f "$TMP_STATUS" "$TMP_BODY"

  printf '%s\n%s' "$STATUS" "$RESPONSE_BODY"
}

require_command curl

HEALTH_RESULT="$(request GET /health)"
HEALTH_STATUS="$(printf '%s' "$HEALTH_RESULT" | sed -n '1p')"
HEALTH_BODY="$(printf '%s' "$HEALTH_RESULT" | sed -n '2,$p')"
[ "$HEALTH_STATUS" = "200" ] || { echo "health failed"; echo "$HEALTH_BODY"; exit 1; }
assert_contains "$HEALTH_BODY" '"database":"up"'

REGISTER_RESULT="$(request POST /auth/register "{\"email\":\"$EMAIL\",\"username\":\"$USERNAME\",\"password\":\"$PASSWORD\",\"full_name\":\"$FULL_NAME\"}")"
REGISTER_STATUS="$(printf '%s' "$REGISTER_RESULT" | sed -n '1p')"
REGISTER_BODY="$(printf '%s' "$REGISTER_RESULT" | sed -n '2,$p')"
[ "$REGISTER_STATUS" = "201" ] || { echo "register failed"; echo "$REGISTER_BODY"; exit 1; }
TOKEN="$(printf '%s' "$REGISTER_BODY" | extract_json_value token)"
[ -n "$TOKEN" ] || { echo "failed to extract token"; echo "$REGISTER_BODY"; exit 1; }

LOGIN_RESULT="$(request POST /auth/login "{\"email_or_username\":\"$USERNAME\",\"password\":\"$PASSWORD\"}")"
LOGIN_STATUS="$(printf '%s' "$LOGIN_RESULT" | sed -n '1p')"
LOGIN_BODY="$(printf '%s' "$LOGIN_RESULT" | sed -n '2,$p')"
[ "$LOGIN_STATUS" = "200" ] || { echo "login failed"; echo "$LOGIN_BODY"; exit 1; }

WALLET_CREATE_RESULT="$(request POST /wallets "{\"name\":\"Primary Wallet\"}" "$TOKEN")"
WALLET_CREATE_STATUS="$(printf '%s' "$WALLET_CREATE_RESULT" | sed -n '1p')"
WALLET_CREATE_BODY="$(printf '%s' "$WALLET_CREATE_RESULT" | sed -n '2,$p')"
[ "$WALLET_CREATE_STATUS" = "201" ] || { echo "wallet create failed"; echo "$WALLET_CREATE_BODY"; exit 1; }
WALLET_ID="$(printf '%s' "$WALLET_CREATE_BODY" | extract_json_value id)"
[ -n "$WALLET_ID" ] || { echo "failed to extract wallet id"; echo "$WALLET_CREATE_BODY"; exit 1; }

WALLET_LIST_RESULT="$(request GET "/wallets?page=1&page_size=10" "" "$TOKEN")"
WALLET_LIST_STATUS="$(printf '%s' "$WALLET_LIST_RESULT" | sed -n '1p')"
WALLET_LIST_BODY="$(printf '%s' "$WALLET_LIST_RESULT" | sed -n '2,$p')"
[ "$WALLET_LIST_STATUS" = "200" ] || { echo "wallet list failed"; echo "$WALLET_LIST_BODY"; exit 1; }
assert_contains "$WALLET_LIST_BODY" '"total_items":1'

DEBT_CREATE_RESULT="$(request POST /debts "{\"name\":\"Laptop Loan\",\"principal_amount\":\"1200.00\",\"tenor_value\":12,\"tenor_unit\":\"month\",\"payment_amount\":\"100.00\"}" "$TOKEN")"
DEBT_CREATE_STATUS="$(printf '%s' "$DEBT_CREATE_RESULT" | sed -n '1p')"
DEBT_CREATE_BODY="$(printf '%s' "$DEBT_CREATE_RESULT" | sed -n '2,$p')"
[ "$DEBT_CREATE_STATUS" = "201" ] || { echo "debt create failed"; echo "$DEBT_CREATE_BODY"; exit 1; }
DEBT_ID="$(printf '%s' "$DEBT_CREATE_BODY" | extract_json_value id)"
[ -n "$DEBT_ID" ] || { echo "failed to extract debt id"; echo "$DEBT_CREATE_BODY"; exit 1; }

INCOMING_RESULT="$(request POST /mutations "{\"wallet_id\":\"$WALLET_ID\",\"type\":\"masuk\",\"amount\":\"2000.00\",\"description\":\"salary\",\"related_to_debt\":false,\"happened_at\":\"${TODAY}T08:00:00Z\"}" "$TOKEN")"
INCOMING_STATUS="$(printf '%s' "$INCOMING_RESULT" | sed -n '1p')"
INCOMING_BODY="$(printf '%s' "$INCOMING_RESULT" | sed -n '2,$p')"
[ "$INCOMING_STATUS" = "201" ] || { echo "incoming mutation failed"; echo "$INCOMING_BODY"; exit 1; }

OUTGOING_RESULT="$(request POST /mutations "{\"wallet_id\":\"$WALLET_ID\",\"debt_id\":\"$DEBT_ID\",\"type\":\"keluar\",\"amount\":\"100.00\",\"description\":\"installment\",\"related_to_debt\":true,\"happened_at\":\"${TODAY}T09:00:00Z\"}" "$TOKEN")"
OUTGOING_STATUS="$(printf '%s' "$OUTGOING_RESULT" | sed -n '1p')"
OUTGOING_BODY="$(printf '%s' "$OUTGOING_RESULT" | sed -n '2,$p')"
[ "$OUTGOING_STATUS" = "201" ] || { echo "outgoing mutation failed"; echo "$OUTGOING_BODY"; exit 1; }

MUTATION_LIST_RESULT="$(request GET "/mutations?type=keluar&related_to_debt=true&page=1&page_size=10" "" "$TOKEN")"
MUTATION_LIST_STATUS="$(printf '%s' "$MUTATION_LIST_RESULT" | sed -n '1p')"
MUTATION_LIST_BODY="$(printf '%s' "$MUTATION_LIST_RESULT" | sed -n '2,$p')"
[ "$MUTATION_LIST_STATUS" = "200" ] || { echo "mutation list failed"; echo "$MUTATION_LIST_BODY"; exit 1; }
assert_contains "$MUTATION_LIST_BODY" '"type":"keluar"'

DASHBOARD_RESULT="$(request GET /dashboard "" "$TOKEN")"
DASHBOARD_STATUS="$(printf '%s' "$DASHBOARD_RESULT" | sed -n '1p')"
DASHBOARD_BODY="$(printf '%s' "$DASHBOARD_RESULT" | sed -n '2,$p')"
[ "$DASHBOARD_STATUS" = "200" ] || { echo "dashboard failed"; echo "$DASHBOARD_BODY"; exit 1; }
assert_contains "$DASHBOARD_BODY" '"total_assets"'

SUMMARY_RESULT="$(request GET "/summary?from=$TODAY&to=$TODAY" "" "$TOKEN")"
SUMMARY_STATUS="$(printf '%s' "$SUMMARY_RESULT" | sed -n '1p')"
SUMMARY_BODY="$(printf '%s' "$SUMMARY_RESULT" | sed -n '2,$p')"
[ "$SUMMARY_STATUS" = "200" ] || { echo "summary failed"; echo "$SUMMARY_BODY"; exit 1; }
assert_contains "$SUMMARY_BODY" '"net_flow"'

echo "phase 9 smoke test passed"
