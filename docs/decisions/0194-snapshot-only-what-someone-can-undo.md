# ADR 0194 — The undo ring: snapshot only what someone can undo

**Status:** Accepted (2026-10-15) · proposed 2026-10-14 · S73 — Big boards: layout and per-action cost
**Issue:** [#3136](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3136)
**Owner decisions (2026-10-15):** every recommendation accepted, see [Owner decisions](#owner-decisions-2026-10-15). The view cost is split out to its own issue.
**Numbering:** checked on 2026-10-14 across every remote head (`develop`, `main`, `cost-ledger`, `docs/adr-demonstrate-a-loop`, `docs/issue-audit`, `feat/3536-graveyard-exile-cast-cost`, `feat/3546-target-crewed-or-saddled`, `feat/3547-gitrog-ravenous-ride-fortune`, `feat/3560-skip-extra-turns-ugins-nexus`, `feat/750-conditional-block-restrictions`, `feat/playmats`, `fix/2947-resolution-copy-casts`, `fix/2950-darkness-replacement-continuation`, `fix/caddy-reload-admin-off`, `issue-3016-reuse-trigger-order`, `wip/836-one-click-default`, `pr/2326`) and open PRs. The highest in use is 0193 (PR #3699). 0190 is a gap below merged numbers and is left alone, as #3699 did. This is **0194**.
**Builds on:** [ADR 0054](0054-dice-rolls-and-coin-flips.md) decision 4 (undo rewinds the random stream), [ADR 0057](0057-win-and-lose-by-effect.md) (the end is final for all but the admin), [ADR 0075](0075-table-settings-and-host-controls.md) (the undo budget), [ADR 0127](0127-answering-repeated-prompts-for-you.md) §4 and §6 (automatic answers are free undo entries), [ADR 0148](0148-capacity-limits-for-a-public-launch.md) (per-table cost and the capacity cap), [ADR 0180](0180-secret-council.md) decision 4 (undo stops at a revealed secret vote).

---

## Context

Every commit on a room (`ws.Room.apply`, `applyBundle`, and each automatic answer in `autoAnswerThenCaptureLocked`) takes `Game.Clone()` before it mutates and pushes the clone onto a ring of `undoStackCap = 32` entries. #3131 found no low-risk way to make the clone cheaper, and #3136 asked whether the fix is sparse snapshots or copy-on-write card storage. #3041 (PR #3698) has since made the mana pool run-length, which removed the pathological case (a 2^20 pool took Clone from 9.5 µs to 30 ms). The issue asked for a re-measure first.

### Measurements, 2026-10-14, `develop` at `52bb194c4`

**Soak profile.** #2113's recipe: three random-bot games, seed 20260923000, `GOMAXPROCS=2`, `CMDCTRL_PRINTED_CACHE_CHECKS=0`, four seats (`AISEAT_SOAK_GAMES=3 AISEAT_SOAK_SEED=20260923000 go test ./internal/aiseat/ -run TestRandomBotSoak -cpuprofile … -memprofile …`).

| | CPU (cumulative) | Bytes allocated |
|---|---|---|
| `Game.Clone` (all of it the undo ring) | **8.9 %** | **26 %** (2.7 GB of 10.4 GB) |
| of which `cloneZone` (the `[]Card` arrays) | 7.3 % | 23.5 % |
| `ws.Room.captureLocked` → `protocol.ViewOfGame` | **49.8 %** | **67 %** |
| GC background mark (`gcBgMarkWorker`) | 31 % | — |

The soak runs with no dump directory, so that 50 % is the view alone, built once per commit; a production room adds the restore-point JSON marshal and write on top (irobinson010's note on #3136, from #3147).

**Clone on a realistic board.** `BenchmarkCloneMidgame` (added with this ADR, `server/internal/game/clone_bench_test.go`): four seats, 100-card decks, about 25 permanents a seat (a third with counters and extra keywords), 15 cards in each graveyard, 7 in each hand, 5,000 log events.

| | |
|---|---|
| time per Clone | 105–280 µs (the shared host was at load 16–30; the spread is the host) |
| bytes per Clone | **649 KB** |
| allocations per Clone | 243 |
| `unsafe.Sizeof(game.Card{})` | **1,432 bytes**, 116 fields |

The event log is no longer a cost (#629 made it shared and truncated on restore), and neither is the mana pool. What is left is the cards: about 400 `Card` values at 1.4 KB each, copied whole into fresh arrays, of which roughly half are library cards that almost never change between two actions.

**Memory.** The ring keeps 32 of those: **32 × 649 KB ≈ 21 MB per table**, retained for the table's life. ADR 0148 measured about 40 MB per four-bot table on the prod VM, and memory was the second resource to give out. The ring is plausibly half of it. (The comment over `undoStackCap` still says "each clone is ~10-100 KB".)

**A targeted fix was tried and is not enough.** Copying each zone's array once and deepening each card in its slot (instead of `cloned.Cards[i] = cloneCard(c)`, which moves the 1.4 KB value through the range variable, the parameter, a local and the return) saves the redundant struct copies. A/B on the benchmark: 170–284 µs before, 209–271 µs after, about 10–15 % on a noisy host, with bytes and allocations unchanged. Bytes are the cost (the allocation, the zeroing and the GC that follows), and only not copying the cards removes them.

### Who can actually reach a ring entry

`Room.undo` lets a seated player pop **only the top entry**, and only if they made it (or it is nil-stamped, which today means a bot improvisation). The budget is `DefaultUndoLimit = 1` per player per turn unless the host raises it. So:

- An entry made by a **bot seat** (every bot action is stamped with the bot's seat id through `ApplyAction`) can be popped only by the **admin**: the bot never undoes, and no human may pop it.
- Every entry **under** a bot's entry is therefore out of every player's reach until the admin pops the bot's.
- A human reaches their own entries only while they are on top, one per turn by default.

In a four-bot table, which is what ADR 0148's weights were measured on, every one of the 32 snapshots is reachable by the admin alone. In a practice table (one human, three bots), about three in four are.

## Options

1. **Micro: deepen in place.** Measured above: 10–15 % of Clone's time, no change to bytes or memory. Safe and local, but it does not move the soak profile measurably (exit criterion 3).
2. **A bot's commit leaves a barrier, not a snapshot.** When the commit's undo stamp is a bot seat, push an entry with no clone. A barrier cannot be popped by anyone; `undo` answers a new `ErrUndoPastBot` ("a bot has acted since"). A bundle still clones, because a failing step rolls it back; it drops the clone once it commits under a bot stamp. Removes the clone from nearly every commit at a bot table: on the soak, the 8.9 % CPU and the 26 % of allocated bytes (with the GC work that follows them), and the whole 21 MB ring at a four-bot table. It is about 30 lines in `ws/room.go` and touches nothing in the engine. What is lost: the admin can no longer rewind a bot's move. Players lose nothing they have today.
3. **Fewer full snapshots: a shallower ring.** Keep full clones for the top K entries only; older entries keep their metadata but drop `pre` and become barriers. K = 8 is 5 MB a table instead of 21 MB, and still allows eight consecutive take-backs with the budget raised. CPU is unchanged; memory falls by 4×. What is lost: the admin's rewind past K actions.
4. **Copy-on-write card storage.** Zones share their `[]Card` with the newest snapshot until written, so a clone copies only the zones that changed (usually one or two). The cut in bytes is large (roughly the library half and every untouched zone), but every path that writes `zone.Cards[i]` in place, or through a `*Card` into a zone, must first take a writable copy. That is hundreds of call sites across `internal/game` and `internal/cards/effects`, and a missed one is a silent undo bug: the undo restores a card that the live game already changed. It needs an aliasing guard in tests (a debug build that write-protects shared arrays, or a reflection diff of every snapshot after every soak commit) before it is safe.
5. **Sparse snapshots with replay forward.** Snapshot every N commits and, on undo, restore the nearest one and re-run the commits after it. The random stream already rewinds (ADR 0054), so replay is deterministic in principle. But a commit is a Go closure (`fn func() error`, a bundle's steps, a lobby step), not a record that can be re-run later, and the closures capture pointers into the game they ran on. Making every commit a replayable record is a rewrite of the room's commit path, and undo latency becomes N commits of engine work.
6. **Shrink `Card`.** Move the rarely set fields (most of the 116: meld, craft, goads, face-down, copy baselines, permissions refs, prepared, saddled, …) into a lazily allocated `*CardExtra`. Bytes per card might fall from 1.4 KB to a few hundred. It helps the view and every zone move too, but it touches every reader of those fields and the snapshot encoder, and it is the largest change here.

## Recommendation

Take **2 and 3 together, and 1 with them**: a bot's commit leaves a barrier, the ring keeps full snapshots for the top 8 entries, and `cloneZone` deepens in place. All three live in `ws/room.go` and `game/clone.go`; none changes what a player can undo today, none changes the engine's aliasing rules, and none changes the snapshot schema (the undo ring is never persisted). Expected effect: Clone leaves the all-bot soak profile almost entirely, its 26 % share of allocated bytes with it, and a table's ring falls from about 21 MB to at most 5 MB (zero at a four-bot table). The ADR 0148 load test should be re-run afterwards; the per-table weights may move.

Do **not** start 4, 5 or 6 for this issue. After 2 and 3, Clone is paid only for human commits, which are a few per second at most, at roughly 0.1–0.3 ms each. Copy-on-write would buy little more and carries the largest correctness risk in the engine.

**The bigger per-action cost is the view, not the clone.** `ViewOfGame` is half the soak's CPU and two thirds of its allocation, built once per commit (and then, in production, marshalled and written as the restore point and the replay line). Exit criterion 3 of S73 names both. That half should be its own issue: building the view once per commit and sharing it with every subscriber whose seat sees the same thing, and skipping the restore-point write for commits that do not need one (a bot's priority pass), are the first candidates.

## Owner decisions (2026-10-15)

The owner approved the recommendations on 2026-10-15. Each one is accepted as written:

1. **May a bot's move become impossible to undo, even for the admin?** Yes. Option 2 is taken: a bot's commit leaves
   a barrier, and nobody can pop it. The admin's rewind of a bot's move is given up. **Accepted.**
2. **How deep should full snapshots go?** K = 8 full snapshots for the top entries, older entries become barriers.
   **Accepted.**
3. **Should the view and restore-point cost be split out of #3136** into its own issue? Yes. It is filed as its own
   issue with a profile-driven fix asked for, and #3136 is decided by options 1–3 alone. **Accepted.**

## Consequences

- `undoEntry` gains a barrier kind (`pre == nil`), and `undo`, `stampUndoBlockedLocked` and the undo tooltip learn to explain it. `revealedSecretVoteSinceLocked` already treats `pre == nil` as "no reveal".
- `BenchmarkCloneMidgame` stays as the measure. The soak profile is the acceptance test (S73 exit criterion 3), alongside the existing undo tests, which must pass unchanged.
- The comment over `undoStackCap` is corrected with the measured size.
