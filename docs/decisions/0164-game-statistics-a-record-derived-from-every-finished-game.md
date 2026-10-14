# ADR 0164 — Game statistics: a record derived from every finished game

**Status:** Accepted · 2026-10-13 · outside a numbered sprint (owner request, 2026-10-13; see "Sprint" below)
**Issues:** [#3363](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3363) (tracker), [#3364](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3364) (this PR: the record at game end), [#3365](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3365) (backfill from replays), [#3366](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3366) (a free stats page)
**Numbering:** I listed `docs/decisions/` and `server/internal/db/migrations/` across every remote head's history on 2026-10-13. The highest ADR anywhere is **0163** (PR #3331) and the highest migration is **0021** (`0021_lfg_board.sql` on `feat/lfg-server`, PR #3361). This ADR takes **0164** and its migration **0022**. **This PR must merge after #3361**: `db.migrate` skips every version at or below the current one, so a database that ran 0022 first would never run 0021.
**Builds on:** [ADR 0041](0041-game-persistence.md) (restore points and the replay log), [ADR 0033](0033-ai-bot-seat.md) §4 (the public game log, `protocol.LogEvent`), [ADR 0051](0051-user-database.md) (games and seats rows), [ADR 0149](0149-privacy-notice-and-account-deletion.md) (the deletion inventory and the privacy notice), [ADR 0158](0158-signed-in-pages-tables-with-history-decks-and-the-catalog.md) (Tables with History, where stats will show).

---

## Context

A finished game is kept: its `games` and `seats` rows (who sat where, with which deck, who won and how) and its replay, one full snapshot per action (ADR 0041). Nothing reads them for patterns. A player who wants to know how their deck does, how long their games run, or which of their cards they actually cast has to scroll replays.

Everything needed is already recorded:

- **The engine's event history**, `game.Game.Events`, is complete for the whole game and survives restarts (it is in every restore point, `game/snapshot.go`). `protocol.publicLogOf` projects it into `LogEvent`s (cast, draw, life, damage, attack, token, eliminated, game over, step and round), but keeps only the last `PublicLogMax` (200) for the wire.
- **The final omniscient view**, `protocol.ViewOfGame`, has every seat's life, commander damage taken (keyed by commander instance), mulligans, elimination, and every card's name, owner and commander flag.
- **The replay JSONL** has one omniscient `SnapshotPayload` per action. Each carries the last 200 log entries, so merging them by `Seq` rebuilds the log, except where one action emitted more than 200 events.

## Decision

### 1. A stats record is derived, versioned and replaceable

Stats are **computed** from a finished game, never entered or edited. Each record carries `stats_version` (`gamestats.Version`, starting at 1). A change to any definition below bumps it, and the backfill (§6) recomputes older records. A record can always be thrown away and derived again, so it is never the only copy of anything.

### 2. When: once, when the game ends, off the lobby lock

`lobby.syncStateLocked` already sees each game's end exactly once (it stamps `ended_at`, `winner_seat` and `outcome`). On that transition, for a game that is not a practice table, the lobby starts a goroutine. It builds the view with `protocol.ViewOfGame` and the **complete** log with a new `protocol.FullLogOf` (the same fold as the public log, with an unbounded ring), runs `gamestats.Compute`, and writes the result. ViewOfGame and FullLogOf each take the game's read lock themselves. Nothing holds `l.mu` while stats are computed, and a failure is logged and never affects the game or the lobby.

A server with no database, or a game with no stored row, writes nothing (`gamestats.NoStore`).

### 3. What is stored (migration 0022)

Three tables, all keyed by the game and deleted with it (`ON DELETE CASCADE` on `games.id`). **No names and no user ids:** a seat's person is `seats.user_id` through the join on `(game_id, seat)`, so deletion and display follow `seats` (§5).

- **`game_stats`** (one row per game): `stats_version`, `source` (`live` | `replay`), `log_complete` (0 when a replay merge found a gap in event `Seq`s), `turns` (the last turn number in the log), `rounds`, `starting_seat`, `computed_at`.
- **`seat_stats`** (one row per seat):
  - `result`: `win`, `loss` or `draw`, or `none` for a table that ended without an outcome (closed).
  - `place`: 1 for the winner, then by reverse elimination order. Surviving seats in a draw share 1. NULL when unknown.
  - `eliminated_turn`, and `conceded` (0/1).
  - `final_life`, `life_gained` and `life_lost` (sums of positive and negative life changes).
  - `damage_to_players` and `combat_damage_to_players`: damage this seat's sources dealt to *other* players.
  - `commander_damage_dealt` and `commander_damage_taken`: from the final view's commander-damage maps, attributed by the commander card's owner.
  - `spells_cast`, `cards_drawn`, `attacks` (one per attacking creature), `tokens_created`, `mulligans`.
  - `commanders`: a JSON array of the seat's commander names, from cards with `is_commander` that the seat owns.
- **`seat_card_stats`** (one row per seat and card name it cast): `name`, `scryfall_id` (any printing, for a picture), `casts`, `first_cast_turn`. Keyed by name, because the view carries names and Scryfall ids but not oracle ids. Two printings of one card share a name.

### 4. Which games

Every ended, non-practice game with a `games` row, including bot and guest seats. Readers filter: a human's stats are the `seat_stats` rows whose `seats.user_id` is theirs, and `seats.bot_tier` / `agent_client` tell bot and agent seats apart. A game that ended before it started (closed in the lobby) gets a `game_stats` row with `turns = 0` and `result = none` seats, so "every ended game has a record" holds and readers need no special case.

### 5. Privacy and deletion (ADR 0149)

- The tables hold numbers and card names about games, keyed by game and seat. **They are part of the game record the other players keep**, so `accountdelete.Inventory` lists every column as **Kept**, with that note. A deleted person's seats already point at the tombstone ("Deleted player"), and so do these rows through the join.
- The privacy notice gains a "Game statistics" row. `PRIVACY_SCHEMA_VERSION` moves to 22.
- Nothing here is shown to anyone yet. Showing stats (#3366) must show a person their own seats, and other players only in aggregate or by consent. Rankings or leaderboards will be **opt-in**, in their own ADR.

### 6. Backfill (#3365, next PR)

A one-shot `cmd/statsbackfill` reads each ended game's replay JSONL, merges the `log` arrays by `Seq`, takes the last line's view as the final view, and runs the same `Compute` with `source = replay`. A gap in `Seq` sets `log_complete = 0`. It skips games with a current-version record, and recomputes records with an older `stats_version`.

## Consequences

- A game's stats exist seconds after it ends, from the complete log, with no new engine state and no change to the wire protocol.
- `FullLogOf` walks the whole event history once per finished game. That's a few thousand events for a long game, and it's off the lock and once only.
- Card stats are by name. A stat by oracle id (e.g. one row for every printing of a reprint with a changed name) would need the card index and is not worth it yet.
- The definitions are version 1, and changing one is a version bump plus a backfill run, not a migration.

## Sprint

Outside the numbered sprints: the owner asked for game statistics on 2026-10-13, ahead of S72/S73. It touches no engine rules and no client yet, so it does not compete with S70/S71's front-door work.
