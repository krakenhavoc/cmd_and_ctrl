#!/usr/bin/env bash
# pin-host-key.sh — decide how a CI job trusts a deploy host's SSH key
# (#1739), and export the ssh options that say so.
#
#   HOST_KEY="<public key>" scripts/pin-host-key.sh <host> <key-variable-name>
#
# Every job that sshes to a VM runs on ANY self-hosted runner. Trusting
# what each runner's ~/.ssh/known_hosts happened to record (accept-new)
# put the trust decision on every runner, and a VM rebuild — which
# changes the host key — broke all of them until each was hand-edited.
#
# So the key is pinned in a repo variable instead: the host's PUBLIC key,
# one "<type> <base64>" per line (a line of `ssh-keyscan` output, host
# field included, is accepted too). This writes it to a known_hosts file
# private to the job and checks strictly against that file only. A
# changed key is still refused; the fix after a rebuild is one variable,
# set after checking the fingerprint on the VM's console. Runbook:
# docs/environments.md#ssh-host-keys.
#
# HOST_KEY unset: falls back to the runner's own known_hosts with
# accept-new, and says so with a ::warning::.
#
# Writes SSH_OPTS to $GITHUB_ENV, so every later step reads it. A job
# must NOT also declare SSH_OPTS in its `env:` — a job-level value wins
# over $GITHUB_ENV.
set -euo pipefail

host="${1:?usage: pin-host-key.sh <host> <key-variable-name>}"
key_var="${2:?usage: pin-host-key.sh <host> <key-variable-name>}"
host_key="${HOST_KEY:-}"
: "${GITHUB_ENV:?GITHUB_ENV is not set; this runs inside a GitHub Actions job}"
: "${RUNNER_TEMP:?RUNNER_TEMP is not set; this runs inside a GitHub Actions job}"

base="-o ConnectTimeout=10"

if [ -z "${host_key}" ]; then
  echo "::warning::Repo variable ${key_var} is not set, so ${host}'s SSH host key is trusted from this runner's own ~/.ssh/known_hosts (accept-new). After a VM rebuild that fails with 'REMOTE HOST IDENTIFICATION HAS CHANGED' on every runner. Set ${key_var} to the host's public key; see docs/environments.md#ssh-host-keys."
  echo "SSH_OPTS=${base} -o StrictHostKeyChecking=accept-new" >> "${GITHUB_ENV}"
  exit 0
fi

kh="${RUNNER_TEMP}/known_hosts.${host}"
printf '%s\n' "${host_key}" \
  | awk -v h="${host}" 'NF && $1 !~ /^#/ { if (NF >= 3 && $2 ~ /^(ssh-|ecdsa-)/) print h, $2, $3; else print h, $1, $2 }' \
  > "${kh}"

if [ ! -s "${kh}" ] || ! ssh-keygen -lf "${kh}" >/dev/null 2>&1; then
  echo "::error::Repo variable ${key_var} is set but is not a usable public host key. It should hold a line like 'ssh-ed25519 AAAA...' (the output of 'ssh-keyscan -t ed25519 ${host}' is fine)."
  exit 1
fi

echo "SSH_OPTS=${base} -o StrictHostKeyChecking=yes -o UserKnownHostsFile=${kh} -o GlobalKnownHostsFile=/dev/null" >> "${GITHUB_ENV}"
echo "Host key:      pinned by ${key_var}"
ssh-keygen -lf "${kh}" | sed 's/^/               /'
