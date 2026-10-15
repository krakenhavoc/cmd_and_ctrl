#!/usr/bin/env bash
# release-notes.sh: post a short changelog to Discord after a push to main
# deploys to production (#3326). Run by the release-notes job in ci-cd.yml.
#
# The changelog is the commit subjects between the previous main tip (BEFORE)
# and the deployed commit (AFTER). chore, ci, test, build, style and merge
# commits are left out. Only the type and scope prefix is stripped.
#
# Env: GH_TOKEN, REPO (owner/name), BEFORE, AFTER; DISCORD_WEBHOOK_URL (optional,
# via discord-post.sh). A failure here is a warning, never a failed job.
set -uo pipefail

here=$(cd "$(dirname "$0")" && pwd)
: "${REPO:?REPO is not set}" "${BEFORE:?BEFORE is not set}" "${AFTER:?AFTER is not set}"

if [[ ! "$BEFORE" =~ ^[0-9a-f]{40}$ ]]; then
  echo "No previous main commit to diff against; no release notes."
  exit 0
fi

subjects=$(gh api "repos/${REPO}/compare/${BEFORE}...${AFTER}" \
  --jq '.commits[].commit.message | split("\n")[0]' 2>/dev/null) || {
  echo "::warning::Could not read the commits for the release notes; nothing was posted."
  exit 0
}

lines=()
while IFS= read -r subject; do
  [ -z "$subject" ] && continue
  case "$subject" in
    "Merge "*) continue ;;
  esac
  if [[ "$subject" =~ ^(chore|ci|test|build|style)(\(.*\))?!?: ]]; then
    continue
  fi
  text=$(sed -E 's/^[a-z]+(\([^)]*\))?!?: //' <<<"$subject")
  lines+=("- ${text}")
done <<<"$subjects"

if [ "${#lines[@]}" -eq 0 ]; then
  echo "No player-facing commits in this release; nothing to post."
  exit 0
fi

shown=25
msg="**CMD & CTRL updated** (\`${AFTER:0:7}\`)"$'\n'
for ((i = 0; i < ${#lines[@]} && i < shown; i++)); do
  msg+="${lines[i]}"$'\n'
done
if [ "${#lines[@]}" -gt "$shown" ]; then
  msg+="- ...and $((${#lines[@]} - shown)) more"$'\n'
fi

printf '%s' "$msg" | "$here/discord-post.sh"
