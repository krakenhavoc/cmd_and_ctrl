#!/usr/bin/env bash
# smoke-release.sh — on a deploy host, prove the running server answers as
# the expected environment AND honours an authenticated request (#3339).
# GET /config only proves the process is listening; this also needs the
# database open and the session machinery working.
#
#   ssh host 'ENV_FILE=... LOCAL_PORT=... APP_ENV=prod bash -s' < scripts/smoke-release.sh
#
# 1. GET /config, retried while the server loads the card index, must
#    report APP_ENV.
# 2. POST /admin/login with CMDCTRL_ADMIN_TOKEN from ENV_FILE (the call
#    the Discord bot makes), then GET /games with the session.
#
# The token is read with sudo and never printed or put on a command line.
set -euo pipefail

: "${ENV_FILE:?ENV_FILE is not set}"
: "${LOCAL_PORT:?LOCAL_PORT is not set}"
: "${APP_ENV:?APP_ENV is not set}"
base="http://127.0.0.1:${LOCAL_PORT}"

if ! body=$(curl -fsS --retry 30 --retry-delay 2 --retry-connrefused --max-time 120 "${base}/config"); then
  echo "::error::The server never answered GET /config on port ${LOCAL_PORT}. Check journalctl; a release rolled back past a database migration fails here (docs/environments.md)."
  exit 1
fi
case "${body}" in
  *'"env":"'"${APP_ENV}"'"'*) ;;
  *) echo "::error::GET /config did not report env=${APP_ENV}: ${body}"; exit 1 ;;
esac
echo "GET /config -> env=${APP_ENV}"

token=$(sudo sed -n 's/^CMDCTRL_ADMIN_TOKEN=//p' "${ENV_FILE}" | tail -n 1)
if [ -z "${token}" ]; then
  echo "::error::CMDCTRL_ADMIN_TOKEN is missing from ${ENV_FILE}, so the authenticated smoke request cannot run."
  exit 1
fi

# curl reads the JSON body from stdin (-d @-), so the token is not in argv.
login=$(printf '{"token":"%s"}' "${token}" \
  | curl -fsS --max-time 30 -H 'Content-Type: application/json' -d @- "${base}/admin/login") || {
  echo "::error::POST /admin/login failed."
  exit 1
}
session=$(printf '%s' "${login}" | sed -n 's/.*"token":"\([^"]*\)".*/\1/p')
if [ -z "${session}" ]; then
  echo "::error::POST /admin/login returned no session token."
  exit 1
fi

# The Authorization header goes through a curl config on stdin, not argv.
if ! printf 'header = "Authorization: Bearer %s"\n' "${session}" \
  | curl -fsS --max-time 30 -o /dev/null -K - "${base}/games"; then
  echo "::error::GET /games with an admin session failed."
  exit 1
fi
echo "Authenticated smoke request: admin login and GET /games -> 200."
