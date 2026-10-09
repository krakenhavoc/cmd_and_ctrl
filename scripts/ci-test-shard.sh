#!/usr/bin/env bash
# ci-test-shard.sh — what one CI test shard runs (#2766).
#
#   scripts/ci-test-shard.sh --names    the shard names, one per line, in
#                                       the order ci-cd.yml's matrix lists them
#   scripts/ci-test-shard.sh <name>     that shard's `go test` arguments, one
#                                       per line: a -run or -skip flag first
#                                       if the shard has one, then its
#                                       packages' import paths, sorted
#
# `server-test` in .github/workflows/ci-cd.yml is a matrix with one job per
# shard, each on its own runner, each running
#   go test -race -cover -timeout 30m <that shard's arguments>
# A test binary's wall clock is its package's, so the job used to last as
# long as internal/game. The heavy packages are PINNED below to shards of
# their own, chosen by measured package time (#2766); every other package
# lands in the last shard, `rest`, which `go list ./...` fills at run time.
# A new package is therefore tested with no edit here.
#
# A package too slow for one runner is SPLIT across two shards by test
# name with one regular expression: the first shard runs it with
# `-skip RE`, the second with `-run RE`. Go applies both to the same
# top-level names, so the two halves are complements and no test can fall
# between them. A split shard runs that one package and nothing else,
# because -run and -skip apply to every package on the command line.
#
# TestCITestShardsCoverEveryPackageOnce (server/internal/docsguard) holds
# the contract: every package in exactly one shard, or in exactly the two
# halves of one split; the union is exactly `go list ./...`; and the
# workflow's matrix names every shard.
#
# To rebalance: move a package between the PINNED lines, move a split's
# letter range, or add a line (and its name to the matrix in ci-cd.yml).
# A pinned package that no longer exists fails this script, so a rename
# cannot drop a package silently.
set -euo pipefail

MODULE=github.com/krakenhavoc/cmd_and_ctrl/server

# name, then the packages it runs, relative to the module.
PINNED=(
  "effects internal/cards/effects"
  "lobby internal/lobby internal/botarena"
)

# package, the regular expression, the -skip shard, the -run shard.
# internal/game: about 4,200 small serial tests. ^Test[N-Z] splits its
# measured time roughly in half (TestEveryCarriedFieldSurvivesTheSnapshot
# is in the first half).
SPLITS=(
  "internal/game ^Test[N-Z] game-1 game-2"
)

REST=rest

die() {
  echo "ci-test-shard: $*" >&2
  exit 2
}

usage() {
  sed -n '2,9p' "${BASH_SOURCE[0]}" | sed 's/^# \{0,1\}//' >&2
  exit 2
}

[ $# -eq 1 ] || usage
want="$1"

if [ "$want" = "--names" ]; then
  for line in "${SPLITS[@]}"; do
    read -r -a f <<<"$line"
    echo "${f[2]}"
    echo "${f[3]}"
  done
  for line in "${PINNED[@]}"; do
    echo "${line%% *}"
  done
  echo "$REST"
  exit 0
fi

server_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/../server" && pwd)"
all="$(cd "$server_dir" && go list ./...)" || die "go list ./... failed"

declare -A taken=() # package -> shard(s) that run it
declare -A names=()

claim() { # claim <package-relative> <owner>
  local pkg="$MODULE/$1"
  [ -z "${taken[$pkg]:-}" ] || die "$1 is in both ${taken[$pkg]} and $2"
  grep -qxF "$pkg" <<<"$all" || die "$1 is pinned to $2 but go list ./... has no such package"
  taken[$pkg]="$2"
}

name_ok() {
  [ "$1" != "$REST" ] || die "a pinned or split shard may not be called '$REST'"
  [ -z "${names[$1]:-}" ] || die "two shards are called '$1'"
  names[$1]=1
}

for line in "${SPLITS[@]}"; do
  read -r -a f <<<"$line"
  [ "${#f[@]}" -eq 4 ] || die "a SPLITS line is <package> <regexp> <skip shard> <run shard>: '$line'"
  name_ok "${f[2]}"
  name_ok "${f[3]}"
  claim "${f[0]}" "${f[2]}+${f[3]}"
  if [ "$want" = "${f[2]}" ]; then
    split_out=$(printf '%s\n' "-skip=${f[1]}" "$MODULE/${f[0]}")
  elif [ "$want" = "${f[3]}" ]; then
    split_out=$(printf '%s\n' "-run=${f[1]}" "$MODULE/${f[0]}")
  fi
done

for line in "${PINNED[@]}"; do
  read -r -a f <<<"$line"
  name_ok "${f[0]}"
  for rel in "${f[@]:1}"; do
    claim "$rel" "${f[0]}"
  done
done

if [ -n "${split_out:-}" ]; then
  echo "$split_out"
  exit 0
fi

if [ "$want" = "$REST" ]; then
  while IFS= read -r pkg; do
    [ -n "${taken[$pkg]:-}" ] || echo "$pkg"
  done <<<"$all" | sort
  exit 0
fi

[ -n "${names[$want]:-}" ] || die "no shard named '$want' (see --names)"
for pkg in "${!taken[@]}"; do
  [ "${taken[$pkg]}" != "$want" ] || echo "$pkg"
done | sort
