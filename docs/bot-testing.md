# AI bot seat: the harness and its test knobs

The commands and the whole-game test gates for the bot seat. The seat itself is in [bot.md](bot.md).

## boteval commands

- `boteval suite run [--dir DIR] [--policy heuristic|assisted|strong] [--max-think 20s] [--think] [--max-tokens N] [--note text] [--parallel N] [--out report.json] [--md]` — asks a policy every labelled position in `server/internal/aiseat/suite/testdata/positions/` and reports agreement overall and per tag, plus reject-hits, malformed replies, out-of-range indices and timeouts. Exits non-zero on a gated miss. The `heuristic` run is also an ordinary Go test and runs on every CI run.
- `boteval suite harvest --from 'dir/*.decisions.jsonl' --to inbox/ [--escalated] [--disagree] [--fallback a,b] [--layer A,B,C] [--seat 0,1] [--tag block,attack] [--limit N] [--seed N]` — pulls candidate windows out of decision logs into an inbox of UNLABELLED positions. Deterministic under `--seed`.
- `boteval suite render --pos path/to/position.json [--deck ID]` — prints the exact prompt a model would see for one position and the move list with `<- accept / reject / heuristic / model@capture` markers. This is the labelling screen. See [docs/bot.md](bot.md#position-suite).
- `boteval arena --seats a,b,c,d [--decks …] --games N --rotate --out DIR [--think] [--max-tokens N]` — headless bot-vs-bot games with the report block ADR 0052 asks every bot PR to carry: win rate with a **Wilson 95% interval** against the table's null rate (1/seats), per policy and per policy-and-deck, ADR 0126's Cards section (each card's offers and uses, the `never` list, the A2/A3 canaries), the funnel's layer/escalation/timeout counters, and decision + model-call latency tails. Rotation seats contestant `k` at position `(k+i)%n` in game `i`, so turn order cancels. `heuristic-baseline` seats the heuristic frozen before S66 (`heuristic.BaselineConfig`, an arena name only). A model tier with **no endpoint is refused, not downgraded** (an `assisted` seat with no client plays the heuristic under a model tier's name). A stall is **reported, not fatal**. Artifacts land in `<out>/<RFC3339 start>/`: `summary.md`, `summary.json`, `games.jsonl` (streamed per game), `decisions/`, `replays/`. Wall clock: ~0.2 s per two-seat heuristic game, ~3.5 s per four-seat curated-deck game, 5–15 min per game with one local-model seat. See [docs/bot.md](bot.md#arena).

## Defensive answers with nothing to answer (#4318, S75)

`heuristic-noidleanswers` is today's heuristic with `PriceAnswerOnlyRows`
off. It restores only the previous prices of answer-only defensive rows,
while retaining response saves and value-activation pricing. Compare it
with `heuristic` on the same seeds and rotated chairs; the pool2 decks
include Selfless Spirit and Ranger-Captain of Eos. Count their activations
with an empty stack from decision logs, and check the Cards table for
responses still taken. The whole-engine tests keep both creatures through
an idle postcombat main phase and sacrifice Selfless Spirit to save two
creatures from a real Wrath of God.

The 2026-10-26 comparison used develop `587c02cdf` plus #4318, the
four pool2 decks, seeds 2000–2047, rotation and lockstep. All four seats
used either `heuristic` or `heuristic-noidleanswers` (48 games each).
All 96 games ended with a winner and none stalled. Empty-stack
sacrifices fell from **14 to 0**: Selfless Spirit 8 → 0 and Ranger-Captain
6 → 0. Thirteen old sacrifices were in postcombat main and one was in
another empty-stack window. Activations with a nonempty stack fell from
4 to 0; those counts alone do not say whether an answer was useful.
The separate whole-engine Wrath regression confirms a useful response
still works. The labelled suite retained 42/42 agreement with no errors
or timeouts.

The Cards table's activation-use games, among games where offered:

| Deck / card | Back-out | Fixed |
|---|---:|---:|
| boros-voltron / Selfless Spirit | 6 / 7 | 0 / 7 |
| mono-white-weenie / Selfless Spirit | 5 / 6 | 0 / 6 |
| mono-white-weenie / Ranger-Captain of Eos | 7 / 9 | 0 / 9 |

| Deck | Back-out wins / 48 (Wilson 95%) | Fixed wins / 48 (Wilson 95%) |
|---|---|---|
| boros-voltron | 9, 18.8% (10.2–31.9%) | 10, 20.8% (11.7–34.3%) |
| mono-white-weenie | 27, 56.2% (42.3–69.3%) | 26, 54.2% (40.3–67.4%) |
| mono-green-stompy | 7, 14.6% (7.2–27.2%) | 6, 12.5% (5.9–24.7%) |
| wubrg-dragons | 5, 10.4% (4.5–22.2%) | 6, 12.5% (5.9–24.7%) |

The four-seat null is 25%. These wide intervals do not establish a
strength gain. Reports, the streaming counter and compressed decision
logs are local under `data/arena-4318/` (gitignored).

## Reconfigure comparison (#4321, S75)

`heuristic-noreconfigure` is today's heuristic with `PriceReconfigure`
turned off. Compare it with `heuristic` using the same seeds and rotated
chairs. The curated `boros-voltron` and `mono-white-weenie` decks include
Rabbit Battery and Lion Sash; synthetic battle decks do not exercise this
change. Decision logs expose the row's `reconfigure` flag and the source's
`attached_to` relation, so activations onto the current host can be counted
without guessing from labels. The position tests cover better and worse
hosts, first attachment and the back-out. The ungated whole-engine test
plays Rabbit Battery's turn through the real enumerator and policy.

The 2026-10-26 evaluation used base `01645dea7` plus #4321, the four
decks listed below, seeds 2000–2047, rotation and lockstep, with all four seats
using either `heuristic` or `heuristic-noreconfigure` (48 games each).
All 96 games ended with a winner and none stalled. Attachments onto the
current host fell from **27 to 0** (Rabbit Battery 21 → 0, Lion Sash
6 → 0); first attachments remained in use (8 → 10). Neither run chose
an unattach. The 42-position labelled suite retained 100% agreement,
with no errors, timeouts or rejected choices.

| Deck | Back-out wins / 48 (Wilson 95%) | Fixed wins / 48 (Wilson 95%) |
|---|---|---|
| boros-voltron | 9, 18.8% (10.2–31.9%) | 7, 14.6% (7.2–27.2%) |
| mono-white-weenie | 27, 56.2% (42.3–69.3%) | 28, 58.3% (44.3–71.2%) |
| mono-green-stompy | 7, 14.6% (7.2–27.2%) | 7, 14.6% (7.2–27.2%) |
| wubrg-dragons | 5, 10.4% (4.5–22.2%) | 6, 12.5% (5.9–24.7%) |

The four-seat null rate is 25%. These runs establish that the repeated
attachment stops; their wide intervals do not establish a strength gain.
Local reports, the streaming counter and decision logs are under
`data/arena-4321/` (gitignored). Full decision logs stay local.

## Test knobs

- **The whole-game tests are gated off by default** and the package owns the longest tests in the tree (CI runs `go test` with a 30m timeout because of them; `server-test` is sharded across five runners by `scripts/ci-test-shard.sh`, the same flags on each, and `aiseat` is in the `rest` shard, #2766):
  - `AISEAT_GAME_TESTS=1` — the master gate. Without it every whole-game test in `internal/aiseat` skips. The nightly `bot-games` job runs `go test ./internal/aiseat/... -race -timeout 30m -skip 'TestFourRandomBotsPlayToAWinner|TestFourHeuristicBotsPlayToAWinner|TestTwoHeuristicBotsPlayToAWinner'`. The catalog soak is its own job, `catalog-soak` (#1456; it needs `CMDCTRL_SCRYFALL_DUMP`, fetched through the day-keyed Scryfall cache; like every test job it runs on a GitHub-hosted runner, because the self-hosted ones hold the deploy credentials, #3335). The skipped tests run without `-race` in their own jobs, side by side (#2661): the random-table run is `bot-random`, four shards of five games, and the heuristic gates are `bot-soak-4p` and `bot-soak-2p` (`.github/workflows/e2e-nightly.yml`).
  - `AISEAT_HEURISTIC_GAMES=N` / `AISEAT_H2H_GAMES=N` — widen the four-heuristic and heuristic-vs-random samples (defaults 3 and small, for CI). The nightly `bot-soak-4p` and `bot-soak-2p` jobs set `AISEAT_HEURISTIC_GAMES=20`, which is S31 exit criterion 2 (#685); the seeds are fixed at 101..120, so each plays the same twenty tables every night. Neither sets `AISEAT_H2H_GAMES`. Both tests play **lockstep** since #1409 (`playLockstepGame` in `heuristic_game_test.go`: the seats' ordinary act-loop called in seat order on one goroutine, through `aiseat.NewStepped` / `Runner.Step` since #1503), so a seed replays move for move — and `boteval arena --lockstep` plays the same schedule; the soak keeps the production one-goroutine-per-seat schedule, because concurrent liveness is what it tests (discussion #1390).
  - `AISEAT_HEURISTIC_SCHEDULE=concurrent` — plays the heuristic gate the old way, one runner goroutine per seat. For comparing the two schedules; the nightly does not set it.
  - `AISEAT_RANDOM_GAMES=N` / `AISEAT_RANDOM_SEED=<uint64>` — `TestFourRandomBotsPlayToAWinner`, S31's four-`random`-bots test: N consecutive games to a winner, default 3, from base seed 31000. The nightly runs it without `-race` as the `bot-random` job: four shards at `AISEAT_RANDOM_GAMES=5` with `AISEAT_RANDOM_SEED` 31000, 31005, 31010 and 31015, which are the twenty seeds 31000..31019 a single run of twenty plays (#2661).
  - `AISEAT_REPLAY_DIR=<path>` — where that test writes each game's replay JSONL. It sets the location only, not whether replays are kept. Unset uses a temp dir. Either way a passing game's replay is deleted and a failing one kept, since one game is hundreds of MiB. The nightly points it at the workspace and uploads it on failure.
  - `AISEAT_FUNNEL_GAMES=N` — widen the Layer A absorption / model-funnel whole-game run.
  - `AISEAT_SOAK_GAMES=N` — independent of the master gate; runs `TestRandomBotSoak` for N games. `AISEAT_SOAK_POLICY=random|heuristic|mixed` picks what fills the seats (default `random`), and `AISEAT_SOAK_SEED=<uint64>` pins the base seed (default: the clock). Game *i* uses `base+i`, and the run logs the base seed and the replay command before the first game, so a green artifact still says which hundred tables were played. A stall prints its own seed too. The nightly `bot-soak-4p` job sets `AISEAT_SOAK_GAMES=100` — S31 exit criterion 4 (#685) — with `AISEAT_SOAK_SEED` derived from the UTC date times 1000, so consecutive nights fuzz disjoint tables rather than overlapping in 99 of 100 games; `bot-soak-2p` plays 100 two-seat games from that seed plus 500. Both take their seeds and the commit they test from the `bot-soak-seed` job.
  - `CMDCTRL_PRINTED_CACHE_CHECKS=0` — set it when **profiling or benchmarking** a test binary (the 3-game soak at `AISEAT_SOAK_SEED=20260923000`, `GOMAXPROCS=2` is the usual recipe). In every test binary each hit of `game`'s printed-characteristic cache is rebuilt and compared ([#1498](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1498), `game/printed_cache.go`), which is the whole suite doubling as the cache's invalidation test — and exactly the cost the cache saves, so a profile taken with the check on shows it (`checkPrintedEntry`). It cannot turn the check on in production.
  - `AISEAT_DEBUG=1` — per-move log in the runner tests.
  - `AISEAT_LIVE_MODEL_TESTS=1` — **the only test in the tree that talks to a model over a network** (`aiseat/model/improvise_live_test.go`, #686): it sends one real improvisation call and checks the bundle against the rail the runner would put it through. Needs a transport too (`CMDCTRL_ANTHROPIC_API_KEY`, or `CMDCTRL_OPENAI_ENDPOINT` plus `CMDCTRL_BOT_MODEL`) and skips without one. **No workflow sets it and none should**: CI has no key, and a test whose result depends on a model's judgement is not a gate. Everything else in that package runs against `model.FakeClient`.
  - `AISEAT_DECISION_LOG=<dir>` — writes a per-game decision log for every game played through `playGame`/`playGameIn` (the same writer the server uses, `aiseat/decisionlog`). Unset (the default, and CI) costs nothing: the runner builds no event without an observer. `AISEAT_DECISION_LOG_MODE` takes the same three values as the server variable. This is how the position corpus gets harvested, and what `TestDecisionLogReplaysOffline` uses to prove a recorded window re-decides identically offline. **Mind the disk**: a four-seat game writes a few thousand windows and a four-player board view is ~50 KiB, so one game is 100–250 MiB. `escalated` only saves anything for policies that run Layer A (the `heuristic` TIER, `rules.NewFilter(heuristic.New(), …)`); the bare `heuristic.New()` the whole-game tests seat reports every window as Layer B and keeps them all in full.
  - `AISEAT_CATALOG_GAMES=N` / `AISEAT_CATALOG_SEED=<uint64>` — the catalog soak (`TestCatalogSoak`, #601): N four-bot games on decks dealt from the catalog itself rather than from the hand-written vanilla decks the other whole-game tests use. The seed defaults to one derived from the UTC date, so each night deals new decks and coverage accumulates; the log prints the base seed, and every failure prints the seed to replay. Needs `CMDCTRL_SCRYFALL_DUMP` as well (a `Spec` carries an oracle ID and a name, not a type line or a mana cost) and skips without it. It **fails on any `EventEffectError`** — a card whose primitive threw mid-resolution, which the engine logs and survives, so nothing else in the tree goes red over it.
  - `AISEAT_CATALOG_POLICY=random|heuristic|mixed` (#1456) — what fills the catalog soak's seats, modelled on `AISEAT_SOAK_POLICY` (the random soak, below): `random` (default, unchanged from before this variable existed) is the widest exploration of the catalog; `heuristic` seats `heuristic.New()` at every seat, which builds board states — blockers up, equipment attached, auras out — that `random` rarely sets up and most catalog cards actually need to be reached at all; `mixed` alternates by seat index. An unrecognised value fails the test outright rather than falling back silently, because a catalog soak run is expensive enough that a misspelled flag deserves a loud failure. The nightly's `catalog-soak` job sets `heuristic`.
  - `AISEAT_CATALOG_REPORT=<path>` — where the catalog soak writes its report JSON: `{"cards": [...], "census": {...}}`. `cards` is the per-card table (cast / resolved / entered / triggered / errored, per oracle ID) it always wrote. `census` is ADR 0041 P7's measurement (#1497, #1558 item 4) — the catalog soak captures a restore point (`game.Game.CaptureSnapshot`, the same call `ws.Room.writeRestorePointLocked` makes) after every observed action of every game in the run, and tallies what it found: `actions` and `captured` (how many of those captures were restorable), `by_kind` (blocked captures per census kind, summed over the whole run — `game.ContinuationCensus.Kinds`'s keys), and `longest_run_by_kind` (the longest run of CONSECUTIVE actions any one kind stayed the blocker — the number the shutdown census can't give, since it only ever reads one instant). It is not a gate: the shutdown census (#524) reads one instant per deploy and every deploy so far has caught an idle table, so this is the stand-in evidence P7 asks for to order phase 3's remaining tiers. Capture is sampled at the poll rate the stall detector already uses, so a burst of actions inside one poll tick coalesces into one capture — an undercount that can only make a stale run look shorter than it was, never longer. The nightly uploads the whole file as an artifact on every run, green included: the useful half of `cards` is the list of cards no bot game reached, and the useful half of `census` is whether tier 4 (stack effects and target specs, by a wide margin in every run so far) is worth hurrying.
  - `AISEAT_STALL` / `AISEAT_WALLCLOCK` — Go durations. Since [#685](https://github.com/krakenhavoc/cmd_and_ctrl/issues/685) **every bot-table budget in the package reads them**, so there is one pair of knobs for a loaded runner and no literal left to edit:
    - `TestFourRandomBotsPlay` — stall detector and wall clock, defaults `15s` / `300s`.
    - `TestCatalogSoak` — the same, same defaults.
    - `TestRandomBotSoak` — per-game stall detector and wall clock, defaults `3s` / `60s`.
    - `playGame` / `playGameIn`, which is every other whole-game test (four-heuristic, random-to-a-winner, head-to-head, latency, funnel, model, life-cost, decision-log replay) — stall detector, default `5s`.
    - `TestManagerPlaysALobbySeatedTable` — `AISEAT_WALLCLOCK` as its wait for the table to finish, default `300s`. Its wait for the runners to EXIT is no longer a budget at all: that is `waitForRunner` now (#1048).

    **`AISEAT_WALLCLOCK` is a floor on the `playGame` / `playGameIn` wall clock, not a replacement.** Each caller passes its own — 120s for a 50-turn heuristic table, 25 turns for the decision-log replay — and those numbers say something about the game being played, so the variable may raise them and never cuts them. Everywhere else it is the value. Unset (the default, and PR CI) changes nothing anywhere.

    Raise them on a loaded runner rather than editing a test. The nightly `bot-soak-4p` and `bot-soak-2p` jobs set `AISEAT_STALL=30s` and `AISEAT_WALLCLOCK=300s`, and both are `workflow_dispatch` inputs.
  - `CMDCTRL_SCRYFALL_DUMP=<path>` — gates the manual bot-deck test that validates the four curated decks against the real Scryfall dump. Also used by other packages.
