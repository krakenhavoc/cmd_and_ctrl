#!/usr/bin/env bash
# provision-dev.sh — stand up the develop preview environment on the
# VPS that already runs production. Run once, as root. Idempotent:
# re-running fixes drift and never clobbers a secret it did not
# generate.
#
#   sudo deploy/provision-dev.sh --check    # report only, change nothing
#   sudo deploy/provision-dev.sh
#
# Covers everything except the reverse proxy, TLS certificate, the
# extra Discord redirect URI, and the GitHub repo variable — those
# are listed at the end for you to do by hand. See
# docs/environments.md and docs/decisions/0023-develop-environment.md.
#
# SAFETY: this script must never be able to damage production. It
# only ever creates paths with a "-dev" suffix, reads production's
# env file without writing to it, and refuses outright if any target
# path resolves to a production one. The one shared resource is the
# Scryfall bulk dump, which it exposes read-only.
set -euo pipefail

# --- configuration ---------------------------------------------------
# Every path is an override so a rehearsal on a scratch host (or the
# test at the bottom of this file's history) can point them elsewhere.
DEV_USER="${DEV_USER:-cmdctrl-dev}"
DEPLOY_USER="${DEPLOY_USER:-krkn}"          # the user CI rsyncs as
DEV_ROOT="${DEV_ROOT:-/opt/cmd_and_ctrl-dev}"
DEV_WEB_ROOT="${DEV_WEB_ROOT:-/var/www/cmdctrl-client-dev}"
DEV_DATA_DIR="${DEV_DATA_DIR:-/var/lib/cmd_and_ctrl-dev}"
DEV_ENV_FILE="${DEV_ENV_FILE:-/etc/cmd_and_ctrl/dev.env}"
PROD_ENV_FILE="${PROD_ENV_FILE:-/etc/cmd_and_ctrl/env}"
DEV_SERVICE="${DEV_SERVICE:-cmd-and-ctrl-dev}"
PROD_GROUP="${PROD_GROUP:-cmdctrl}"
DEV_ADDR="${DEV_ADDR:-:8081}"
DEV_PUBLIC_URL="${DEV_PUBLIC_URL:-https://dev.cmd.labxp.io}"
# SKIP_SYSTEM=1 skips useradd / systemctl, for exercising the file
# half of this script somewhere that isn't the real host.
SKIP_SYSTEM="${SKIP_SYSTEM:-0}"

CHECK=0
[[ "${1:-}" == "--check" ]] && CHECK=1

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
CHANGED=0
PENDING=0
MANUAL=()

# --- output helpers --------------------------------------------------
ok()   { printf '  \033[32m✓\033[0m %s\n' "$*"; }
skip() { printf '  \033[90m·\033[0m %s\n' "$*"; }
die()  { printf '\033[31mERROR:\033[0m %s\n' "$*" >&2; exit 1; }
head_() { printf '\n\033[1m%s\033[0m\n' "$*"; }

# did <base-verb-phrase> — reports one mutation. Under --check it says
# what would happen and counts as pending, never as done: a dry run
# that claims to have created things is worse than no dry run.
did() {
  if (( CHECK )); then
    printf '  \033[90m·\033[0m would %s\n' "$*"
    PENDING=$((PENDING+1))
  else
    printf '  \033[33m+\033[0m %s\n' "$*"
    CHANGED=$((CHANGED+1))
  fi
}

# run CMD... — a no-op under --check. Reporting is did()'s job, so
# this stays silent and the two can't disagree.
run() {
  (( CHECK )) && return 0
  "$@"
}

# --- preflight -------------------------------------------------------
head_ "Preflight"

if [[ "$(id -u)" != "0" ]]; then
  # SKIP_SYSTEM=1 means this is a rehearsal on something that is not
  # the real host, so a non-root run is expected and the file half is
  # still worth exercising. Anywhere else, fail with a clear message
  # rather than a confusing permission error three steps in.
  if (( CHECK )) || (( SKIP_SYSTEM )); then
    skip "not root — ownership and systemd steps will be skipped"
  else
    die "must run as root (try: sudo $0)"
  fi
fi

# The load-bearing guard. Every one of these must differ from the
# production path it shadows; a config override that collapses them
# would have this script chowning or symlinking inside production.
for pair in \
  "DEV_ROOT=$DEV_ROOT:/opt/cmd_and_ctrl" \
  "DEV_ENV_FILE=$DEV_ENV_FILE:$PROD_ENV_FILE" \
  "DEV_WEB_ROOT=$DEV_WEB_ROOT:/var/www/cmdctrl-client" \
  "DEV_DATA_DIR=$DEV_DATA_DIR:/var/lib/cmd_and_ctrl" \
  "DEV_SERVICE=$DEV_SERVICE:cmd-and-ctrl"
do
  name_val="${pair%:*}"; prod="${pair##*:}"; val="${name_val#*=}"
  [[ "$val" == "$prod" ]] && die "${name_val%%=*} is set to production's value ($prod). Refusing."
done
ok "no dev path collides with a production path"

if [[ ! -f "$PROD_ENV_FILE" ]]; then
  die "production env file $PROD_ENV_FILE not found — is this the right host?"
fi
ok "production env file present (read-only to this script)"

# Production's data dir is the source of the Scryfall dump. Read it
# rather than guessing: the runbook used to say "replace this with
# production's real CMDCTRL_DATA_DIR", which is an invitation to get
# it wrong.
PROD_DATA_DIR="$(sed -n 's/^CMDCTRL_DATA_DIR=//p' "$PROD_ENV_FILE" | tail -1)"
if [[ -z "$PROD_DATA_DIR" ]]; then
  skip "production sets no CMDCTRL_DATA_DIR; Scryfall sharing will be skipped"
else
  ok "production data dir: $PROD_DATA_DIR"
fi

# --- service user ----------------------------------------------------
head_ "Service user"
if id "$DEV_USER" >/dev/null 2>&1; then
  ok "user $DEV_USER exists"
elif (( SKIP_SYSTEM )); then
  skip "SKIP_SYSTEM=1, not creating $DEV_USER"
else
  run useradd --system --no-create-home --shell /usr/sbin/nologin "$DEV_USER" \
    && did "create system user $DEV_USER"
fi

# --- directories -----------------------------------------------------
head_ "Directories"
# Owned by the deploy user because CI rsyncs into them as that user.
for d in "$DEV_ROOT/server/bin" "$DEV_ROOT/scripts" "$DEV_WEB_ROOT"; do
  if [[ -d "$d" ]]; then ok "$d"; else run mkdir -p "$d" && did "create $d"; fi
done
if id "$DEPLOY_USER" >/dev/null 2>&1; then
  run chown -R "$DEPLOY_USER":"$DEPLOY_USER" "$DEV_ROOT" "$DEV_WEB_ROOT" && ok "owned by $DEPLOY_USER"
else
  skip "deploy user $DEPLOY_USER not found; leaving ownership alone"
  MANUAL+=("chown $DEV_ROOT and $DEV_WEB_ROOT to whichever user CI rsyncs as")
fi
# The data dir is created by systemd (StateDirectory=) on first start,
# but the Scryfall symlink below needs it to exist now.
if [[ -d "$DEV_DATA_DIR" ]]; then ok "$DEV_DATA_DIR"; else run mkdir -p "$DEV_DATA_DIR" && did "create $DEV_DATA_DIR"; fi
if ! (( SKIP_SYSTEM )) && id "$DEV_USER" >/dev/null 2>&1; then
  run chown "$DEV_USER":"$DEV_USER" "$DEV_DATA_DIR"
  run chmod 0750 "$DEV_DATA_DIR"
fi

# --- env file --------------------------------------------------------
head_ "Environment file"
SETENV="$SCRIPT_DIR/../scripts/set-server-env.sh"
[[ -x "$SETENV" ]] || SETENV="$DEV_ROOT/scripts/set-server-env.sh"

if [[ -f "$DEV_ENV_FILE" ]]; then
  ok "$DEV_ENV_FILE exists (values left as they are)"
else
  run mkdir -p "$(dirname "$DEV_ENV_FILE")"
  # root:$DEV_USER 0640 is what the unit's EnvironmentFile= expects.
  # The fallbacks exist so a rehearsal on a scratch host (not root, no
  # such group) still exercises the rest of the script rather than
  # aborting here.
  run install -o root -g "$DEV_USER" -m 0640 /dev/null "$DEV_ENV_FILE" 2>/dev/null \
    || run install -o root -g root -m 0640 /dev/null "$DEV_ENV_FILE" 2>/dev/null \
    || { run install -m 0640 /dev/null "$DEV_ENV_FILE"
         MANUAL+=("chown root:$DEV_USER $DEV_ENV_FILE — could not set ownership"); }
  did "create $DEV_ENV_FILE (0640, group $DEV_USER)"
fi

# Upsert only the keys this environment cannot work without, and only
# when they are missing. An existing value is never overwritten —
# re-running must not rotate a token out from under a live session.
set_if_absent() {
  local key="$1" val="$2"
  if grep -q "^${key}=" "$DEV_ENV_FILE" 2>/dev/null; then
    ok "$key already set"
    return
  fi
  if (( CHECK )); then did "set $key"; return; fi
  printf '%s=%s\n' "$key" "$val" >> "$DEV_ENV_FILE"
  did "set $key"
}

# A DIFFERENT admin token from production, generated here so nobody is
# tempted to paste production's in.
if ! grep -q '^CMDCTRL_ADMIN_TOKEN=' "$DEV_ENV_FILE" 2>/dev/null; then
  GENERATED_TOKEN="$(openssl rand -hex 24 2>/dev/null || head -c 32 /dev/urandom | base64 | tr -d '/+=' | head -c 48)"
  set_if_absent CMDCTRL_ADMIN_TOKEN "$GENERATED_TOKEN"
  if (( CHECK )); then
    MANUAL+=("a dev admin token will be generated on the real run — distinct from production's")
  else
    MANUAL+=("dev admin token generated — read it with: sudo grep CMDCTRL_ADMIN_TOKEN $DEV_ENV_FILE")
  fi
else
  ok "CMDCTRL_ADMIN_TOKEN already set"
fi

set_if_absent CMDCTRL_ENV dev
set_if_absent CMDCTRL_ADDR "$DEV_ADDR"
set_if_absent CMDCTRL_DATA_DIR "$DEV_DATA_DIR"
set_if_absent CMDCTRL_SECURE_COOKIES 1
set_if_absent CMDCTRL_TRUST_FORWARDED 1
set_if_absent CMDCTRL_PUBLIC_BASE_URL "$DEV_PUBLIC_URL"
set_if_absent CMDCTRL_DISCORD_REDIRECT_URI "$DEV_PUBLIC_URL/auth/discord/callback"

# Discord client id/secret are copied from production deliberately —
# same application, extra redirect URI (ADR 0023 §6). Copied rather
# than prompted so there is nothing to mistype.
for key in CMDCTRL_DISCORD_CLIENT_ID CMDCTRL_DISCORD_CLIENT_SECRET; do
  if grep -q "^${key}=" "$DEV_ENV_FILE" 2>/dev/null; then
    ok "$key already set"
  else
    val="$(sed -n "s/^${key}=//p" "$PROD_ENV_FILE" | tail -1)"
    if [[ -n "$val" ]]; then
      set_if_absent "$key" "$val"
    else
      skip "$key not set in production either; Discord sign-in stays disabled on dev"
    fi
  fi
done

# NOT copied: CMDCTRL_GITHUB_TOKEN. CD provisions it, and a dev
# deployment filing bug reports into the real tracker is noise.

# --- systemd ---------------------------------------------------------
head_ "systemd unit"
UNIT_SRC="$SCRIPT_DIR/${DEV_SERVICE}.service"
UNIT_DST="/etc/systemd/system/${DEV_SERVICE}.service"
if [[ ! -f "$UNIT_SRC" ]]; then
  skip "unit source $UNIT_SRC not found; skipping"
  MANUAL+=("install ${DEV_SERVICE}.service from deploy/ by hand")
elif (( SKIP_SYSTEM )); then
  skip "SKIP_SYSTEM=1, not installing the unit"
elif cmp -s "$UNIT_SRC" "$UNIT_DST" 2>/dev/null; then
  ok "$UNIT_DST is current"
else
  run cp "$UNIT_SRC" "$UNIT_DST" && did "install $UNIT_DST"
  run systemctl daemon-reload
  run systemctl enable "$DEV_SERVICE" && did "enable $DEV_SERVICE"
fi

# --- Scryfall dump, shared read-only ---------------------------------
head_ "Scryfall bulk dump (shared read-only with production)"
if [[ -z "$PROD_DATA_DIR" ]]; then
  skip "no production data dir; give dev its own scryfall-refresh cron instead"
  MANUAL+=("run scripts/scryfall-refresh.sh for the dev data dir, or symlink it later")
elif [[ ! -d "$PROD_DATA_DIR/scryfall" ]]; then
  skip "$PROD_DATA_DIR/scryfall does not exist yet"
  MANUAL+=("re-run this script after production's first scryfall-refresh")
elif [[ -e "$DEV_DATA_DIR/scryfall" ]]; then
  ok "$DEV_DATA_DIR/scryfall already present"
else
  run ln -s "$PROD_DATA_DIR/scryfall" "$DEV_DATA_DIR/scryfall" && did "symlink the dump"
  # Read + traverse only. Never g+w: dev must not be able to corrupt
  # the file production reads.
  run chgrp -R "$PROD_GROUP" "$PROD_DATA_DIR/scryfall" || true
  run chmod -R g+rX "$PROD_DATA_DIR/scryfall" || true
  if ! (( SKIP_SYSTEM )); then
    run usermod -aG "$PROD_GROUP" "$DEV_USER" && did "add $DEV_USER to $PROD_GROUP (read-only access to the dump)"
  fi
fi

# --- what is left ----------------------------------------------------
head_ "Summary"
if (( CHECK )); then
  printf '  --check: nothing was changed. %d change(s) pending.\n' "$PENDING"
else
  printf '  %d change(s) applied.\n' "$CHANGED"
fi

head_ "Still to do by hand"
cat <<MANUALSTEPS
  1. Reverse proxy: a vhost for dev.cmd.labxp.io -> 127.0.0.1${DEV_ADDR},
     static root ${DEV_WEB_ROOT}, WebSocket upgrade on /ws.
     BOTH vhosts (dev and production) need /config proxied; dev also
     needs /dev. Until then the client falls back to production
     defaults, which hides every dev feature.
  2. TLS certificate for dev.cmd.labxp.io.
  3. Discord: add ${DEV_PUBLIC_URL}/auth/discord/callback to the
     existing application's redirect URIs. Same client id/secret.
  4. GitHub repo variable CMDCTRL_DEV_PUBLIC_BASE_URL=${DEV_PUBLIC_URL}
     and branch protection on 'develop'.
  5. Push the branches, merge into develop, and watch the deploy.
MANUALSTEPS
for m in "${MANUAL[@]:-}"; do [[ -n "$m" ]] && printf '  ! %s\n' "$m"; done

head_ "Verify after the first deploy"
cat <<VERIFY
  systemctl status ${DEV_SERVICE}
  curl -s http://127.0.0.1${DEV_ADDR}/config     # {"env":"dev",...}
  curl -s https://cmd.labxp.io/config            # {"env":"prod",...}
VERIFY
