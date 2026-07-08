#!/usr/bin/env bash
# set-server-env.sh — idempotently upsert one KEY=VALUE into the game
# server's env file (/etc/cmd_and_ctrl/env by default — the path the
# HomeLab cloud-init template writes and the systemd units read via
# EnvironmentFile=).
#
# The file is root:cmdctrl 0640, so run this under sudo (the deploy
# user's cloud-init sudoers grant covers it). The VALUE arrives on
# stdin, not argv, so secrets never show up in `ps` output or shell
# history on the host:
#
#   printf '%s' "$VALUE" | sudo set-server-env.sh KEY [env-file]
#
# Called by the CD job ("Sync server env" step in ci-cd.yml) after
# every deploy, before the service restart picks the file up. Safe to
# run repeatedly: the write is atomic (tmp file + rename) so a
# concurrent restart never reads a torn file, and the existing file's
# owner/group/mode are preserved across the swap.
set -euo pipefail

KEY="${1:?usage: printf '%s' \"\$VALUE\" | sudo set-server-env.sh KEY [env-file]}"
ENV_FILE="${2:-/etc/cmd_and_ctrl/env}"

# Keys are CMDCTRL_*-shaped identifiers; anything else is a caller bug.
if ! [[ "$KEY" =~ ^[A-Z_][A-Z0-9_]*$ ]]; then
  echo "set-server-env: invalid key: ${KEY}" >&2
  exit 1
fi

VALUE="$(cat)"
# Env files are line-oriented; a value with a newline would smuggle
# in a second assignment.
if [[ "$VALUE" == *$'\n'* ]]; then
  echo "set-server-env: value must be a single line" >&2
  exit 1
fi

mkdir -p "$(dirname "$ENV_FILE")"

TMP="$(mktemp "${ENV_FILE}.XXXXXX")"
trap 'rm -f "$TMP"' EXIT
if [[ -e "$ENV_FILE" ]]; then
  # Drop any existing assignment of KEY (commented lines survive)
  # and carry the file's owner/group/mode onto the replacement so
  # the swap is invisible to systemd and the service user.
  grep -v -E "^${KEY}=" "$ENV_FILE" >"$TMP" || true
  chmod --reference="$ENV_FILE" "$TMP"
  chown --reference="$ENV_FILE" "$TMP"
else
  chmod 0640 "$TMP"
fi
printf '%s=%s\n' "$KEY" "$VALUE" >>"$TMP"
mv "$TMP" "$ENV_FILE"
trap - EXIT
echo "set-server-env: ${KEY} updated in ${ENV_FILE}"
