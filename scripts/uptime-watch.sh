#!/usr/bin/env bash
# Off-node dead-man check on the monitoring VM (ADR 0123 section 8, #2281,
# #598 item 3).
#
# Cloudflare Bot Fight Mode challenges GitHub runner IPs and cannot be skipped
# by a WAF rule, so this does not curl the sites. The monitoring VM (HomeLab)
# writes the Actions variable CMDCTRL_MONITORING_HEARTBEAT, Unix epoch seconds,
# every 5 minutes while its Prometheus and Alertmanager are ready. This script
# reads it. A heartbeat older than 20 minutes means the VM has stopped
# checking in: the Proxmox node, the house's internet or the monitoring stack
# is down. The VM's own alerts cover the sites the rest of the time.
#
# State is one GitHub issue labelled `outage`, titled "`monitoring` heartbeat lost":
#   stale, no open issue -> open one, post ONE Discord message
#   stale, issue open    -> nothing (no repeat posts)
#   fresh, issue open    -> comment the gap, close it, post ONE recovery
#   fresh, none open     -> nothing
#
# Env: GH_TOKEN, GITHUB_REPOSITORY (set by Actions);
#      HEARTBEAT        the variable's value; unset, empty or non-numeric is
#                       "not configured": a warning, nothing else, exit 0;
#      DRY_RUN=true     check and print, change nothing (default);
#      DISCORD_WEBHOOK  optional, empty skips posting with a warning.
set -uo pipefail

DRY_RUN="${DRY_RUN:-true}"
WEBHOOK="${DISCORD_WEBHOOK:-}"
HEARTBEAT="${HEARTBEAT:-}"
REPO="${GITHUB_REPOSITORY:?GITHUB_REPOSITORY is not set}"
STALE_AFTER=1200 # 20 minutes
FUTURE_SLACK=300 # a clock this far ahead is a warning, not an alarm

dry() { [ "$DRY_RUN" = "true" ]; }

if [[ ! "$HEARTBEAT" =~ ^[0-9]+$ ]]; then
  echo "::warning::heartbeat not configured: CMDCTRL_MONITORING_HEARTBEAT is unset, empty or not epoch seconds (got '${HEARTBEAT:0:40}'). Nothing to check."
  exit 0
fi

now_s=$(date -u +%s)
age=$((now_s - HEARTBEAT))
mins=$((age / 60))
hb_time=$(date -u -d "@$HEARTBEAT" +%Y-%m-%dT%H:%M:%SZ)
echo "heartbeat: $HEARTBEAT ($hb_time), age ${age}s (stale after ${STALE_AFTER}s)"

if [ "$age" -lt "-$FUTURE_SLACK" ]; then
  echo "::warning::heartbeat is $((-age)) s in the future; the VM's clock is ahead. Treating it as fresh."
fi

if [ -z "$WEBHOOK" ]; then
  echo "::warning::CMDCTRL_ALERT_DISCORD_WEBHOOK is empty; Discord posts are skipped, the outage issue is still managed."
fi

if dry; then
  echo "DRY RUN: nothing will be created, commented, closed or posted."
else
  gh label create outage --repo "$REPO" --color B60205 \
    --description "An outage the uptime watcher noticed" 2>/dev/null || true
fi

post() {
  local msg="$1" payload
  if [ -z "$WEBHOOK" ]; then
    echo "  (no webhook; would post: $msg)"
    return 0
  fi
  if dry; then
    echo "  DRY RUN would post to Discord: $msg"
    return 0
  fi
  payload=$(jq -n --arg c "$msg" '{content: $c}')
  curl -sS -o /dev/null --max-time 15 -H 'Content-Type: application/json' \
    -d "$payload" "$WEBHOOK" || echo "::warning::Discord post failed"
}

# shellcheck disable=SC2016 # the backticks are literal Markdown
title='`monitoring` heartbeat lost'
# In a dry run the label may not exist yet; an empty list is the right answer.
open_json=$(gh issue list --repo "$REPO" --label outage --state open --limit 100 \
  --json number,title,createdAt,url 2>/dev/null) || open_json='[]'
issue=$(jq -c --arg t "$title" '[.[] | select(.title == $t)][0] // empty' <<<"$open_json")
stamp=$(date -u +%Y-%m-%dT%H:%M:%SZ)

if [ "$age" -gt "$STALE_AFTER" ]; then
  echo "result: STALE"
  if [ -n "$issue" ]; then
    echo "issue #$(jq -r .number <<<"$issue") already open: nothing to do."
    exit 0
  fi
  # shellcheck disable=SC2016 # the backticks are literal Markdown
  body=$(printf "The monitoring VM has not checked in for %s minutes (checked %s).\n\nThe last heartbeat was at %s. The Proxmox node, the house's internet or the monitoring stack is down. While the VM is gone its own alerts (\`SiteDown\`, \`ServerNotReporting\`) cannot fire, so check the sites by hand.\n\nThis issue closes itself when the heartbeat is fresh again." \
    "$mins" "$stamp" "$hb_time")
  msg_tail="The monitoring VM hasn't checked in for $mins minutes: the Proxmox node, the house's internet or the monitoring stack is down. Otherwise the VM's own alerts cover the sites, and they cannot fire now."
  if dry; then
    echo "DRY RUN would open issue \"$title\" and post a down message."
    post "HEARTBEAT LOST: $msg_tail <issue link>"
  else
    if ! issue_url=$(gh issue create --repo "$REPO" --label outage --title "$title" --body "$body"); then
      echo "::error::could not open the heartbeat issue"
      exit 1
    fi
    echo "opened $issue_url"
    post "HEARTBEAT LOST: $msg_tail $issue_url"
  fi
else
  echo "result: FRESH"
  if [ -z "$issue" ]; then
    echo "no open heartbeat issue: nothing to do."
    exit 0
  fi
  num=$(jq -r .number <<<"$issue")
  issue_url=$(jq -r .url <<<"$issue")
  created=$(jq -r .createdAt <<<"$issue")
  secs=$((now_s - $(date -u -d "$created" +%s)))
  gap=$(printf '%dh %dm' $((secs / 3600)) $((secs % 3600 / 60)))
  if dry; then
    echo "DRY RUN would comment \"heartbeat back after $gap\", close #$num and post a recovery message."
    post "HEARTBEAT BACK: the monitoring VM is checking in again after about $gap. $issue_url"
  else
    rc=0
    gh issue comment "$num" --repo "$REPO" \
      --body "The heartbeat is fresh again as of $stamp. The gap was about $gap." || rc=1
    gh issue close "$num" --repo "$REPO" --reason completed || rc=1
    post "HEARTBEAT BACK: the monitoring VM is checking in again after about $gap. $issue_url"
    exit "$rc"
  fi
fi
