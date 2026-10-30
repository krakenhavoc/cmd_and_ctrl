#!/usr/bin/env bash
# Open a themed sprint: its milestone, then its tracker issue in it.
#
#   new-sprint.sh "S81 — Theme words" <tracker-body-file> "<milestone description>"
#
# The title is "S<n> — <theme>", with an em dash; the tracker issue has the
# same title. Write the body with `sprints_md.py render <spec> --part tracker`.
# The milestone description is one or two sentences ending with "Opened at
# the <date> triage." Tracker labels default to enhancement,tier:2-standard,P2
# (TRACKER_LABELS overrides them). No one is assigned.
#
# Safe to rerun: an existing milestone or tracker with that exact title is
# reused, not duplicated. DRY_RUN=1 prints the writes instead. Prints the
# tracker's number last, for the spec's "tracker" field.
# shellcheck source=lib.sh
source "$(dirname "$(readlink -f "$0")")/lib.sh"

[ $# -eq 3 ] || die 'usage: new-sprint.sh "S81 — Theme words" <tracker-body-file> "<milestone description>"'
title=$1; body=$2; desc=$3
[[ $title =~ ^S[0-9]+(\.[0-9]+)?\ —\ .+ ]] || die "title must look like 'S81 — Theme words' (em dash): $title"
[ -s "$body" ] || die "empty or missing tracker body: $body"
labels=${TRACKER_LABELS:-enhancement,tier:2-standard,P2}
R=$(triage_repo)

ms=$(read_retry gh api --paginate "repos/$R/milestones?state=all&per_page=100" \
  --jq ".[] | select(.title == \"${title//\"/\\\"}\") | .number")
if [ -n "$ms" ]; then
  echo "milestone exists: #$ms $title"
else
  mutate "milestone $title" gh api -X POST "repos/$R/milestones" -f title="$title" -f description="$desc"
fi

key=${title%% —*}
tracker=$(read_retry gh issue list -R "$R" --state all --limit 50 --search "\"$key\" in:title" \
  --json number,title --jq ".[] | select(.title == \"${title//\"/\\\"}\") | .number" | head -1)
if [ -n "$tracker" ]; then
  echo "tracker exists: #$tracker"
else
  mutate "tracker $title" gh issue create -R "$R" --title "$title" --body-file "$body" \
    --label "$labels" --milestone "$title"
  tracker=$(grep -oE '/issues/[0-9]+' <<<"$MUTATE_OUT" | grep -oE '[0-9]+' | tail -1 || true)
  [ "${DRY_RUN:-0}" = 1 ] && tracker="(dry run)"
fi
echo "tracker: ${tracker:-unknown}"
