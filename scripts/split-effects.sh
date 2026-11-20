#!/usr/bin/env bash
# split-effects.sh — run scripts/split-effects (#5906, ADR 0225), the tool that
# splits server/internal/cards/effects and keeps it split.
#
#   scripts/split-effects.sh --place FILE...   move card/test files written for the old
#                                              one-package layout into place
#   scripts/split-effects.sh --check           list misplaced files
#   scripts/split-effects.sh                   the split itself (scripts/split-effects/REGENERATE.md)
#
# Run it from the repository root. FILE paths are relative to the directory
# you run it in. It uses the Go on PATH; without one (AGENTS.md §5) it runs
# under scripts/go-docker.sh, with the repository mounted at /work.
set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root="$(cd "$here/.." && pwd)"

if command -v go >/dev/null 2>&1; then
  bin="${TMPDIR:-/tmp}/split-effects-$(id -u)"
  (cd "$here/split-effects" && go build -o "$bin" .)
  exec "$bin" -root "$root" "$@"
fi

# No local Go: map every file argument to its path in the container.
args=()
for a in "$@"; do
  if [ -e "$a" ]; then
    abs="$(cd "$(dirname "$a")" && pwd)/$(basename "$a")"
    case "$abs" in
      "$root"/*) a="/work/${abs#"$root"/}" ;;
      *) echo "split-effects: $a is outside the repository" >&2; exit 2 ;;
    esac
  fi
  args+=("$a")
done
exec "$here/go-docker.sh" go -C /work/scripts/split-effects run . -root /work "${args[@]}"
