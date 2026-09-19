# ADR 0049 — Architecture review of the card ↔ engine seam (September 2026)

**Status:** Accepted · 2026-09-15 · Discussions [#557](https://github.com/krakenhavoc/cmd_and_ctrl/discussions/557) (findings), [#558](https://github.com/krakenhavoc/cmd_and_ctrl/discussions/558) (D1), [#559](https://github.com/krakenhavoc/cmd_and_ctrl/discussions/559) (D2), [#560](https://github.com/krakenhavoc/cmd_and_ctrl/discussions/560) (D3), [#561](https://github.com/krakenhavoc/cmd_and_ctrl/discussions/561) (D4, open) · Tracking issue #585

## Context

The card roadmap paused after batch 36 with 1,534 specs on `main` and
about 2,400 more scheduled. Before those went in, the question was
whether the way cards plug into the engine is the way to keep building:
cards should hook in where things happen, carry flags for the hooks they
use so the engine consults only relevant cards, wire new references into
existing machinery rather than adding parallel paths, and never need the
same change in twelve places.

The review (#557) found that the design already exists — `effects.Spec`
slots are the flags, `Watches` on triggers and replacements is the
pre-filter, and the event log is the hook surface — and that neither
runtime nor compile cost is a problem: on an 80-permanent board a tap
event costs about 6 µs through the whole listener chain and a full layer
recompute about 46 µs; an incremental build after touching one card file
takes under half a second.

What had drifted was the card layer. A clone detector over the effects
package found 87 exact-duplicate groups, 218 same-code-different-literal
groups and 326 structural groups (about 8,700 redundant lines) across
3,708 function and closure bodies. 757 batch-prefixed helper functions
existed against 46 shared primitives; 64 were referenced by nothing and
532 by exactly one file. 262 of the 559 files declaring a trigger were
one trigger wrapping one primitive. The root cause was the parallel-agent
rule "never edit shared files", which made every batch grow its own
vocabulary. Separately, `Spec.OnETB` ran off the stack and was used for
printed "When ~ enters" triggers in nine files.

## Decisions

### D1 — Authoring model (#558, option A)

Stay in Go. Add constructor functions returning `game.TriggeredAbility`
for the printed shapes (`WhenThisEnters(...)`, `WheneverYouCast(...)`,
`AtYourUpkeep(...)`); move shared code into mechanic-named, append-only
vocabulary files; run a promotion pass that merges the duplicate groups
and drops the `bNN` prefix; add a CI gate that fails a PR introducing a
new exact duplicate of six or more lines. A data-driven card format was
considered and deferred: it would add a second interpreter and a
1,500-file conversion for a benefit the constructors mostly deliver.
Issues #579, #583, #584.

### D2 — Engine seams that delete card-side copies (#559, all approved)

1. A typed per-player `TurnTally` on `Game`, reset with the existing
   per-turn tallies, replacing ~30 helpers that rescan the event log
   (#586).
2. `OncePerBatch` on `TriggeredAbility` plus a `BatchID` on events,
   replacing four hand-rolled "one or more" dedup helpers (#587).
3. Let the harvester watch the generic `EventStepBegan`; retire the
   per-step event kinds (#588).
4. A token table on `TokenSpec` replacing 111 near-identical
   constructors (#581).
5. Delete the 64 unreferenced helpers; run golangci-lint in CI (#582).
6. `docs/engine-seams.md` as the seam registry, and a `seam` label
   (#580).

### D3 — ETB paths (#560, option A) — this PR

`Spec.OnETB` is renamed `Spec.AsEnters` and is the CR 614.12 slot only:
"As this permanent enters, choose …" effects, which are not triggers and
correctly happen off the stack. Every printed "When ~ enters" declares a
`Triggered` ability watching `EventETB`, so it uses the stack and can be
answered. The two remaining tap-after-entry workarounds use
`SelfEntersTapped()`. `game.ETBEffectHook` keeps its name until D4
replaces the hook variables. Issue #578.

### D4 — Plumbing (#561, option A, approved 2026-09-16) — issue #622

One precomputed `game.CardDef` per card, built at `effects.Register`;
`game.CatalogLookup` is the only hook the catalog sets. The per-slot
variables remain as defaults that read the `CardDef`, because two dozen
test files stub them individually; nothing in the catalog assigns them.
The review's benchmark is committed as `bench_test.go`.

## Consequences

- The batch brief and AGENTS.md §7 change under D1: "append to a
  vocabulary file" replaces "never edit shared files"; the `bNN` prefix
  goes.
- The roadmap resumes at batch 37 after the constructors (#579) land, so
  new batches write constructors rather than closures.
- Nine cards that resolved their ETB instantly now give every player a
  response window; three edicts marked complete were wrong before and
  are right after.
- Things the review looked at and rejected, so they are not re-proposed:
  splitting the effects package (compile times say no), a per-event-kind
  index of battlefield permanents (invalidation cost exceeds the
  microseconds saved), an instance-ID index over zones (too many
  slice-moving writers).

## Amendment (2026-09-17, #829): D2.2's `BatchID` shipped, and the in-flight check was wrong

D2.2 above specified "`OncePerBatch` on `TriggeredAbility` **plus a
`BatchID` on events**". #594 shipped the flag without the batch id and
stood something else in its place: `dispatchTriggerLocked` declined a
matching event while an item of the same ability was in flight — on
`PendingTriggers`, on `StackMeta`, or waiting on its trigger or target
prompt. The queue and the prompts are bounded by the next priority
boundary and were a fair proxy for "this batch". **The stack is not.**
An item on the stack outlives the batch that put it there, so a
second, separate batch arriving while the first batch's trigger waited
to resolve was swallowed as if it were part of the first: Dour
Port-Mage drew one card for two bounces, which is weaker than printed
(#829, and #587's own description of the failure).

The batch id ships now, as specified:

- **`Event.Batch uint64`**, stamped by `EmitEvent` from a monotonic
  `Game.eventBatch`.
- **One boundary.** `beginEventBatchLocked`
  (`server/internal/game/event_batch.go`) advances the counter at the
  two points where play moves on, and nowhere else: a stack item
  beginning to resolve (`resolveTopOfStackLocked`) and the turn cursor
  entering a new step (`advanceCursorLocked`). One resolution is one
  batch; one turn-based action is one batch however many engine calls
  the sandbox splits it across, which is what keeps a three-creature
  attack declared one `DeclareAttacker` at a time to a single Adeline
  trigger with no special case for the declaration.
- **One guard.** `oncePerBatchAllowsLocked(batch, source, key)` is a
  test-and-set against `Game.oncePerBatchFired`, a
  `TallyKey(source, key) → batch` map. The in-flight scan is gone from
  the dispatch path.
- **Undo and restore.** The counter and the map are carried together
  by `Clone` / `RestoreFrom` and by the persisted snapshot, for the
  reason `TurnTally` is: rewinding one without the other would either
  double-fire a trigger or swallow it. No schema bump — both read as
  "no batch has fired yet" from an older file, which is the free
  direction.

What is deliberately **not** covered, and stays with #784: a dedup key
computed PER EVENT cannot ride the static `TriggeredAbility.Key`, so
Breena, the Demagogue (per attacked opponent) and Nature's Will (per
damaged player) still call `Game.TriggerInFlightForEffect` and still
carry this failure in their own dimension. That helper survives for
those two callers only, with its doc narrowed to say so; #784's
per-event batch key retires it.

One gap remains and is declared rather than papered over: two
SANDBOX-MANUAL mutations in a row, with nothing resolving and no step
change between them, share a batch. The old check behaved the same
way, and hand-shoving cards between zones has no rules occurrence to
count.

## Amendment (2026-09-18, #784): the key gets its second dimension, and the helper is gone

The amendment above closed WHEN a batch is and left open WHAT the
guard counts within one. CR 603.2c has both halves: "an ability
triggers only once each time its trigger event occurs. However, it
can trigger repeatedly if one event contains multiple occurrences."
A clause that NAMES AN OBJECT contains one occurrence per object, and
the object the catalog keeps meeting is a player — "whenever one or
more creatures you control deal combat damage to **a player**"
(Keeper of Fables' ruling, 2019-10-04: "if non-Human creatures you
control deal combat damage to two or more players at the same time,
Keeper of Fables's ability triggers for each of those players"), and
"whenever you attack **a player**" (Horizon Explorer: "will trigger
once for each player you attack"; Neyali: "triggers for each player
you are attacking with one or more tokens").

`TallyKey(source, key)` could not say that: `TriggeredAbility.Key` is
static catalog data and the player is only known from the event. So
three creatures hitting three opponents in one damage step made ONE
Treasure, and the two cards that needed the player dimension had
hand-rolled it out of `TriggerInFlightForEffect`.

- **`TriggeredAbility.BatchKey func(ev, source, g) string`** — the
  second dimension, read off the event. `oncePerBatchKeyLocked`
  appends it to `Key` and the guard is unchanged underneath:
  `oncePerBatchAllowsLocked(batch, source, key)` is still one
  test-and-set against one map. One key, two dimensions, one guard —
  **no second dedupe path**, which is the constraint this whole area
  is held to.
- **One reading in the catalog.** `effects.OncePerBatchPerPlayer`
  sets `BatchKey` to `effects.PerPlayer`: the damaged player for a
  damage event, the defending player for an attack declaration (read
  through `b17DefendingPlayer`, so a planeswalker or battle counts as
  its controller and not as a second player). Cards take it through
  `WheneverOneOrMoreCreaturesYouControlDealCombatDamageToAPlayer` or,
  when they build the item from the event, by setting the two fields.
- **The other batch boundary the rules ask for.** CR 510.4 gives a
  combat with first strike in it TWO combat damage steps. The engine
  runs both inside the cursor's single `combat_damage` step, so
  `resolveCombatDamageLocked` now opens a batch between them: a
  first-striker and a regular attacker connecting with the same
  player are two triggers, as in paper. The boundary rule is still
  one sentence — "a stack item begins to resolve, or the cursor
  enters a new step" — with the one step the cursor does not see
  named beside it.

  **Superseded in part (2026-09-18, [#717](https://github.com/krakenhavoc/cmd_and_ctrl/issues/717)):**
  the cursor now sees it. `first_strike_damage` is a real step (see
  ADR 0045's 2026-09-18 amendment), so the ordinary "the cursor
  enters a new step" boundary produces both batches and the
  hand-rolled `beginEventBatchLocked` between the passes is deleted.
  The outcome is unchanged — a first-striker and a regular attacker
  hitting the same player are still two triggers — and the boundary
  rule is now one sentence with nothing named beside it.
- **`Game.TriggerInFlightForEffect` and its doc are deleted.** Breena
  (per attacked opponent) and Nature's Will (per damaged player) were
  its only callers and both now carry a static `Key` plus
  `BatchKey`; their stack labels stay per player, because that is
  what tells two simultaneous triggers apart on the stack and what
  the once-per-turn tally reads. Breena's declared retreat — a second
  opponent swallowed while the first opponent's creature pick was
  open — goes with the helper.
- **No new state.** The dimension rides the existing
  `Game.oncePerBatchFired` key, so `clone.go`, `snapshot.go` and
  `snapshot_drift_test.go` are untouched: the same map, a longer
  string.

Cards moved onto the per-player key: Professional Face-Breaker,
Keeper of Fables, Grazilaxx, Rapacious Guest, Thopter Spy Network,
Olivia (caveat retired, now `full`), Alela (her per-player target
clause works as printed), Nature's Will, Breena, Horizon Explorer
(caveat retired) and Neyali. Every other `OncePerBatch` user was
re-read: "whenever you attack with N creatures" (Aurelia, Chivalric
Alliance, Firemane Commando), "whenever you attack" (Adeline,
Hermes, Mavren Fein, The Earth King), "is attacked" (Curse of
Opulence) and the zone-move family (Dour Port-Mage, Laelia, Teval,
Tormod, Satoru, Sidisi, On Wings of Gold) name no second object and
keep the plain guard.


## Amendment (2026-09-18, #936): the tally's other dimension is the OBJECT (CR 400.7)

Decision 1's `TurnTally` keys its two per-ability counts —
`Resolved` and `Triggered`, the "only once each turn" gates — by
`TallyKey(source, label)`, where `source` is an instance ID. An
instance ID is the identity of the CARD and survives a zone change, so
a permanent that left the battlefield and came back the same turn
still counted the object before it: its "whenever …, if this is the
first time this has happened this turn" clause could not fire again,
where CR 400.7 says the returning permanent is a new object and it
should.

The sibling registries were fixed at the one battlefield exit in #630
(`forgetPerObjectTurnStateLocked`). These could not be, and the reason
is the one dependency this key has: **`TurnTally.LoopRun` and
`TurnTally.LoopAllowance` share it**, and they are ADR 0055's loop
detector. A blink loop leaves and re-enters on every iteration, so
clearing the tally at the exit would have reset the run every
iteration and the breaker would never have reached its threshold — the
escape hatch created by exactly the loops it exists for.

So the key gets a dimension rather than a cleanup, the same way #784's
did:

- **`Card.ObjectEpoch`**, bumped once per zone change in `MoveCard`,
  next to the rest of CR 400.7's forgetting. It is a serial number for
  the OBJECT; nothing reads its value, only whether two readings are
  equal.
- **`ObjectTallyKey(source, epoch, label)`** is the per-object
  projection and **`TallyKey(source, label)`** stays the per-card one.
  One (source, label) pair, two projections, and the reader's question
  decides which:

  | Reader | Projection | Why |
  |---|---|---|
  | `TurnTally.Resolved` / `Triggered`, through `Game.ResolvedThisTurn` / `TriggeredThisTurn` (and so `b15ResolvedThisTurn`, `b11TriggeredThisTurn` and every catalog gate behind them) | per OBJECT | CR 400.7 — the clause is about this permanent |
  | `TurnTally.LoopRun` / `LoopAllowance` (`loopSuspectedLocked`, `grantLoopShortcutLocked`, `notePlayerActivationLocked`) | per CARD | a loop is a loop whichever object is running it |
  | `Game.oncePerBatchFired` (`oncePerBatchAllowsLocked`) | per CARD | each entry is compared against the live batch, so a stale one cannot match |

- **No catalog change.** The gates ask `Game.TriggeredThisTurn` /
  `ResolvedThisTurn`, which take the epoch of whichever object the
  source names now, so `batch16_helpers.go`, `exemplar_of_light.go`,
  `nykthos_paragon.go`, `breena_the_demagogue.go` and the rest are
  correct without being touched.
- **Nothing is deleted at the exit.** The old object's entries stay in
  the map, unreachable, until the turn boundary flushes the tally with
  everything else — so #935's one battlefield-exit seam is unchanged
  and gains no fourth clearing site.

State: `ObjectEpoch` is `carried` (snapshot and clone), because
nothing can re-derive how many times a card has moved. A game restored
from a file written before this shipped reads every card at epoch
zero, which merges the current turn's counts for a permanent that had
already returned — one turn, in a game that was mid-turn when the
server went down.
