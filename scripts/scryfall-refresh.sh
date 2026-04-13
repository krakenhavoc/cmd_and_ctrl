#!/usr/bin/env bash
# scryfall-refresh.sh — download the Scryfall "default_cards" bulk dump
# into $CMDCTRL_DATA_DIR/scryfall/default-cards.json.
#
# The default-cards dataset is the one every card (most recent printing)
# dump; it's what the server's cards.Index loads on boot to resolve
# card metadata + image URIs by UUID.
#
# Intended cron line (Sundays at 05:00 server time):
#     0 5 * * 0  /opt/cmd_and_ctrl/scripts/scryfall-refresh.sh >> /var/log/cmdctrl/scryfall.log 2>&1
#
# Idempotent: rewrites default-cards.json atomically via a tmp file +
# mv, and updates data/scryfall/last-refresh with the current UTC
# timestamp. Exit non-zero on any curl/jq failure — cron will mail the
# operator.
#
# Environment:
#   CMDCTRL_DATA_DIR — root of the server's data dir. Default "./data".
#
# Dependencies: curl, jq.

set -euo pipefail

DATA_DIR="${CMDCTRL_DATA_DIR:-./data}"
SCRYFALL_DIR="$DATA_DIR/scryfall"
TARGET="$SCRYFALL_DIR/default-cards.json"
STAMP="$SCRYFALL_DIR/last-refresh"

mkdir -p "$SCRYFALL_DIR"

# Step 1 — fetch the bulk-data catalog and pull out the download_uri
# for the default_cards dataset. Scryfall rotates the URL daily, so
# we never hardcode it.
catalog_uri="https://api.scryfall.com/bulk-data"
download_uri="$(curl --fail --silent --show-error --max-time 30 "$catalog_uri" \
  | jq -r '.data[] | select(.type == "default_cards") | .download_uri')"

if [[ -z "$download_uri" ]]; then
  echo "scryfall-refresh: could not locate default_cards in bulk-data catalog" >&2
  exit 1
fi

echo "scryfall-refresh: downloading $download_uri"

# Step 2 — stream download into a tmp file under the same directory
# so the atomic rename below is within one filesystem.
tmp="$(mktemp "$SCRYFALL_DIR/default-cards.XXXXXX.json.tmp")"
trap 'rm -f "$tmp"' EXIT

curl --fail --silent --show-error --location --max-time 600 \
     --output "$tmp" \
     "$download_uri"

# Step 3 — sanity check that the downloaded file is a JSON array
# (Scryfall occasionally serves HTML error pages on maintenance).
if ! head -c 1 "$tmp" | grep -q '\['; then
  echo "scryfall-refresh: download is not a JSON array, refusing to publish" >&2
  exit 1
fi

mv -f "$tmp" "$TARGET"
trap - EXIT
date -u +'%Y-%m-%dT%H:%M:%SZ' > "$STAMP"

size="$(stat --format='%s' "$TARGET" 2>/dev/null || stat -f '%z' "$TARGET")"
echo "scryfall-refresh: wrote $TARGET ($size bytes)"
