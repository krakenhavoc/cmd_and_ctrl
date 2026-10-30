# shellcheck shell=bash
# Shared helpers for the triage scripts. Source it from a script; don't run it.
#
# Environment:
#   TRIAGE_REPO     owner/name; default: `gh repo view` in this checkout
#   TRIAGE_DIR      snapshot directory; default: $CLAUDE_JOB_DIR/tmp/triage
#   DRY_RUN=1       print each GitHub write instead of making it
#   TRIAGE_PACE     seconds to wait after each write (default 2)
#   TRIAGE_BACKOFF  seconds to wait after a 403 or 429 (default 60)
#   TRIAGE_TRIES    attempts per write before giving up (default 4)

set -euo pipefail

TRIAGE_ROOT=$(git -C "$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")" rev-parse --show-toplevel)
TRIAGE_PACE=${TRIAGE_PACE:-2}
TRIAGE_BACKOFF=${TRIAGE_BACKOFF:-60}
TRIAGE_TRIES=${TRIAGE_TRIES:-4}

die() { echo "$*" >&2; exit 2; }

# The repository the issues live in. `gh repo view` reads this checkout's
# remote, so the scripts work from any working directory.
triage_repo() {
  if [ -z "${TRIAGE_REPO:-}" ]; then
    TRIAGE_REPO=$(cd "$TRIAGE_ROOT" && gh repo view --json nameWithOwner --jq .nameWithOwner </dev/null)
    export TRIAGE_REPO
  fi
  printf '%s\n' "$TRIAGE_REPO"
}

# The snapshot directory. Scratch belongs in the job's tmp dir, never /tmp.
triage_dir() {
  local d=${TRIAGE_DIR:-${CLAUDE_JOB_DIR:+$CLAUDE_JOB_DIR/tmp/triage}}
  [ -n "$d" ] || die "set TRIAGE_DIR (or CLAUDE_JOB_DIR) to a scratch directory"
  mkdir -p "$d"
  printf '%s\n' "$d"
}

is_rate_limited() { grep -qiE 'HTTP (403|429)|rate limit|secondary' <<<"$1"; }

# read_retry <cmd...>: a GitHub read that backs off when rate-limited.
read_retry() {
  local try out
  for ((try = 1; try <= TRIAGE_TRIES; try++)); do
    if out=$("$@" 2>&1 </dev/null); then
      printf '%s\n' "$out"
      return 0
    fi
    if is_rate_limited "$out"; then
      echo "rate-limited (read, try $try/$TRIAGE_TRIES), backing off ${TRIAGE_BACKOFF}s" >&2
      sleep "$TRIAGE_BACKOFF"
      continue
    fi
    printf '%s\n' "$out" >&2
    return 1
  done
  echo "gave up after $TRIAGE_TRIES rate-limited reads" >&2
  return 1
}

# mutate <what> <cmd...>: one GitHub write, paced. Waits TRIAGE_PACE seconds
# after a success, so a loop of writes stays near one every two seconds. On a
# 403 or 429 it waits TRIAGE_BACKOFF seconds and tries again. The command's
# output is left in MUTATE_OUT. With DRY_RUN=1 it prints the command instead.
# shellcheck disable=SC2034 # read by the scripts that source this
MUTATE_OUT=
mutate() {
  local what=$1; shift
  MUTATE_OUT=
  if [ "${DRY_RUN:-0}" = 1 ]; then
    printf 'dry-run %s:' "$what"; printf ' %q' "$@"; printf '\n'
    return 0
  fi
  local try out
  for ((try = 1; try <= TRIAGE_TRIES; try++)); do
    if out=$("$@" 2>&1 </dev/null); then
      # shellcheck disable=SC2034
      MUTATE_OUT=$out
      echo "ok $what"
      sleep "$TRIAGE_PACE"
      return 0
    fi
    if is_rate_limited "$out"; then
      echo "rate-limited $what (try $try/$TRIAGE_TRIES), backing off ${TRIAGE_BACKOFF}s" >&2
      sleep "$TRIAGE_BACKOFF"
      continue
    fi
    echo "FAIL $what: $out" >&2
    return 1
  done
  echo "GAVE UP $what after $TRIAGE_TRIES rate-limited tries" >&2
  return 1
}

# resolve_milestone <title-or-prefix>: the full title of the open milestone
# named exactly that, or the one open milestone whose title starts with it
# followed by a space ("S79", "Backlog").
resolve_milestone() {
  local want=$1 titles matches count
  titles=$(read_retry gh api --paginate "repos/$(triage_repo)/milestones?state=open&per_page=100" --jq '.[].title')
  if grep -qxF -- "$want" <<<"$titles"; then
    printf '%s\n' "$want"
    return 0
  fi
  matches=$(awk -v w="$want" 'index($0, w) == 1 && substr($0, length(w) + 1, 1) == " "' <<<"$titles")
  count=$(grep -c . <<<"$matches" || true)
  [ "$count" -eq 1 ] || die "milestone '$want' matches $count open milestones${matches:+: $(tr '\n' ';' <<<"$matches")}"
  printf '%s\n' "$matches"
}
