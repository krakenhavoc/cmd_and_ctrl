#!/usr/bin/env bash
# curated-arena.sh OUT_DIR DECK,DECK,DECK,DECK
#
# Plays 16 four-seat, lockstep, rotated heuristic games on the given
# curated decks and exits non-zero if any game stalled (#3347). boteval
# arena reports a stall without failing on it, so this reads the
# summary it wrote. Needs CMDCTRL_SCRYFALL_DUMP and jq. Run from
# server/. Seed 1000 is fixed: a red night replays with the same call.
set -euo pipefail

out=${1:?usage: curated-arena.sh OUT_DIR DECK,DECK,DECK,DECK}
decks=${2:?usage: curated-arena.sh OUT_DIR DECK,DECK,DECK,DECK}

go run ./cmd/boteval arena \
  --seats heuristic,heuristic,heuristic,heuristic \
  --decks "$decks" \
  --games 16 --seed 1000 --rotate --lockstep \
  --turn-budget 60 --wall 10m \
  --out "$out" --md

summary=$(find "$out" -name summary.json | head -n 1)
[ -n "$summary" ] || { echo "no summary.json under $out" >&2; exit 1; }
# The run dir is named by a timestamp with colons, which upload-artifact
# refuses; keep the summaries at a fixed, flat path for it.
cp "$summary" "$out/summary.json"
cp "$(dirname "$summary")/summary.md" "$out/summary.md"
stalls=$(jq '[.games[] | select(.stalled)] | length' "$summary")
played=$(jq '.games | length' "$summary")
if [ "$played" -ne 16 ]; then
  echo "expected 16 games, summary has $played" >&2
  exit 1
fi
if [ "$stalls" -ne 0 ]; then
  jq -r '.games[] | select(.stalled) | "stalled: seed \(.seed) turn \(.turns)"' "$summary" >&2
  echo "$stalls of $played games stalled on $decks" >&2
  exit 1
fi
echo "no stalls in $played games on $decks"
