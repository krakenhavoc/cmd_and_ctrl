#!/usr/bin/env bash
# load-deploy-key.sh — give a CI job the SSH key it logs in to a deploy
# host with, from an environment secret, for this job only (#3335).
#
#   DEPLOY_SSH_KEY="<private key>" scripts/load-deploy-key.sh
#
# Run it AFTER scripts/pin-host-key.sh, which writes SSH_OPTS: this
# appends to that value and writes the result back to $GITHUB_ENV, so
# every later step's ssh and rsync use the key.
#
# Why: the deploy jobs used to log in with whatever key the self-hosted
# runner's user kept in ~/.ssh, so the key sat on the runner hosts all
# the time, readable by every job that ran there, nightly test jobs and
# their npm dependencies included. Held in the `dev` and `prod`
# environments' CMDCTRL_DEPLOY_SSH_KEY secret instead, it exists only
# inside a deploy job (written to $RUNNER_TEMP, removed by the job's
# last step), and only a run on `develop` or `main` can read it, because
# of each environment's deployment branch policy.
#
# DEPLOY_SSH_KEY unset, or not a usable private key: the job fails
# before any ssh (ADR 0181). It used to fall back to the runner's own
# key with a ::warning::, which let #3335 merge before the secrets
# existed. They exist now, so a fallback could only ever mean a job that
# lost its environment, or a runner that still holds a key it should
# not, and both are worth a red run. Runbook:
# docs/environments.md#runners-and-credentials.
#
# With the key loaded, ssh offers ONLY this key (IdentitiesOnly=yes,
# IdentityAgent=none): a stale key left in the runner's ~/.ssh or agent
# is never tried, so a deploy that works proves the secret works.
set -euo pipefail

: "${GITHUB_ENV:?GITHUB_ENV is not set; this runs inside a GitHub Actions job}"
: "${RUNNER_TEMP:?RUNNER_TEMP is not set; this runs inside a GitHub Actions job}"
: "${SSH_OPTS:?SSH_OPTS is not set; run scripts/pin-host-key.sh first}"

key="${DEPLOY_SSH_KEY:-}"

if [ -z "${key}" ]; then
  echo "::error::Environment secret CMDCTRL_DEPLOY_SSH_KEY is empty. A deploy logs in only with its environment's key, never the runner's own. Check that the job names its environment (dev or prod), that the run is on that environment's branch, and that the secret is set there; see docs/environments.md#runners-and-credentials."
  exit 1
fi

key_file="${RUNNER_TEMP}/deploy_ssh_key"
umask 077
printf '%s\n' "${key}" > "${key_file}"
chmod 600 "${key_file}"

if ! ssh-keygen -y -f "${key_file}" >/dev/null 2>&1; then
  rm -f "${key_file}"
  echo "::error::CMDCTRL_DEPLOY_SSH_KEY is set but is not a usable private key. It must be an unencrypted OpenSSH private key, the whole file from '-----BEGIN OPENSSH PRIVATE KEY-----' to the END line."
  exit 1
fi

echo "SSH_OPTS=${SSH_OPTS} -o IdentitiesOnly=yes -o IdentityAgent=none -i ${key_file}" >> "${GITHUB_ENV}"
echo "Deploy key:    from CMDCTRL_DEPLOY_SSH_KEY"
ssh-keygen -lf "${key_file}" | sed 's/^/               /'
