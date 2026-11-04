# ADR 0217 — Advanced player stats, per deck and overall

**Status:** Accepted · 2026-11-03 · outside S79's card scope (owner request; see "Sprint" below)
**Issue:** [#4880](https://github.com/krakenhavoc/cmd_and_ctrl/issues/4880) (epic)
**Owner decisions:** Ian, 2026-11-03, in the [owner-decisions comment on #4880](https://github.com/krakenhavoc/cmd_and_ctrl/issues/4880). Its nine decisions are **decided** and are recorded below as [Owner decisions](#owner-decisions). Everything else this ADR fixes is a call made here, listed under [Calls made here](#calls-made-here), and can be overturned in review.
**Numbering:** checked on 2026-11-03 with the command in [the ADR README](README.md#picking-a-number) over every remote head. `develop` ends at 0215, and `docs/4870-planechase-adr` (PR #4872) holds 0216. No other head has a number above 0215. This is **0217**.
**Builds on:**
- [ADR 0165](0165-game-statistics-a-record-derived-from-every-finished-game.md): the stats record (`game_stats`, `seat_stats`, `seat_card_stats`), `gamestats.Compute`, `cmd/statsbackfill`, `GET /me/stats`, `GET /admin/stats` and the §9 privacy pass. This ADR is its version 2 and adds a dated note to it.
- [ADR 0149](0149-privacy-notice-and-account-deletion.md): `accountdelete.Inventory` and `Delete`, the privacy page, `PRIVACY_SCHEMA_VERSION`.
- [ADR 0126](0126-bots-that-play-their-decks.md) §6: the catalog's declared `Purpose`.
- [ADR 0148](0148-capacity-limits-for-a-public-launch.md): the beta's size, about 30 tables on one VM.
- [ADR 0041](0041-game-persistence.md) (the replay JSONL), [ADR 0055](0055-loop-breaker.md) (the loop breaker) and [ADR 0193](0193-demonstrate-a-loop-record-once-replay-n-times.md) (demonstrated loops), for combo detection.
- [ADR 0124](0124-admin-views-accounts-games-and-who-is-on-now.md) (admin Accounts, and the `admin action` audit line), [ADR 0158](0158-signed-in-pages-tables-with-history-decks-and-the-catalog.md) (the Decks page and the signed-in page frame), [ADR 0213](0213-friends-recently-played-with-and-live-presence.md) (friends and blocks).

---

## Context

ADR 0165 stores, per game and seat, the result, place, elimination turn, life, damage, commander damage, spells cast, cards drawn, attacks, tokens, mulligans and casts per card. It shows them at `#/stats` (`GET /me/stats`, by deck and by commander, `routes/Stats.svelte`) and to operators at `#/admin/stats`. The owner wants a second round: "Pretty much if there is a useful stat I want it" (#4880), shown as a well-designed visual breakdown that helps people improve their play and their decks.

What the code gives us to work with, read on 2026-11-03:

- **The projected log is enough for most play stats.** `protocol.FullLogOf` folds the whole of `game.Game.Events` into `LogEvent`s. They include `cast`, `activate`, `trigger`, `attack`, `damage` (with `Combat` and `TargetSeat`), `life`, `zone` (with `OldZone` and `NewZone`), `step` (with `Step` and `Turn`), `loop` (a demonstrated loop's one replay line, ADR 0193 §7), `eliminated` (with `Cause`: `life`, `empty_draw`, `poison`, `commander_damage`, `effect`, `concede`) and `game_over` (with `Cause`: `last_standing`, `effect`, `all_lost`, `mandatory_loop`, and the winning source's `CardID` for an effect win).
- **The log is not enough for anything about a hand.** Draws never name their cards in the log, and nothing in it says which cards sat in a hand on a given turn. Screw, flood, colour screw and dead cards all need the hand.
- **The replay has the hand.** `replays/<id>.jsonl` (ADR 0041) holds one omniscient `SnapshotPayload` per committed action. Each one carries every hand, the battlefield, the turn and step, and each seat's `LandsPlayedThisTurn`, `CommanderCasts` and `Poison` (`protocol.PlayerView`). `cmd/statsbackfill` already reads these files.
- **The loop breaker's event is not in the log.** `game.EventLoopSuspected` (ADR 0055) is not projected by `publicLogOf`, so a stats fold cannot see it today.
- **Card facts are in the card index.** `cards.Index` (loaded from Scryfall's `default_cards` dump by `scripts/scryfall-refresh.sh`, weekly) has `OracleID`, `ManaCost`, `ProducedMana`, `ColorIdentity` and the type line for every printing, keyed by Scryfall id. A `CardView` carries the Scryfall id, so a frame's card can be looked up.
- **Decks** (`decks`, migration 0005) keep `source_text`, `updated_at` and the owner. A seat keeps `deck_id` and `deck_name`, but nothing records which version of a deck a seat played.
- **Settings** in `user_settings` are a client-synced JSON body. The server does not read it, so it cannot gate a server route.
- **Admin routes** go through `requireAdmin`, which already writes the `admin action` audit line for every call (`lobby/admins.go`, `logAdminAction`).
- **No pod or group concept exists.** ADR 0213 adds mutual friends (one pair at a time). Nothing names a group.
- **The client has no chart dependency** (`client/package.json` has no runtime dependencies at all). Inline SVG is already the house style for small graphics (`components/decks/CoverageRing.svelte`, `board/ManaSymbol.svelte`).

## Decision

### 1. The record, version 2

`gamestats.Version` moves from 1 to **2**. The record stays what ADR 0165 §1 made it: derived, versioned, replaceable, and never the only copy of anything.

**Two inputs, one `Compute`.** Version 2 reads the same final view and full log as version 1, plus a **turn sampler** fed from the replay's frames:

- `gamestats.Input` gains `Frames`, an iterator over the replay's omniscient views. `Compute` keeps, for each seat and each of that seat's own turns, two samples: the frame at the **start of its precombat main phase** (after its draw) and the **last frame of the turn** (the frame before the turn number changes). It also keeps the frame at each `declare_attackers` step for the attack-share stat, and the set of every card id that was ever in the seat's hand.
- **An undo** shows as a frame whose turn and step go backwards. The sampler drops every sample at or after that point and takes the frames that follow, so an undone draw or an undone land play counts as not having happened.
- **The live path** (ADR 0165 §2) still runs once, off the lobby lock, when a game ends. It now runs on **one** stats worker goroutine with a short queue rather than one goroutine per game, because reading a replay costs more than folding the log. It waits for the room's final frame to be flushed and reads `Room.ReplayPath()`.
- **The backfill** (`cmd/statsbackfill`) already reads the replay. It now passes the frames to `Compute` as well.
- **No replay** (a file lost or never written) leaves every version-2 field NULL and sets the new `game_stats.frames_complete` to 0. The version-1 fields are still written from the log, so no game loses its basic record.
- **The loop breaker.** `FullLogOf` gains a stats-only option that also projects `game.EventLoopSuspected` as a `LogLoop` entry with `Call: "suspected"`. It is never sent on the wire, so the public log and the client are unchanged. A replay's log has no such entry, so a backfilled game relies on the repetition heuristic of §3 alone. That is why the combo rule is versioned.

**Schema: one migration** (the next free number when slice 2 starts; `develop` is at 0028 on 2026-11-03). Additive only. Every new table has the same `ON DELETE CASCADE` chain to `games` as ADR 0165 §3. As in version 1: **no names of people and no user ids**, because a seat's person is `seats.user_id` through `(game_id, seat)`.

`game_stats` gains:

| Column | Meaning |
|---|---|
| `frames_complete` | 1 when the replay was read whole, 0 when it was missing or cut short |
| `win_kind` | §3's win kind, NULL for a draw or a game with no outcome |
| `win_turn` | the turn the last opponent left |
| `combo_rule` | which part of §3's combo rule fired (`loop`, `repeats`, `actions`, `combats`), NULL otherwise |

`seat_stats` gains:

| Column | Meaning |
|---|---|
| `turn_position` | 1–4: the seat's place in turn order, from `starting_seat` |
| `deck_hash` | §5's content hash of the deck the seat played |
| `out_kind` | how this seat left, in §3's kinds (NULL if it did not leave) |
| `out_by_seat` | the seat whose source dealt the last damage or loss before this seat left, NULL when none |
| `own_turns` | how many turns the seat took |
| `land_drops_hit`, `land_drops_missed` | over its own turns (§4) |
| `screw_turns`, `flood_turns`, `colour_screw_turns` | §4's turn counts |
| `screw_game`, `flood_game`, `colour_screw_game` | 0/1, §4's game thresholds |
| `creatures_eligible`, `creatures_attacked` | the sums behind §3's attack share |
| `commander_casts`, `commander_first_cast_turn` | from `PlayerView.CommanderCasts` and the first `cast` of a commander card |
| `life_lost_to_combat` | the part of `life_lost` that was combat damage |

`seat_turn_stats` is new: one row per seat and own turn, for the curves and the "over turns" charts:

| Column | Meaning |
|---|---|
| `game_id`, `seat`, `own_turn` | key; `own_turn` counts 1, 2, 3… for this seat |
| `turn` | the game's turn number |
| `land_played` | land plays this turn (`LandsPlayedThisTurn` at the turn's last frame) |
| `land_drops` | land plays so far this game, this turn included |
| `lands_controlled`, `mana_sources` | at the start of main (§4) |
| `hand_size`, `lands_in_hand`, `spells_in_hand` | at the start of main |
| `missed`, `screw`, `flood`, `colour_screw` | 0/1, §4 |
| `attackers`, `eligible` | attacking creatures, and the creatures that could have attacked |
| `damage_to_players`, `spells_cast` | this turn |

`seat_target_stats` is new: one row per seat and opponent seat, for the spread of aggression (`attacks`, `damage`, `combat_damage`, `commander_damage`). The row names a seat number, never a person.

`seat_card_seen` is new: one row per seat and card **name** that was ever in its hand, cast or not (§3, dead cards). Keyed by name like `seat_card_stats`, with no Scryfall id, because a picture is found by name in the card index:

| Column | Meaning |
|---|---|
| `game_id`, `seat`, `name` | key |
| `copies` | how many instances with this name were in hand |
| `in_opening_hand` | 0/1, the kept hand |
| `first_seen_turn` | the turn it was first in hand (0 for the opening hand) |
| `held_turns` | own turns it sat in hand at the start of main |
| `affordable_turns` | of those, turns it was castable by amount and colour (§4) |
| `cast` | 0/1, whether `seat_card_stats` has the card for this seat |

`seat_card_stats` does not change. It stays the list of cards cast, so ADR 0165's readers and `gamestats.Tally` keep their meaning. A separate table for cards held means no reader has to learn that a zero-cast row exists.

**Backfill.** `cmd/statsbackfill` already recomputes any record with an older `stats_version` (ADR 0165 §6). After slice 2 deploys, one run on each host brings every ended game to version 2. It is a one-shot an operator runs, as before, and it is safe to run while the server is up.

**Storage.** The estimate per game, for a four-seat game of about ten rounds:

| Part | Rows | Bytes |
|---|---|---|
| new `seat_stats` and `game_stats` columns | 4 + 1 | ~0.3 KB |
| `seat_turn_stats` (~9 own turns a seat) | ~36 | ~4.7 KB at ~130 B a row with its key index |
| `seat_target_stats` | 12 | ~1.2 KB |
| `seat_card_seen` (~25 names a seat) | ~100 | ~15 KB at ~150 B a row with its key index |
| **Total** | | **~21 KB, say 25 KB with SQLite page slack** |

At the beta's real pace (20–50 players, about **30 recorded games a day**), that is about **900 games and 22 MB a month**. At ADR 0148's ceiling, 30 tables running all day (about 320 games a day), it is about **9,600 games and 240 MB a month**. The SQLite file and its nightly backup grow by the same. The backfill adds the same per game, once, for every game already stored.

**Retention.** The new rows live as long as their game's row and go with it by cascade, exactly like ADR 0165's tables. They are derived, so they can be dropped and rebuilt at any time. If storage ever matters, `seat_turn_stats` and `seat_card_seen` are the first to drop: the replays can rebuild them. Nothing here expires on its own.

### 2. Who gets advanced stats

- **Basic stats** (ADR 0165 §7) stay on for every account.
- **Advanced stats** (everything new in this ADR) are a **free per-person setting**, off by default.
- **Computed for every game, shown by the setting.** The version-2 record is computed for every seat of every recorded game, as version 1 is, and the setting decides only whether the person's advanced pages are shown. The owner's "switching them on later fills in past games" is then immediate: the rows are already there, the backfill has already run, and turning the setting on queues nothing. It also keeps one pipeline: the game-level facts (win kind, who dealt the last blow, aggression spread) need every seat anyway. The rows hold nothing the replay does not, and the replay is kept for the same games.
- **The setting is stored on the server**, in a new `stats_prefs` table, because `user_settings` is a client body the server does not read. `stats_prefs(user_id PRIMARY KEY, advanced, advanced_set_by, share_pods, share_link_hash, share_link_at, updated_at)`:
  - `advanced` is NULL, 0 or 1. **NULL means "the default"**.
  - `advanced_set_by` is `self` or `admin`.
- **The default is one constant:** `gamestats.AdvancedDefault = false`. If Luke decides advanced stats should be automatic, this one line becomes `true`. Everyone who never touched the setting (a NULL or a missing row) gets advanced stats at once. Anyone who switched it off keeps it off. No migration and no data change.
- **The person's switch:** under Settings → Stats, and an opt-in card on `#/stats` itself ("Turn on advanced stats"). Routes: `GET /me/stats/prefs` and `PUT /me/stats/prefs`.
- **An admin's switch:** `PUT /admin/users/{id}/advanced-stats` with `{"on": true|false}`, behind `requireAdmin`. The middleware's `admin action` audit line records it. It is a switch on the person's Accounts detail (ADR 0124). The admin sets the same `advanced` column with `advanced_set_by = admin`, and the person can still switch it off. The admin switch reads no stats.
- **Guests get none.** A guest seat has no `user_id`, so no route can read it. Its rows exist only as part of the game's record, for the other seats' game-level facts.
- **Bot and agent seats** are never a person's record (ADR 0165 §9), even when they carry the person's `user_id`. A stand-in bot (ADR 0177) keeps the person's seat. Its turns count as the person's, and the per-deck page says how many games had a stand-in.

### 3. What each stat means

Every stat is per deck (§5) and overall. Every stat can be filtered by **table makeup** (below) and by a date range (all time, 90 days, 30 days). A rate shows its count beside it ("38%, 5 of 13 games"), and no rate is shown on fewer than **3 games**: the page shows "Not enough games yet" instead.

**Table makeup.** Read from the seats at read time, so nothing new is stored. A seat is a **bot** when `bot_tier` or `agent_client` is set, and a **person** otherwise. Guests count as people. The filter is:

- **All** (the default);
- **People only**: every other seat is a person;
- **With bots**: at least one other seat is a bot, and at least one is a person;
- **Bots only**: every other seat is a bot.

**How a game was won** (`game_stats.win_kind`). It is the kind of the **last opponent's exit**, in this order. The first rule that matches wins:

1. **`combo`** if any of these hold in the turn the last opponent left:
   - **`loop`**: a demonstrated loop was replayed (a `loop` entry from ADR 0193), or the loop breaker fired (§1's stats-only entry);
   - **`repeats`**: the winner's one source (by card name) was cast, activated or triggered **10 or more** times;
   - **`actions`**: the winner cast and activated **30 or more** times in all;
   - **`combats`**: the turn had **4 or more** combat phases (counted by `begin_combat` step entries). This is the infinite-combat loop. Two or three combats is an ordinary extra-combat card, and the win then falls through to its own kind.

   A **non-combat** exit is a combo by any of the four rules. A **combat** exit is a combo only by `loop` or `combats`, so a big attack after a turn of many spells stays a combat win. The rule that fired is stored in `combo_rule`, so the page can say why.
2. **`concession`** when every opponent still standing at the last exit conceded (`eliminated` with `Cause: concede`).
3. **`win_effect`**: a "you win the game" card (`game_over` with `Cause: effect`), or an "opponents lose" card (the last exit has `Cause: effect`).
4. **`poison`**: `Cause: poison`.
5. **`commander_damage`**: `Cause: commander_damage`.
6. **`combat`**: `Cause: life`, and the last damage to that player before the exit was combat damage (`damage` with `Combat`).
7. **`noncombat`**: `Cause: life` otherwise (burn, drain, life payment).
8. **`decking`**: `Cause: empty_draw`.

When several opponents leave in one batch, the kind is the first of the list above that matches any of them. A draw, or a game with no outcome, has no win kind. The same kinds, on the seat that left, are `seat_stats.out_kind`, so a person can also read how they lose.

The thresholds (10, 30, 4) are **version-2 constants** in `gamestats`. They are a heuristic, and changing one is a version bump and a backfill, like any other definition.

**Finishing place.** A breakdown of `place` (ADR 0165 §3) into 1st, 2nd, 3rd and 4th, per table size. Two- and three-seat games are shown on their own scale, never mixed into four-seat percentages. "First out" is the last place of its table.

**First-turn effect.** The win rate and average place by `turn_position` (1st to 4th to act), for four-seat games.

**Game length.** Turns, rounds, and minutes from `games.started_at` to `ended_at`. Shown as the median and the spread, because a few long games skew a mean.

**Mulligans.** Per game, and the win rate after 0, 1, and 2 or more mulligans.

**Commander.** `commander_casts` per game, the first-cast turn, and the win rate when the commander was cast by turn 5 against later or never. By commander (pair) as in ADR 0165 §7, with the win kinds.

**Aggression.**

- **Attacks per turn:** the sum of `attackers` over own turns, divided by own turns.
- **Share of creatures that attack:** `attackers` / `eligible` on each `declare_attackers` frame. A creature is **eligible** if it is an untapped creature the seat controls that was already on the battlefield at the start of the turn, so summoning-sick creatures are left out without reading haste.
- **Damage to players per turn:** `damage_to_players` / own turns.
- **Spread across opponents:** from `seat_target_stats`, the share of damage taken by the most-hit opponent ("52% at one opponent") and an evenness score from 0 (all at one) to 1 (even). Opponents are never named (§6).
- **"Over turns":** attacks and damage by own-turn number, 1 to 12+.

**Cards.**

- **Most-cast:** from `seat_card_stats`, per deck and overall, by games cast in and total casts.
- **Dead cards:** from `seat_card_seen`, a non-land card that was in hand and never cast in a game. Per deck, a card's dead rate is games dead over games it was in hand. Each card is marked **"could afford"** when `affordable_turns > 0` in at least half of its dead games, and **"couldn't afford"** otherwise. A held land is not a dead card, because land use is §4's business.
- **Cards in the most decks:** from the person's saved decks at read time. It counts decks per oracle card, leaves out basic lands, and is never computed from game rows.
- **Average mana value:** the mean mana value of the deck's non-land cards, commanders included, from the deck list and the card index. Overall, it is the mean over the person's decks, weighted by games played.
- **Colour share:** per deck, each colour's share of the coloured mana symbols in its non-land cards' costs (hybrid counts half to each, Phyrexian counts as its colour). Overall, the share of the person's games whose deck's colour identity has each colour. A five-colour deck counts for all five, so the overall shares are shown as bars, not slices adding to 100%.
- **Composition:** card types and §7's roles, counted from the deck list.

### 4. Mana: land drops, screw, flood and colour screw

All of this reads the turn samples of §1, on the seat's **own turns** only.

- **A land drop** is hit when `land_played ≥ 1` at the turn's last frame. A turn with `land_played = 0` is a **missed drop**.
  - Its **land number** is `land_drops + 1`, the drop the seat was going for, and it is stored with the turn (`own_turn`, `turn`).
  - The page shows missed drops by land number, 1–5 apart from 6+, because the owner said the early ones hurt most.
  - A turn where the seat could not play a land (`PlayerView.CantPlayLands` set at the start of main) is not a missed drop.
- **Screw:** a missed drop with no land card in hand on any frame of that turn.
- **Flood:** a turn with a land drop hit, where at the turn's last frame `lands_in_hand > spells_in_hand`.
- **A land card** is any card with a land face. A modal double-faced card with a land back face counts as a land in hand for screw (it could have been played), and as neither a land nor a spell for flood.
- **Colour screw**, at the start of main:
  - **Mana sources** are the untapped permanents the seat controls whose card has a non-empty `ProducedMana` and that are not creatures that entered this turn. Each one counts as one mana of any colour in its `ProducedMana`. This undercounts a Sol Ring and ignores conditional mana, and version 2 accepts that.
  - **A spell is castable by amount** when its mana value (X as 0) is at most the number of sources.
  - **A spell is castable by colour** when each coloured symbol in its cost can be matched to a different source that makes that colour. The match is a small bipartite matching (at most a few dozen sources). A hybrid symbol matches either colour, a Phyrexian symbol always matches (life can pay), and generic is paid from the sources left.
  - A turn is a **colour screw** turn when some non-land card in hand is castable by amount but not by colour.
  - The same check, run at the start of every own main phase while a card is in hand, gives `seat_card_seen.affordable_turns`.
- **A screwed, flooded or colour-screwed game:**
  - **screwed:** 2 or more screw turns among the seat's first 6 own turns;
  - **flooded:** 3 or more flood turns;
  - **colour-screwed:** 2 or more colour-screw turns.

  These are version-2 constants too.
- **The land-drop reliability curve**, per deck: for N = 1 to 7+, P(land drop N hit by own turn N), meaning the share of games, among those where the seat took at least N own turns, in which `land_drops ≥ N` at the end of own turn N. **7+** pools N = 7 and up. A second, dimmer line shows **lands controlled** by own turn N, which includes ramp. A ramp deck that skips drops but has lands in play should not read as screwed.
- **Beside the mana curve.** The curve is shown in the same column as the deck's **mana curve**: the count of non-land cards at mana value 0, 1, … 7+, from the deck list. Both share one x-axis (N = land N = mana value N). A deck whose 6- and 7-drops sit over a falling land curve shows the problem the owner described without any arithmetic.

### 5. Deck versions

- **A deck's version** is a content hash. `gamestats.DeckHash` is the first 16 hex characters of SHA-256 over the sorted lines `oracle_id × count` of the deck's main deck and commanders. It leaves out the sideboard, the companion, tokens and anything not in the deck list.
- **For a seat**, it is computed in `Compute` from the cards the seat owned in the final view (every zone, tokens excluded), with the oracle ids taken from the card index by Scryfall id. It is stored in `seat_stats.deck_hash`. The backfill fills it for old games too.
- **For a saved deck**, it is computed at read time from `decks.source_text` with the deck parser the Decks page already uses, then hashed with the same function. It is cached in memory by `(deck_id, updated_at)`.
- **Why a hash and not `updated_at`.** A rename, a new printing or a tag change moves `updated_at` but is not a new deck. A hash survives all three, and it can be computed for games played before this ADR. `updated_at` alone could not say which version an old game used.
- **The default view is "Since last edit":** the seats whose `deck_hash` equals the saved deck's current hash. If none do, the page says "No games since your last edit on <date>" and offers **All time**.
- **All time** is every seat with the deck's `deck_id`, or, for a deleted deck or an upload, the same `deck_name` (ADR 0165 §7's fallback).
- **A version picker** lists each hash the person has played with this deck, with its date range and game count. Games without a deck id group by name and have no version picker.

### 6. Sharing, and pods

- **Private by default.** A person's stats are shown to them alone (ADR 0165 §9), until they choose otherwise.
- **A share link.** Shipped first, because it needs nothing else:
  - `POST /me/stats/share-link` mints an unguessable token (128 bits, base64url). Only its SHA-256 goes in `stats_prefs.share_link_hash`.
  - The link opens `#/stats/shared/<token>`, a read-only copy of the person's overview and deck pages. It needs no sign-in, so it can be posted in Discord.
  - `DELETE /me/stats/share-link` revokes it. Minting again rotates it.
  - `GET /stats/shared/{token}` is rate limited per address, like the other unauthenticated routes.
  - A shared view respects the owner's makeup and date filters, and the viewer can change them.
- **What a share link shows:**
  - the person's display name;
  - every stat on their own pages, including deck names, commanders, most-cast and dead cards. That reveals much of a decklist, and the share dialog says so in one sentence.
- **What it never shows:**
  - opponents' names;
  - table names (a host may have named a table after people);
  - game ids or replay links;
  - anything about any other seat.
- **Pods: a separate ADR.** No group concept exists, and a pod brings things this ADR should not decide in passing: who creates one, how members join (link or request), its name as a free-text surface to moderate, leaving, size limits and deletion. This ADR scopes itself to the stats side and **requires a pod ADR before slice 7**. What the stats side needs from a pod:
  - **Membership:** a pod is a named set of accounts. Guests, bots and agents cannot be members. Deleted accounts leave it, and ADR 0213 blocks apply.
  - **Consent:** a member's stats are visible to the pod only when that member has switched on `stats_prefs.share_pods` **and** belongs to the pod. Joining a pod is not consent to share. The switch is. A member who has not opted in shows as "not sharing" with no numbers.
  - **Pod stats** (`GET /pods/{id}/stats`, members only):
    - each sharing member's overview;
    - the pod's own games, meaning games where two or more sharing members sat;
    - win kinds and turn order in those games;
    - head-to-head between two sharing members (wins when both sat).
  - **A leaderboard** (`GET /pods/{id}/leaderboard`, members only). It ranks sharing members with at least **5 counted games** in the range, by win rate, then average place, then games. Each row shows its count, so a 3–2 record doesn't look like a 60–40 one. It has two scopes: "Pod games" (the default) and "All games".
- **The rules every reader keeps** (ADR 0165 §9):
  - Deleted accounts never appear (`users.deleted_at IS NULL`).
  - A seat with `bot_tier` or `agent_client` never ranks and is never a human's record.
  - Opponents are never named unless they are sharing members of the same pod and the view is that pod's.
  - A person blocked under ADR 0213 never sees the blocker's stats through a pod. The share link is the one exception, because it is a link the person chose to post.

### 7. Card roles

**Roles.** A card has zero or more of these:

`removal` (spot), `board_wipe`, `counterspell`, `fog`, `ramp`, `mana_rock`, `mana_dork`, `draw`, `tutor`, `protection`, `recursion`, `graveyard_hate`, `burn`, `lifegain`, `token_maker`, `sacrifice_outlet`, `wincon`.

**Card types** (land, creature, artifact, enchantment, instant, sorcery, planeswalker, battle) come from the type line and are not roles.

**Source: Scryfall's community function tags (Tagger).** They are reached with ordinary search queries, `otag:<tag>`, on the Scryfall API, which is already the project's card-data source. Tags are not in the bulk dumps, so they are fetched by search:

- **The fetch.** A new `cmd/cardroles` (Go, so the role → tag mapping is code with a test) runs after `scripts/scryfall-refresh.sh` on the same weekly cron. For each tag it walks `GET https://api.scryfall.com/cards/search?q=otag:<tag>&unique=cards` through `has_more` and `next_page`, at 175 cards a page. It keeps only `oracle_id`.
- **Politeness.** It follows Scryfall's API rules: a descriptive `User-Agent` (`cmd_and_ctrl-cardroles/1.0`, as `scryfall-refresh.sh` sends) and `Accept: application/json`. It waits **100 ms between requests** (Scryfall asks for 50–100 ms, about 10 requests a second at most), backs off on HTTP 429, and stops after three failures. About 20 tags of up to a few thousand cards each is a few hundred requests, a minute or two once a week.
- **The output** is a generated file, `data/scryfall/card-roles.json`: `{"fetched_at": …, "tags": {"<tag>": count}, "roles": {"<oracle_id>": ["removal", …]}}`. It is written atomically like the dump. It is runtime data under `data/`, so it is gitignored, never hand-edited and never in the database. It holds no personal data, so it adds nothing to the privacy inventory.
- **The mapping** of roles to tags is a table in `cmd/cardroles` (for example, `board_wipe ← otag:sweeper`, `counterspell ← otag:counterspell`, `ramp ← otag:ramp`, `draw ← otag:draw`). The tag names cannot be checked from the sandbox this ADR was written in, which can't reach Scryfall. Slice 4 confirms each one against the live API, and `cmd/cardroles` **fails a tag that returns no cards** (Scryfall answers 404 for an empty search) rather than writing an empty role.
- **Loading.** `cards.Index` loads the file beside the dump and exposes `Roles(oracleID)`. A missing file means no tag roles: the composition chart then shows card types only and says roles are unavailable.

**Overrides, in this order:**

1. **A catalog `Purpose`** (ADR 0126 §6), where a catalog card declares one: `Sweep` → `board_wipe`, `Tutors` → `tutor`, `Lands` → `ramp`, `Draws` → `draw`. A declared purpose adds the role even when the tags lack it.
2. **A short code list**, `cards/roles_overrides.go`, of oracle ids with roles added or removed, reviewed in PRs like a catalog card, for a tag that is plainly wrong for our purposes. It is expected to stay short.

**Never Archidekt.** It is not scraped, and no other deck site is either.

**Drift.** Roles are joined at **read time** and never stored in game rows. A refresh therefore changes the composition of every deck at once, consistently, and nothing has to be backfilled. The page shows "Card roles from Scryfall Tagger, updated <date>" with a link, which is also the credit.

### 8. The pages

The owner wants a stats experience that is "pretty" and useful for improving play and decks: "feeling something is not the same as having hard data". It follows DESIGN.md. Charts sit on the flat navy-steel surfaces and use tokens only. Colour carries meaning, never decoration, and every label renders as `L.<key>` from `labels.ts`.

**Pages:**

- **`#/stats`, the overview** (the existing route, rebuilt).
  - A filter bar: makeup, range.
  - A row of stat tiles: games, win rate, average place, median game length, most-played deck.
  - Sections in this order:
    - **Results**: place distribution, win kinds, first-turn effect.
    - **Play**: aggression over turns and the spread across opponents.
    - **Mana**: missed drops by land number, and screw, flood and colour screw rates.
    - **Cards**: most-cast with card images, and cards in the most decks.
    - **Colours.**
    - **Decks**: one row per deck with its record, a results sparkline, and a link.
  - Without advanced stats, the page is ADR 0165's basic record plus the opt-in card.
- **`#/stats/deck/<deck_id>`, a deck.** Reached from a **Stats** link on the deck's card on `#/decks` (`DeckCard.svelte`) and from the overview's Decks rows.
  - A header with the commander art (`CommanderArt.svelte`), the version picker (§5) and the deck's stat tiles.
  - Then:
    - **Mana**: the mana curve and the land-drop curve in one column on one axis, then missed drops by land number, then screw, flood and colour screw.
    - **Cards**: dead cards as a grid of card images, each with its dead rate and a "could afford" or "couldn't afford" tag in words; then most-cast.
    - **Composition**: roles and types against the person's average over their other decks.
    - **Play**: win kinds, places, aggression over turns.
    - **Commander**: casts per game and the first-cast turn.
- **`#/stats/shared/<token>`**, the shared view (§6). The same components, read-only, with the owner's name in the header.
- **`#/pods/<id>/stats`**, the pod page (slice 7, after the pod ADR). The leaderboard table, then the sharing members, then the pod's games.

**Which chart for which stat:**

| Stat | Chart |
|---|---|
| Land-drop reliability against mana curve | two aligned panels on one x-axis (1 … 7+): the mana-curve histogram above, the reliability line below (0–100%) with the lands-controlled line dimmer. No dual y-axis |
| Place distribution | one horizontal stacked bar per table size, segments labelled 1st … 4th with their percentages |
| Win kinds | horizontal bars, sorted, with counts |
| First-turn effect | four bars, turn positions 1–4, win rate with counts |
| Aggression over turns | a line of attacks per turn by own-turn number, with damage per turn as a second small panel |
| Spread across opponents | a single 100% bar of the most-hit opponent against the rest, plus the evenness number |
| Missed drops by land number | a bar per land number, 1–5 emphasised, 6+ muted |
| Screw, flood, colour screw | three stat tiles with the rate, its count, and a small trend over the last 20 games |
| Colour share | horizontal bars using the fixed `--identity-*` tokens, each bar labelled with its letter and percentage |
| Composition | horizontal bars per role, with a tick for the person's average |
| Dead and most-cast cards | card-image grids (Scryfall images, artist credit kept, as everywhere) |
| Game length | a small histogram of turns, with the median marked |
| Deck row | a sparkline of the last 20 results |

**How the charts are built.** Inline SVG Svelte components in `lib/components/stats/` (`BarChart`, `StackedBar`, `LineChart`, `Histogram`, `Sparkline`, `StatTile`). The geometry (scales, ticks, bars) lives in a pure `lib/statsCharts.ts`, unit-tested in node like `lib/myStats.ts`. That follows `CoverageRing.svelte`, adds **no dependency**, and keeps the bundle as it is. A chart library would bring its own styling to fight and its own accessibility to audit, for about six chart shapes.

**Colour.** Categorical colours for win kinds and roles come from new `--chart-1` … `--chart-6` tokens in `app.css` (and DESIGN.md), checked against every skin and High contrast. Seat colours are not reused: they mean a seat at the table.

**Accessibility** (DESIGN.md §8):

- Each chart is a `<figure>`. Its `<figcaption>` says the finding in one sentence ("You hit your 4th land drop by turn 4 in 71% of games"), and its SVG has `role="img"` and `aria-labelledby` pointing at that caption.
- Every chart has a **Show as table** disclosure that renders the same numbers as a `<table>` with headers.
- Bars and points carry their own labels, so nothing is told by colour alone.
- There is no animation, so reduced motion has nothing to stop.
- The filters are a labelled radio group, and a filter change is announced once through the polite live region.

**Layout.** Site pages in `Page.svelte`, `wide` column. Two columns of sections at 820 px and up. Below the one 819 px breakpoint (DESIGN.md §5) everything is one column, which is also what a phone gets: no new breakpoint. The charts scale with their `viewBox` and keep readable minimum label sizes. A wide table scrolls inside its own container, never the page.

**Labels.** New keys go through `labels.ts` and `docs/labels.md`: the page names, the filter names and options, the opt-in, the share dialog, the version picker, and the "could afford" and "couldn't afford" tags. The empty states say what to do ("Play 3 games with this deck to see its land curve").

**Screenshots.** Every client slice attaches the pages at 1440 px, 820 px and 390 px.

### 9. Privacy (ADR 0149)

Every new stored item, and what deleting an account does to it:

| Data | Where | Kept for | Deleting the account |
|---|---|---|---|
| Version-2 columns on `game_stats` and `seat_stats` | DB | as long as the game | **Kept**: part of the game record the other players keep, reached through the tombstone, as ADR 0165 §5 |
| `seat_turn_stats`, `seat_target_stats`, `seat_card_seen` | DB | as long as the game | **Kept**, as above |
| `stats_prefs` (the setting, who set it, pod sharing, the share link's hash and date) | DB | until changed | **Deleted**. The share link stops working in the same transaction |
| Pod tables | the pod ADR | | the pod ADR |
| Card roles | `data/scryfall/card-roles.json` | until the next refresh | not personal |
| The deck-hash cache | memory | until restart | not stored |

- **`accountdelete.Inventory`** lists every new column. The game tables are marked Kept with ADR 0165's note, and `stats_prefs` is marked Deleted. `Delete` removes the `stats_prefs` row, and the audit line gains `stats_prefs_rows`.
- **`PRIVACY_SCHEMA_VERSION`** takes each slice's migration number (slices 2 and 3 each add one).
- **The privacy page** (`routes/Privacy.svelte`, numbers in `lib/privacy.ts`) changes in the slice that ships each part:
  - **Slice 2:** the "Game statistics" row gains: "For every game we also work out, from its replay, how each seat used its lands and mana, which cards it held and cast, and how it attacked. It is part of the game's record, kept with the game, and shown only to you."
  - **Slice 3:** a new "Stats settings" row: "Whether advanced stats are on for you, and whether you or an admin set it. Kept until you change it or delete your account." It also says an admin can switch advanced stats on for you, and you can switch them off.
  - **Slice 6:** a new "Sharing your stats" entry under "Who it is shared with": "If you make a share link, anyone with the link can see your stats pages under your display name, including your deck names, commanders and the cards you cast or held. They never show the other players' names. You can revoke the link at any time." Slice 7 adds the pod sentence.
- **Admins.** `GET /admin/stats` gains aggregates only: win kinds, table makeup totals and version-2 coverage (games at each `stats_version`, with `frames_complete`). An admin gets **no per-person advanced page**. The privacy page's existing "server admins can read them" stays true and needs no new wording.

## Calls made here

Each is overturnable in review:

- **Computed for everyone, shown by the setting** (§2), rather than computed only for people who opted in.
- **The combo thresholds:** 10 repeats of one source, 30 casts and activations, 4 combat phases.
- **The game thresholds:** screwed = 2 screw turns in the first 6, flooded = 3 flood turns, colour-screwed = 2 colour-screw turns.
- **Small samples:** no rate on fewer than 3 games. The leaderboard needs 5 games.
- **Mana sources count one each** (§4), versioned.
- **Deck versions by content hash**, not `updated_at`.
- **A share link readable without sign-in.**
- **Pods need their own ADR**, and pod sharing is a separate switch from joining.
- **Card roles from Scryfall Tagger**, fetched weekly into a generated file, with Purpose and a code list as overrides.
- **Inline SVG**, no chart library.
- **No per-person advanced page for admins.**

## Alternatives considered

- **Amend ADR 0165 instead of a new ADR.** The new schema, setting, sharing, card-role source and pages are more than a 0165 amendment should carry. ADR 0165 gets a dated note pointing here.
- **Record hand samples in the engine as the game runs.** This would add engine state and snapshot fields (a schema bump under #522) for something the replay already holds. It also would not help past games.
- **Compute only for people who opted in.** This stores a little less, but turning the setting on would then queue a per-person backfill. Game-level facts would also need every seat anyway. The rows add nothing the kept replay doesn't hold.
- **`updated_at` for deck versions.** It cannot tell a rename from an edit, and it cannot place old games.
- **Hand-tagged roles in the catalog only.** The catalog covers a fraction of the cards people bring, and ADR 0126 §6 declares Purpose for the bot's priced classes only. Tagger covers every card, and Purpose stays as the override.
- **Archidekt or another deck site's tags.** Scraping is out (owner decision 6), and Scryfall is already our source and credit.
- **A chart library** (Chart.js, ECharts, Layer Cake). These weigh more than the page. Their styling and accessibility would have to be bent to DESIGN.md, and they would be the client's first runtime dependency for about six simple shapes.
- **One flag on `users` for the setting.** That has nowhere to put who set it or the sharing state. `stats_prefs` keeps them together and deletes as one row.

## Risks

- **The combo heuristic will be wrong sometimes.** A storm turn that ends in burn reads as a combo, which is arguably right. A slow non-combat drain over a long turn of many triggers may read as one too. A backfilled game cannot see the loop breaker. Mitigations: the rule that fired is stored and shown ("combo: one card repeated 14 times"); the thresholds are versioned, so tuning them is a bump and a backfill run; and slice 2 tests the classifier against hand-built logs and the fixture corpus' finished games.
- **Mana and colour screw are approximations.** Counting each source as one mana, ignoring conditional mana and ignoring cost reductions all mislabel some turns. The page states the definition beside every number ("a spell you could pay for in amount but not colour, at the start of your main phase"), so it reads as a measure, not a verdict.
- **Storage.** At most about 240 MB a month at the ADR 0148 ceiling, and about 22 MB a month at the real pace. The backup grows with it. It is derived, so the two big tables can be dropped and rebuilt if it ever matters.
- **Replay reading at game end** costs more CPU than the log fold. One worker with a queue keeps it from competing with live tables. A burst of endings waits seconds, which nobody sees.
- **Tag data drift and availability.** Tagger is community data and changes weekly, so a deck's composition can shift without an edit. The page shows the refresh date. If Scryfall stops answering `otag:` searches, or a tag is renamed, `cmd/cardroles` fails loudly and keeps the last good file. Roles then degrade to Purpose and card types and say so. Before slice 4 ships, the owner confirms that using Tagger data this way fits Scryfall's API terms (it is read through the public API, credited and not resold, but it is community work).
- **A share link spreads.** Anyone it is passed to can read it, and it reveals decklists. It is revocable, the dialog says what it shows, and it never shows another person.

## Slices

Each slice is one mergeable PR with its tests, against `develop`.

| # | Slice | Schema | Server | Client | Privacy | Tier |
|---|---|---|---|---|---|---|
| 1 | **This ADR** | — | — | — | — | — |
| 2 | **Record v2** | the §1 migration | `gamestats.Version = 2`; `Input.Frames` and the turn sampler; win kinds and the combo rule; `out_kind`, `out_by_seat`, `turn_position`; land drops, screw, flood and colour screw (with the card index); `seat_card_seen`; `seat_target_stats`; `DeckHash`; the stats-only loop entry in `FullLogOf`; the single stats worker; `cmd/statsbackfill` passing frames; `GET /admin/stats` coverage and win-kind aggregates; docs/lobby.md. Tests: the classifier on hand-built logs, the sampler on a fixture replay with an undo, the matching, the hash | — | `Inventory` (Kept), the Game statistics row, `PRIVACY_SCHEMA_VERSION` | 3 |
| 3 | **The setting and the admin switch** | `stats_prefs` | `GET`/`PUT /me/stats/prefs`; `PUT /admin/users/{id}/advanced-stats`; `gamestats.AdvancedDefault`; `GET /me/stats` reports `advanced` | Settings → Stats switch; the opt-in card on `#/stats`; the Accounts switch; labels | `Inventory` and `Delete` (Deleted), the Stats settings row, the schema bump | 2 |
| 4 | **Card roles** | none | `cmd/cardroles` with its tag mapping (each tag confirmed live) and its test against a recorded search response; `data/scryfall/card-roles.json`; `cards.Index.Roles`; the Purpose override and `roles_overrides.go`; the cron line and docs/commands.md | — | none (not personal) | 2 |
| 5 | **The read routes** | none | `GET /me/stats/advanced?makeup=&range=` (overview) and `GET /me/stats/decks/{deck_id}?version=` (a deck), both refusing when the setting is off; deck composition, mana curve, average mana value and colour share from the deck list; cards in the most decks; the makeup filter; docs/lobby.md | — | none | 2 |
| 6 | **The pages and the share link** | none (`stats_prefs` has the columns) | `POST`/`DELETE /me/stats/share-link`; `GET /stats/shared/{token}`, rate limited | `lib/statsCharts.ts`, `lib/components/stats/*`, `--chart-*` tokens and DESIGN.md; the rebuilt `#/stats`; `#/stats/deck/<id>` and the Stats link on `DeckCard`; `#/stats/shared/<token>`; the share dialog; labels and e2e; screenshots at 1440, 820 and 390 px | the Sharing your stats entry | 3 |
| 7 | **Pod stats and the leaderboard** (after the pod ADR) | per the pod ADR, plus `share_pods` already in `stats_prefs` | `GET /pods/{id}/stats`, `GET /pods/{id}/leaderboard` | `#/pods/<id>/stats`, the pod-sharing switch | the pod sentence | 3 |

Slice 6 may split into the overview and the deck page if the PR grows past review size. Slices 3 and 4 are independent of each other and of slice 2's code. Slice 5 needs 2 and 3, slice 6 needs 4 and 5, and slice 7 needs 6 and the pod ADR.

## Consequences

- One run of `cmd/statsbackfill` gives every past game the version-2 record. A person who switches advanced stats on sees their whole history at once.
- No wire change and no engine change. The one protocol change is a stats-only option on `FullLogOf`.
- Every definition is versioned, so tuning a threshold is a constant, a bump and a backfill run, never a migration.
- The client gains a small chart kit that later pages (admin stats, the bot arena report) can reuse.
- Luke's call on automatic advanced stats is a one-line change to `gamestats.AdvancedDefault`.

## Back-out

| To undo | Do |
|---|---|
| Slice 2 | set `gamestats.Version` back to 1 and rerun the backfill; the new columns read as NULL and the new tables can be dropped by hand after rolling the binary back (an older binary refuses the newer schema, `ErrSchemaTooNew`; see docs/environments.md) |
| Slice 3 | remove the routes; `stats_prefs` can be dropped the same way |
| Slice 4 | stop the cron and delete `card-roles.json`; composition falls back to types |
| Slices 5 and 6 | remove the routes and pages; `#/stats` returns to ADR 0165's page; revoked links answer 404 |

## Owner decisions

Ian, 2026-11-03, on #4880. **Decided.**

1. **Who gets advanced stats.** Basic stats stay on for everyone. Advanced stats are a free per-person setting, and an admin can switch them on for a player. Luke may decide they should be automatic, and the design makes that a one-line default change (§2, `gamestats.AdvancedDefault`). Switching them on later fills in past games (§2). Guests get no stats.
2. **Combo win.** The last opponent dies outside combat in a turn with many repeated actions or a detected loop, and repeated extra combats also count. Combat damage, commander damage, poison, a "you win the game" card and concessions are their own kinds (§3).
3. **Mana troubles.** Screw is a missed land drop with no land in hand. Flood is more lands than spells in hand while still hitting drops. Colour screw is a spell in hand that couldn't be cast for colour though there was enough mana. The land curve shows how reliably a deck hits land 1 to 7+, next to its mana curve (§4).
4. **Missed land drop.** No land played on your own turn, recorded with the land number and turn (§4).
5. **Dead cards.** Not answered on the issue, so its default was taken: a card held but never cast, marked separately for whether it could have been afforded (§3).
6. **Card roles.** Tag cards properly, as Archidekt does. Evaluate Scryfall's community function tags through the Scryfall API, with catalog `Purpose` as an override. Don't scrape Archidekt (§7).
7. **Deck changes.** "Since last edit" by default, with all-time as an option (§5).
8. **Sharing and pods.** A person can share their stats. A pod sees its members' stats and a leaderboard. Sharing is opt-in per person, and deleted accounts never appear (§6).
9. **A good-looking stats experience.** Visual, with charts and clear comparisons, following DESIGN.md, with accessible text equivalents for every chart (§8).

## Owner answers to the open questions (Ian, 2026-11-04)

These supersede the "still open" list this ADR was drafted with.

1. **Default: not automatic, and an admin can change it.** Advanced stats stay opt-in for now. Instead of a code constant, the default is an **admin setting** stored on the server: `gamestats.AdvancedDefault` becomes the fallback when the setting is unset, and `PUT /admin/stats/advanced-default` (behind `requireAdmin`, audited) flips it. Slice 3 also adds an admin page listing accounts with their advanced-stats state (on, off, or following the default) and who set it. On that page an admin can switch it on or off for a person, as §2's `PUT /admin/users/{id}/advanced-stats` already allows. Luke can then make it automatic without a deploy. Anyone who switched it off themselves stays off.
2. **Combo thresholds** (10 repeats, 30 actions, 4 combats): accepted.
3. **Screw, flood and colour-screw game thresholds** (2 in the first 6 turns, 3, 2): accepted.
4. **Compute version 2 for every seat**, with the setting gating only what is shown: accepted, **on one condition**. Slice 2 measures the cost: the time and memory `gamestats.Compute` v2 adds per finished game, a backfill rate on the fixture corpus, and the stored bytes per game. It reports the numbers in its PR. If the cost is negligible at beta scale (ADR 0148), nothing more is needed. If not, compute only for opted-in players and backfill when someone opts in.
5. **Pods** get their own ADR. A pod can be formed from friends (ADR 0213), and/or joined by an invite link. C2 may also **suggest** a pod to players who often play together. A suggestion forms nothing on its own: each member accepts it, consistent with §6's per-person opt-in.
6. **A share link may be read without signing in.** The person sharing has already chosen to reveal what it shows, including decklists. Revoking the link still makes it answer 404.
7. **Leaderboard:** ranked by win rate, then average place, with a minimum of 5 games. It covers games within the pod only; there is no "All games" leaderboard.
8. **Scryfall Tagger tags:** check Scryfall's API terms before slice 4. If the terms don't allow this use, the roles come from C2's own tagging (catalog `Purpose` plus a hand-curated table) instead.
9. **Screen sizes:** PC and tablet are the primary targets and must look good there. Phones are supported too: the pages must work at 390 px, with the one-column layout below 819 px.

## Sprint

Outside the numbered sprints' scope. The active sprint is S79 — Card coverage: every Standard set, newest first, and this is an owner request (Ian, 2026-11-03, #4880) outside its card scope, as ADR 0165 was. It touches no engine rules. The slices are tracked on #4880.
