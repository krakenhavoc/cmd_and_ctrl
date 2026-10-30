#!/usr/bin/env bash
# Set an issue's milestone and labels, paced for GitHub's write limits.
#
#   edit.sh <issue> <milestone|-> [add-labels-csv] [remove-labels-csv]
#
#   edit.sh 4326 S78 bug,P2,tier:2-standard
#   edit.sh 4407 Backlog question
#   edit.sh 4497 - P2 P3                   # labels only: P3 -> P2
#
# <milestone> is a full open-milestone title or a unique prefix of one ("S79",
# "Backlog"); "-" or "" leaves it alone. Setting a milestone clears
# needs-triage by itself (.github/workflows/issue-triage.yml), so don't remove
# that label by hand.
#
# For a decisions file of tab-separated rows (issue, milestone, add, remove):
#   while IFS=$'\t' read -r n ms add rm; do edit.sh "$n" "$ms" "$add" "$rm"; done < decisions.tsv
#
# Safe to rerun: adding a label an issue has, or setting the milestone it is
# already in, changes nothing. DRY_RUN=1 prints the command instead.
# shellcheck source=lib.sh
source "$(dirname "$(readlink -f "$0")")/lib.sh"

[ $# -ge 2 ] || die "usage: edit.sh <issue> <milestone|-> [add-labels-csv] [remove-labels-csv]"
n=${1#\#}; ms=$2; add=${3:-}; rm=${4:-}
[[ $n =~ ^[0-9]+$ ]] || die "not an issue number: $1"

args=()
if [ -n "$ms" ] && [ "$ms" != "-" ]; then
  args+=(--milestone "$(resolve_milestone "$ms")")
fi
[ -n "$add" ] && args+=(--add-label "$add")
[ -n "$rm" ] && args+=(--remove-label "$rm")
[ ${#args[@]} -gt 0 ] || { echo "nothing to change on #$n"; exit 0; }

mutate "#$n" gh issue edit "$n" -R "$(triage_repo)" "${args[@]}"
