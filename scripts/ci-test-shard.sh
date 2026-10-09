#!/usr/bin/env bash
# ci-test-shard.sh — which server packages one CI test shard runs (#2766).
#
#   scripts/ci-test-shard.sh --names    the shard names, one per line, in
#                                       the order ci-cd.yml's matrix lists them
#   scripts/ci-test-shard.sh <name>     that shard's packages, one import
#                                       path per line, sorted
#
# `server-test` in .github/workflows/ci-cd.yml is a matrix with one job per
# shard, each on its own runner, each running
#   go test -race -cover -timeout 30m <that shard's packages>
# A test binary's wall clock is its package's, so the job used to last as
# long as internal/game. The heavy packages are PINNED below to shards of
# their own, chosen by measured package time (#2766); every other package
# lands in the last shard, `rest`, which `go list ./...` fills at run time.
# A new package is therefore tested with no edit here.
#
# TestCITestShardsCoverEveryPackageOnce (server/internal/docsguard) holds
# the contract: the shards are disjoint, their union is exactly
# `go list ./...`, and the workflow's matrix names every shard.
#
# To rebalance: move a package between the PINNED lines (or add a line for
# a new pinned shard and its name to the matrix in ci-cd.yml). A pinned
# package that no longer exists fails this script, so a rename cannot drop
# a package silently.
set -euo pipefail

MODULE=github.com/krakenhavoc/cmd_and_ctrl/server

# name, then the packages it runs, relative to the module.
PINNED=(
  "game internal/game"
  "effects internal/cards/effects"
  "lobby internal/lobby internal/botarena"
)
REST=rest

die() {
  echo "ci-test-shard: $*" >&2
  exit 2
}

usage() {
  sed -n '2,8p' "${BASH_SOURCE[0]}" | sed 's/^# \{0,1\}//' >&2
  exit 2
}

[ $# -eq 1 ] || usage
want="$1"

if [ "$want" = "--names" ]; then
  for line in "${PINNED[@]}"; do
    echo "${line%% *}"
  done
  echo "$REST"
  exit 0
fi

server_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/../server" && pwd)"
all="$(cd "$server_dir" && go list ./...)" || die "go list ./... failed"

declare -A pinned_to=()
found=0
for line in "${PINNED[@]}"; do
  read -r -a fields <<<"$line"
  name="${fields[0]}"
  [ "$name" != "$REST" ] || die "a pinned shard may not be called '$REST'"
  for rel in "${fields[@]:1}"; do
    pkg="$MODULE/$rel"
    [ -z "${pinned_to[$pkg]:-}" ] || die "$rel is pinned to both ${pinned_to[$pkg]} and $name"
    grep -qxF "$pkg" <<<"$all" || die "$rel is pinned to shard $name but go list ./... has no such package"
    pinned_to[$pkg]="$name"
  done
  if [ "$name" = "$want" ]; then found=1; fi
done

if [ "$want" = "$REST" ]; then
  while IFS= read -r pkg; do
    [ -n "${pinned_to[$pkg]:-}" ] || echo "$pkg"
  done <<<"$all" | sort
  exit 0
fi

[ "$found" = 1 ] || die "no shard named '$want' (see --names)"
for pkg in "${!pinned_to[@]}"; do
  [ "${pinned_to[$pkg]}" != "$want" ] || echo "$pkg"
done | sort
