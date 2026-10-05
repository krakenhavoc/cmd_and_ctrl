#!/usr/bin/env bash
# Off-node uptime watcher (ADR 0123 section 8, #2281, #598 item 3).
#
# Runs from .github/workflows/uptime.yml. Curls each site's /healthz from the
# internet: up to 3 tries 20 s apart, any single 2xx is up. The state is a
# GitHub issue labelled `outage`, titled "`<env>` is down":
#   down, no open issue -> open one, post ONE Discord message
#   down, issue open    -> nothing (no repeat posts)
#   up,   issue open    -> comment the downtime, close it, post ONE recovery
#   up,   none open     -> nothing
#
# Env: GH_TOKEN, GITHUB_REPOSITORY (set by Actions);
#      DRY_RUN=true  check and print, change nothing (default);
#      DISCORD_WEBHOOK  optional, empty skips posting with a warning;
#      UPTIME_TRIES / UPTIME_GAP  override 3 / 20 (local testing only).
set -uo pipefail

DRY_RUN="${DRY_RUN:-true}"
TRIES="${UPTIME_TRIES:-3}"
GAP="${UPTIME_GAP:-20}"
WEBHOOK="${DISCORD_WEBHOOK:-}"
REPO="${GITHUB_REPOSITORY:?GITHUB_REPOSITORY is not set}"

SITES=(
  "prod|https://cmd.labxp.io/healthz"
  "dev|https://cmd-dev.labxp.io/healthz"
)

dry() { [ "$DRY_RUN" = "true" ]; }

if [ -z "$WEBHOOK" ]; then
  echo "::warning::CMDCTRL_ALERT_DISCORD_WEBHOOK is empty; Discord posts are skipped, outage issues are still managed."
fi

if dry; then
  echo "DRY RUN: nothing will be created, commented, closed or posted."
else
  gh label create outage --repo "$REPO" --color B60205 \
    --description "A site's /healthz stopped answering (uptime watcher)" 2>/dev/null || true
fi

# probe URL: sets PROBE_DETAIL. Returns 0 up, 1 down, 2 blocked by a Cloudflare
# challenge (the check could not see the site, which is not an outage).
probe() {
  local url="$1" i code rc err hdr
  PROBE_DETAIL=""
  local challenged=0
  for ((i = 1; i <= TRIES; i++)); do
    err=$(mktemp)
    hdr=$(mktemp)
    code=$(curl -sS -o /dev/null -D "$hdr" -w '%{http_code}' --connect-timeout 5 --max-time 10 -A 'cmd-and-ctrl-uptime/1 (+https://github.com/krakenhavoc/cmd_and_ctrl)' "$url" 2>"$err")
    rc=$?
    if [ "$rc" -eq 0 ] && [[ "$code" =~ ^2[0-9][0-9]$ ]]; then
      rm -f "$err" "$hdr"
      PROBE_DETAIL="HTTP $code on try $i"
      return 0
    fi
    PROBE_DETAIL="try $i: HTTP ${code:-000}, curl exit $rc: $(tr '\n' ' ' <"$err")"
    echo "  $PROBE_DETAIL"
    # Who answered: a Cloudflare or proxy block shows up here, not in the code.
    grep -iE '^(server|cf-mitigated|cf-ray|via):' "$hdr" | tr -d '\r' | sed 's/^/    /'
    if grep -qi '^cf-mitigated:' "$hdr"; then challenged=$((challenged + 1)); fi
    rm -f "$err" "$hdr"
    if [ "$i" -lt "$TRIES" ]; then sleep "$GAP"; fi
  done
  if [ "$challenged" -eq "$TRIES" ]; then return 2; fi
  return 1
}

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

rc=0
for site in "${SITES[@]}"; do
  env="${site%%|*}"
  url="${site#*|}"
  title="\`$env\` is down"
  echo "== $env: $url"

  probe "$url"
  case $? in
    0) up=true ;;
    2)
      msg="$env: Cloudflare challenged the GitHub runner on every try, so the check cannot see the site. Skip the challenge for this check (a WAF skip rule). No issue is opened or closed."
      if dry; then echo "::warning::$msg"; else echo "::error::$msg"; rc=1; fi
      continue
      ;;
    *) up=false ;;
  esac
  echo "  result: $([ "$up" = true ] && echo UP || echo DOWN) ($PROBE_DETAIL)"

  # In a dry run the label may not exist yet; an empty list is the right answer.
  open_json=$(gh issue list --repo "$REPO" --label outage --state open --limit 100 \
    --json number,title,createdAt,url 2>/dev/null) || open_json='[]'
  issue=$(jq -c --arg t "$title" '[.[] | select(.title == $t)][0] // empty' <<<"$open_json")

  now=$(date -u +%Y-%m-%dT%H:%M:%SZ)
  if [ "$up" = false ]; then
    if [ -n "$issue" ]; then
      echo "  issue #$(jq -r .number <<<"$issue") already open: nothing to do."
      continue
    fi
    # shellcheck disable=SC2016 # the backticks are literal Markdown
    body=$(printf 'The uptime watcher could not reach `%s` at %s.\n\n- URL: %s\n- Tries: %s, %s s apart\n- Last result: %s\n\nThis issue closes itself when the site answers again.' \
      "$env" "$now" "$url" "$TRIES" "$GAP" "$PROBE_DETAIL")
    if dry; then
      echo "  DRY RUN would open issue \"$title\" and post a down message."
      post "DOWN: \`$env\` ($url) is not answering. <issue link>"
    else
      if ! issue_url=$(gh issue create --repo "$REPO" --label outage --title "$title" --body "$body"); then
        echo "::error::could not open the outage issue for $env"
        rc=1
        continue
      fi
      echo "  opened $issue_url"
      post "DOWN: \`$env\` ($url) is not answering. $issue_url"
    fi
  else
    if [ -z "$issue" ]; then
      echo "  up, no open outage issue: nothing to do."
      continue
    fi
    num=$(jq -r .number <<<"$issue")
    issue_url=$(jq -r .url <<<"$issue")
    created=$(jq -r .createdAt <<<"$issue")
    secs=$(($(date -u +%s) - $(date -u -d "$created" +%s)))
    dur=$(printf '%dh %dm' $((secs / 3600)) $((secs % 3600 / 60)))
    if dry; then
      echo "  DRY RUN would comment \"down for $dur\", close #$num and post a recovery message."
      post "RECOVERED: \`$env\` ($url) is up again after $dur. $issue_url"
    else
      gh issue comment "$num" --repo "$REPO" \
        --body "\`$env\` answered again at $now ($PROBE_DETAIL). Down for about $dur." || rc=1
      gh issue close "$num" --repo "$REPO" --reason completed || rc=1
      post "RECOVERED: \`$env\` ($url) is up again after $dur. $issue_url"
    fi
  fi
done
exit "$rc"
