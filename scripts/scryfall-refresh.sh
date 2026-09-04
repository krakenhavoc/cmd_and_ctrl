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
# Freshness: if last-refresh is less than 24 hours old (and the dump
# exists), the download is skipped — the ~400-500 MB dump only changes
# daily. Pass --force to refresh regardless.
#
# Usage: scryfall-refresh.sh [--force]
#
# Environment:
#   CMDCTRL_DATA_DIR — root of the server's data dir. Default "./data".
#
# Dependencies: curl, jq.

set -euo pipefail

FORCE=0
for arg in "$@"; do
  case "$arg" in
    --force) FORCE=1 ;;
    *)
      echo "scryfall-refresh: unknown argument: $arg" >&2
      echo "usage: scryfall-refresh.sh [--force]" >&2
      exit 2
      ;;
  esac
done

DATA_DIR="${CMDCTRL_DATA_DIR:-./data}"
SCRYFALL_DIR="$DATA_DIR/scryfall"
TARGET="$SCRYFALL_DIR/default-cards.json"
STAMP="$SCRYFALL_DIR/last-refresh"

mkdir -p "$SCRYFALL_DIR"

# Freshness check — skip the download when the last refresh was under
# 24h ago and the dump is still present. last-refresh holds a UTC
# ISO-8601 timestamp (see the write at the bottom of this script);
# parse with GNU date, falling back to BSD date for macOS devs. An
# unparseable stamp counts as stale, so we fail open into a refresh.
if [[ "$FORCE" -ne 1 && -s "$TARGET" && -f "$STAMP" ]]; then
  last="$(cat "$STAMP")"
  last_epoch="$(date -u -d "$last" +%s 2>/dev/null \
    || date -j -u -f '%Y-%m-%dT%H:%M:%SZ' "$last" +%s 2>/dev/null \
    || echo 0)"
  now_epoch="$(date -u +%s)"
  age=$(( now_epoch - last_epoch ))
  if [[ "$last_epoch" -gt 0 && "$age" -ge 0 && "$age" -lt 86400 ]]; then
    echo "scryfall-refresh: $TARGET refreshed ${age}s ago (<24h); skipping. Use --force to refresh anyway."
    exit 0
  fi
fi

# Step 1 — fetch the bulk-data catalog and pull out the download URI
# for the default_cards dataset. Scryfall rotates the URL daily, so
# we never hardcode it.
#
# As of Sept 2026 Scryfall dropped the JSON-array `download_uri` and
# only publishes `jsonl_download_uri` (gzipped JSON Lines). The server
# (cards.Index.Load) streams a top-level JSON array, so we convert on
# the fly below. `download_uri` is still preferred if it ever returns.
catalog_uri="https://api.scryfall.com/bulk-data"
catalog="$(curl --fail --silent --show-error --max-time 30 \
  -H 'Accept: application/json' -A 'cmd_and_ctrl-scryfall-refresh/1.0' \
  "$catalog_uri")"
entry="$(jq -c '.data[] | select(.type == "default_cards")' <<<"$catalog")"
download_uri="$(jq -r '.download_uri // empty' <<<"$entry")"
jsonl_uri="$(jq -r '.jsonl_download_uri // empty' <<<"$entry")"

if [[ -z "$download_uri" && -z "$jsonl_uri" ]]; then
  echo "scryfall-refresh: could not locate a default_cards download URI in bulk-data catalog" >&2
  exit 1
fi

# Step 2 — stream download into a tmp file under the same directory
# so the atomic rename below is within one filesystem.
tmp="$(mktemp "$SCRYFALL_DIR/default-cards.XXXXXX.json.tmp")"
trap 'rm -f "$tmp"' EXIT

if [[ -n "$download_uri" ]]; then
  echo "scryfall-refresh: downloading $download_uri"
  curl --fail --silent --show-error --location --max-time 600 \
       --output "$tmp" \
       "$download_uri"
else
  echo "scryfall-refresh: downloading $jsonl_uri (JSONL → JSON array)"
  # Streamed: gunzip → wrap lines in [ , ] without buffering the whole
  # ~500 MB dump in memory (jq -s would). Blank lines are skipped.
  curl --fail --silent --show-error --location --max-time 600 \
       "$jsonl_uri" \
    | gzip -dc \
    | awk 'BEGIN { printf "[" } NF { if (n++) printf ","; printf "%s", $0 } END { print "]" }' \
    > "$tmp"
fi

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
