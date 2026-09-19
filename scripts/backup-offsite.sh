#!/usr/bin/env bash
# backup-offsite.sh — nightly off-node backup of the game server's data
# dir to a restic repository (Cloudflare R2 in production and on the
# preview; any restic backend works). #1031, ADR 0051 decision 1.
#
# Run by deploy/cmd-and-ctrl-backup.service (User=cmdctrl, data dir
# read-only), started nightly by cmd-and-ctrl-backup.timer. Both units
# and /etc/cmd_and_ctrl/backup.env are installed by the CD step
# "Ensure off-site backup" in .github/workflows/ci-cd.yml. The server
# process never reads backup.env.
#
# What it does, in order, stopping at the first failure:
#   1. Initialises the repository if it does not exist yet
#      (`restic cat config` fails). The first run on a fresh bucket
#      does this; every later run skips it.
#   2. Backs up, relative to the data dir, whichever of these exist:
#        db/cmdctrl.backup.sqlite  the server's hourly VACUUM INTO copy.
#                                  NOT the live cmdctrl.sqlite or its
#                                  -wal/-shm: a file copy of a live WAL
#                                  database is not a consistent database.
#        restore/ replays/ lobby/ bugreports/ games/
#      scryfall/, images/ and avatars/ are re-downloadable caches and
#      are excluded explicitly. Snapshots are tagged `cmdctrl` and the
#      environment (CMDCTRL_BACKUP_ENV: prod or dev).
#   3. `restic forget --prune` with 7 daily, 4 weekly, 6 monthly.
#      Retention is restic's, not an R2 lifecycle rule: a lifecycle rule
#      deleting pack files under restic corrupts the repository.
#   4. On Sundays (or with --check), `restic check --read-data-subset=5%`,
#      which downloads and verifies a random 5% of the data.
#
# Exits non-zero on any failure, so systemd marks the unit failed and
# `systemctl --failed` / `journalctl -u cmd-and-ctrl-backup` show it.
# restic's exit 3 (snapshot written but some files unreadable) counts as
# a failure too. Ends with one summary line: snapshot ID, bytes added.
#
# Usage: backup-offsite.sh [--check]
#
# Environment (backup.env, via the unit's EnvironmentFile=):
#   RESTIC_REPOSITORY     e.g. s3:https://<account_id>.r2.cloudflarestorage.com/<bucket>
#   RESTIC_PASSWORD       the repository key. NEVER CHANGE IT once the
#                         repository is initialised: restic cannot open
#                         the repository with a different password, and
#                         every backup after the change fails. It lives
#                         in a GitHub secret and in the owner's password
#                         manager; the latter is the recovery path.
#   AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY
#                         an R2 API token scoped to this environment's bucket
#   CMDCTRL_BACKUP_ENV    prod or dev; the snapshot tag
# Environment (set by the unit):
#   CMDCTRL_DATA_DIR      default /var/lib/cmd_and_ctrl/data
#   RESTIC_CACHE_DIR      the unit's CacheDirectory=
#
# Dependencies: restic (apt; installed by CD if missing).

set -euo pipefail

FORCE_CHECK=0
for arg in "$@"; do
  case "$arg" in
    --check) FORCE_CHECK=1 ;;
    *)
      echo "backup-offsite: unknown argument: $arg" >&2
      echo "usage: backup-offsite.sh [--check]" >&2
      exit 2
      ;;
  esac
done

DATA_DIR="${CMDCTRL_DATA_DIR:-/var/lib/cmd_and_ctrl/data}"

missing=()
for var in RESTIC_REPOSITORY RESTIC_PASSWORD CMDCTRL_BACKUP_ENV; do
  if [[ -z "${!var:-}" ]]; then
    missing+=("$var")
  fi
done
if ((${#missing[@]} > 0)); then
  echo "backup-offsite: not configured, missing: ${missing[*]} (see /etc/cmd_and_ctrl/backup.env)" >&2
  exit 1
fi

if ! [[ "$CMDCTRL_BACKUP_ENV" =~ ^[a-z][a-z0-9-]*$ ]]; then
  echo "backup-offsite: CMDCTRL_BACKUP_ENV must be a lowercase tag (prod, dev), got: ${CMDCTRL_BACKUP_ENV}" >&2
  exit 1
fi

if [[ ! -d "$DATA_DIR" ]]; then
  echo "backup-offsite: data dir ${DATA_DIR} does not exist" >&2
  exit 1
fi

# 1. Initialise on first run. `cat config` is the cheapest call that
# proves the repository exists and the password opens it. Only "there
# is no repository here" leads to init: restic >= 0.17 says so with exit
# code 10, and 0.16 (Ubuntu noble's) with "Is there a repository at the
# following location?". Anything else -- a wrong password -- fails the
# run with restic's own message. 0.16 prints the same hint for a network
# or credentials error, so those reach `restic init` too; that is safe,
# because init refuses to run unless it can see there is no config
# already, and it fails with the underlying error.
rc=0
cat_err="$(restic cat config 2>&1 >/dev/null)" || rc=$?
if ((rc != 0)); then
  if ((rc == 10)) || [[ "$cat_err" == *"Is there a repository at the following location?"* ]]; then
    printf '%s\n' "$cat_err" >&2
    echo "backup-offsite: no repository found at the configured location; running restic init"
    restic init
  else
    printf '%s\n' "$cat_err" >&2
    echo "backup-offsite: cannot open the repository (exit ${rc}). A 'wrong password' here means RESTIC_PASSWORD no longer matches the one the repository was created with; see docs/environments.md#backups." >&2
    exit 1
  fi
fi

# 2. Back up. Paths are relative to the data dir so a restore lays out
# the same tree under whatever target it is given.
cd "$DATA_DIR"
candidates=(
  db/cmdctrl.backup.sqlite
  restore
  replays
  lobby
  bugreports
  games
)
paths=()
for p in "${candidates[@]}"; do
  if [[ -e "$p" ]]; then
    paths+=("$p")
  else
    echo "backup-offsite: skipping ${p} (not present)"
  fi
done
if ((${#paths[@]} == 0)); then
  echo "backup-offsite: nothing to back up under ${DATA_DIR}" >&2
  exit 1
fi

# The excludes are belt and braces: none of the listed paths contains
# them today. They keep the caches, the live database and a half-written
# VACUUM INTO temp file out even if someone later widens the list to
# `db/` or the whole data dir.
#
# --json gives one machine-readable summary line to pull the snapshot
# ID and byte count from. Every other non-progress line (errors,
# warnings) is passed through to stderr, i.e. the journal.
rc=0
backup_out="$(
  restic backup --json --tag cmdctrl --tag "$CMDCTRL_BACKUP_ENV" \
    --exclude scryfall --exclude images --exclude avatars \
    --exclude 'cmdctrl.sqlite' --exclude 'cmdctrl.sqlite-*' \
    --exclude '*.tmp' \
    "${paths[@]}"
)" || rc=$?
grep -v -e '"message_type":"status"' -e '"message_type":"summary"' \
  <<<"$backup_out" >&2 || true
if ((rc != 0)); then
  echo "backup-offsite: restic backup exited ${rc} (3 = snapshot written but some files were unreadable)" >&2
  exit "$rc"
fi
summary_json="$(grep '"message_type":"summary"' <<<"$backup_out" | tail -n 1 || true)"
snapshot_id="$(sed -n 's/.*"snapshot_id":"\([0-9a-f]*\)".*/\1/p' <<<"$summary_json")"
data_added="$(sed -n 's/.*"data_added":\([0-9]*\).*/\1/p' <<<"$summary_json")"
if [[ -z "$snapshot_id" ]]; then
  echo "backup-offsite: restic backup reported no snapshot" >&2
  exit 1
fi

# 3. Retention. Grouped by tags, not restic's default host+paths: the
# path list above changes when a directory first appears (lobby/ on a
# fresh host), and a path-grouped forget would then keep every old
# group forever. One repository per environment, so the tags are the
# group.
restic forget --prune --quiet --tag cmdctrl --group-by tags \
  --keep-daily 7 --keep-weekly 4 --keep-monthly 6

# 4. Weekly integrity check, Sundays (date +%u == 7).
checked="no"
if ((FORCE_CHECK == 1)) || [[ "$(date +%u)" == "7" ]]; then
  restic check --read-data-subset=5%
  checked="yes"
fi

echo "backup-offsite: env=${CMDCTRL_BACKUP_ENV} snapshot=${snapshot_id:0:8} added_bytes=${data_added:-0} paths=${#paths[@]} check=${checked}"
