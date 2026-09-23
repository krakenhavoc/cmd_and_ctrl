# ADR 0084 — Phasing: a permanent that is treated as though it does not exist (CR 702.26)

**Status:** Accepted · 2026-09-23 · S46 — Permanents that change what they are
**Issues:** [#1199](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1199) (phasing),
under tracker [#889](https://github.com/krakenhavoc/cmd_and_ctrl/issues/889)
(permanents that change what they are). Filed out of
[#1197](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1197) (player protection),
whose Teferi's Protection shipped with phasing as its first caveat.
**Numbering:** swept on 2026-09-23 with `git ls-tree origin/develop docs/decisions/`
and with `git ls-tree -r --name-only <ref> docs/decisions` over all 405 remote heads.
The highest number present anywhere was **0083** (`0083-token-abilities.md`) and
nothing held an `0084-*` file. 0005, 0024, 0029 and 0030 stay permanently unused per
AGENTS.md §4.

**Related:** [ADR 0079](0079-transforming-a-permanent.md) (the sibling verb — a
permanent that changes without changing zones, and the argument for why that is not
a `MoveCard`), [ADR 0036](0036-attachments.md) (`Card.AttachedTo`, one direction
only, and why there is no reverse index), [ADR 0070](0070-untap-step-choices.md)
(the untap step's one exit and its pause), [ADR 0063](0063-durations-and-control.md)
(the `Duration` model and the one sweep), [ADR 0039](0039-layer-4-authoritative.md)
(why a change to what is on the battlefield has to invalidate the layer cache),
[ADR 0041](0041-game-persistence.md) (the snapshot schema and its version-skew
policy), [ADR 0010](0010-card-effect-catalog.md) (the `Spec` slots the proof cards
declare through)

**A note on rule numbers.** Phasing is **CR 702.26** in the pinned August 7, 2026
edition. 702.24 is Cumulative Upkeep and 702.25 is Flanking, so the three existing
`CR 702.25f` citations in the tree — `game/activation_tally.go`,
`game/exhaust_test.go`, `cards/effects/teferis_protection.go`, and the table row in
[ADR 0020](0020-activated-abilities.md) — point at a rule that does not exist and
are corrected here to **CR 702.26d**, which is the sentence they were reaching for
("the phasing event doesn't actually cause a permanent to change zones or
control"). The untap step's phasing turn-based action is **CR 502.1**, the *first*
of the step's three, ahead of the day/night check (502.2) and the untap
determination (502.3) that ADR 0070 already cites.

---

## Context

Phasing is not modelled anywhere in this engine. The only mention of it in the tree
before this change was a comment in `game/activation_tally.go` saying so, and one
catalog caveat on Teferi's Protection saying so again. Both are already right about
the constraint the work starts from:

> a PHASE-OUT does not [bump the object epoch]. Phasing is not a zone change […]
> when it is [modelled], the rule it must respect to keep this true is "phasing does
> not bump `ObjectEpoch`".

CR 702.26d is the sentence behind that: *"The phasing event doesn't actually cause a
permanent to change zones or control, even though it's treated as though it's not on
the battlefield and not under its controller's control while it's phased out.
Zone-change triggers don't trigger when a permanent phases in or out. Tokens continue
to exist on the battlefield while phased out. Counters and stickers remain on a
permanent while it's phased out."*

So phasing is the sibling of ADR 0079's transform, not of a `MoveCard`: the same
object, with a **status** (CR 110.5 — tapped/untapped, flipped/unflipped, face
up/face down, **phased in/phased out**) rather than a location. Everything keyed on
the object survives it: `InstanceID`, `ObjectEpoch` and every `{instance, epoch}`
registry, `Counters`, `DamageMarked`, `Tapped`, `AttachedTo` / `AttachedAt`, the
CR 613.7 timestamp, `SummonedThisTurn`, `Owner` and `Controller`.

The other half of the rule is the hard half:

> **CR 702.26b** If a permanent phases out, its status changes to "phased out."
> Except for rules and effects that specifically mention phased-out permanents, a
> phased-out permanent is treated as though it does not exist. It can't affect or be
> affected by anything else in the game. A permanent that phases out is removed from
> combat.

"It can't affect or be affected by anything else in the game" is a statement about
**every** read of the battlefield, and this engine has a lot of them. Counted on
`origin/develop` at the time of writing:

| surface | battlefield walks |
|---|---|
| `internal/game` (open-coded `for … range g.Battlefield.Cards`) | 96 |
| `internal/legal` | 11 |
| `internal/aiseat` (over `protocol.GameView.Battlefield`) | 16 |
| `internal/protocol` | 2 |
| `internal/cards/effects` through `BattlefieldCardsForEffect` | 151 |
| `internal/cards/effects` open-coded | 38 |
| single-card lookups (`findBattlefieldCard`, `findCardOnBattlefield`, `findCardZoneLocked`, `battlefieldCardLocked`, `findBattlefieldCardLocked`, `controllerOfBattlefieldCardLocked`, `legal.findBattlefield`, `legal.cardByID`) | 8 helpers, 144 callers |

`game.Zone` has no predicate, no filter hook and no "visible cards" view; every
consumer reaches into `.Cards` directly. Issue #1199 proposed "one predicate asked
at each battlefield walk rather than a filter applied in one place". That is the
right *rule* and the wrong *mechanism* at this scale: 125 engine-side edit sites, of
which a missed one is a silent rules bug (a phased-out creature that still blocks, a
Wrath that still kills it, a Lord that still pumps from it), with no test that can
notice the omission and no way for a reviewer to prove the list is complete.

---

## Decisions

### 1. A phased-out permanent leaves the battlefield **slice**, not the battlefield

`Game.PhasedOut *Zone` holds phased-out permanents. Phasing moves the `Card` value
between `g.Battlefield.Cards` and `g.PhasedOut.Cards` and does nothing else.

CR 702.26b then holds **by construction** for all 125 engine walks, all 189 catalog
walks, all 11 enumerator walks, all 16 bot walks, every one of the 144 single-card
lookups, the layer pass, the trigger harvester, the SBA sweep, combat, the
autotapper and the view builder — none of them can see a card that is not in the
slice they read. There is nothing to add at a call site and therefore nothing to
forget at the next one. This is the same argument `game/leave_game.go`'s
`removeObjectsOwnedByLocked` already makes for CR 800.4a ("leaving the game is not a
zone change"), and the same one ADR 0036 §12 makes for not keeping a reverse
attachment index: *a second thing that can disagree with the first*.

**It is still not a zone change**, and the code is written so that this is checkable
rather than asserted. `phaseOutLocked` / `phaseInLocked` call **neither**
`MoveCard`, **nor** `battlefieldExitLocked`, **nor** `pruneChoicesAfterArrivalLocked`,
and emit **no** `EventZoneMove` / `EventLTB` / `EventETB`. `ObjectEpoch` is not
touched (CR 702.26d, and the promise `activation_tally.go` has been making since
S38). `MoveCard`'s battlefield-exit block — which clears `Tapped`, `Counters`,
`DamageMarked`, `AttachedTo`, `AttachedAt`, `EnteredBattlefieldAt`,
`SummonedThisTurn`, `RegenerationShields`, `BattleX/Y`, the combat targets and the
face-down state — is exactly the list CR 702.26d says to preserve, which is the
clearest possible statement of why phasing must not travel through it.

`ZonePhasedOut` is a new `ZoneKind`, and it is a **label, not a zone in the CR 400
sense**. `findCardZoneLocked` deliberately does not look there, so
`FindCardZoneForEffect` on a phased-out permanent answers `nil` — which is the right
answer to "where is this object", given that as far as the game is concerned there
is no object. Nothing routes a card there through `MoveCard`, `zoneFromRefLocked` or
`executeZoneRouteLocked`; the two functions in `phasing.go` are the only writers.
What the kind buys is that `cloneZone`, `snapshotZone` / `restoreZone` and
`viewOfZone` all work on it unchanged.

**Rejected: a `Card.PhasedOut bool` consulted at each walk.** It is what the issue
asked for and it is what the client-side precedent (face-down permanents stay in
every walk and are redacted on projection) suggests. It loses on the number above,
and on one structural fact the face-down precedent does not share: a face-down
permanent *is* affected by things (it can be targeted, blocked, destroyed and
counted), so it genuinely belongs in the walks. A phased-out one is affected by
nothing.

**Rejected: a `Zone`-level predicate or an iterator.** It would mean rewriting 125
open-coded `for … range` loops into a closure-taking helper to get the same
guarantee this gets for free, and every loop that was missed would still be wrong.

### 2. Three fields on `Card`, and each answers a question the rules ask

```go
PhasedOutBy       uuid.UUID  // CR 702.26a — under whose control it phased out
PhaseInLockedBy   uuid.UUID  // "phases out until <source> leaves the battlefield"
PhasedOutIndirect bool       // CR 702.26g — phased out with its host
```

- **`PhasedOutBy`** is not `Controller` and is not derivable from it. CR 702.26a
  phases in *"all phased-out permanents that the active player controlled **when
  they phased out**"*, and CR 702.26f allows a control-changing continuous effect to
  expire while the permanent is phased out. An Act of Treason creature phased out
  under **your** control phases in during **your** untap step even though
  `Controller` has reverted by then. One `uuid.UUID`, written once by the phase-out
  and read once by the untap step.

- **`PhasedOutIndirect`** is CR 702.26g's *"An Aura, Equipment, or Fortification that
  phased out indirectly won't phase in by itself, but instead phases in along with
  the permanent it's attached to"*, and CR 702.26h's *"If an object would
  simultaneously phase out directly and indirectly, it just phases out indirectly"* —
  which is why the flag is decided once, for the whole batch, by the expansion in
  Decision 3 rather than per card by whoever asked.

- **`PhaseInLockedBy`** is the "phases out until …" family — Oubliette, Out of Time.
  A permanent carrying one is **off the untap step's clock entirely**: CR 502.1's
  phase-in half skips it, and it phases back in the *moment* the named object stops
  being on the battlefield, in the sweep at the top of `stateBasedActionsLocked`.
  That is what the cards do rather than a simplification — Out of Time's whole point
  is that every creature returns at once when the last time counter comes off,
  mid-turn, and Oubliette's "Tap that creature as it phases in this way" only means
  anything if the return is not an untap step (where CR 502.3 would untap it again
  a moment later).

  A bare `uuid` rather than an ADR 0063 `Duration`: presence is the whole condition,
  `ForAsLongAsOnBattlefieldDuration` is exactly it, and a ten-field `Duration` on
  every `Card` in the game would be a large price for one printed phrase. The sweep
  runs where it does — first in the SBA pass, ahead of the layer recompute and
  ahead of `attachmentSBALocked` — so an Aura that phases in onto a host that died
  while it was away is graveyarded in the same settling (CR 702.26i, CR 704.5m).
  It is **not** a state-based action, and CR 704.3's "SBAs ignore phased-out
  permanents" needs no code: a phased-out permanent is not in the slice the SBA
  pass walks.

`Card.TapOnPhaseIn bool` is the fourth and smallest: Oubliette's *"Tap that creature
as it phases in this way"*, a rider on one phase-in, cleared as it fires. It is a
bool on the card rather than a delayed trigger because CR 702.26a's phase-in is a
turn-based action that uses no stack, so there is nothing for a trigger to sit on.

All four live in the flat printed-adjacent block or the bool block per `card.go`'s
padding rule (`TestCardHasNoInteriorPadding`), are cleared by `MoveCard`'s
battlefield-exit block (a permanent that really leaves is a new object, CR 400.7),
are carried by `cloneCard` and by `cardSnapshot`, and are **not** copiable values:
CR 707.2 does not copy status.

### 3. One phase-out primitive, and it takes the attachment subtree

`(*Game).phaseOutLocked(ids []uuid.UUID, opts phaseOutOptions)` is the only way a
permanent phases out. It does four things in the order CR 702.26 gives them:

1. **Expand.** Each named permanent drags everything attached to it, transitively
   (CR 702.26g): an Equipment on the creature, an Aura on that Equipment. ADR 0036
   keeps one direction only and no reverse index, so the expansion is a worklist over
   `g.Battlefield.Cards` per level — the idiomatic shape here, and cycle-safe because
   the visited set is the result set. A permanent reached both directly and
   indirectly is marked **indirect** (CR 702.26h).
2. **Remove from combat** (CR 506.4, named in CR 702.26b's own text): clear
   `AttackingTarget` / `BlockingTarget` and drop the card from `announcedAttacks`,
   `announcedBlocks` and `blockedAttackers`. Deliberately **not**
   `forgetPerObjectTurnStateLocked`, which would also drop
   `LoyaltyActivatedThisTurn` — CR 702.26d says an effect that checks a phased-in
   permanent's history is not to treat the phasing as having changed anything, and
   CR 606.3's once-per-turn is exactly such a history.
3. **Move**, in one batch (CR 702.26a: "This all happens simultaneously"), choosing
   the whole set before moving any of it for the reason `untapStepSetLocked` chooses
   its set up front and `#529`'s mill chooses its cards up front: the moves emit, and
   emitting runs listeners, which is not a walk to hold `*Card` pointers across.
4. **Emit** `EventPhaseOut` per permanent and nil each card's `effective` cache.

Card-facing entry point: `(*Game).PhaseOutForEffect(source uuid.UUID, ids ...uuid.UUID) error`
in `effect_api.go`, plus `PhaseOutUntilLeavesForEffect` for the Oubliette shape.

### 4. CR 502.1 is one statement at the top of the untap step's turn-based action

`performPhasingLocked(activePlayer)` runs as the first statement of
`performUntapStepLocked`, ahead of the CR 302.6 summoning-sickness clear and well
ahead of `untapStepSetLocked` — a permanent phasing in has to be in the untap set,
and one phasing out has to be out of it.

It is inside `performUntapStepLocked` rather than in the `case StepUntap:` arm for
the reason ADR 0070 gives about `consumeUntapSkipsLocked`: the step can **pause** on
the CR 502.3 determination, and its continuation (`finishUntapStepLocked`) only
untaps and exits. Anything that runs before the pause runs exactly once. Anything
moved into the arm would have to be re-reasoned about the day a second pause is
added.

Both halves, chosen before either is applied:

- **phase in** — every card in `g.PhasedOut` with `PhasedOutBy == activePlayer`,
  `!PhasedOutIndirect` and no live `PhaseInLockedBy`, plus its indirectly-phased-out
  attachments, transitively.
- **phase out** — every phased-in permanent the active player controls with the
  `phasing` keyword (CR 702.26a's static-ability half).

CR 702.26m ("if an effect causes a player to skip their untap step, the phasing
event simply doesn't occur that turn") falls out: `performUntapStepLocked` is not
called for a skipped step, so neither is this.

### 5. `phasing` joins `canonicalKeywords`

`KeywordPhasing = "phasing"`, in the same change that teaches the engine to honour
it — which is the closedness rule `keywords.go` states for the table. Its consumer
is the phase-out half of Decision 4, read off `Effective().Abilities` so that a
granted phasing (Shimmer's "each land of the chosen type has phasing", Vanishing's
aura) works the same as a printed one, and so that a permanent that has lost all
abilities stops phasing.

The deck importer stamps it from Scryfall's `keywords` array like every other
canonical token, so ~70 printed phasing cards — Tolarian Serpent, Taniwha, Teferi's
Imp, Sandbar Crocodile, Cloak of Invisibility and the rest of the Mirage/Visions
cycle — become correct with **no catalog entry at all**. That is the whole point of
the keyword being in the table rather than in a `Spec`.

### 6. The wire: a shared `phased_out` zone, and the bot is correct by construction

`GameView.phased_out` is a `ZoneView` beside `battlefield`, `stack` and `exile` —
shared, owner-less, public. Phased-out permanents are **not** merged into
`battlefield`, and that is the load-bearing half of the decision:

- `stampNoUntap` and `stampCombatTargets` index `view.Battlefield.Cards[i]`
  positionally against `g.Battlefield.Cards[i]`. Merging would have desynchronised
  every positional stamp.
- `internal/aiseat` never touches `game.Game`; its 16 battlefield walks are over
  `protocol.GameView.Battlefield`. A merged view with a `phased_out` flag would have
  meant teaching all 16 — and a heuristic that counts a phased-out creature as a
  blocker is the same silent class of bug Decision 1 exists to avoid. A separate
  zone the bot does not read means the bot is right for free.

The client renders it as a per-seat **phased pile** in `PileBar`, beside graveyard
and exile, opening the existing `ZoneBrowserModal` — the house pattern for "cards
that are somewhere other than the board", which is what a phased-out permanent looks
like to a player. Each card carries `phased_out: true` so a future in-place render
has the bit without another wire change.

### 7. Snapshot, clone and undo

`PhasedOut` is `carried` in `snapshot_drift_test.go` and round-tripped by
`snapshot_carried_test.go`: it is a `*Zone` of plain `Card`s and reuses
`snapshotZone` / `restoreZone` verbatim, so it costs no new census counter and
nothing in it is a continuation.

`SnapshotSchemaVersion` goes **5 → 6**. Not because a v5 file decodes wrongly — an
absent `phasedOut` key zero-values to "nothing is phased out", which is right — but
for the direction the emblem bump (v2) and the token-key bump (v5) exist for: a v5
binary handed a v6 file would drop the key it does not know and restore a game with
somebody's whole board **silently gone**. There is no per-field way to refuse that,
so the version is it.

`cloneLocked` / `RestoreFrom` carry it the same way, which is what makes an undo
across a phase-out and a phase-in exact.

### 8. What this deliberately does not build

- **CR 702.26e / 702.26f, the continuous-effect corners.** A "gain control until end
  of turn" or a "+3/+3 until end of turn" whose object phases out keeps its
  `ScopedStatic` registration, so it applies again if the permanent phases back in
  inside the duration. CR 702.26e excludes a phased-out permanent from the *set* of
  affected objects and CR 702.26f ends "for as long as" durations that track it.
  What the slice move does give correctly is the common case: while the permanent is
  out of the battlefield slice the layer pass cannot see it, so it contributes
  nothing and receives nothing. The corner is the *return*, and no catalogued card
  reaches it. Noted at `game/phasing.go`.
- **CR 702.26n**, a phased-out permanent whose controller has left the game.
  CR 702.26k's ordinary case is built — `leave_game.go` sweeps `PhasedOut` alongside
  the battlefield — but the "phases in during the next untap step after that
  player's next turn would have begun" clause is not, because that player has no
  next turn in this engine.
- **`Game.TurnScopedReplacements` durations and "your life total can't change"**
  ([#1200](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1200), tracker
  [#882](https://github.com/krakenhavoc/cmd_and_ctrl/issues/882)). See below.

---

## What this leaves in place for #1200

#1200 is the other half of Teferi's Protection and it is **not** on this design. It
is left out deliberately, and the shape it should take is clearer after this work
rather than despite it:

- The registry #1200 names, `Game.TurnScopedReplacements`, holds `ReplacementEffect`
  values with **closures** (`AppliesTo` / `Replace`). It is counted in
  `ContinuationCensus.TurnScopedReplacements` and marked `dropped` in the drift test,
  which means a "until your next turn, your life total can't change" stored there
  would be the one thing on the board that stops the table having a restore point —
  for the whole turn cycle the card is famous for.
- `Player.Statics` (ADR 0072's 2026-09-22 amendment, `game/player_statics.go`) is
  the registry that already solved exactly this problem for the *other* half of the
  same card: plain data, no closure, a CR 611.2 `Duration` swept through
  `durationExpiredLocked`, cloned by value, mirrored into the snapshot, classified
  `carried`. "Your life total can't change" is a **player-scoped** rule
  (CR 119.7 "can't gain life" and CR 119.8 "can't lose life", applied in the CR 614
  window), so it is the same kind of thing as "you have protection from everything"
  and wants the same home.
- Its consumer already exists and is already one function:
  `ChangePlayerLifeThenForEffect` routes **every** life change in the engine through
  the CR 614 window (#482), and its own doc comment names `"your life total can't
  change"` as the zero-delta case a continuation has to be told about.

So #1200 is: one more payload on `PlayerStatic` beside `Keyword` and `Timing`, one
predicate in the life pipeline, and — separately — the `Duration` generalisation of
`TurnScopedReplacements` that ADR 0063 Decision 8 deferred. Phasing changes none of
that and blocks none of it. Bundling it here would have doubled a PR that already
moves the battlefield out from under 125 walks, and it would have put two unrelated
registries in one ADR.

Teferi's Protection therefore goes from **two** caveats to **one**, and the one that
remains is #1200's.

---

## Consequences

**Good**

- CR 702.26b is true of the whole engine at once, including the parts nobody
  remembered to change, and including the parts written after this ADR.
- The bot, the legal-move enumerator and the client's `legal_targets` are correct
  with no edits, because all three read surfaces the phased-out permanent is absent
  from.
- The phasing keyword makes ~70 printed cards correct with no catalog work.
- Undo, clone and the snapshot are exact across a phase cycle, because the state is
  a `Card` value in a slice rather than a flag several subsystems have to agree about.

**Bad / accepted**

- `g.Battlefield` is no longer "every permanent on the battlefield" in the CR sense;
  it is "every permanent that currently exists as far as the game is concerned".
  That is the *useful* set for 100% of readers, but it is a renaming of a concept
  and the doc comments on `Game.Battlefield` and `Game.PhasedOut` have to carry it.
- Anything that genuinely needs the phased-out permanents — the untap step, the
  leave-the-game sweep, the view — has to say so explicitly. That is three sites,
  and each names the rule it is implementing.
- A `*Card` pointer into `g.Battlefield.Cards` held across a phase-out dangles, the
  same way it does across a destroy. The primitive chooses its set up front for
  exactly this reason.
- The schema bump costs every pre-v6 restore point, phasing or not. That is the
  policy ADR 0041 set and the cost v2 and v5 already paid.

**Cards unlocked** (all `full`): Vodalian Illusionist, Reality Ripple, Clever
Concealment, Oubliette. **Card improved:** Teferi's Protection, from two caveats to
one. **Cards reachable with no further engine work:** every printed-phasing
permanent through Decision 5, and the "phases out until this leaves the battlefield"
family (Out of Time) through Decision 2's `PhaseInLockedBy`.
