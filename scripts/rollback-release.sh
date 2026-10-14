#!/usr/bin/env bash
# rollback-release.sh — on a deploy host, swap the live release with the
# one scripts/keep-previous-release.sh set aside, and restart (#3339).
# The rollback-prod workflow streams it over ssh:
#
#   ssh host 'DEPLOY_ROOT=... WEB_ROOT=... SERVICE=... bash -s' < scripts/rollback-release.sh
#
# A swap, not a delete: the release rolled back FROM becomes the new
# .prev, so running it twice rolls forward again. Server, client and (if
# a bot .prev exists and its unit is enabled) bot move together.
#
# This does NOT touch the database or the data directory. A release that
# ran a migration needs the by-hand steps in docs/environments.md
# ("Rolling back past ...") before its predecessor will start; the
# workflow's smoke test then fails and says so.
set -euo pipefail

: "${DEPLOY_ROOT:?DEPLOY_ROOT is not set}"
: "${WEB_ROOT:?WEB_ROOT is not set}"
: "${SERVICE:?SERVICE is not set}"

server="${DEPLOY_ROOT}/server/bin/cmd_and_ctrl-server"
bot="${DEPLOY_ROOT}/bot/bin/cmd_and_ctrl-bot"
prev="${DEPLOY_ROOT}/prev"

if [ ! -f "${server}.prev" ] || [ -z "$(ls -A "${prev}/client" 2>/dev/null)" ]; then
  echo "::error::No previous release is kept on this host (needs ${server}.prev and ${prev}/client). The next deploy writes it; until then a rollback is a revert PR."
  exit 1
fi

swap_file() {
  local f="$1"
  mv -f "${f}" "${f}.swap"
  mv -f "${f}.prev" "${f}"
  mv -f "${f}.swap" "${f}.prev"
}

swap_file "${server}"
echo "server: swapped with its .prev."

swap_bot=0
if [ -f "${bot}" ] && [ -f "${bot}.prev" ]; then
  swap_file "${bot}"
  swap_bot=1
  echo "bot: swapped with its .prev."
fi

# Client: live -> tmp, prev -> live, tmp -> prev.
tmp="${prev}/swap"
rm -rf "${tmp}"
mkdir -p "${tmp}"
rsync -a "${WEB_ROOT}/" "${tmp}/"
rsync -a --delete "${prev}/client/" "${WEB_ROOT}/"
rsync -a --delete "${tmp}/" "${prev}/client/"
rm -rf "${tmp}"
# The recorded hash described the dist that was live; it no longer is.
# Dropping it makes the next deploy snapshot whatever is live then.
rm -f "${prev}/client.sha"
echo "client: swapped with prev/client."

sudo systemctl reset-failed "${SERVICE}" 2>/dev/null || true
sudo systemctl restart "${SERVICE}"
echo "${SERVICE}: restarted."

if [ "${swap_bot}" = 1 ] && systemctl is-enabled cmd-and-ctrl-bot >/dev/null 2>&1; then
  sudo systemctl reset-failed cmd-and-ctrl-bot 2>/dev/null || true
  sudo systemctl restart cmd-and-ctrl-bot
  echo "cmd-and-ctrl-bot: restarted."
fi
