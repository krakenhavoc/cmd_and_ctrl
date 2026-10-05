#!/usr/bin/env bash
# Run the Go server toolchain in Docker, for machines with no local Go (#2035).
#
# This is THE way to run Go here. It owns the three shared cache volumes and
# nothing else may create one: per-PR volumes filled a 178 GB partition.
#
#   cmdctrl-gomod     /go/pkg/mod
#   cmdctrl-gobuild   /root/.cache/go-build
#   cmdctrl-golangci  /root/.cache/golangci-lint
#
# The container's exit status is the script's exit status. Do not pipe the
# script through tail or head: that replaces the status with the pipe's.
set -euo pipefail

GO_IMAGE="${CMDCTRL_GO_IMAGE:-golang:1.27}" # matches server/go.mod; lint runs on it too
LINT_PKG="github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0" # keep in step with server/Makefile
MAX_GB="${CMDCTRL_GOCACHE_MAX_GB:-15}"

VOL_MOD=cmdctrl-gomod
VOL_BUILD=cmdctrl-gobuild
VOL_LINT=cmdctrl-golangci

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

usage() {
  cat <<'USAGE'
Usage: scripts/go-docker.sh <command> [args...]

Runs the Go server toolchain in Docker with the repo mounted at /work and the
three shared cache volumes (cmdctrl-gomod, cmdctrl-gobuild, cmdctrl-golangci).
Never create any other cache volume. The exit status is the command's own.

Every command runs on golang:1.27, the version server/go.mod declares.

Commands:
  test [args]    go test -race -cover [args]      (default ./...)
  vet [args]     go vet [args]                    (default ./...)
  lint [args]    golangci-lint v2.14.0 run [args] (default ./...)
  go <args>      go <args>, run in server/
  prune          empty the build cache and the lint cache
  -h, --help     this text

Environment:
  CMDCTRL_GOCACHE_MAX_GB   build cache cap in GB (default 15). Above it, the
                           cache is cleaned before the command runs.

Passing variables into the container (#2157). Nothing leaks in by default;
these are the only routes, and each passes a NAME, never a value on the
command line:
  CMDCTRL_SCRYFALL_DUMP    if set, the file is mounted read-only at
                           /scryfall/<basename> and the variable is
                           rewritten to that path inside the container. A
                           relative path is resolved against your cwd. A
                           missing file is an error.
  AISEAT_*                 AISEAT_GAME_TESTS, AISEAT_HEURISTIC_GAMES,
                           AISEAT_H2H_GAMES, AISEAT_FUNNEL_GAMES,
                           AISEAT_SOAK_GAMES, AISEAT_SOAK_POLICY,
                           AISEAT_SOAK_SEED, AISEAT_DEBUG, AISEAT_STALL and
                           AISEAT_WALLCLOCK are forwarded when set.
  GO_DOCKER_ENV            space-separated names of any other variables to
                           forward when set, e.g.
                           GO_DOCKER_ENV="CMDCTRL_DEV_SKIP_DECK_VALIDATION" \
                             scripts/go-docker.sh test ./...
USAGE
}

die() {
  echo "go-docker: $*" >&2
  exit 2
}

command -v docker >/dev/null 2>&1 || die "docker not found"

case "$MAX_GB" in
  '' | *[!0-9]*) die "CMDCTRL_GOCACHE_MAX_GB must be a whole number, got '$MAX_GB'" ;;
esac

tty_flag=()
if [ -t 0 ] && [ -t 1 ]; then tty_flag=(-t); fi

# Cheap size check: a throwaway container runs du on the volume, read-only.
# Over the cap, clean the build cache first and say so in one line.
enforce_cap() {
  local kb
  kb="$(docker run --rm -v "$VOL_BUILD:/c:ro" "$GO_IMAGE" du -sk /c | cut -f1)"
  if [ "$kb" -gt $((MAX_GB * 1024 * 1024)) ]; then
    echo "go-docker: build cache is $((kb / 1024 / 1024)) GB, over the ${MAX_GB} GB cap; running go clean -cache"
    docker run --rm -v "$VOL_BUILD:/root/.cache/go-build" "$GO_IMAGE" go clean -cache
  fi
}

# Environment pass-through (#2157). Only names are given to docker (-e NAME),
# so values never appear in argv or the process list, and an unset name is
# forwarded as unset rather than as empty.
env_flags=()
forward_env() {
  local name
  for name in "$@"; do
    case "$name" in
      '' | *[!A-Za-z0-9_]* | [0-9]*) die "GO_DOCKER_ENV: '$name' is not a variable name" ;;
    esac
    if [ -n "${!name+x}" ]; then env_flags+=(-e "$name"); fi
  done
}
forward_env AISEAT_GAME_TESTS AISEAT_HEURISTIC_GAMES AISEAT_H2H_GAMES \
  AISEAT_FUNNEL_GAMES AISEAT_SOAK_GAMES AISEAT_SOAK_POLICY AISEAT_SOAK_SEED \
  AISEAT_DEBUG AISEAT_STALL AISEAT_WALLCLOCK
# shellcheck disable=SC2086 # word-splitting the list is the point
forward_env ${GO_DOCKER_ENV:-}

# The dump is a ~630 MB file outside the repo mount: bind it read-only and
# point the variable at the in-container path.
if [ -n "${CMDCTRL_SCRYFALL_DUMP:-}" ]; then
  dump_host="$CMDCTRL_SCRYFALL_DUMP"
  [ -f "$dump_host" ] || die "CMDCTRL_SCRYFALL_DUMP '$dump_host' is not a file"
  dump_host="$(cd "$(dirname "$dump_host")" && pwd)/$(basename "$dump_host")"
  env_flags+=(-v "$dump_host:/scryfall/$(basename "$dump_host"):ro" \
    -e "CMDCTRL_SCRYFALL_DUMP=/scryfall/$(basename "$dump_host")")
fi

cmd="${1:-}"
if [ $# -gt 0 ]; then shift; fi

image="$GO_IMAGE"
flags=()
case "$cmd" in
  '' | -h | --help | help)
    usage
    exit 0
    ;;
  test)
    [ $# -gt 0 ] || set -- ./...
    run_cmd=(go test -race -cover "$@")
    ;;
  vet)
    [ $# -gt 0 ] || set -- ./...
    run_cmd=(go vet "$@")
    ;;
  lint)
    [ $# -gt 0 ] || set -- ./...
    run_cmd=(go run "$LINT_PKG" run "$@")
    flags=(-e GOFLAGS=-buildvcs=false)
    ;;
  go)
    run_cmd=(go "$@")
    ;;
  prune)
    # Empty the contents, keep the volumes: the names are the convention.
    docker run --rm \
      -v "$VOL_BUILD:/root/.cache/go-build" \
      -v "$VOL_LINT:/root/.cache/golangci-lint" \
      "$GO_IMAGE" find /root/.cache/go-build /root/.cache/golangci-lint -mindepth 1 -delete
    echo "go-docker: cleared $VOL_BUILD and $VOL_LINT"
    exit 0
    ;;
  *)
    usage >&2
    exit 2
    ;;
esac

enforce_cap

# Not exec, so the status is passed through explicitly.
rc=0
docker run --rm "${tty_flag[@]}" \
  -v "$ROOT:/work" -w /work/server \
  -v "$VOL_MOD:/go/pkg/mod" \
  -v "$VOL_BUILD:/root/.cache/go-build" \
  -v "$VOL_LINT:/root/.cache/golangci-lint" \
  "${env_flags[@]}" "${flags[@]}" "$image" "${run_cmd[@]}" || rc=$?
exit "$rc"
