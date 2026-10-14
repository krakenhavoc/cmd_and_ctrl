#!/usr/bin/env bash
# keep-previous-release.sh — on a deploy host, set aside the release that
# is about to be replaced, so scripts/rollback-release.sh can put it back
# (#3339). CD streams this to the host over ssh just before it copies the
# new files, so the host needs no copy of the repo:
#
#   ssh host 'DEPLOY_ROOT=... WEB_ROOT=... bash -s -- <server-sha> <bot-sha|-> <client-sha>' \
#     < scripts/keep-previous-release.sh
#
# The three arguments are the sha256 of the NEW server binary, bot binary
# ("-" on a host that gets no bot) and client dist (a hash of its file
# hashes, computed by the workflow). What it keeps, all under krkn's own
# ${DEPLOY_ROOT}:
#
#   server/bin/cmd_and_ctrl-server.prev   the server binary being replaced
#   bot/bin/cmd_and_ctrl-bot.prev         the bot binary, if there is one
#   prev/client/                          a copy of ${WEB_ROOT}
#   prev/client.sha                       hash of the dist now live
#
# One release back, no more. A file is set aside only if the new one
# differs from what is live, so re-running a deploy (a failed job, the
# weekly Scryfall refresh's restart) never overwrites the good .prev with
# the release that was already replaced.
set -euo pipefail

: "${DEPLOY_ROOT:?DEPLOY_ROOT is not set}"
: "${WEB_ROOT:?WEB_ROOT is not set}"
usage="usage: keep-previous-release.sh <server-sha> <bot-sha|-> <client-sha>"
new_server="${1:?${usage}}"
new_bot="${2:?${usage}}"
new_client="${3:?${usage}}"

keep_file() {
  local file="$1" new="$2" name cur
  name=$(basename "${file}")
  if [ "${new}" = "-" ] || [ ! -f "${file}" ]; then
    return 0
  fi
  cur=$(sha256sum "${file}" | cut -d' ' -f1)
  if [ "${cur}" = "${new}" ]; then
    echo "${name}: already live; keeping the existing .prev."
    return 0
  fi
  cp -p "${file}" "${file}.prev.tmp"
  mv -f "${file}.prev.tmp" "${file}.prev"
  echo "${name}: previous build kept as ${name}.prev."
}

keep_file "${DEPLOY_ROOT}/server/bin/cmd_and_ctrl-server" "${new_server}"
keep_file "${DEPLOY_ROOT}/bot/bin/cmd_and_ctrl-bot" "${new_bot}"

prev="${DEPLOY_ROOT}/prev"
mkdir -p "${prev}/client"
recorded=$(cat "${prev}/client.sha" 2>/dev/null || true)
if [ "${recorded}" = "${new_client}" ]; then
  echo "client: already live; keeping the existing prev/client."
elif [ -n "$(ls -A "${WEB_ROOT}" 2>/dev/null)" ]; then
  rsync -a --delete "${WEB_ROOT}/" "${prev}/client/"
  echo "client: previous dist kept in ${prev}/client."
fi
# Recorded now rather than after the copy: if the deploy dies midway and
# is re-run, the snapshot above is the good one and must not be redone.
printf '%s\n' "${new_client}" > "${prev}/client.sha"
