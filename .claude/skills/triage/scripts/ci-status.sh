#!/usr/bin/env bash
# One look at a pull request's checks. Run it about every 90 seconds while
# waiting; never `gh run watch` without --interval, which polls every 3
# seconds and trips GitHub's secondary rate limit when sessions overlap.
#
#   ci-status.sh <pr>
#
# Exit status: 0 green (everything passed or was skipped), 1 a check failed or
# was cancelled, 2 still running, 3 GitHub refused the read (a rate limit:
# wait and look again; it says nothing about the run).
#
# This gh has no `pr checks --json`, so it reads the plain tab-separated output.
# A failed server-check whose log says "Timeout exceeded" is the known
# golangci-lint flake: `gh run rerun <run-id> --failed`.
# shellcheck source=lib.sh
source "$(dirname "$(readlink -f "$0")")/lib.sh"
set +e

[ $# -eq 1 ] || die "usage: ci-status.sh <pr>"
out=$(gh pr checks "${1#\#}" -R "$(triage_repo)" 2>&1 </dev/null)
if is_rate_limited "$out"; then
  echo "rate-limited: look again in a few minutes (not a red run)"
  exit 3
fi
if grep -qi 'no checks reported' <<<"$out"; then
  echo "no checks yet"
  exit 2
fi

count() { awk -F'\t' -v s="$1" '$2 == s' <<<"$out" | grep -c .; }
pass=$(count pass); fail=$(count fail); pending=$(count pending); skip=$(count skipping); cancel=$(count cancel)
if [ $((pass + fail + pending + skip + cancel)) -eq 0 ]; then
  echo "could not read the checks:"; echo "$out"
  exit 3
fi
echo "pass $pass, fail $fail, cancelled $cancel, pending $pending, skipped $skip"
awk -F'\t' '$2 == "fail" || $2 == "cancel" {
  run = $4; sub(/.*\/runs\//, "", run); sub(/\/.*/, "", run)
  printf "  %s %s: run %s %s\n", $2, $1, run, $4
}' <<<"$out"

[ $((fail + cancel)) -eq 0 ] || exit 1
[ "$pending" -eq 0 ] || exit 2
exit 0
