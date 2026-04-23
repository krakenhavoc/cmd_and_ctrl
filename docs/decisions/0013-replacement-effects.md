# ADR 0013 — Replacement effects engine (S17)

**Status:** Accepted · 2026-04-22 · Sprint S17

## Context

S16 shipped the CR 613 layer engine — the catalog can now describe
continuous effects that exist while a permanent is on the battlefield.
What the engine still cannot do is *intercept* events before they
happen. A creature can't "enter the battlefield tapped", Doubling
Season can't double +1/+1 counters, Stasis can't skip untap steps,
and S13.1's commander-zone rewrite is hand-rolled inline with no
generalisation path.

Replacement effects (CR 614) are the structural answer: effects that
watch for candidate events and substitute a different event — or
cancel it — before it resolves. They are the second-most complex
rules subsystem after the stack, and they unblock most of Phase 7's
remaining card catalog (S18 combat keywords, S21 tokens, S22 draw,
S30 damage prevention all depend on the hooks).

S17 ships the engine plus twelve cards (seven listed on the sprint
doc + three enters-tapped finishes from S14 + Library of Leng +
Fog), refactors S13.1's commander-zone replacement into a built-in,
and closes the S14–S16 deferrals that were bundled here. Mind
Control's aura infrastructure, cost-replacement effects (Trinisphere
/ Thalia / Spellshift / Kambal), and damage-prevention shields with
charges are explicitly deferred — tracked in S24, S28, and S30
respectively.

## Decisions

### 1. Pre-event pipeline with a tagged-union value type

Replacement effects fire *before* a mutation runs, not after. The
architecture introduces `ReplacementEvent` — a mutable tagged-union
struct with per-kind payload fields (draw / move / counter / life /
damage / step-transition) — and `applyReplacementsLocked(ev)` as the
central pipeline call. The five rules-visible mutation functions
(`AddCounter`, `MoveCardByID`, `DrawCard`, `ChangePlayerLife`,
`MarkDamage`) each construct a `ReplacementEvent`, call the
pipeline, and branch on the result: apply (possibly mutated),
cancel, or bail out pending a CR 616 prompt.

**Why a tagged union, not an interface:** mirrors the `Event` struct
that `EmitEvent` already uses. A typed-fields struct is the shape
replacement-effect cards want to read — Doubling Season reads
`CounterDelta`, Kismet reads `NewZone` — and matches the existing
wire projection pattern. An interface would force every card to
unwrap via type assertion, and the field surface is small and
bounded.

**Why pre-event, not post-event:** CR 614 is explicit — "substitute
a different event before it happens." Post-event interception would
require emit-then-rollback on cancellation, which is intractable
once any listener has already observed the original event. The
pipeline runs synchronously under the same game write lock the
mutation already holds; no observer sees partial state.

### 2. `ReplacementEffect` is a declarative struct sibling of `StaticAbility`

```go
type ReplacementEffect struct {
    Watches         []EventKind
    AppliesTo       func(*ReplacementEvent, *Game, *Card) bool
    Replace         func(*ReplacementEvent, *Game, *Card) error
    Controller      func(*ReplacementEvent, *Game, *Card) uuid.UUID
    SelfReplacement bool
    Label           string
}
```

Same func-field shape as S16's `StaticAbility`. Cards register via
`effects.Spec.Replacements []game.ReplacementEffect` — a parallel
field to `Static`, `OnETB`, `OnResolve`, `ManaAbilities`. A seventh
function-var hook `CatalogReplacements` joins the existing six,
following the cycle-break pattern every sprint since S14 has used.

**Why not an interface:** card files already import `game` for
`Card` / `Game` / `Characteristic`; a parallel interface adds zero
information. Func fields are what the catalog needs.

### 3. CR 616.1 iterative apply-loop, centrally enforced

`applyReplacementsLocked` runs a bounded loop. Each iteration:
gather applicable replacements (filtered by `Watches`, `AppliesTo`,
and the per-event already-fired map), pick one, apply it, iterate.
Stop when no more apply or the iteration cap is hit (emit
`EventEffectError`). If two or more apply in the same iteration,
queue a `replacement_order` PendingChoice and return — the caller
handles re-entry.

**Why centrally:** CR 614/616 correctness is cross-cutting. A
card-level apply-loop would repeat the logic in every card file
and drift. Cards declare `AppliesTo` + `Replace`; the engine owns
the iteration, the once-per-event map, and the ordering prompt.

### 4. Once-per-event tracking applies to all fired replacements, not just `SelfReplacement`

CR 614.5 prevents "an otherwise infinite number of applicable
replacement effects from the same source" by capping self-
replacement at one per event. In practice the apply-loop needs the
same guarantee for *all* effects — once Doubling Season doubles a
counter-placement event, the loop must not pick it again the next
iteration. (A pure "re-evaluate AppliesTo every iteration" approach
is brittle: Doubling Season's predicate is "any creature gets
counters," which stays true after it fires.)

Decision: the once-per-event map keys on
`(ReplacementEventID, ReplacementEffectID)` and tracks every fired
effect. `SelfReplacement` stays on the struct as a hook /
documentation flag for future differentiation (e.g. if the "same
source, different event" distinction ever matters for a specific
card), but the apply-loop does not treat self-replacements
specially in S17.

Map scope is **per-call**: `defer delete(g.replacementsAppliedThisEvent, ev.ID)`
at the outermost entry point. The map survives across CR 616 prompt
pauses because the outermost entry is the pipeline function, which
doesn't return until the whole event settles (or is canceled).

### 5. CR 616 affected-player-chooses-order via `replacement_order` PendingChoice

When ≥2 replacements apply in the same iteration, the affected
player picks the order. Implementation: a new `PendingChoiceKind`
value alongside the existing `discard_from_hand` / `mana_pick` /
reserved `mode_pick` / `mill_reveal`. The PendingChoice carries
`ReplacementEffectIDs []ReplacementEffectID`; the wire projection
adds `ReplacementOptions []ReplacementOptionView` (id + label +
source card id).

The engine stashes a server-only `replacementResume` frame on the
PendingChoice — the in-flight `ReplacementEvent` plus the gathered
applicable list. When the client submits `resolve_choice` with
`order: [id1, id2, …]`, `ResolveReplacementOrder` validates the
permutation, locks in the order, and re-invokes the pipeline body
via a `resumeXxxLocked` helper keyed on the original event kind.

**Why queue-and-return, not block:** the existing PendingChoice
infrastructure (since S13 for targeting, S15 for mana picks) is
synchronous queue-then-return. The game-write-lock releases when
the pipeline function returns; the client renders the modal; the
client's `resolve_choice` acquires the lock and resumes. No
goroutines, no channels — same shape Thoughtseize / Arcane Signet
already use.

**Load-bearing invariant:** the pipeline function emits no events
and performs no observable mutation before the prompt queues. A
partial state is impossible because the actual mutation
(`actuallyDrawCardLocked`, `applyCounterLocked`, etc.) only runs
after the replacement loop settles.

### 6. Six pipeline integration points (five mutations + step transition)

The core five mutations named in the sprint plan are the rules-
visible event emitters:

| Function | File:Line | Kind |
|---|---|---|
| `drawCardLocked` | `mutations.go:114` | `RepEventDraw` |
| `MoveCardByIDAsCommander` | `mutations.go:1777` | `RepEventMove` |
| `ChangePlayerLife` | `mutations.go:2772` | `RepEventLife` |
| `MarkDamage` (+ new `MarkCombatDamage` wrapper) | `mutations.go:1460` | `RepEventDamage` |
| `AddCounter` | `mutations.go:2797` | `RepEventCounter` |

A sixth integration point — `runStepEntryHooksLocked`
(`game.go:503`) — fires `RepEventStepTransition` at the top of the
step-entry path so Stasis can cancel `StepUntap`. Step advance does
not emit a wire `Event` today; the new kind is engine-internal only
and the apply-loop short-circuits on the cancel path (replacements
for skip-step are order-irrelevant).

Parallel hooks in the `effect_api.go` `*ForEffect` helpers
(`AddCounterForEffect`, `ChangePlayerLifeForEffect`,
`DealDamageToCreatureForEffect`, `DealDamageToPlayerForEffect`)
route catalog-effect-driven mutations through replacements too, so
Doubling Season works whether a counter lands via the public action
or via an effect's primitive.

### 7. `enterBattlefieldLocked` shared helper

All battlefield-entry sites (`mutations.go:364`, `:780`, `:1828`,
plus `SearchLibraryForEffect`'s battlefield branch) consolidate
into one helper. Required so the `RepEventMove.EntersTapped` and
`EntersWithCounters` replacements fire uniformly from every entry —
cast resolution, admin direct-drop, commander-zone rewrite, land
fetch, token creation. Refactoring blast radius is the riskiest
leg of sub-PR 2; factored into its own commit with explicit
regression tests.

### 8. S13.1 commander-zone replacement refactored into a built-in

Today the commander-zone rewrite lives inline in
`MoveCardByIDAsCommander` via `applyCommanderZoneReplacementLocked`
(`mutations.go:1833`). After S17 it is one built-in
`ReplacementEffect` in `builtin_replacements.go`, registered at
`NewGame` alongside S16's `layerVersionBump` listener.
`AppliesTo` gates on `ev.asCommanderMove && card.IsCommander &&
NewZone ∈ {graveyard, exile, hand, library}`; `Replace` rewrites
`NewZone = ZoneCommand`. The bespoke function is deleted; existing
`mutations_test.go` commander-zone tests must still pass unchanged.

**Why built-in, not a catalog entry:** it's a rules-engine
replacement (CR 903.9), not a card-defined one. Registering it
through the catalog would require a synthetic oracle ID per
commander. Built-in is the right shape.

### 9. Enters-tapped for fetched lands uses a primitive flag, not the pipeline

Cultivate, Path to Exile, and Solemn Simulacrum fetch lands to the
battlefield. Their S14 simplification ("enters untapped") predated
the replacement pipeline; the obvious S17 path is to declare each
card with an enters-tapped replacement. That's heavier than the
text warrants — Cultivate's "tapped" is hard-coded for the card,
not a general effect that watches other moves.

Decision: add `TappedOnEntry bool` to the `SearchLibrary` primitive
(`server/internal/cards/effects/primitives.go`) + parallel flag on
`SearchLibraryForEffect`. Set `Tapped = true` after push. Simpler,
matches the card text, and doesn't grow the generic replacement
surface.

Kismet — "creatures, artifacts, and lands opponents control enter
the battlefield tapped" — stays in the generic replacement path
because it watches arbitrary other moves.

### 10. Library of Leng scope limited to cleanup-step discard

CR 701.8a/c distinguishes voluntary vs involuntary discard. Library
of Leng's "top or bottom of library instead" only replaces
voluntary discards. Threading `IsVoluntary` through every discard
caller (cleanup, Mind Rot, Liliana's Caress, cast-trigger
discard-to-draw) is out of scope for this sprint.

Decision: S17 ships Library of Leng wired into the cleanup-step
discard path only (where voluntariness is implicit). An
involuntary-discard follow-up issue tracks the gap; proper CR
701.8a/c detection lands when a second voluntary-discard
replacement enters the catalog.

### 11. Damage prevention hook only, no shield mechanic

Fog ships as a proof-of-concept damage-prevention card: a turn-
scoped replacement cleared at `StepCleanup` that cancels combat
damage. The `MarkCombatDamage(source, target, delta)` wrapper
around `MarkDamage` flags `IsCombatDamage=true` in the
`ReplacementEvent` so combat-only prevention distinguishes cleanly
from spell damage.

Full CR 615 damage-prevention with charges (Shield of the Oversoul,
Story Circle) — a stateful "shield has N uses, tick on apply" —
stays in S30. The S17 turn-scoped `DamagePreventionUntilEndOfTurn`
slice is a scaffold S30 extends, not a full implementation.

### 12. Additive wire protocol — no version bump

`PendingChoiceView.ReplacementOptions` is a new optional field;
pre-S17 clients ignore it. The new `replacement_order` kind only
ships on the wire when an S17 card creates one. No message
version increment; the protocol stays v0 per the existing
conventions.

### 13. Mycosynth Lattice clauses — provisional sub-PR with deferred scope decision

Lattice's "lands tap for any color" + "no land's mana ability adds
non-colorless" clauses were deferred from S16 and slotted for S17.
Both overlap S15 mana-ability machinery. Two paths evaluated in
sub-PR 6:

- **Route through replacement pipeline**: introduce
  `RepEventManaProduced` fired from `ActivateManaAbility` before
  the token drops into the pool. Seventh integration point outside
  the core S17 scope; manageable if cheap.
- **Defer to S18**: hold the clauses. Lattice stays type-adder
  only; the new clauses land in S18 alongside the mana-ability
  rewrite.

Scope decision deferred to sub-PR 6 PR-open time with an explicit
user prompt. Either outcome is tracked (§Consequences below).

## Out of scope (explicit deferrals)

- **Aura / Equipment attachment infrastructure** (Mind Control) →
  **S24** [#76](https://github.com/krakenhavoc/cmd_and_ctrl/issues/76). The S16 doc bundled this into S17; re-homed to
  S24 because aura attachment + combat-state coupling fits
  S24's equipment/aura scope more naturally.
- **Cost-replacement effects** (Trinisphere / Thalia / Spellshift /
  Kambal) → **S28** [#93](https://github.com/krakenhavoc/cmd_and_ctrl/issues/93). The S17 pipeline hooks touch
  event-path mutations; cost replacement touches the S15 cost-
  computation engine. Separate architectural surface.
- **Damage prevention shields with charges** (CR 615) → **S30**
  [#95](https://github.com/krakenhavoc/cmd_and_ctrl/issues/95). Fog
  in S17 is atomic cancel; stateful shields are a different shape.
- **Dependency detection** (CR 613.8) + Layer 1 copy effects →
  **S16.5** layer-system follow-ups. Not blocking for replacements.
- **Library of Leng voluntary-vs-involuntary discard precision** —
  S17 ships cleanup-path only; proper CR 701.8a/c detection
  tracked as a follow-up issue.
- **Non-self repeat replacements** — if a future card needs
  "replace the event again" behavior beyond CR 614/616, flag for
  rules review.

## Consequences

- **12 catalog cards ship**: Doubling Season, Hardened Scales,
  Branching Evolution, Champion of Lambholt (counter half),
  Hangarback Walker, Stasis, Kismet, Fog, Library of Leng, plus
  the three enters-tapped finishes for Cultivate / Path to Exile /
  Solemn Simulacrum.
- **Seventh catalog hook** `CatalogReplacements` joins the
  S14/S15/S16 six.
- **`Spec.Replacements []game.ReplacementEffect`** new field on the
  catalog spec; parallel to `Static`.
- **New PendingChoiceKind** `replacement_order` + new
  `ResolveReplacementOrder` method + dispatcher leg for `order`
  payload. Wire projection adds `ReplacementOptions`.
- **S13.1 commander-zone replacement deleted**; built-in
  replacement takes over. Behavior unchanged.
- **`enterBattlefieldLocked`** new shared helper replacing ad-hoc
  calls at four battlefield-entry sites.
- **`MarkCombatDamage(source, target, delta)`** new wrapper around
  `MarkDamage`; flags `IsCombatDamage=true` for Fog / combat-only
  prevention.
- **`RepEventStepTransition`** engine-only event kind fired at the
  top of `runStepEntryHooksLocked` so Stasis can cancel StepUntap.
  Step advance still does not emit a wire `Event`.
- **`SearchLibrary` primitive gains `TappedOnEntry bool`**;
  Cultivate / Path to Exile / Solemn Simulacrum stop shipping the
  "enters untapped" sandbox note.
- **Library of Leng ships cleanup-path only**; an involuntary-
  discard follow-up issue remains open.
- **Mycosynth Lattice clauses** land in sub-PR 6 iff the scope
  decision clears; otherwise re-homed to S18.
- **S18** inherits aura attachment (via S24), combat keyword
  behavior, and potentially the Mycosynth Lattice clauses.
- **S24** inherits Mind Control + aura attachment infrastructure.
- **S28** inherits cost-replacement effects.
- **S30** inherits damage-prevention shields with charges, built on
  top of the S17 `MarkCombatDamage` hook + turn-scoped prevention
  pattern.
- **No protocol version bump.** `replacement_order` is additive.

## Decision log (planning round, 2026-04-22)

1. Aura / Mind Control → S24, not S17.
2. Cost-replacement effects → S28.
3. Damage prevention split — S17 ships the hook + Fog; S30 ships
   the shield mechanic with charges.
4. Library of Leng → cleanup-path only; strict voluntariness
   deferred.
5. Mycosynth Lattice clauses → own sub-PR, scope decision deferred
   to PR-open.
6. Enters-tapped for fetched lands → `SearchLibrary.TappedOnEntry`
   primitive flag, not the generic replacement pipeline.
7. Once-per-event tracking applies to all fired replacements, not
   only `SelfReplacement`.
8. No protocol version bump; `replacement_order` is additive.
