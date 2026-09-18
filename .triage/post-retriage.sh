#!/usr/bin/env bash
# Post the re-triage comment on every card-batch issue that does not
# already carry one. Idempotent: safe to re-run, skips what is done.
#
# Claude Code's auto-mode classifier blocks this as a loop, and blocks
# rapid individual `gh issue comment` calls after about nine, so the
# remainder is handed to you rather than ground out one call at a time.
#
#   cd /Users/luke/devops/repos/cmd_and_ctrl && ./.triage/post-retriage.sh
#
# Delete the .triage/ directory when it is done; it is untracked.
set -u
cd "$(dirname "$0")"
REPO=krakenhavoc/cmd_and_ctrl
MARK="Re-triage against today's engine"
ok=0; skip=0; fail=0
for f in c*.md; do
  n="${f#c}"; n="${n%.md}"
  if gh issue view "$n" --repo "$REPO" --json comments \
       --jq '.comments[].body' 2>/dev/null | grep -qF "$MARK"; then
    skip=$((skip+1)); printf "s"; continue
  fi
  if gh issue comment "$n" --repo "$REPO" --body-file "$f" >/dev/null 2>&1; then
    ok=$((ok+1)); printf "."
  else
    fail=$((fail+1)); printf "\nFAILED #%s\n" "$n"
  fi
  sleep 1
done
printf "\nposted=%d already-had-it=%d failed=%d\n" "$ok" "$skip" "$fail"
