#!/usr/bin/env bash
# ensure-monitoring-agent.sh — install and configure Grafana Alloy, the
# monitoring agent on the app hosts, or stand it down (ADR 0123 §4, §10).
# Runbook: docs/environments.md#monitoring-agent.
#
# Run as root by the CD step "Ensure monitoring agent" in
# .github/workflows/ci-cd.yml, on both hosts, on every deploy:
#
#   ensure-monitoring-agent.sh off
#       Creates the textfile directory, then stops and disables Alloy if
#       it is installed. CD runs this when the monitoring URL or this
#       environment's push password is unset or a REPLACE_ME placeholder.
#
#   printf '%s' "$ENV_FILE_CONTENT" | ensure-monitoring-agent.sh on SRC_DIR
#       Creates the textfile directory, adds Grafana's apt repository (key
#       checked against its pinned fingerprint), installs the pinned Alloy
#       and holds it there, then installs from SRC_DIR (deploy/alloy/ in
#       this repo, rsynced to the host) and stdin:
#         SRC_DIR/config.alloy -> /etc/alloy/config.alloy   root:alloy 0640
#         stdin                -> /etc/alloy/env            root:alloy 0640
#         SRC_DIR/cmdctrl.conf -> /etc/systemd/system/alloy.service.d/cmdctrl.conf
#       The config is checked with `alloy validate` before it is installed.
#       Alloy is restarted only when a file or the package changed, and
#       started if it was stopped. stdin carries the push password, so it
#       never appears in argv.
#
# Exits non-zero on a real failure (apt, a key with the wrong
# fingerprint, an invalid config). Alloy that is not ready after its
# restart, and a server that does not answer on its metrics address, are
# GitHub `::warning::` lines and exit 0: the game does not depend on its
# monitor.

set -euo pipefail

# The pinned Alloy. To move it: change this, check deploy/alloy/config.alloy
# with the same version's `alloy validate` (the grafana/alloy image of the
# same tag), and deploy. The package is held, so nothing else moves it.
ALLOY_VERSION="1.20.1-1"
# Grafana's apt signing key (https://apt.grafana.com/gpg.key). A key with a
# different fingerprint fails the run: a rotation is for a human to check.
GRAFANA_KEY_FPR="B53AE77BADB630A683046005963FA27710458545"
GRAFANA_KEY_URL="https://apt.grafana.com/gpg.key"
KEYRING=/etc/apt/keyrings/grafana.asc
SOURCES=/etc/apt/sources.list.d/grafana.list

# root:cmdctrl 0775: the backup unit (User=cmdctrl) writes backup.prom here,
# and Alloy's textfile collector reads it.
TEXTFILE_DIR=/var/lib/cmdctrl-metrics

CONFIG=/etc/alloy/config.alloy
ENV_FILE=/etc/alloy/env
DROPIN_DIR=/etc/systemd/system/alloy.service.d
DROPIN="${DROPIN_DIR}/cmdctrl.conf"
ALLOY_READY_URL=http://127.0.0.1:12345/-/ready
# CMDCTRL_METRICS_ADDR, which CD's "Sync server env (metrics listener)" sets.
SERVER_METRICS_URL=http://127.0.0.1:9464/metrics

usage() {
  echo "usage: ensure-monitoring-agent.sh off" >&2
  echo "       ensure-monitoring-agent.sh on SRC_DIR < env-file-content" >&2
  exit 2
}

log() { echo "ensure-monitoring-agent: $*"; }

if [[ "${EUID}" -ne 0 ]]; then
  echo "ensure-monitoring-agent: must run as root (sudo)" >&2
  exit 1
fi

ensure_textfile_dir() {
  install -d -o root -g cmdctrl -m 0775 "$TEXTFILE_DIR"
  log "${TEXTFILE_DIR} is root:cmdctrl 0775"
}

# Every helper below is called as a plain command, never as an `if` or
# `||` condition: bash ignores `set -e` inside a function called that way,
# so a failed apt-get there would be carried on past. Results go out
# through the globals below instead of return codes.
CHANGED=0     # set to 1 by anything that needs Alloy restarted
INSTALLED=0   # install_if_changed's answer for its last call
ENV_NEW=""    # the staged env file; global so the EXIT trap can remove it
trap 'if [[ -n "$ENV_NEW" ]]; then rm -f "$ENV_NEW"; fi' EXIT

# install_if_changed SRC DST OWNER GROUP MODE: installs SRC at DST unless
# DST already has the same content, owner, group and mode. Sets INSTALLED.
install_if_changed() {
  local src="$1" dst="$2" owner="$3" group="$4" mode="$5"
  INSTALLED=0
  if [[ -f "$dst" ]] && cmp -s "$src" "$dst" &&
    [[ "$(stat -c '%U:%G %a' "$dst")" == "${owner}:${group} ${mode#0}" ]]; then
    log "${dst} unchanged"
    return 0
  fi
  install -o "$owner" -g "$group" -m "$mode" "$src" "$dst"
  INSTALLED=1
  log "${dst} installed"
}

# The installed Alloy's version, or nothing. A package that was removed
# but left its conffiles behind counts as not installed.
installed_version() {
  dpkg-query -W -f='${db:Status-Status} ${Version}' alloy 2>/dev/null |
    sed -n 's/^installed //p' || true
}

# Grafana's key, refreshed on every run so an extended expiry reaches the
# host before apt needs it, and checked against the pinned fingerprint
# before apt ever sees it.
ensure_repo() {
  local tmp got
  tmp="$(mktemp)"
  if ! curl -fsSL --max-time 60 "$GRAFANA_KEY_URL" -o "$tmp"; then
    rm -f "$tmp"
    if [[ "$(installed_version)" == "$ALLOY_VERSION" && -f "$KEYRING" ]]; then
      echo "::warning title=Monitoring agent::Could not fetch ${GRAFANA_KEY_URL} on $(hostname); kept the installed key. Alloy ${ALLOY_VERSION} is already installed, so this deploy does not need it."
      return 0
    fi
    echo "ensure-monitoring-agent: could not fetch ${GRAFANA_KEY_URL}" >&2
    exit 1
  fi
  if ! command -v gpg >/dev/null 2>&1; then
    DEBIAN_FRONTEND=noninteractive apt-get install -y -qq gpg >/dev/null
  fi
  got="$(gpg --show-keys --with-colons "$tmp" 2>/dev/null |
    awk -F: '$1 == "fpr" { print $10; exit }' || true)"
  if [[ "$got" != "$GRAFANA_KEY_FPR" ]]; then
    rm -f "$tmp"
    echo "ensure-monitoring-agent: ${GRAFANA_KEY_URL} has fingerprint '${got}', not the pinned ${GRAFANA_KEY_FPR}. If Grafana rotated its key, check the new one and update GRAFANA_KEY_FPR in scripts/ensure-monitoring-agent.sh." >&2
    exit 1
  fi
  install -d -m 0755 "$(dirname "$KEYRING")"
  install_if_changed "$tmp" "$KEYRING" root root 0644
  echo "deb [signed-by=${KEYRING}] https://apt.grafana.com stable main" >"$tmp"
  install_if_changed "$tmp" "$SOURCES" root root 0644
  rm -f "$tmp"
}

# Installs the pinned version unless it is already there, and holds it.
ensure_package() {
  local have
  have="$(installed_version)"
  if [[ "$have" == "$ALLOY_VERSION" ]]; then
    apt-mark hold alloy >/dev/null
    log "alloy ${have} already installed (held)"
    return 0
  fi
  # Only Grafana's list: the rest of apt's lists are not this step's business.
  apt-get update -qq \
    -o Dir::Etc::sourcelist="$SOURCES" \
    -o Dir::Etc::sourceparts=- \
    -o APT::Get::List-Cleanup=0
  # confold: /etc/alloy/config.alloy is the package's conffile and ours
  # replaces it; keep ours on an upgrade, and never prompt.
  DEBIAN_FRONTEND=noninteractive apt-get install -y -qq \
    --allow-change-held-packages --allow-downgrades \
    -o Dpkg::Options::=--force-confdef -o Dpkg::Options::=--force-confold \
    "alloy=${ALLOY_VERSION}"
  apt-mark hold alloy >/dev/null
  CHANGED=1
  log "alloy ${ALLOY_VERSION} installed and held (was: ${have:-not installed})"
}

cmd_off() {
  ensure_textfile_dir
  if [[ -n "$(installed_version)" ]]; then
    systemctl disable --now alloy >/dev/null 2>&1 || true
    log "alloy stopped and disabled"
  else
    log "alloy is not installed; nothing to stop"
  fi
}

cmd_on() {
  local src="$1" i
  if [[ ! -f "${src}/config.alloy" || ! -f "${src}/cmdctrl.conf" ]]; then
    echo "ensure-monitoring-agent: ${src} must hold config.alloy and cmdctrl.conf" >&2
    exit 1
  fi
  # mktemp makes it 0600, so the password is never readable by anyone else
  # while it waits here.
  ENV_NEW="$(mktemp)"
  cat >"$ENV_NEW"
  if [[ ! -s "$ENV_NEW" ]]; then
    echo "ensure-monitoring-agent: no env file content on stdin" >&2
    exit 1
  fi

  ensure_textfile_dir
  ensure_repo
  ensure_package

  # Refuse a config this Alloy cannot load before it replaces a working one.
  alloy validate "${src}/config.alloy"
  log "config.alloy is valid for alloy $(installed_version)"

  install_if_changed "${src}/config.alloy" "$CONFIG" root alloy 0640
  ((INSTALLED == 0)) || CHANGED=1
  install_if_changed "$ENV_NEW" "$ENV_FILE" root alloy 0640
  ((INSTALLED == 0)) || CHANGED=1
  install -d -m 0755 "$DROPIN_DIR"
  install_if_changed "${src}/cmdctrl.conf" "$DROPIN" root root 0644
  if ((INSTALLED == 1)); then
    systemctl daemon-reload
    CHANGED=1
  fi

  systemctl enable --quiet alloy
  if ((CHANGED == 1)); then
    systemctl restart alloy
    log "alloy restarted"
  elif ! systemctl is-active --quiet alloy; then
    systemctl start alloy
    log "alloy started"
  else
    log "nothing changed; alloy left running"
  fi

  for i in $(seq 1 20); do
    if curl -fsS --max-time 2 -o /dev/null "$ALLOY_READY_URL" 2>/dev/null; then
      log "alloy ready on 127.0.0.1:12345 (${i} s)"
      break
    fi
    if ((i == 20)); then
      echo "::warning title=Monitoring agent not ready::Alloy on $(hostname) is not ready 20 s after this deploy (state: $(systemctl is-active alloy || true)). Check journalctl -u alloy on that host (docs/environments.md#monitoring-agent)."
    fi
    sleep 1
  done

  if curl -fsS --max-time 5 -o /dev/null "$SERVER_METRICS_URL"; then
    log "the server answers on ${SERVER_METRICS_URL}"
  else
    echo "::warning title=Server metrics not served::The game server on $(hostname) does not answer ${SERVER_METRICS_URL}, so Alloy has nothing of the game's to scrape. Check CMDCTRL_METRICS_ADDR in /etc/cmd_and_ctrl/env and journalctl -u cmd-and-ctrl (docs/environments.md#monitoring-agent)."
  fi
}

case "${1:-}" in
  off)
    [[ $# -eq 1 ]] || usage
    cmd_off
    ;;
  on)
    [[ $# -eq 2 ]] || usage
    cmd_on "$2"
    ;;
  *) usage ;;
esac
