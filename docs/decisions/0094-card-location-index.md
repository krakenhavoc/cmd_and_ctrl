# ADR 0094 — Card lookups go through a checked hint table, not a walk of every zone

**Status:** Accepted · 2026-09-24 · S47 — Bot seat, round 2: wire it, measure it, sharpen it (tracker #890); operability/hygiene tracker #892
**Issue:** [#1479](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1479) (the view looks each card up with a linear scan of every zone)
**Numbering:** reserved on #1479 on 2026-09-24 after sweeping the ADR file names in the history of
every remote branch (`git log --remotes=origin --name-only -- docs/decisions`). The highest number
present anywhere was **0093**, and no open issue's claim comment had reserved 0094.

**Related:** [ADR 0033](0033-ai-bot-seat.md) amendment 2026-09-24 (#1401, the public log's resumable
fold — the same "a derived structure next to the game, and a test that holds its assumptions against
the source" shape), [ADR 0041](0041-game-persistence.md) (snapshots), [ADR 0084](0084-phasing.md)
(why `PhasedOut` is not a zone a lookup sees).

---

## Context

Every "which zone is this card in" question in package `game` came down to `findCardZoneLocked`
(`mutations.go`), which ran `Zone.Contains` over the battlefield, the stack, exile, and then each
seat's library, hand, graveyard and command zone in turn. `LookupCardForEffect`, `ControllerOfCard`,
`findCardByIDLocked` and `findCardAndZoneLocked` each did that walk, or one like it, and then walked
the zone they found a second time to get the card.

The view asks once per card it projects (`castSourceOf`, `liveCardForView`, `liveCardForAbilityRows`,
`stampActivatedAbilities`, `stampLibraryTopSpecialActions`), so every view was O(cards²). After #1401
took the public log out of the way, a 3-game random soak (seed 20260923000, `GOMAXPROCS=2`) put
`LookupCardForEffect` at **8.0%** of all CPU and `Zone.Contains` alone at 6.6%. Every sample of it
came from the view.

Instrumenting the same soak showed 1.13 million lookups: 40% battlefield, 34% hand, 22% graveyard or
command zone, 4% library, and **none** for a card no zone held.

## Decision

### 1. An index from instance ID to (zone slot, seat index, position), used as a hint

`Game.cardIndex` (`game/card_index.go`) holds a table from instance ID to where the card was the last
time the table was built: a zone **slot** (battlefield, stack, exile, library, hand, graveyard,
command), a **seat index** for the per-seat slots, and the **position** in that zone's `Cards`.
`locateCardLocked` answers every lookup:

1. Resolve the slot and seat index against the live game, and read the card at the position. If it
   has the ID, that is the answer. This is the fast path: an atomic load, a map read, one comparison.
2. If not, search outward from the old position in the same zone. A removal lower in a zone shifts a
   card by one or two places, and a put-on-the-bottom shifts it the other way.
3. If the card is not in that zone, fall back to the reference walk: the old `findCardZoneLocked`,
   unchanged except that it also returns the position.

`findCardZoneLocked`, `LookupCardForEffect`, `ControllerOfCard`, `findCardByIDLocked` and
`findCardAndZoneLocked` all go through it now, and read the card at the returned position rather than
walking the zone again.

### 2. Nothing invalidates it

The table is never trusted, so it never needs invalidating. A zone move, a token created or ceasing to
exist, a flicker that mints a new instance ID, phasing, `RestoreFrom`, `Clone`, a snapshot load: none
of them touch the index. A hint they make stale costs its lookup one check and a fallback. It cannot
produce a wrong answer, as long as three things hold:

- **A hint holds no pointer.** It names its zone by slot and seat index and is resolved through the
  live `Game` on every read. This is what makes undo safe. `RestoreFrom` swaps every zone pointer, and
  the discarded zones still hold their cards, so a hint that kept a `*Zone` would pass its check
  against a zone the game no longer has.
- **The index and the reference walk cover the same zones in the same order.** Both go through
  `eachIndexedZoneLocked`. `PhasedOut` is in neither (ADR 0084: a phased-out permanent is treated as
  though it does not exist).
- **An instance ID is in at most one of those zones at a time.** This is an engine invariant. The index
  cannot check it cheaply, so the cross-check mode (§4) is there to catch a breach.

`RestoreFrom` keeps the live game's table. Most cards are where they were before the undone action,
so the table keeps answering on the fast path. A `Clone` or a restored snapshot is a new `*Game` with
an empty table. Its first lookup builds one.

### 3. Rebuilt when the fallbacks add up; published whole for concurrent readers

Views run under the **read** lock, and several of them (the room's and each bot seat's) run at once. So
a published table is immutable, and a new one replaces it whole through an `atomic.Pointer`. Readers on
the fast path write nothing shared.

The table is rebuilt when the off-fast-path work charged against it reaches **8×** its size, counted
in cards examined. Building a table touches every card and inserts it into a map, several times the
per-card cost of a fallback's comparison. At that factor a rebuild costs no more than the fallbacks it
ends. Two things are not charged:

- **The first build**, which happens on a game's first lookup.
- **A fallback that found nothing.** That is a card no zone holds: a token that has ceased to exist, or
  the old ID of a card that changed zones. A new table would not answer it any faster.

Two readers that cross the threshold together may both rebuild. The tables they build are identical,
since nobody can move a card while they hold the read lock, and whichever is published last wins.

### 4. Cross-check mode, on for five test packages

`game.SetCardIndexCrossCheck(report)` makes every lookup also run the reference walk and pass any
disagreement to `report`. The `game`, `protocol`, `cards/effects`, `legal` and `actions` test packages
turn it on in `TestMain` with a panic. Every lookup those suites make, including the view's lookups in
#1401's enumerator-driven played games with undos and restores, is checked against the walk it
replaced. It is off in production and in `aiseat`, whose nightly budgets are measured without it.

## What pins it

- **`TestCardIndexAgreesWithTheZoneWalk`**, the property. It runs three seeds × 400 steps of random
  4-seat games covering every class of change the index could be stale after: sandbox moves between
  every pair of zones and to the bottom, draws, mills, shuffles, tokens created and dying, flickers
  that mint a new ID, phasing out and in, undo to a random depth through `Clone`/`RestoreFrom`, play
  continued on a clone, and persisted-snapshot restores. After every step, every card in a zone, every
  phased-out card, a sample of departed IDs and some never-used IDs are looked up through the index. Each
  answer must equal a reference walk written out in the test, zone by zone in the pre-#1479 order,
  rather than borrowed from `card_index.go`.
- **`card_index_source_test.go`**, the source checks, in the style of #1401's
  `TestEventLogIsReplacedOnlyWhereTheGenerationMoves`:
  - `TestCardIndexHintsHoldNoPointers` holds the first assumption by reflection.
  - `TestCardIndexIsReadOnlyThroughItsCheckedLookup` fails on any code outside `card_index.go` that
    touches `Game.cardIndex`, and on any read of the table outside `locateIndexedLocked`.
  - `TestZoneCardsAreWrittenOnlyAtKnownPoints` lists every write to a `Cards` slice, a whole element of
    one, or an element's `InstanceID`, outside the `Zone` primitives and `MoveCard`. A new writer fails
    until it is added with the reason it keeps one ID in one zone. That is the assumption the index
    cannot check, and the cross-check mode's dynamic net. The check is syntactic. A write to another
    type's `Cards` field is listed separately as "not a Zone".
- Behaviour tests: the table survives an undo and answers for the restored zones; a clone and a
  restored snapshot start empty; a shifted card is found by the outward search, not a fallback; a
  rebuild comes on the lookup that crosses the threshold, not the first fallback and not never; eight
  concurrent readers on the rebuild path agree under `-race`; the cross-check reports a card put in two
  zones.
- **`protocol.TestViewCardLookupsAreLinearInTheBoard`**, the cost test, in the style of #1261's
  `view_cost_test.go`. It counts the lookups one steady-state view makes and the cards they examine, on
  boards of n permanents plus 2n graveyard cards at n = 10 and n = 40. A lookup must examine about one
  card at both sizes. With the walk, it examined 26.5 cards at n = 10 and 70.7 at n = 40.

## Rejected

- **A per-view index** (#1479's first shape, and the triage's suggested first step). Build an ID → zone
  map once per `ViewOfGame` and pass it down. It catches the view's lookups, which are all of the
  measured cost, but it would add a parameter to every view helper that looks a card up. It would also
  leave the engine's own lookups (effects, the enumerator, `ControllerOfCard` on every action's
  authorisation) on the walk, and it rebuilds on every view even when nothing moved. The triage's worry
  about the game-level shape was the maintenance cost across clone, undo and snapshot. A table that is
  checked on every read and never invalidated has no such cost, which is §2.
- **An index maintained at every mutation point.** Keep the table exact by updating it in `Zone`'s
  methods and at every direct write to `Cards`. `Zone` has no back-pointer to its `Game`. Package
  `game` writes `Cards` directly in about a dozen functions, and the test suites in many more. Every
  one would have to keep the index in step, or a lookup would return a wrong card. With the checked hint
  the worst a missed write can do is cost a fallback.
- **A mutable map behind a mutex, patched on every fallback.** Moved cards would be O(1) after their
  first miss, without a rebuild. But the fast path would take a lock that every concurrent view
  contends on, and the fast path is almost every lookup.
- **Positions left out: ID → zone only.** Simpler to keep fresh, but a lookup would still walk the zone
  it names. Battlefield cards are 40% of lookups and were already found by walking the battlefield
  first, so that shape leaves the battlefield half of the quadratic in place.
