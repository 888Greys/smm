#!/usr/bin/env bash
# Test configured SMM providers from the server.
#
# Default mode is read-only:
#   - checks balances
#   - lists TikTok follower candidate services
#   - optionally checks existing order IDs
#
# Live order mode:
#   PLACE_TEST_ORDERS=1 TEST_LINK='https://www.tiktok.com/@yourhandle' bash deploy/test_providers.sh

set -euo pipefail

TEST_LINK="${TEST_LINK:-}"
TEST_QTY="${TEST_QTY:-10}"
PLACE_TEST_ORDERS="${PLACE_TEST_ORDERS:-0}"

SMMWIZ_API_URL="${SMMWIZ_API_URL:-https://smmwiz.com/api/v2}"
MTP_API_URL="${MTP_API_URL:-https://morethanpanel.com/api/v2}"
MTP_API_KEY="${MTP_API_KEY:-${MORETHANPANEL_API_KEY:-}}"

SMMWIZ_TEST_SERVICE="${SMMWIZ_TEST_SERVICE:-18612}"
MTP_TEST_SERVICE="${MTP_TEST_SERVICE:-9408}"
MTP_TEST_QTY="${MTP_TEST_QTY:-100}"

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "missing command: $1" >&2
    exit 1
  }
}

post() {
  local url="$1"
  local key="$2"
  local action="$3"
  shift 3
  curl -fsS -X POST "$url" -d "key=$key&action=$action" "$@"
}

print_json_line() {
  python3 -c 'import sys,json; print(json.dumps(json.load(sys.stdin), ensure_ascii=False))'
}

provider_balance() {
  local name="$1"
  local url="$2"
  local key="$3"
  echo
  echo "== $name balance =="
  post "$url" "$key" balance | print_json_line
}

provider_tiktok_services() {
  local name="$1"
  local url="$2"
  local key="$3"
  echo
  echo "== $name TikTok follower candidates =="
  post "$url" "$key" services | python3 -c '
import sys,json
services=json.load(sys.stdin)
rows=[]
for s in services:
    text=((s.get("name","") or "")+" "+(s.get("category","") or "")).lower()
    if "tiktok" in text and "follower" in text:
        rows.append(s)
for s in rows[:80]:
    print("{} | rate={} | min={} | max={} | {}".format(
        s.get("service"),
        s.get("rate"),
        s.get("min"),
        s.get("max"),
        s.get("name"),
    ))
if len(rows) > 80:
    print(f"... {len(rows)-80} more omitted")
'
}

provider_order_status() {
  local name="$1"
  local url="$2"
  local key="$3"
  local order="$4"
  if [[ -z "$order" ]]; then
    return
  fi
  echo
  echo "== $name order $order status =="
  post "$url" "$key" status -d "order=$order" | print_json_line
}

provider_place_test_order() {
  local name="$1"
  local url="$2"
  local key="$3"
  local service="$4"
  local qty="$5"

  if [[ "$PLACE_TEST_ORDERS" != "1" ]]; then
    return
  fi
  if [[ -z "$TEST_LINK" ]]; then
    echo "PLACE_TEST_ORDERS=1 requires TEST_LINK" >&2
    exit 1
  fi

  echo
  echo "== $name placing test order =="
  echo "service=$service qty=$qty link=$TEST_LINK"
  post "$url" "$key" add \
    -d "service=$service" \
    --data-urlencode "link=$TEST_LINK" \
    -d "quantity=$qty" | print_json_line
}

main() {
  require_cmd curl
  require_cmd python3

  if [[ -f /opt/smm/.env ]]; then
    set -a
    # shellcheck disable=SC1091
    source /opt/smm/.env
    set +a
  fi

  MTP_API_KEY="${MTP_API_KEY:-${MORETHANPANEL_API_KEY:-}}"

  echo "Provider test started at $(date -Is)"
  echo "Read-only mode: $([[ "$PLACE_TEST_ORDERS" == "1" ]] && echo no || echo yes)"

  if [[ -n "${SMMWIZ_API_KEY:-}" ]]; then
    provider_balance "SMMWiz" "$SMMWIZ_API_URL" "$SMMWIZ_API_KEY"
    provider_order_status "SMMWiz" "$SMMWIZ_API_URL" "$SMMWIZ_API_KEY" "${SMMWIZ_ORDER_ID:-}"
    provider_tiktok_services "SMMWiz" "$SMMWIZ_API_URL" "$SMMWIZ_API_KEY"
    provider_place_test_order "SMMWiz" "$SMMWIZ_API_URL" "$SMMWIZ_API_KEY" "$SMMWIZ_TEST_SERVICE" "$TEST_QTY"
  else
    echo
    echo "== SMMWiz skipped: SMMWIZ_API_KEY not set =="
  fi

  if [[ -n "$MTP_API_KEY" ]]; then
    provider_balance "MoreThanPanel" "$MTP_API_URL" "$MTP_API_KEY"
    provider_order_status "MoreThanPanel" "$MTP_API_URL" "$MTP_API_KEY" "${MTP_ORDER_ID:-}"
    provider_tiktok_services "MoreThanPanel" "$MTP_API_URL" "$MTP_API_KEY"
    provider_place_test_order "MoreThanPanel" "$MTP_API_URL" "$MTP_API_KEY" "$MTP_TEST_SERVICE" "$MTP_TEST_QTY"
  else
    echo
    echo "== MoreThanPanel skipped: MTP_API_KEY or MORETHANPANEL_API_KEY not set =="
  fi

  echo
  echo "Provider test finished at $(date -Is)"
}

main "$@"
