#!/usr/bin/env bash
# discord-post.sh: post stdin to the Discord webhook named by DISCORD_WEBHOOK_URL
# (#3326). Used by the release-notes and deploy-failure jobs in ci-cd.yml.
#
# It never fails its caller: an unset URL or an empty message is a skip with a
# notice, and a failed post is a ::warning::, so a notification can never fail a
# deploy. The URL is never printed, and curl's stderr is discarded because it
# can carry the URL.
set -uo pipefail

url="${DISCORD_WEBHOOK_URL:-}"
if [ -z "$url" ]; then
  echo "Discord post skipped: the webhook secret is not set."
  exit 0
fi

msg=$(cat)
if [ -z "$msg" ]; then
  echo "Discord post skipped: the message is empty."
  exit 0
fi
# Discord caps content at 2000 characters; stay well under it.
msg="${msg:0:1900}"

payload=$(jq -n --arg c "$msg" '{content: $c}') || {
  echo "::warning::Discord post skipped: could not build the payload."
  exit 0
}

if curl --fail -sS -o /dev/null --max-time 15 \
    -H 'Content-Type: application/json' -d "$payload" "$url" 2>/dev/null; then
  echo "Posted to Discord."
else
  echo "::warning::Discord post failed; this does not affect the deploy."
fi
exit 0
