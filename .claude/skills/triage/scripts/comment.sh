#!/usr/bin/env bash
# Comment on an issue, or close it with that comment, paced for GitHub's
# write limits.
#
#   comment.sh <issue> <body-file>                          # comment only
#   comment.sh <issue> <body-file> completed                # close as completed
#   comment.sh <issue> <body-file> "not planned"
#   comment.sh <issue> <body-file> duplicate <kept-issue>   # close as a duplicate
#
# Safe to rerun: a comment whose text is already on the issue is not posted
# again, and an issue that is already closed is left alone. DRY_RUN=1 prints
# the command instead.
# shellcheck source=lib.sh
source "$(dirname "$(readlink -f "$0")")/lib.sh"

[ $# -ge 2 ] || die "usage: comment.sh <issue> <body-file> [completed|\"not planned\"|duplicate <kept-issue>]"
n=${1#\#}; f=$2; reason=${3:-}; kept=${4:-}
kept=${kept#\#}
[[ $n =~ ^[0-9]+$ ]] || die "not an issue number: $1"
[ -s "$f" ] || die "empty or missing body file: $f"
case "$reason" in
  ""|completed|"not planned") ;;
  duplicate) [[ $kept =~ ^[0-9]+$ ]] || die "duplicate needs the kept issue's number" ;;
  *) die "unknown close reason: $reason" ;;
esac

R=$(triage_repo)
view=$(read_retry gh issue view "$n" -R "$R" --json state,comments)
state=$(jq -r .state <<<"$view")
posted=$(jq --rawfile b "$f" \
  '[.comments[] | select((.body | sub("\\s+$"; "")) == ($b | sub("\\s+$"; "")))] | length' <<<"$view")

if [ -z "$reason" ]; then
  if [ "$posted" -gt 0 ]; then echo "skip #$n: that comment is already there"; exit 0; fi
  mutate "#$n comment" gh issue comment "$n" -R "$R" --body-file "$f"
  exit 0
fi

if [ "$state" = CLOSED ]; then echo "skip #$n: already closed"; exit 0; fi
close=(gh issue close "$n" -R "$R")
case "$reason" in
  duplicate) close+=(--duplicate-of "$kept") ;;
  *) close+=(--reason "$reason") ;;
esac
# A rerun after the comment went up but the close failed only closes.
[ "$posted" -gt 0 ] || close+=(--comment "$(cat "$f")")
mutate "#$n close ($reason)" "${close[@]}"
