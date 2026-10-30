#!/usr/bin/env bash
# Read-only snapshot of the open issues, labels, milestones and recent pull
# requests, so a triage can be decided offline and the writes made in one pass.
#
#   snapshot.sh [since-date]
#
# Writes, under $TRIAGE_DIR (default $CLAUDE_JOB_DIR/tmp/triage):
#   issues.json      every open issue: labels, milestone, assignees, dates
#   iss/<n>.json     one issue with its body and comments
#   labels.json      the label set
#   milestones.json  every milestone, open and closed
#   prs.json         pull requests updated since <since-date> (default: 45
#                    days ago), with bodies and merge commits. GitHub search
#                    returns at most 1000 results, so it reads in windows.
#                    `query.py evidence <n>` searches further back on its own.
#
# It also fetches origin, so origin/develop is current for the code checks.
# Rerun it at the end of a triage: issues keep arriving during a run. Only
# issues that changed since the last snapshot are fetched again. A full first
# run takes a minute or two (one read per issue); give it a long timeout.
# shellcheck source=lib.sh
source "$(dirname "$(readlink -f "$0")")/lib.sh"

since=${1:-$(date -d '-45 days' +%F)}
[[ $since =~ ^[0-9]{4}-[0-9]{2}-[0-9]{2}$ ]] || die "since-date must be YYYY-MM-DD"
R=$(triage_repo)
D=$(triage_dir)
mkdir -p "$D/iss"

git -C "$TRIAGE_ROOT" fetch -q origin

read_retry gh issue list -R "$R" --state open --limit 1000 \
  --json number,title,labels,milestone,assignees,createdAt,updatedAt,author >"$D/issues.json.tmp"
mv "$D/issues.json.tmp" "$D/issues.json"
total=$(jq length "$D/issues.json")
[ "$total" -lt 1000 ] || echo "warning: 1000 open issues listed; the list may be cut off" >&2

read_retry gh label list -R "$R" --limit 200 --json name,description >"$D/labels.json"
read_retry gh api --paginate "repos/$R/milestones?state=all&per_page=100" \
  --jq '.[] | {number, title, state, open_issues, closed_issues, description}' | jq -s . >"$D/milestones.json"

fetched=0
while read -r n upd; do
  f="$D/iss/$n.json"
  if [ -f "$f" ] && [ "$(jq -r .updatedAt "$f")" = "$upd" ]; then continue; fi
  if read_retry gh issue view "$n" -R "$R" \
      --json number,title,body,comments,labels,milestone,assignees,author,createdAt,updatedAt,state >"$f.tmp"; then
    mv "$f.tmp" "$f"
    fetched=$((fetched + 1))
  else
    rm -f "$f.tmp"
    echo "could not fetch #$n" >&2
  fi
done < <(jq -r '.[] | "\(.number) \(.updatedAt)"' "$D/issues.json")

# Each PR has one updatedAt, so windows on it never overlap. A window that
# hits the 1000-result cap is split in half until it doesn't.
pr_fields=number,title,state,isDraft,mergedAt,baseRefName,headRefName,author,body,mergeCommit,createdAt
fetch_prs() {
  local from=$1 to=$2 out days mid
  out=$(read_retry gh pr list -R "$R" --state all --limit 1000 --search "updated:$from..$to" --json "$pr_fields")
  if [ "$(jq length <<<"$out")" -ge 1000 ]; then
    days=$(( ($(date -d "$to" +%s) - $(date -d "$from" +%s)) / 86400 ))
    if [ "$days" -ge 1 ]; then
      mid=$(date -d "$from +$((days / 2)) days" +%F)
      fetch_prs "$from" "$mid"
      fetch_prs "$(date -d "$mid +1 day" +%F)" "$to"
      return
    fi
    echo "warning: 1000 PRs updated on $from alone; some are missing" >&2
  fi
  jq -c '.[]' <<<"$out" >>"$D/prs.jsonl"
}
: >"$D/prs.jsonl"
today=$(date +%F)
start=$since
while [[ ! $start > $today ]]; do
  end=$(date -d "$start +6 days" +%F)
  fetch_prs "$start" "$end"
  start=$(date -d "$end +1 day" +%F)
done
jq -s 'unique_by(.number)' "$D/prs.jsonl" >"$D/prs.json"
rm -f "$D/prs.jsonl"

echo "snapshot of $R in $D"
echo "  open issues: $total ($fetched fetched this run)"
jq -r '
  def lab(l): any(.labels[]; .name == l);
  "  needs-triage: \([.[] | select(lab("needs-triage"))] | length)",
  "  no milestone: \([.[] | select(.milestone == null)] | length)",
  "  player-report: \([.[] | select(lab("player-report"))] | length)",
  "  P0 or P1: \([.[] | select(lab("P0") or lab("P1")) | "#\(.number)"] | join(" ") | if . == "" then "none" else . end)"
' "$D/issues.json"
echo "  pull requests updated since $since: $(jq length "$D/prs.json")"
