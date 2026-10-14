#!/usr/bin/env bash
# restore-rehearsal.sh — monthly proof that the off-site backup restores
# into a working application (#3340). The weekly `restic check` proves the
# repository is intact; this proves the data in it is usable.
#
# Run by deploy/cmd-and-ctrl-restore-rehearsal.service, started monthly by
# its .timer, as the same user and with the same /etc/cmd_and_ctrl/backup.env
# as the backup. Both units are installed by the CD step "Ensure off-site
# backup". It touches nothing in the live data dir: it restores into a
# scratch directory it creates and removes.
#
# What it does, stopping at the first failure:
#   1. `restic restore latest --tag <env>` of the three paths the check
#      needs (the database copy, playmats/, deletions/) into a scratch dir
#      under the unit's StateDirectory.
#   2. `cmd_and_ctrl-restorecheck verify -dir <scratch>`: copies the
#      database copy into place, opens it with the CURRENT binary (which
#      applies every pending migration, so a migration that cannot run on
#      real backed-up data fails here, not during a disaster), runs
#      integrity_check and foreign_key_check, re-applies the deletion
#      ledger, and checks every playmat row has its image file.
#
# Whatever the outcome, it records the result for the monitoring agent
# (ADR 0123 section 4) in /var/lib/cmdctrl-metrics/restore-rehearsal.prom:
#   cmdctrl_restore_rehearsal_last_exit_code                  this run's exit code
#   cmdctrl_restore_rehearsal_last_success_timestamp_seconds  when a run last
#                                                             exited 0; a failed
#                                                             run carries the
#                                                             previous value over
# alert rule RestoreRehearsalStale fires on a failure and on 35 days
# without a pass. The file is replaced by write-then-rename. A missing
# directory or a failed write is logged and never changes the exit code.
#
# Usage: restore-rehearsal.sh
#
# Environment (backup.env, via the unit's EnvironmentFile=):
#   RESTIC_REPOSITORY, RESTIC_PASSWORD, AWS_ACCESS_KEY_ID,
#   AWS_SECRET_ACCESS_KEY, CMDCTRL_BACKUP_ENV   as for backup-offsite.sh
# Environment (set by the unit):
#   CMDCTRL_DATA_DIR      default /var/lib/cmd_and_ctrl/data; the path
#                         prefix a snapshot may carry
#   RESTIC_CACHE_DIR      the backup unit's cache
#   STATE_DIRECTORY       the unit's StateDirectory=, where the scratch dir goes
# Environment (optional):
#   CMDCTRL_RESTORECHECK          default /opt/cmd_and_ctrl/server/bin/cmd_and_ctrl-restorecheck
#   CMDCTRL_METRICS_TEXTFILE_DIR  default /var/lib/cmdctrl-metrics
#   CMDCTRL_REHEARSAL_SNAPSHOT    default `latest`; a snapshot ID to rehearse an older one
#
# Dependencies: restic.

set -euo pipefail

if (($# > 0)); then
  echo "usage: restore-rehearsal.sh" >&2
  exit 2
fi

DATA_DIR="${CMDCTRL_DATA_DIR:-/var/lib/cmd_and_ctrl/data}"
CHECK_BIN="${CMDCTRL_RESTORECHECK:-/opt/cmd_and_ctrl/server/bin/cmd_and_ctrl-restorecheck}"
SNAPSHOT="${CMDCTRL_REHEARSAL_SNAPSHOT:-latest}"
STATE_DIR="${STATE_DIRECTORY:-/var/lib/cmd-and-ctrl-restore-rehearsal}"
METRICS_DIR="${CMDCTRL_METRICS_TEXTFILE_DIR:-/var/lib/cmdctrl-metrics}"
METRICS_FILE="${METRICS_DIR}/restore-rehearsal.prom"
SCRATCH=""

# Same shape as backup-offsite.sh's write_metrics; always called as
# `write_metrics "$rc" || true`.
write_metrics() {
  local rc="$1" last_success="" tmp
  if [[ ! -d "$METRICS_DIR" ]]; then
    echo "restore-rehearsal: ${METRICS_DIR} does not exist; not recording this run for the monitoring agent"
    return 0
  fi
  if ((rc == 0)); then
    last_success="$(date +%s)"
  elif [[ -r "$METRICS_FILE" ]]; then
    last_success="$(sed -n 's/^cmdctrl_restore_rehearsal_last_success_timestamp_seconds \([0-9][0-9]*\)$/\1/p' "$METRICS_FILE" | tail -n 1)"
  fi
  if ! tmp="$(mktemp "${METRICS_FILE}.XXXXXX")"; then
    echo "restore-rehearsal: cannot create a file in ${METRICS_DIR}; not recording this run" >&2
    return 0
  fi
  if {
    if [[ -n "$last_success" ]]; then
      echo "# HELP cmdctrl_restore_rehearsal_last_success_timestamp_seconds Unix time the restore rehearsal (scripts/restore-rehearsal.sh) last exited 0."
      echo "# TYPE cmdctrl_restore_rehearsal_last_success_timestamp_seconds gauge"
      echo "cmdctrl_restore_rehearsal_last_success_timestamp_seconds ${last_success}"
    fi
    echo "# HELP cmdctrl_restore_rehearsal_last_exit_code Exit code of the restore rehearsal's most recent run (0 success)."
    echo "# TYPE cmdctrl_restore_rehearsal_last_exit_code gauge"
    echo "cmdctrl_restore_rehearsal_last_exit_code ${rc}"
  } >"$tmp" && chmod 0644 "$tmp" && mv -f "$tmp" "$METRICS_FILE"; then
    echo "restore-rehearsal: recorded exit=${rc} in ${METRICS_FILE}"
  else
    rm -f "$tmp"
    echo "restore-rehearsal: could not write ${METRICS_FILE}; not recording this run" >&2
  fi
  return 0
}
on_exit() {
  local rc=$?
  if [[ -n "$SCRATCH" && -d "$SCRATCH" ]]; then
    rm -rf "$SCRATCH"
  fi
  write_metrics "$rc" || true
  exit "$rc"
}
trap on_exit EXIT

missing=()
for var in RESTIC_REPOSITORY RESTIC_PASSWORD CMDCTRL_BACKUP_ENV; do
  if [[ -z "${!var:-}" ]]; then
    missing+=("$var")
  fi
done
if ((${#missing[@]} > 0)); then
  echo "restore-rehearsal: not configured, missing: ${missing[*]} (see /etc/cmd_and_ctrl/backup.env)" >&2
  exit 1
fi
if [[ ! -x "$CHECK_BIN" ]]; then
  echo "restore-rehearsal: ${CHECK_BIN} is missing or not executable; deploy ships it with the server" >&2
  exit 1
fi

mkdir -p "$STATE_DIR"
SCRATCH="$(mktemp -d "${STATE_DIR}/scratch.XXXXXX")"

# 1. Restore the three paths the check needs. Include patterns are given
# both relative to the data dir and under its absolute path, because a
# snapshot may record either (restic matches against the stored path);
# a pattern that matches nothing is not an error. --no-lock: the
# rehearsal only reads, and a read-only token cannot lock.
echo "restore-rehearsal: restoring ${SNAPSHOT} (tag ${CMDCTRL_BACKUP_ENV}) into ${SCRATCH}"
restic --no-lock restore "$SNAPSHOT" --tag "$CMDCTRL_BACKUP_ENV" --target "$SCRATCH" \
  --include "/db/cmdctrl.backup.sqlite" --include "/playmats" --include "/deletions" \
  --include "${DATA_DIR}/db/cmdctrl.backup.sqlite" --include "${DATA_DIR}/playmats" --include "${DATA_DIR}/deletions"

# The tree root is wherever the database copy landed.
dbcopy="$(find "$SCRATCH" -type f -path '*/db/cmdctrl.backup.sqlite' -print -quit)"
if [[ -z "$dbcopy" ]]; then
  echo "restore-rehearsal: the snapshot has no db/cmdctrl.backup.sqlite; the nightly backup is not capturing the database" >&2
  exit 1
fi
root="$(dirname "$(dirname "$dbcopy")")"

# 2. The application half.
"$CHECK_BIN" verify -dir "$root"
echo "restore-rehearsal: env=${CMDCTRL_BACKUP_ENV} snapshot=${SNAPSHOT} ok"
