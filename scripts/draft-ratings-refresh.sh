#!/usr/bin/env bash
# draft-ratings-refresh.sh — write one small ratings file per draftable
# set from 17Lands' public card ratings, into
# server/internal/draftratings/data/<set>.json (ADR 0222, #5703).
#
# 17Lands data may be used with credit on the page (the owner's decision
# of 2026-11-11); the draft screens say "Card ratings from 17Lands" and
# link to 17lands.com. Nothing else is sent: a generic User-Agent, no
# account, no cookie, one request per set (a second and third only for a
# set whose first format has too little data), and a pause between every
# request.
#
# For each set the script asks, in order, PremierDraft, QuickDraft and
# TradDraft, from the set's first day on 17Lands (GET /data/filters
# knows it) to today, and keeps the first format whose games-in-hand
# counts add up to at least MIN_TOTAL_GAMES. Per card it keeps:
#
#   n  the card's name as 17Lands spells it
#   w  GIH WR, the win rate of games with the card in hand (opening
#      hand or drawn), as a fraction, 4 places
#   g  the games behind w (17Lands' "# GIH")
#   s  average last seen at (pick number)
#   p  average pick (pick number the card was taken at)
#
# A card with fewer than MIN_GAMES (500) games in hand is dropped. 17Lands
# itself publishes no win rate under 500 games (the field is null), and
# at 500 the standard error of a win rate is still about 2.2 points
# against a set's spread of about 3.5, so a lower bar would only be
# noise. A card with no rate (a basic land, a card nobody plays) has no
# row, and the draft shows the in-house rating for it. A set left with
# fewer than MIN_CARDS (40) cards gets no file: a tier is read off the
# set's own spread, and a few cards do not make one. Run the script again
# as a set matures and more of its cards pass the bar.
#
# --from-dataset builds the same file from 17Lands' public datasets
# instead (#5708), for a set the live endpoint has thinned: on
# 2026-11-11 it served a few hundred games for each older set, where the
# datasets hold every game and pick 17Lands logged. For each format in
# the same order the script downloads the set's two files from
# https://17lands-public.s3.amazonaws.com/analysis_data/ (game_data and
# draft_data, 10 to 300 MB each, gzip), and server/cmd/draftdataset
# reads them as a stream into rows of the live endpoint's shape: per card
# the games with it in hand and their win rate, its average pick and its
# average last seen. The same builder below then applies the same
# threshold and writes the same file; "start" is the day of the set's
# first draft in the dataset. A set whose datasets do not exist gets no
# file. The downloads are deleted once a set is built, unless they were
# put in --dataset-dir, where a later run reuses them.
#
# Sets: the arguments, else every draftable set of $CMDCTRL_SETS_URL
# (GET /draft/sets, default https://cmd.labxp.io/draft/sets). A set
# 17Lands has no usable data for gets no file, and a stale file of it is
# left alone.
#
# Usage: draft-ratings-refresh.sh [--out DIR] [--keep-raw DIR] [--from-raw DIR]
#                                 [--from-dataset [--dataset-dir DIR]] [SET ...]
#   --out DIR         where the files go (default server/internal/draftratings/data)
#   --keep-raw DIR    also save each raw 17Lands answer there (with
#                     --from-dataset, the rows computed from the datasets)
#   --from-raw DIR    rebuild from answers saved with --keep-raw; no requests
#   --from-dataset    build from 17Lands' public datasets, not the live endpoint
#   --dataset-dir DIR keep the downloaded datasets there, and use any
#                     already there instead of downloading them again
#
# Environment: MIN_GAMES (500), MIN_CARDS (40), MIN_TOTAL_GAMES (20000),
# PAUSE (3, seconds), DRAFTDATASET (a built server/cmd/draftdataset; else
# it is built with go, or scripts/go-docker.sh where there is no go),
# CMDCTRL_SCRYFALL_DUMP (with --from-dataset: the Scryfall dump, from
# which a card name the datasets mangled, "Bespoke B?" for "Bespoke Bō",
# is put back; without it such a card keeps the dataset's spelling).
# Dependencies: curl, python3; go or Docker for --from-dataset.

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUT="$ROOT/server/internal/draftratings/data"
KEEP_RAW=""
FROM_RAW=""
FROM_DATASET=0
DATASET_DIR=""
SETS=()
while [[ $# -gt 0 ]]; do
  case "$1" in
    --out) OUT="$2"; shift 2 ;;
    --keep-raw) KEEP_RAW="$2"; shift 2 ;;
    --from-raw) FROM_RAW="$2"; shift 2 ;;
    --from-dataset) FROM_DATASET=1; shift ;;
    --dataset-dir) DATASET_DIR="$2"; shift 2 ;;
    -h|--help) sed -n '2,/^set -euo/p' "$0" | sed '$d'; exit 0 ;;
    -*) echo "draft-ratings-refresh: unknown option: $1" >&2; exit 2 ;;
    *) SETS+=("$1"); shift ;;
  esac
done

MIN_GAMES="${MIN_GAMES:-500}"
MIN_CARDS="${MIN_CARDS:-40}"
MIN_TOTAL_GAMES="${MIN_TOTAL_GAMES:-20000}"
PAUSE="${PAUSE:-3}"
UA='draft-ratings-refresh/1.0'
BASE='https://www.17lands.com'
FORMATS=(PremierDraft QuickDraft TradDraft)
TODAY="$(date -u +%Y-%m-%d)"

mkdir -p "$OUT"
[[ -n "$KEEP_RAW" ]] && mkdir -p "$KEEP_RAW"
WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT

if [[ ${#SETS[@]} -eq 0 ]]; then
  curl --fail --silent --show-error --max-time 60 -A "$UA" \
    "${CMDCTRL_SETS_URL:-https://cmd.labxp.io/draft/sets}" -o "$WORK/sets.json"
  mapfile -t SETS < <(python3 -I -c '
import json,sys
for s in json.load(open(sys.argv[1]))["sets"]:
    if s.get("draftable"): print(s["code"])' "$WORK/sets.json")
fi

if [[ -n "$FROM_RAW" && "$FROM_DATASET" -eq 1 ]]; then
  echo "draft-ratings-refresh: --from-raw and --from-dataset do not mix" >&2
  exit 2
fi

if [[ -n "$FROM_RAW" ]]; then
  cp "$FROM_RAW/filters.json" "$WORK/filters.json"
elif [[ "$FROM_DATASET" -eq 1 ]]; then
  # The datasets carry their own first day; www.17lands.com is not asked.
  echo '{"start_dates": {}}' >"$WORK/filters.json"
  [[ -n "$DATASET_DIR" ]] && mkdir -p "$DATASET_DIR"
  if [[ -n "${DRAFTDATASET:-}" ]]; then
    TOOL="$DRAFTDATASET"
  elif command -v go >/dev/null; then
    TOOL="$WORK/draftdataset"
    (cd "$ROOT/server" && go build -o "$TOOL" ./cmd/draftdataset)
  else
    # No local Go: build it in Docker into the gitignored data/ (#2035).
    mkdir -p "$ROOT/data"
    "$ROOT/scripts/go-docker.sh" go build -o ../data/draftdataset ./cmd/draftdataset
    TOOL="$ROOT/data/draftdataset"
  fi
else
  curl --fail --silent --show-error --max-time 60 -A "$UA" "$BASE/data/filters" -o "$WORK/filters.json"
  [[ -n "$KEEP_RAW" ]] && cp "$WORK/filters.json" "$KEEP_RAW/filters.json"
  sleep "$PAUSE"
fi

# build writes $OUT/<set>.json from one raw answer, or exits 3 when the
# answer has too little data to rate anything.
build() {
  python3 -I - "$@" <<'PY'
import json, sys
raw, out, set_code, fmt, start, today = sys.argv[1:7]
min_games, min_total, min_cards = int(sys.argv[7]), int(sys.argv[8]), int(sys.argv[9])
try:
    rows = json.load(open(raw))
except ValueError:
    sys.exit(3)
if not isinstance(rows, list):
    sys.exit(3)
total = sum(r.get("ever_drawn_game_count") or 0 for r in rows)
if total < min_total:
    sys.exit(3)
cards = []
for r in rows:
    g, w = r.get("ever_drawn_game_count") or 0, r.get("ever_drawn_win_rate")
    if g < min_games or w is None:
        continue
    cards.append({"n": r["name"], "w": round(w, 4), "g": g,
                  "s": round(r.get("avg_seen") or 0, 2), "p": round(r.get("avg_pick") or 0, 2)})
cards.sort(key=lambda c: c["n"])
if len(cards) < min_cards:
    sys.exit(3)
head = {"set": set_code.lower(), "source": "17lands", "format": fmt, "start": start,
        "fetched": today, "min_games": min_games}
with open(out, "w") as f:
    f.write("{\n")
    for k, v in head.items():
        f.write(f"  {json.dumps(k)}: {json.dumps(v)},\n")
    f.write('  "cards": [\n')
    f.write(",\n".join("    " + json.dumps(c, ensure_ascii=False, separators=(",", ":")) for c in cards))
    f.write("\n  ]\n}\n")
print(f"{set_code} {fmt}: {len(cards)} cards of {len(rows)}, {total} games in hand")
PY
}

DATASETS='https://17lands-public.s3.amazonaws.com/analysis_data'

# dataset writes to $3 the rows of set $1 (upper case) in format $2,
# from its two public dataset files, and prints the dataset's first day.
# It fails, quietly, when 17Lands has no dataset for that format.
dataset() {
  local upper="$1" fmt="$2" rows="$3" dir="${DATASET_DIR:-$WORK/datasets}" kind name status=0
  local paths=()
  mkdir -p "$dir"
  for kind in game_data draft_data; do
    name="${kind}_public.$upper.$fmt.csv.gz"
    if [[ ! -s "$dir/$name" ]]; then
      curl --head --fail --silent --max-time 60 -A "$UA" "$DATASETS/$kind/$name" >/dev/null || return 1
      curl --fail --silent --show-error --max-time 3600 -A "$UA" "$DATASETS/$kind/$name" -o "$dir/$name.part" ||
        { rm -f "$dir/$name.part"; return 1; }
      mv "$dir/$name.part" "$dir/$name"
      sleep "$PAUSE"
    fi
    paths+=("$dir/$name")
  done
  "$TOOL" -games "${paths[0]}" -drafts "${paths[1]}" -out "$rows" || status=$?
  # The downloads are a few hundred MB a set: gone once read, unless kept.
  [[ -n "$DATASET_DIR" ]] || rm -f "${paths[@]}"
  return "$status"
}

for set in "${SETS[@]}"; do
  upper="$(tr '[:lower:]' '[:upper:]' <<<"$set")"
  if [[ "$FROM_DATASET" -eq 0 ]]; then
    start="$(python3 -I -c '
import json,sys
v=json.load(open(sys.argv[1])).get("start_dates",{}).get(sys.argv[2])
print(v[:10] if v else "")' "$WORK/filters.json" "$upper")"
    if [[ -z "$start" ]]; then
      echo "draft-ratings-refresh: $set: 17Lands has no such set; no file"
      continue
    fi
  fi
  done_set=0
  found=0
  for fmt in "${FORMATS[@]}"; do
    raw="$WORK/$set-$fmt.json"
    if [[ "$FROM_DATASET" -eq 1 ]]; then
      start="$(dataset "$upper" "$fmt" "$raw")" || continue
      found=1
      if [[ -n "$KEEP_RAW" ]]; then
        cp "$raw" "$KEEP_RAW/$set-$fmt.json"
        # Record the first day where --from-raw looks for it.
        python3 -I -c '
import json,os,sys
p,code,day=sys.argv[1:4]
d=json.load(open(p)) if os.path.exists(p) else {}
d.setdefault("start_dates",{})[code]=day
with open(p,"w") as f: json.dump(d,f,indent=1)' "$KEEP_RAW/filters.json" "$upper" "$start"
      fi
    elif [[ -n "$FROM_RAW" ]]; then
      [[ -f "$FROM_RAW/$set-$fmt.json" ]] || continue
      cp "$FROM_RAW/$set-$fmt.json" "$raw"
    else
      curl --fail --silent --show-error --max-time 120 -A "$UA" \
        "$BASE/card_ratings/data?expansion=$upper&format=$fmt&start_date=$start&end_date=$TODAY" \
        -o "$raw" || { echo "draft-ratings-refresh: $set $fmt: request failed" >&2; sleep "$PAUSE"; continue; }
      [[ -n "$KEEP_RAW" ]] && cp "$raw" "$KEEP_RAW/$set-$fmt.json"
      sleep "$PAUSE"
    fi
    if build "$raw" "$OUT/$set.json" "$set" "$fmt" "$start" "$TODAY" "$MIN_GAMES" "$MIN_TOTAL_GAMES" "$MIN_CARDS"; then
      done_set=1
      break
    fi
  done
  if [[ "$done_set" -eq 0 && "$FROM_DATASET" -eq 1 && "$found" -eq 0 ]]; then
    echo "draft-ratings-refresh: $set: 17Lands has no public dataset for it; no file"
  elif [[ "$done_set" -eq 0 ]]; then
    echo "draft-ratings-refresh: $set: too little data in any format; no file"
  fi
done
