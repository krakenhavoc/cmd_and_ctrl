# ADR 0035 — "Until end of turn" continuous effects (S32)

**Status:** Accepted · 2026-09-11 · Sprint S32 · Issue #279
**Turn-identity note:** [ADR 0059](0059-turn-machinery.md) sub-PR 1
renames the table's round counter to `Turn.Round` and adds `Turn.Seq`.
ADR 0063 has since replaced this ADR's turn stamps with the shared
`Duration` vocabulary and per-seat `TurnsBegun` counters.

## Context

ADR 0012 built the CR 613 layer system around a single assumption:
every continuous effect is sourced from a permanent on the
battlefield. `activeStaticAbilitiesLocked` walks `g.Battlefield`,
looks each card's statics up via `CatalogStaticAbilities`, and
rebuilds the whole effect list from scratch on every recompute pass.
That is exactly right for Glorious Anthem — the anthem's +1/+1
exists for as long as the anthem does (CR 113.6) — and it leaves no
room at all for an effect that outlives the object that made it.

Giant Growth is in the graveyard a moment after it resolves. Its
+3/+3 has to last the rest of the turn. There was nowhere for it to
live, which is why no pump spell had ever appeared in the catalog,
why Aang, the Last Airbender shipped with its Lesson/lifelink
clause omitted entirely (granting lifelink permanently would have
been stronger than printed, the one direction that is never
allowed), and why two of The Wandering Emperor's three loyalty
abilities are still unwritten.

The replacement pipeline had already solved the same shape once:
`Game.TurnScopedReplacements` is the Fog slot — a plain slice,
appended to at resolution, swept at the cleanup step.

## Decisions

### 1. A parallel registry, not a reuse of `TurnScopedReplacements`

`Game.TurnScopedStatics []ScopedStatic` sits beside
`TurnScopedReplacements` and follows it exactly: exported field,
appended under the resolution write lock, swept in the cleanup-step
entry hook, carried by `clone.go` in both directions.

**Why parallel rather than shared:** the two feed different
pipelines. A replacement is consulted per *event* by
`gatherActiveReplacementsLocked` and needs `Watches` / `Replace` /
CR 616 ordering metadata. A static is consulted per *recompute
pass* by `activeStaticAbilitiesLocked` and needs a layer, a
sub-layer and a timestamp. Beyond the word "turn-scoped" they share
no field. A union type would have been two disjoint structs behind
one name.

**What they do share is the discipline**, and it is the part worth
copying: one exported slice, one registration helper, one sweep
site, one clone block.

### 2. `ScopedStatic` carries the ability, an LKI source, a
       timestamp and an expiry

`ScopedStatic.Ability` is the same `game.StaticAbility` a catalog
card declares in `Spec.Static`, so the layer engine learns no second
vocabulary — `staticContinuousEffect` adapts both, and the two sort
together by timestamp within each layer bucket (CR 613.7). A UEOT
+3/+3 and an anthem compose; a UEOT layer-7b "base P/T becomes N/N"
applies before either, because sub-layer beats timestamp.

`Source` is a value copy of the creating card, captured at
registration — last known information, because the real card is
usually in the graveyard by the time the effect is read. It is
handed to `AppliesTo` / `Apply` as their `source` argument, so a
"creatures you control" predicate reads `source.Controller` exactly
as a battlefield static does.

`Timestamp` comes from the same Unix-nano clock the battlefield uses
for `EnteredBattlefieldAt`, so ordering between a floating effect
and a permanent's static is well-defined without a second scheme.

### 3. Expiry is stamped, not implied

`ExpiresAfterTurn` records the turn number current at registration;
the cleanup sweep drops every entry whose `ExpiresAfterTurn` is not
greater than the current turn number.

**This is the end-step case, and it is the reason the field exists.**
CR 514.2 ends "until end of turn" effects during the *cleanup step*.
The end step is not the end of the turn. A grant created during the
end step must therefore die at that same turn's cleanup, a few
moments later — not survive into the next turn. A duration
implemented as "expire at the next cleanup step I have not already
seen" gets this exactly backwards and leaks the grant for a whole
extra turn. Stamping the turn at registration and comparing makes
the correct answer fall out.

**Known caveat, documented on the field:** `Turn.Number` counts
ROUNDS, not seat-turns — it increments only when the cursor wraps
back to seat 0. That is invisible to "until end of turn" (every
cleanup drops everything stamped at or before it, which is the
desired semantics) but it means a future "until your next turn"
duration cannot be expressed by bumping this field alone; it would
need the active seat alongside it.

**Amendment (2026-09-18, #661): the known simplification is gone.** It
read: "an effect created *during* a cleanup step, after the sweep has
already run, survives into the following turn. CR 514.3a would give an
extra cleanup step to catch it. Nothing in the catalog can create an
effect during cleanup — no player gets priority there — so this is
unreachable today." Both halves have changed. A trigger or a
state-based action in the cleanup step now gives the active player
priority there, so a player CAN create an effect during cleanup; and
the extra cleanup step CR 514.3a asks for is built, so the sweep runs
again and catches it in its own turn. No extra bookkeeping was needed
for that: the effect is registered with an ordinary "until end of
turn" duration, and the very next cleanup step it meets — the same
turn's second one — is the sweep that ends it. See
[ADR 0006's 2026-09-18 amendment](0006-priority-foundation.md#amendment-2026-09-18-661-cleanup-grants-priority-when-something-happens-there-cr-5143a)
and `server/internal/game/cleanup.go`.

### 4. The sweep allocates; it never compacts in place

`ClearExpiredTurnScopedStaticsLocked` builds a fresh slice rather
than filtering with `s[:0]`. The backing array is shared with every
undo snapshot `Clone` has taken, so an in-place compaction would
rewrite history — the same class of bug `cloneCard` exists to
prevent on the card side. `clone.go` copies the slice header into a
fresh array in `Clone` and adopts it wholesale in `RestoreFrom`;
`RestoreFrom` already bumps `layerVersion` past
`lastResolvedVersion`, so a restored board recomputes without the
undone grants rather than serving a stale `effective`.

### 5. One-shot effects snapshot their affected set (CR 611.2c)

The card-facing primitives — `BoostUntilEOT`,
`GrantKeywordUntilEOT` — resolve their `Match` predicate ONCE, at
resolution, into a fixed set of instance IDs. Overrun does not pump
the creature you cast after it.

The set is keyed on `(instance ID, EnteredBattlefieldAt)`, not on
the ID alone. CR 400.7: a permanent that leaves the battlefield and
returns is a new object and the effect stops applying to it.
Instance IDs persist across zone changes in this engine, so an
ID-only set would keep pumping a flickered creature. The
battlefield-entry stamp is re-minted on every entry, which catches
exactly that case.

A battlefield static is the opposite and stays that way: Glorious
Anthem's `AppliesTo` genuinely re-runs every pass, because a
creature that enters under an anthem does get +1/+1. `StaticUntilEOT`
is the escape hatch for a duration effect that really is
board-sensitive; it uses the declared `AppliesTo` verbatim.

### 6. Layer 6 grants are only as real as the keyword behind them

`GrantKeywordUntilEOT` appends to `Characteristic.Abilities`, which
`game.HasKeyword` reads. Twelve tokens are honoured by the engine
(flying, reach, first strike, double strike, deathtouch, lifelink,
trample, vigilance, menace, defender, haste, flash). Granting one of
those is real. Granting anything else — hexproof, indestructible,
protection, ward — appends a string nothing reads.

The primitive deliberately does **not** reject unknown tokens. A
declared-but-inert grant is how The Wandering Rescuer is already
written, so that the day hexproof lands in the targeting gate the
card starts working untouched. What the primitive does instead is
require the card comment to say so: a card that ships an inert grant
is weaker than printed and must declare it.

### 7. Layer 1 stays deferred

Copy effects (Clone, Three Steps Ahead) are unaffected by this work.
A turn-scoped duration does not help a layer that does not exist —
layer 1 is still the stub ADR 0012 left it as, and Clone is still
blocked on it, not on durations.

## Consequences

- Giant Growth and Overrun are in the catalog; the pump is on the
  wire the moment the spell resolves and gone after cleanup.
- Aang, the Last Airbender's Lesson/lifelink clause is restored.
  The card no longer ships a printed line it does not have.
- The Wandering Emperor's -2 ("+2/+1 and lifelink until end of
  turn") is now writable as two primitives; only the loyalty-ability
  hook (S27) is missing.
- Katara Water Tribe's Hope's "base power and toughness X/X until
  end of turn" is writable via `StaticUntilEOT` at layer 7b; it is
  still blocked on waterbend.
- The Wandering Rescuer is **not** fixed by this work. Its hexproof
  grant is inert because hexproof is not enforced in the targeting
  path at all — not because of a duration gap. Its simplification
  note stands unchanged.
