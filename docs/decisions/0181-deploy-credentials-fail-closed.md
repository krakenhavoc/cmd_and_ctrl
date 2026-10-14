# ADR 0181 — Deploy credentials: deploy runners, environment keys, and failing closed

**Status:** Accepted · 2026-10-14 · S70 — Public beta: a table strangers can find
**Issue:** [#3335](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3335)
**Numbering:** every ADR file on any remote head was listed (`git log --remotes --name-only -- docs/decisions`, twelve remote branches); the highest was **0180** (`seam/secret-council`), so this is **0181**.
**Builds on:** [ADR 0023](0023-develop-environment.md) (develop and main, two hosts), #1739 (pinned host keys), #3356 (the first half of #3335).

---

## Context

The 2026-10-13 audit (#3335) found two ways to reach production past review:

1. **The deploy runners also ran test code.** The self-hosted runners that SSH to both hosts as `krkn`, who has `sudo`, held that key in the runner user's `~/.ssh`. The same runners ran the nightly Playwright, catalog-soak, realdump and soak-tracker jobs, so a poisoned npm or Go dependency could read the key and be root on both hosts.
2. **A deploy key could push to `main` past every check.** The `main` ruleset let any deploy key bypass it, and the write-enabled census key was a repository secret any same-repository run could read.

#3356 fixed both, but it had to merge before the owner had made the new credentials, so every new control fell back to the old behaviour with a `::warning::`:

- `CMDCTRL_DEPLOY_RUNNER_LABEL` unset meant any self-hosted runner;
- `CMDCTRL_DEPLOY_SSH_KEY` unset meant the runner's own key;
- a host key variable unset meant `accept-new` against the runner's `known_hosts`.

Since then the owner has set all of them (issue comments of 2026-10-14). The old keys have been removed from both hosts' `authorized_keys`, and the repository-level server secrets are deleted. The fallbacks now protect nothing. They only mean that deleting one variable or secret quietly reopens the hole, with a yellow warning nobody reads on a green run.

## Decision

### 1. The model

| What | Where | Who can use it |
|---|---|---|
| Runners that reach the hosts | self-hosted runners labelled `deploy` (`gunner-0`, `gunner-1`) | only jobs that say `runs-on: [self-hosted, deploy]` |
| `krkn`'s SSH key | environment secret `CMDCTRL_DEPLOY_SSH_KEY`, one key per host, in `dev` and `prod` | a job naming that environment, on its branch (`develop` → `dev`, `main` → `prod`). The other host refuses it |
| Host public keys | variables `CMDCTRL_HOST_KEY`, `CMDCTRL_DEV_HOST_KEY` | public |
| Server and bot secrets | `dev` and `prod` environment secrets | as the SSH key |
| Census push key | environment secret `CENSUS_DEPLOY_KEY` in `census`, `develop` only | `census-publish` |
| Ruleset bypass | `develop` only, for deploy keys (the census key is the only write key) | `main` has no bypass |

Test jobs, nightly ones included, run on GitHub-hosted runners and name no environment, so none of these is in reach of dependency code.

### 2. Fail closed

Each control is now required, and its fallback is removed:

- **Runner.** The deploy jobs and `scryfall-refresh.yml` say `runs-on: [self-hosted, deploy]` literally. `CMDCTRL_DEPLOY_RUNNER_LABEL` is no longer read. A label written into the workflow is changed by a reviewed PR, while a variable is changed by anyone with admin, with no record in the repo. With no runner labelled `deploy`, a deploy queues. It never runs elsewhere.
- **SSH key.** `scripts/load-deploy-key.sh` fails the job when `CMDCTRL_DEPLOY_SSH_KEY` is empty. With the key loaded, ssh offers only that key (`IdentitiesOnly=yes`, `IdentityAgent=none`), so a stale key left on a runner is never tried. The job removes the key in an `if: always()` last step.
- **Host key.** `scripts/pin-host-key.sh` fails the job when the host key variable is empty. It no longer falls back to `accept-new`.

A red deploy is the right outcome for each case. It means a job lost its environment, someone deleted a credential, or a runner was relabelled, and each of those needs a person to look.

### 3. Rules for a new job that reaches a host

This is the contract #3339 (rollback) and #3340 (backups) build on. A job that SSHes or rsyncs to a host:

1. runs on `[self-hosted, deploy]`;
2. names `environment: dev` or `environment: prod`, matching the host it touches, so the branch policy decides who may run it;
3. runs `scripts/pin-host-key.sh <host> <var>` and then `scripts/load-deploy-key.sh` before its first ssh, and does not set `SSH_OPTS` in its own `env:`;
4. removes `${RUNNER_TEMP}/deploy_ssh_key` in an `if: always()` last step.

A `workflow_dispatch` job in `prod` can only run from `main`, because of the environment's branch policy. The policy is the gate, so the job adds no `if:` of its own.

Any other job never names `dev` or `prod`, and never asks for a self-hosted runner. A job that seems to need one is a design question, not a `runs-on` line.

### 4. What stays on the host side

`krkn` still has broad `sudo`, so its key is root on its host. Making `${DEPLOY_ROOT}/scripts` root-owned would not help while the grant is broad, and CD cannot install root-owned files without that same `sudo`. The real fix is a sudoers grant narrowed to the exact commands CD runs, plus root-owned scripts, in the HomeLab template. That is HomeLab's change, not this repository's.

## Alternatives considered

- **Required reviewers on `prod`.** This would stop an unreviewed push to `main` from deploying. But the `main` ruleset already requires CI and has no bypass, and promotion is a reviewed `develop` → `main` PR. A reviewer gate would add a manual click to every promotion and would not stop anything the ruleset lets through. The owner can turn it on later in the environment settings with no workflow change.
- **OIDC instead of a stored SSH key.** OIDC would need something on the LAN to exchange a GitHub token for a short-lived SSH certificate (an SSH CA, or Vault). That is a new service to run and secure on a hobby homelab, for little gain over a per-host key that only one environment can read.
- **A separate deploy repository.** This would split the trust boundary cleanly, but every deploy change would then be a two-repo PR. The environment branch policies already give the same boundary.
- **Keep the fallbacks with louder warnings.** A warning on a green run is not read. That is how a fallback survives for months after its reason has gone.

## Consequences

- Deleting `CMDCTRL_DEPLOY_SSH_KEY`, a host key variable, or the last `deploy` label turns deploys red or queued. It never sends them through a weaker path.
- `CMDCTRL_DEPLOY_RUNNER_LABEL` is unused and can be deleted.
- A VM rebuild still needs its host key variable updated (docs/environments.md#ssh-host-keys). It also needs its environment's public key in `authorized_keys`, which HomeLab#127 keeps across a rebuild.
