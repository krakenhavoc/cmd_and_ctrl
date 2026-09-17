# ADR 0058 — "Doesn't untap": untap-step restrictions, a next-untap-step marker, and stun counters

**Status:** Accepted · 2026-09-17 · unscheduled (card-coverage audit, wave 2) · tracked on [#751](https://github.com/krakenhavoc/cmd_and_ctrl/issues/751)
**Numbering:** 0058 was assigned for #751. 0052 is reserved for the
emblems ADR ([#623](https://github.com/krakenhavoc/cmd_and_ctrl/issues/623)).
0056, 0057 and 0059 are being drafted at the same time for #748, #749
and #753. On 2026-09-17, after `git fetch origin`, every remote branch
(207 of them) was checked with the AGENTS.md §4 loop. No branch has a
`docs/decisions/0056-*` or higher file, and the highest on any branch
is 0055.
**Builds on:** [#548](https://github.com/krakenhavoc/cmd_and_ctrl/issues/548)
(`server/internal/game/untap.go`: one untap primitive, and the
`UntapStepPermission` that widens the untap step),
[ADR 0013](0013-replacement-effects.md) (replacement effects, and the
#710 step-transition resume), [ADR 0041](0041-game-persistence.md) and
[ADR 0044](0044-surviving-a-deploy.md) (snapshots, undo and restore points),
[ADR 0033](0033-ai-bot-seat.md) (the bot tiers) and
[ADR 0011](0011-mana-pool-and-auto-tapper.md) (the auto-tapper).
**Related:** [#680](https://github.com/krakenhavoc/cmd_and_ctrl/issues/680)
(Tangle's declared caveat), [#620](https://github.com/krakenhavoc/cmd_and_ctrl/issues/620)
and [#633](https://github.com/krakenhavoc/cmd_and_ctrl/issues/633)
(the `Card` layout guard), [#753](https://github.com/krakenhavoc/cmd_and_ctrl/issues/753)
/ [ADR 0059](https://github.com/krakenhavoc/cmd_and_ctrl/pull/824)
(extra turns, which this design needs nothing from; ADR 0059 Decision 7
requires the marker to be used up by an untap step, not a turn change,
which is what Decision 2 does), [ADR 0056](https://github.com/krakenhavoc/cmd_and_ctrl/pull/818)
Decision 5 (counter events and placers; stun removal is outside it), and
[#826](https://github.com/krakenhavoc/cmd_and_ctrl/issues/826) (the
untap-step choice ADR, filed for Decision 8).
**Owner decisions:** answered 2026-09-17 and recorded under
[Decided (2026-09-17)](#decided-2026-09-17): a "won't untap" badge on
tapped permanents only plus a hover footer line, with no log line
(question 1); the first wave includes the three stun cards (question 2);
Winter Orb, Static Orb and Winter Moon wait for #826 (question 3).

Line references are to `origin/develop` at `684f2786`.

## Context

### The untap step can add permanents but can't hold any back

CR 502.3: "the active player determines which permanents they control
will untap. Then they untap them all simultaneously. […] Normally, all
of a player's permanents untap, but effects can keep one or more of a
player's permanents from untapping."

The engine has half of that sentence:

- `untapStepSetLocked` (`server/internal/game/untap.go:272-296`) is the
  only place the set is chosen. Every tapped permanent the active
  player controls goes in with no checks (`:283-287`).
- `UntapStepPermission` (`untap.go:108-144`), declared as
  `Spec.UntapStep` (`cards/effects/spec.go:451`), bridged by
  `CatalogUntapStepPermissions` (`untap.go:158`, `carddef.go:196`) and
  gathered once per step by `activeUntapStepPermissionsLocked`
  (`untap.go:225-251`) through `CatalogAbilityKey`, can only **add**
  other players' permanents (Seedborn Muse, Unwinding Clock).
- `untap.go:87-94` leaves the other direction out on purpose: "A 'this
  permanent doesn't untap' restriction (Winter Orb, Tangle's rider, Icy
  Manipulator's …) … untapStepSetLocked is where it goes when the first
  such card is written". (Icy Manipulator's oracle text only taps. The
  comment is wrong about that card.)
- `performUntapStepLocked` (`untap.go:320-335`) is called from the
  `StepUntap` case of `finishStepEntryLocked` (`game.go:994-1009`). That
  case runs only when the step actually begins: a step skipped by a
  CR 614 replacement (Stasis, `cards/effects/stasis.go`) takes the
  `canceled` branch at `game.go:905-910`, a CR 616 ordering pause (#710)
  returns before it, and so does the mulligan window (`game.go:925`).

### What the missing half costs today

- **Tangle ships weaker than printed.** Its caveat (`cards/effects/tangle.go:35`)
  reads "Attacking creatures still untap during their controller's next
  untap step; only the combat damage prevention works."
- **Stun counters have a name and nothing else.** `CounterStun`
  (`counter_types.go:37-41`) has a comment that files the rule under
  "cleanup-step turn-based action territory", but CR 122.1d makes it a
  replacement on untapping. `untapPermanentLocked` (`untap.go:189-199`)
  has no hook for it. Proliferate's auto-pick already treats stun as a
  harmful counter (`cards/effects/proliferate.go:79-82`).
- **The audit** (#751) found 17 cards where this is the only core
  blocker and 20 where it is one of them. The verifier estimated about
  12 of those are real.

### The rules, and the rulings that pin the edge cases

- **CR 502.3**, quoted above.
- **CR 122.1d**: "One or more stun counters on a permanent create a
  single replacement effect that stops the permanent from untapping.
  That effect is 'If a permanent with a stun counter on it would become
  untapped, instead remove a stun counter from it.'"
- **CR 701.26b**: only tapped permanents can be untapped. So a stun
  counter on an untapped permanent stays where it is.
- **CR 614.10a**: "Anything scheduled for the 'next' occurrence of
  something waits for the first occurrence that isn't skipped."
- **CR 701.43a-b** (exert): "you choose to have it not untap during
  your next untap step", and several exerts before that step "expire
  during the same untap step".
- **CR 400.7**: a permanent that changes zones is a new object.
- **CR 707.2**: a copy gets copiable values. Effects applied to the
  original are not copiable values.
- **CR 616.1**: when several replacement effects apply to one event,
  the affected object's controller chooses the order.

The official rulings settle the three questions the rules text leaves
open:

1. **"Its controller's next untap step" follows the permanent, not the
   player.** Wall of Frost (2014-07-18), Frost Titan (2010-08-15) and
   Tamiyo, the Moon Sage (2012-05-01) all say: "If the permanent changes
   controllers before its old controller's next untap step, it won't
   untap during its new controller's next untap step."
2. **The effect is used up even if nothing was tapped.** Wall of Frost:
   "If the creature isn't tapped during its controller's next untap step
   … Wall of Frost's ability has no effect at that time. It won't try to
   keep the creature tapped on subsequent turns." Kefnet's Monument and
   Combat Celebrant (exert) say the same thing.
3. **The affected set is fixed when the effect is created.** Tangle
   (2025-09-19): "The creatures that are not going to untap are
   determined at the time this spell resolves." Sleep (2018-07-13) and
   Kiora Bests the Sea God (2020-01-24) say the same. Kiora adds:
   "Any nonland permanents that come under the target player's control
   after the second chapter ability resolves will untap as normal."

"**Your** next untap step" is keyed on the player instead. Combat
Celebrant (2017-04-18): "If you gain control of another player's
creature until end of turn and exert it, it will untap during that
player's untap step."

### The shapes printed on real cards

One pass over the Scryfall dump found every Commander-legal card that
prints "doesn't / don't / won't untap during", with reminder text
removed. The table groups them by shape:

| shape | cards | examples | where it goes |
|---|---:|---|---|
| static, self: "this doesn't untap during your untap step" | 39 (+12 with an "if" condition) | Mana Vault, Basalt Monolith, Grim Monolith, Goblin Sharpshooter, Colossus of Sardia | Decision 1 |
| static, attached: "enchanted creature doesn't untap during its controller's untap step" | 69 (mostly Auras) | Claustrophobia, Dehydration, Paralyze, Bind the Monster | Decision 1 |
| static, filtered: "X don't untap during their controllers' untap steps" | 24 | Meekstone, Back to Basics, Intruder Alarm, Choke, Rising Waters | Decision 1 |
| one-shot, controller-keyed: "… during its / their controller's next untap step" | 101 | Tangle, Wall of Frost, Kefnet's Monument, Frost Titan, Frost Breath | Decision 2 |
| one-shot, player-keyed: "… during your / that player's next untap step" | 29 | Sleep, Kefnet's Last Word, Thalakos Lowlands, Blinding Beam | Decision 2 |
| duration-bound: "… for as long as you control this creature / as this remains tapped" | 23 | Dungeon Geists, Amber Prison, Rust Tick, Wall of Stolen Identity | out of scope (Decision 8) |
| choose which to untap: "can't untap more than N", "you may choose not to untap" | 9 + 45 | Winter Orb, Static Orb, Winter Moon, Rust Tick, Amber Prison | out of scope (Decision 8) |
| stun counters (CR 122.1d) | 82 | Alchemax Slayer-Bots, Dreamdew Entrancer, Cryogen Relic, Baloth Prime | Decision 3 |
| "next two untap steps" | 1 | Telekinesis | out of scope |

One card prints a replacement on untapping of its own: Freyalise's
Winds ("If a permanent with a wind counter on it would untap during its
controller's untap step, remove all wind counters from it instead").

## Decision 1 — `UntapStepRestriction`: the counterpart of `UntapStepPermission`, read at the step

```go
// untap.go
type UntapStepRestriction struct {
    // Restricts reports whether `target` doesn't untap during its
    // controller's untap step because of `source`. Same signature and
    // argument order as StaticAbility.AppliesTo, so selfOnly and
    // AttachedToSource plug in unchanged.
    Restricts func(target *Card, g *Game, source *Card) bool
    Label     string
}

var CatalogUntapStepRestrictions func(oracleID string) []UntapStepRestriction
```

It is declared as `Spec.UntapStepRestrictions []game.UntapStepRestriction`.
Adding the slot takes the four edits AGENTS.md §7 "Adding a `Spec` slot"
lists (Spec field, `CardDef` field, `buildDef` line, engine call site).
The `CatalogUntapStepRestrictions` variable is added as well, because
the game-package untap tests stub it the way they already stub
permissions (`untap_test.go:16-22`).

**How it is read.** `activeUntapStepRestrictionsLocked` walks the
battlefield once per step, as `activeUntapStepPermissionsLocked` does.
It goes through `CatalogAbilityKey`, so a Mana Vault that has lost all
its abilities doesn't keep itself tapped, and a Meekstone that has lost
its abilities restricts nothing. In `untapStepSetLocked` the check sits
**inside the active player's own branch only**:

```go
if c.Controller == activePlayer {
    if restricted(c) || c.hasNextUntapSkipFor(activePlayer) { continue } // Decisions 1, 2
    ids = append(ids, c.InstanceID)
    continue
}
if c.hasPlayerKeyedUntapSkipFor(activePlayer) { continue }             // Decision 2
for _, bp := range permissions { … }                                    // unchanged
```

**Scope is the controller's own untap step, and the engine enforces
it.** Every printed static in the family says "your untap step", "its
controller's untap step" or "their controllers' untap steps". So a
restriction never needs to know which step it is in, and the struct has
no `AppliesTo`. It also gets the known interaction right: a Seedborn
Muse untap during another player's untap step still untaps Mana Vault,
because that step is not the vault controller's.

**Conditions live in the predicate.** "This land doesn't untap during
your untap step if you control a Forest" and Meekstone's "power 3 or
greater" are read when the step asks.

**Layers are recomputed first.** `performUntapStepLocked` calls
`RecomputeLayersIfStaleLocked` before it builds the set. This is needed
for Meekstone's power, Back to Basics' "nonbasic" and ability removal,
and it also fixes a bug that exists today. On a turn wrap the cleanup
hook sweeps until-end-of-turn statics (`game.go:1077`, which bumps
`layerVersion`) and recurses straight into the next untap step, and
nothing recomputes in between. So the permission branch reads stale
layer output. For example, a Seedborn Muse whose "loses all abilities
until end of turn" ended at cleanup doesn't untap anything in the next
player's untap step. A throwaway game-package test confirmed this
(the control without the ability loss passed). The bug is reported
separately so the fix can land before this ADR's code.

**Why a `Spec` slot evaluated at the step, and not a `Restriction` bit.**
S24's restriction vocabulary (`game/restrictions.go`) has the other
obvious home: a sixth bit on `Characteristic.Restrictions`, OR'd in
layer 6 by `RestrictSelf` / `RestrictAttached`, which would put a token
on the wire for free. It is rejected for three reasons:

1. **It reads the wrong P/T.** A layer-6 `AppliesTo` runs before layer
   7. Meekstone's "creatures with power 3 or greater" would miss a 2/2
   pumped to 3/3 by an anthem, and would still catch a 3/3 shrunk to
   2/2. When `restrictions.go` says "the layer does not matter", it
   means the order between restriction effects. That argument doesn't
   cover a predicate that reads P/T.
2. **It is the same question as the permission.** CR 502.3 asks one
   question, which permanents untap, and the permission and the
   restriction are its two directions. Keeping both in
   `untapStepSetLocked`, read the same way at the same moment, means
   they can't drift. `Spec.UntapStep`'s own doc comment
   (`spec.go:421-451`) already argues that a change to what a
   turn-based action does "has no layer to sit in". That argument
   covers this direction too.
3. **It would be computed on every recompute** for a question that is
   asked once per turn per player.

Decision 4 still puts the answer on the wire, so the thing the bit
would have given for free is not lost.

## Decision 2 — The "next untap step" marker is data on `Card`

```go
// card.go
// NextUntapSkips records one-shot "doesn't untap during … next untap
// step" effects on this permanent (CR 502.3). nil for nearly every card.
NextUntapSkips []UntapSkip

type UntapSkip struct {
    // Player is whose next untap step this permanent sits out.
    // uuid.Nil means "its controller's": whoever controls the
    // permanent when an untap step of theirs begins (Wall of Frost,
    // Frost Titan and Tamiyo rulings).
    Player uuid.UUID
}
```

**Keying.** The two printed shapes are two values of one field:

- **Controller-keyed** (`Player == uuid.Nil`): Tangle, Wall of Frost,
  Kefnet's Monument, Frost Titan, Kiora Bests the Sea God. This is
  resolved when the step runs, not when the effect is created, so a
  control change before the step hands the marker to the new
  controller's next untap step, as the rulings require. Nothing has to
  happen when control changes.
- **Player-keyed** (`Player == p`): Sleep ("that player's"), Kefnet's
  Last Word, Thalakos Lowlands and exert ("your"). It applies in `p`'s
  next untap step whoever controls the permanent. That covers a
  Seedborn-style permission untap too: "doesn't untap during that
  player's next untap step" holds whatever the reason for untapping.

**Used up at the step, tapped or not.** `performUntapStepLocked` for
active player `a` removes every entry with `Player == a`, and every
`uuid.Nil` entry on a permanent `a` controls, **from every battlefield
permanent, tapped or untapped**. This happens after the set is computed
and before anything untaps. That is ruling 2: an already-untapped
permanent uses its marker up, and the marker doesn't wait for a turn
when the permanent is tapped.

**Added once.** An entry equal to one already there is not added again,
so two Frost Breaths, or an exert on a creature Tangle already caught,
both end at the same untap step (CR 701.43b).

**Survives what it must.**

- A skipped untap step (Stasis) never reaches the `StepUntap` case, so
  nothing is used up. The marker waits for the next untap step that
  happens (CR 614.10a). The same holds for a #710 ordering pause and for
  the mulligan window. Nothing new has to be written for this. It
  follows from putting the code in `performUntapStepLocked`, and a test
  pins it.
- Extra turns (#753, ADR 0059): an extra turn's untap step is that
  player's next untap step, so it uses the marker up. The design reads no
  `Turn.Number`, so ADR 0059's move to `Seq` can't break it. When ADR
  0059's rotation seam ends a departed active player's turn early, only
  the next player's untap step uses markers up, through
  `performUntapStepLocked` as usual; the seam itself uses nothing up.
- A player-keyed entry naming a player who has left the game is never
  used up and has no effect. It is not swept. Decision 4 leaves it off
  the wire.

**Dropped when it must be.**

- **Zone change (CR 400.7).** `MoveCard`'s battlefield-exit block
  (`zone.go:176-211`) sets `c.NextUntapSkips = nil` next to `Tapped`,
  `Counters` and `GoadedBy`. The exile-to-battlefield path at
  `effect_api.go:2460-2476` resets the entering card field by field, so
  it gets the same line.
- **Copies (CR 707.2).** The marker is not a copiable value.
  `CopiableValuesOf` (`copy.go:131`) doesn't read it, and a token copy
  starts clean. A test pins this.
- Transforming and turning face down or face up keep the marker, because
  the permanent is the same object.

**Carried where state is carried.**

- `cloneCard` (`clone.go:300`) deep-copies the slice, because a value
  copy would alias it into every undo snapshot.
- `cardSnapshot` (`snapshot.go:259`) gets
  `NextUntapSkips []untapSkipSnapshot \`json:"nextUntapSkips,omitempty"\``.
  It is classified `carried` in `cardFields` (`snapshot_drift_test.go:133`).
  A file written before this change has no key and restores with no
  markers. That is correct, because no card could have created one.
- Undo and restore points need nothing more: the marker is plain data.
  That is the reason it is **not** a `ScopedStatic`-style registry entry
  with an expiry predicate. `TurnScopedStatics` is closures, dropped by
  the snapshot and counted in `ContinuationCensus` (`snapshot_drift_test.go:124`),
  so a marker kept there would block restore points for a whole round of
  the table.
- **Layout guard.** A slice header is 24 bytes and 8-aligned, and it
  goes between two 8-aligned fields, so `TestCardHasNoInteriorPadding`
  stays green. `UntapSkip` has no bool, so #633's planned recursion into
  embedded structs has nothing to flag.

**The effect API.**

```go
// game: battlefield only; a card that isn't on the battlefield is a
// no-op (it isn't a permanent, cf. CR 701.43c).
func (g *Game) SkipNextUntapForEffect(cardID, player uuid.UUID) error

// effects: the card-facing primitive.
type DoesntUntapNextUntapStep struct {
    Targets []uuid.UUID // fixed at resolution (the Tangle / Sleep / Kiora rulings)
    Player  uuid.UUID   // uuid.Nil = each permanent's controller's
    Label   string
}
```

Most of the 101 controller-keyed cards print "tap target X. It doesn't
untap during its controller's next untap step", so `TapAndFreeze{Targets}` wraps
`TapTargetForEffect` and the primitive above. It is a helper in
`cards/effects`, not an engine concept.

## Decision 3 — Stun counters are a rule in the untap primitive

`untapPermanentLocked` becomes:

```go
if c == nil || !c.Tapped { return }          // CR 701.26b: nothing would become untapped
if c.Counters[CounterStun] > 0 {             // CR 122.1d
    _ = g.applyCounterLocked(c.InstanceID, CounterStun, -1)
    return                                    // stays tapped; no EventUntapCard
}
c.Tapped = false
g.EmitEvent(…EventUntapCard…)
```

- **It covers every untap.** The primitive is the only way `Tapped`
  goes false (AGENTS.md §7, "Untapping in another player's untap step").
  So the untap step, `UntapTargetForEffect`, "untap all" effects,
  Seedborn Muse and the sandbox `tap` action all get the replacement.
  CR 122.1d is about untapping, not about the untap step, so it
  belongs in the primitive and not in the set computation.
- **Order with Decisions 1 and 2.** A permanent held back by a
  restriction or a marker never reaches the primitive, so it doesn't
  "become untapped" and keeps its stun counter. That is the rules
  result: the stun replacement only applies to an untap that would
  happen.
- **Removing a counter is not "putting" one**, so Vorinclex-style
  replacements on putting counters don't see it. The removal calls
  `applyCounterLocked` directly and **stays outside `RepEventCounter`**:
  no replacement window runs for it. Its `EventCounterPlaced` carries
  the new amount and a **`uuid.Nil` placer** (`Actor`), which is how
  every other rules-driven counter removal shows up. This is the same
  rule [ADR 0056](https://github.com/krakenhavoc/cmd_and_ctrl/pull/818)
  Decision 5 states for removals ("Removals are not placements"), so
  its `CounterPlacer` and `CounterFromCombatDamage` fields never apply
  here.
- **Why not a `ReplacementEffect` in the CR 614 pipeline.** The rule
  has no source permanent for `gatherActiveReplacementsLocked` to find,
  and there is no untap replacement event kind. The only other untap
  replacement printed on a card is Freyalise's Winds. Once that card is
  written, CR 616.1 needs the controller to order it against stun, and
  that PR adds `RepEventUntap` and moves stun into the pipeline. Until
  then an event kind with one hard-coded subscriber is speculative. This
  is the same reasoning `untap.go:87-94` used for not writing the
  restriction early.
- **The sandbox `untap_all` button** (`untapAllForLocked`,
  `untap.go:345-366`) keeps ignoring Decisions 1 and 2, because pressing
  a button is not an untap step. It does get stun, because it untaps.
  A player who really wants a stunned permanent upright removes the
  counter by hand first. That is the same control that fixes any other
  wrong counter.

`counter_types.go:37-40`'s comment is corrected in the same PR.

## Decision 4 — The wire says whether a permanent won't untap

```go
// protocol.CardView
NoUntap *NoUntapView `json:"no_untap,omitempty"`

type NoUntapView struct {
    // Static: an UntapStepRestriction applies to this permanent right
    // now (as of this frame; Meekstone's power can change before the step).
    Static bool `json:"static,omitempty"`
    // Next: the players whose next untap step this permanent sits out.
    // A controller-keyed marker is written as the current controller's ID.
    // Eliminated players are left out. Deduplicated.
    Next []string `json:"next,omitempty"`
}
```

- It is filled in by a stamping pass with a game handle, the way
  `stampCombatTargets` (`protocol/view.go:1773`) is. The pass calls an
  exported `game.UntapStepRestrictedLocked(c)`, which is the same
  predicate the step uses, so the wire and the engine can't disagree.
- It is public board state, like `goaded_by`. `next` goes on the
  face-down allowlist (`protocol/face_down_view_test.go`). `static` is
  derived from card identity and is always false for a face-down
  permanent, because a face-down permanent has no abilities.
- `docs/protocol.md` and `client/src/lib/protocol.ts` get the type.
- **How the client shows it** (owner-decided, [question 1](#decided-2026-09-17),
  option (a)):
  - A small "won't untap" badge on a permanent that is **tapped** and
    has `no_untap` that applies to its controller's next untap step
    (`static`, or `next` containing the current controller). An
    untapped Mana Vault or an untapped nonbasic land under Back to
    Basics wears no badge.
  - A line in the hover-zoom footer for **any** permanent with
    `no_untap`, tapped or not: "doesn't untap during its controller's
    untap step" for `static`, and "doesn't untap during Alice's next
    untap step" for each `next` entry.
  - Stun counters need nothing new: they already show as counter pips.
  - **No log line** and no reveal-strip cue. The log records changes to
    the game's outcome and turn structure, not per-object annotations.
- **No new event kind.** No card triggers on "didn't untap". The untap
  step already emits `EventStepBegan`, and a permanent that stays tapped
  is simply one without an `EventUntapCard`. The log-line option that
  would have added one (question 1 (c)) was not chosen.

## Decision 5 — Bots and the auto-tapper

- **Enumerator:** nothing new to enumerate. No prompt, cost or action
  is added.
- **Heuristic** (`aiseat/heuristic/score.go`): a tapped permanent whose
  `no_untap` says it won't untap during its controller's next untap
  step is valued below an ordinary tapped one. That means a new
  `FrozenManaSource` weight below `TappedManaSource` (`:61`, `0.55`),
  and a `FrozenCreature` multiplier applied on top of `TappedCreature`
  (`:48`). The values are tuned in the PR against the existing
  heuristic tests. `producesMana` doesn't change, because it already
  only counts untapped sources.
- **Model tier:** `aiseat/model/prompt.go:306-308` adds a `won't untap`
  flag next to `tapped`.
- **Auto-tapper** (`game/autotap.go`): a source under a restriction,
  or carrying a marker for its controller, is **sorted last, not
  excluded**. `gatherTapSources` (`:157`) marks it, and the sort at
  `:120` uses "untaps normally" as the primary key, with
  restrictiveness as the tie-break. The solver backtracks, so any cost
  that could be paid can still be paid. It just spends a Mana Vault or
  a nonbasic land under Back to Basics only when nothing else pays.
  Exclusion would follow the "no hidden costs" list at `:244-289` more
  literally, but it would stop auto-tap working at all for a player
  whose nonbasic lands are under an opponent's Back to Basics. Mana
  Vault's draw-step damage is a separate trigger, like City of Brass's
  pain, and the same comment already accepts that.
- **Catalog soak** (#601): the first card wave goes into the random
  catalog decks, and one run's report is attached to the card PR.

## Decision 6 — Tangle loses its caveat in sub-PR 1

`tangle.go` gets `DoesntUntapNextUntapStep` over the attacking creatures
at resolution, controller-keyed. `Completeness` becomes `Full`, and the
comment block at `:11-29` is replaced. Tangle is one of sub-PR 1's
reference cards (see [PR split](#pr-split)): under the owner's policy
(2026-09-17) every new seam path ships with at least one real card, so
the engine PR carries one card per path. `docs/engine-seams.md`'s
"Doesn't untap during your untap step" row (`:118`) moves to Closed
when sub-PR 1 lands.

## Decision 7 — Docs that change with the code

- `untap.go`'s header (`:75-94`) stops listing the restriction as not
  written. It gets a section on the set's two directions, the marker,
  and stun. The Icy Manipulator mention goes.
- AGENTS.md §7 "Untapping in another player's untap step (#74)" gains
  a paragraph on `Spec.UntapStepRestrictions`, the marker helpers and
  the stun rule.
- `docs/protocol.md` covers `no_untap`.

## Decision 8 — Out of scope, stated

- **Choosing which permanents untap** (Winter Orb, Static Orb, Winter
  Moon, Damping Field, Smoke, Stoic Angel, Mungha Wurm, Dovin Baan's
  emblem; and "you may choose not to untap this during your untap
  step": Rust Tick, Amber Prison and 43 more). CR 502.3's "determines
  which permanents … will untap" becomes a real decision. Today the
  untap step runs inside one write lock with no priority and no prompt,
  and every step-entry path recurses through it (`game.go:1008-1009`).
  A choose-N prompt means pausing a turn-based action halfway, resuming
  it through a `finishStepEntryLocked`-style split (as #710 did for the
  replacement window), counting it in `ContinuationCensus`, and giving
  bots and the enumerator a new prompt. That is a follow-up with its
  own ADR. The restriction and marker here don't depend on it, and it
  will read them: a permanent that can't untap is not one of the N
  choices. **Winter Orb, Static Orb and Winter Moon wait for it**
  (owner-decided, [question 3](#decided-2026-09-17)). The ADR is filed
  as [#826](https://github.com/krakenhavoc/cmd_and_ctrl/issues/826) and
  is not scheduled.
- **Duration-bound restrictions** ("for as long as you control this
  creature": Dungeon Geists, Wall of Stolen Identity; "for as long as
  this remains tapped": Amber Prison, Rust Tick). These are an
  `UntapStepRestriction` on the source whose predicate reads a
  remembered target, and the engine has no per-source linked-object
  field yet. Once one exists, they need no further change to this
  design.
- **Exert** (CR 701.43; Combat Celebrant, Pride Sovereign, Arena of
  Glory, 36 cards). It will add a player-keyed marker, but "exert as it
  attacks" is an optional cost at declaration (CR 508.1g), and the
  ability costs are a cost component. Both belong to their own issue.
- **Mana abilities that create a marker** (Thalakos Lowlands' colored
  ability). This needs a mana-ability rider flag so the auto-tapper
  plans around it, like the painlands. It is left for the card that
  needs it.
- **Telekinesis'** "next two untap steps", and **Freyalise's Winds**
  (Decision 3).

## Consequences

- The untap step answers CR 502.3 in both directions from one function.
  A card author writes a predicate, not a trigger or a caveat.
- About 360 Commander-legal cards print a shape this ADR covers (303
  print "doesn't / don't / won't untap during", 82 mention stun
  counters, less the 23 duration-bound ones). Fewer will become
  writable, because the audit found about 30% of the sampled cards
  blocked by something else as well.
- **Behaviour players will notice:**
  - A stunned permanent clicked "untap" in the sandbox loses a stun
    counter and stays tapped.
  - The auto-tapper spends Mana Vault, the Monoliths and restricted
    lands last.
  - A Seedborn Muse whose ability loss expired at cleanup untaps things
    in the next untap step, as it always should have.
- `performUntapStepLocked` gains one layer recompute (a no-op on the
  fast path) and a second battlefield walk, only when a restriction is
  declared or a marker exists. A board with neither pays one nil check
  per permanent.
- `Card` grows by 24 bytes.
- Snapshot files gain an optional key. Old files restore unchanged.
- The wire gains `CardView.no_untap`. Old clients ignore it.
- Players see a "won't untap" badge only on tapped permanents that will
  stay tapped, and the reason in the hover footer. The log doesn't
  change.

## Alternatives considered

- **A `Restriction` bit in layer 6** (`restrictions.go`). Rejected in
  Decision 1: it reads P/T before layer 7, and it splits CR 502.3's one
  question across two mechanisms.
- **Resolving "its controller's" to a player ID when the effect is
  created.** It is simpler, but it is wrong under a control change
  (the Wall of Frost, Frost Titan and Tamiyo rulings).
- **A turn-scoped registry (`TurnScopedStatics`-style) with an
  "expires at X's next untap step" predicate.** Rejected: closures don't
  go into snapshots, so every marker would block restore points until it
  expired, and the registry would have to find its permanent again
  after zone changes, which a field on the card gets for free.
- **Two fields: a `bool` for controller-keyed plus `[]uuid.UUID` for
  player-keyed.** It reads well, but a ninth bool in the block costs 8
  more bytes on every `Card`, for a state nearly no card is ever in.
  One slice of a small struct costs 24 bytes and can grow a count
  (Telekinesis) without a second migration.
- **Stun as a `ReplacementEffect` with a new `RepEventUntap` kind now.**
  Deferred to Freyalise's Winds (Decision 3).
- **Excluding frozen sources from auto-tap.** Rejected in Decision 5:
  it would stop auto-tap working under Back to Basics.
- **Winter Orb's choose-N in this ADR.** Rejected in Decision 8: it is
  a pausable turn-based action with its own persistence and bot
  surface, and nothing here needs it.

## PR split

**Bug fix first, separately:** recompute layers at the start of the
untap step (the stale-read bug in Decision 1), with its regression test.
Sub-PR 1 assumes that line exists and adds it if it doesn't.

**Sub-PR 1 — engine, with one reference card per path.**
- `UntapStepRestriction`, `CatalogUntapStepRestrictions`,
  `activeUntapStepRestrictionsLocked`, `UntapStepRestrictedLocked`.
  `Spec.UntapStepRestrictions`, the `CardDef` field and the `buildDef`
  line.
- `Card.NextUntapSkips` / `UntapSkip`, `SkipNextUntapForEffect`, marker
  use-up in `performUntapStepLocked`, the zone-exit and exile-entry
  clears, `cloneCard`, `cardSnapshot`, and the drift-test row.
- Stun in `untapPermanentLocked`.
- Auto-tapper ordering.
- `cards/effects`: `DoesntUntapNextUntapStep`, `TapAndFreeze`, and the
  restriction constructors `doesntUntapDuringYourUntapStep`,
  `enchantedDoesntUntap` (over `AttachedToSource`) and
  `doesntUntapDuringTheirControllersUntapSteps(match)`, next to
  `untap_step.go`.
- **Reference cards**, so no path lands without a real card (owner
  policy, 2026-09-17): Basalt Monolith (static self, and the
  auto-tapper's ordering), Claustrophobia (static attached), Meekstone
  (static filtered, read after layer 7), Tangle's caveat removal
  (controller-keyed marker, Decision 6), Frost Breath (`TapAndFreeze`),
  Sleep (player-keyed marker) and Alchemax Slayer-Bots (stun).
- Docs: `untap.go` header, the `counter_types.go` comment, AGENTS.md §7,
  and moving the `docs/engine-seams.md` row to Closed.
- Checks: `go test ./internal/game/... ./internal/cards/... ./internal/legal/...`,
  then `go test ./...` and `make lint`.

**Sub-PR 2 — wire, bots, client.**
- `CardView.no_untap` stamping, `docs/protocol.md`, `protocol.ts`, and
  the face-down allowlist.
- Heuristic weights and the model-tier flag.
- The client rendering decided in owner question 1 (Decision 4): the
  "won't untap" badge on tapped permanents only, and the hover-footer
  line. No log line.
- Checks: `go test ./internal/protocol/... ./internal/aiseat/...`,
  `npm run test`, `npm run check`.

**Card PRs.** The rest of the first wave decided in owner question 2,
option (b): Mana Vault, Grim Monolith, Goblin Sharpshooter, Traxos,
Back to Basics, Intruder Alarm, Wall of Frost, Kefnet's Monument,
Dreamdew Entrancer and Cryogen Relic. Each card follows AGENTS.md §7: `Spec`
slots, completeness declared, caveats weaker than printed and never
stronger, and oracle text checked against the dump. The catalog soak
runs with the wave in the random decks.

## Test plan

Engine (`internal/game`, stubbing `CatalogUntapStepRestrictions` the way
`untap_test.go:16` stubs permissions):

1. **Self restriction.** A tapped restricted permanent stays tapped in
   its controller's untap step and gets no `EventUntapCard`. An
   unrestricted neighbour untaps.
2. **Scope under Seedborn Muse.** The same restricted permanent, with a
   permission from its controller's Seedborn-style source, **untaps**
   during another player's untap step.
3. **Ability removal.** A restriction source that has lost all
   abilities (a `RemovesAbilities` static) restricts nothing, including
   when the removal expired at the previous cleanup. This last case is
   the stale-layers regression.
4. **Condition read at the step.** A power-≥3 predicate against a 2/2
   pumped to 3/3 by an anthem is restricted. The same predicate against
   a 3/3 shrunk to 2/2 is not.
5. **Controller-keyed marker.** Tapped, the marker holds the permanent
   for one untap step of its controller, and it untaps in the following
   one.
6. **Used up when untapped.** An untapped permanent with a marker loses
   the marker at its controller's untap step. If it is tapped
   afterwards, it untaps normally next time.
7. **Control change.** A marker placed while seat A controls the
   permanent, then control moves to seat B. It is used up in B's next
   untap step, and A's untap step leaves it alone.
8. **Player-keyed marker.** Keyed on A, with the permanent controlled by
   B. It is used up in A's untap step, and it stops a B-controlled
   Seedborn-style permission from untapping the permanent during that
   step.
9. **Skipped step.** A Stasis-style cancel on `RepEventStepTransition`
   leaves the marker in place, and it is used up at the first untap
   step that happens. The same holds across a #710 CR 616 ordering pause
   and its resume.
10. **Zone change.** The marker is gone after battlefield → hand →
    battlefield, and after an exile-and-return.
11. **Copy.** A clone of a marked permanent (`CopiableValuesOf` and a
    token copy) has no marker.
12. **Added once.** Two identical markers are used up in one untap step.
13. **Undo and snapshot.** `Clone` → change the marker → `RestoreFrom`
    gives back the original. Snapshot capture → encode → decode →
    restore round-trips markers. A file with no `nextUntapSkips` key
    restores with none. The drift test classifies the field.
14. **Layout guard.** `TestCardHasNoInteriorPadding` passes.
15. **Stun.** A tapped permanent with 2 stun counters: an untap removes
    one and it stays tapped with no `EventUntapCard`. The second untap
    removes the last. The third untaps. An untapped permanent with a
    stun counter keeps it after an untap attempt. A restricted or marked
    stunned permanent keeps its counter through its untap step. The
    sandbox `tap` action and `untap_all` both apply stun, and
    `untap_all` ignores restrictions and markers.
16. **Auto-tapper.** Given a frozen source and an ordinary one that can
    each pay a cost, the plan picks the ordinary one. A cost only the
    frozen source can complete still gets a plan.

Wire and bots:

17. `protocol`: `no_untap.static` for a restricted permanent,
    `no_untap.next` resolves a controller-keyed marker to the current
    controller, an eliminated player is left out, and the face-down
    allowlist includes `no_untap`.
18. `aiseat/heuristic`: a tapped frozen mana source and creature score
    below their ordinary tapped equivalents.
19. **Catalog soak** with the first wave: no stalls, and the report
    lists the wave's cards as exercised.

Cards (sub-PR 1's reference cards and the card PRs): each card gets its
behaviour test. Tangle's test changes from "caveat" to "attacking
creatures stay tapped through their controller's next untap step, and
blockers do not". Alchemax Slayer-Bots' test checks that the stun
counter's removal emits `EventCounterPlaced` with a Nil `Actor` and runs
no counter replacement (a Doubling Season-style stub on the battlefield
sees nothing).

Client (sub-PR 2): the badge predicate (tapped, and `no_untap` applies
to the current controller's next untap step) and the footer text are a
pure helper with vitest coverage. Rendering is checked by hand until
#689.

## Card first wave

Owner-decided ([question 2](#decided-2026-09-17), option (b)): every card
in the table ships, the three stun cards included. The oracle text below
was checked against the Scryfall dump on
2026-09-17. "Other needs" is what else each card uses, and every such
piece exists on develop today unless marked.

| card | oracle text (abridged only where noted) | shape | other needs |
|---|---|---|---|
| Tangle | "Prevent all combat damage that would be dealt this turn. Each attacking creature doesn't untap during its controller's next untap step." | controller marker | already in the catalog (caveat removal) |
| Mana Vault | "This artifact doesn't untap during your untap step. At the beginning of your upkeep, you may pay {4}. If you do, untap this artifact. At the beginning of your draw step, if this artifact is tapped, it deals 1 damage to you. {T}: Add {C}{C}{C}." | static self | upkeep `MayPay`, draw-step trigger with an intervening if |
| Basalt Monolith | "This artifact doesn't untap during your untap step. {T}: Add {C}{C}{C}. {3}: Untap this artifact." | static self | — |
| Grim Monolith | "This artifact doesn't untap during your untap step. {T}: Add {C}{C}{C}. {4}: Untap this artifact." | static self | — |
| Goblin Sharpshooter | "This creature doesn't untap during your untap step. Whenever a creature dies, untap this creature. {T}: This creature deals 1 damage to any target." | static self | dies trigger |
| Traxos, Scourge of Kroog | "Trample. Traxos enters tapped and doesn't untap during your untap step. Whenever you cast a historic spell, untap Traxos." | static self | enters tapped, historic cast trigger |
| Meekstone | "Creatures with power 3 or greater don't untap during their controllers' untap steps." | static filtered | — |
| Back to Basics | "Nonbasic lands don't untap during their controllers' untap steps." | static filtered | — |
| Intruder Alarm | "Creatures don't untap during their controllers' untap steps. Whenever a creature enters, untap all creatures." | static filtered | ETB-any trigger |
| Claustrophobia | "Enchant creature. When this Aura enters, tap enchanted creature. Enchanted creature doesn't untap during its controller's untap step." | static attached | Aura |
| Wall of Frost | "Defender. Whenever this creature blocks a creature, that creature doesn't untap during its controller's next untap step." | controller marker | block trigger |
| Kefnet's Monument | "Blue creature spells you cast cost {1} less to cast. Whenever you cast a creature spell, target creature an opponent controls doesn't untap during its controller's next untap step." | controller marker | cost modifier (ADR 0048), cast trigger |
| Frost Breath | "Tap up to two target creatures. Those creatures don't untap during their controller's next untap step." | controller marker | — |
| Sleep | "Tap all creatures target player controls. Those creatures don't untap during that player's next untap step." | player marker | — |
| Alchemax Slayer-Bots | "When this creature enters, tap target creature an opponent controls and put a stun counter on it." | stun | — |
| Dreamdew Entrancer | "Reach. When this creature enters, tap up to one target creature and put three stun counters on it. If you control that creature, draw two cards." | stun | — |
| Cryogen Relic | "When this artifact enters or leaves the battlefield, draw a card. {1}{U}, Sacrifice this artifact: Put a stun counter on up to one target tapped creature." | stun | — |

Tamiyo, the Moon Sage (emblem, #623), Hands of Binding (cipher) and Kiora
Bests the Sea God are not in the wave. Winter Orb, Static Orb and Winter
Moon wait for the untap-step choice ADR,
[#826](https://github.com/krakenhavoc/cmd_and_ctrl/issues/826).

## Decided (2026-09-17)

Answered by the owner on 2026-09-17. The options are kept, the chosen one
is marked **(chosen)**, and the recommendation text is kept for the
record.

1. **How does the board show that a permanent won't untap?**
   - (a) **(chosen)** A small badge on a **tapped** permanent that won't untap during
     its controller's next untap step (static or marker), and a line in
     the hover-zoom footer for any permanent with `no_untap` ("doesn't
     untap during Alice's next untap step" / "doesn't untap during its
     controller's untap step"). No log line.
   - (b) The badge on every permanent with `no_untap`, tapped or not.
     That means every untapped Mana Vault and every nonbasic land under
     Back to Basics wears it all the time.
   - (c) As (a), plus a game-log line when a permanent stays tapped
     through an untap step. This adds a log-only event kind.
   - **Recommendation: (a).** The badge appears exactly when something
     surprising is about to happen, and the footer covers planning
     ("if I attack with this, it stays tapped"). Stun counters already
     show as counter pips.
   - **Decision: (a).** The badge on tapped permanents only, plus the
     hover footer line, and no log line. Applied in Decision 4 and
     sub-PR 2.
2. **Which cards go in the first wave?**
   - (a) Tangle's caveat removal plus the 14 static and marker cards in
     the table.
   - (b) **(chosen)** (a) plus the three stun-counter cards.
   - (c) A minimal wave: Tangle, Mana Vault, Basalt Monolith, Grim
     Monolith, Sleep, Frost Breath (one card per shape), with the rest
     as batch fill.
   - **Recommendation: (b).** Stun ships in the same engine PR, and real
     cards put it through the soak. The three candidates need nothing
     else.
   - **Decision: (b).** Every card in the first-wave table ships,
     including Alchemax Slayer-Bots, Dreamdew Entrancer and Cryogen
     Relic. One card per path goes in sub-PR 1 (PR split).
3. **Do the Winter Orb cards wait for the untap-step choice ADR, or
   ship early with an automatic pick?**
   - (a) **(chosen)** They wait. File "untap-step choices (choose N, may choose not
     to untap)" as its own ADR issue and don't schedule it yet.
   - (b) Ship Winter Orb, Static Orb and Winter Moon now, with the engine
     choosing which N untap (for example, the lands that produce the
     colors the player's hand needs) and a declared caveat, like
     proliferate's auto-pick.
   - (c) As (a), but schedule the choice ADR right after this one.
   - **Recommendation: (a).** Under a Winter Orb lock, which land you
     untap is the decision the card is about, every turn, for everyone.
     A visible automatic pick in that spot is worse than a card that
     honestly isn't there yet. The shape is rare in Commander, so the
     ADR doesn't need scheduling now.
   - **Decision: (a).** Winter Orb, Static Orb and Winter Moon wait. The
     untap-step choice ADR is filed as
     [#826](https://github.com/krakenhavoc/cmd_and_ctrl/issues/826),
     unscheduled.


## Implementation checkpoint — 2026-09-17

The implementation for #751 combines the engine, board projection, bot support
and first-wave cards in one feature PR. Static restrictions use the catalog's
ability key, next-step markers survive undo and snapshots, and the common untap
primitive replaces attempts with stun-counter removal. The auto-tapper reserves
restricted sources until needed; the board and bot prompt explain held
permanents using the same public projection.

All 17 cards in the table are represented: 16 new entries and Tangle's completed
untap clause. Mana Vault and Claustrophobia explicitly declare a remaining
post-departure last-known-information limitation for their draw-step and entry
abilities, respectively. Their untap restrictions are implemented. Winter Orb
and the other choose-N cards still wait on #826; exert's action/cost and
source-linked durations remain outside this implementation.
