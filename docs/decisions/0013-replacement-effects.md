# ADR 0013 — Replacement effects engine (S17)

**Status:** Implemented · 2026-04-22 (planned), 2026-04-23 (shipped) · Sprint S17
**Amended:** 2026-09-16 · Branch `docs/discard-rules-library-of-leng` — §10 withdrawn, see [§10a](#10a-amendment-2026-09-16-10-misread-the-card-and-the-rules)

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

**Withdrawn 2026-09-16: this section misstates both the card and the
rules. Read [§10a](#10a-amendment-2026-09-16-10-misread-the-card-and-the-rules) instead.**
The original text stays below as the record of what was decided.

> CR 701.8a/c distinguishes voluntary vs involuntary discard. Library
> of Leng's "top or bottom of library instead" only replaces
> voluntary discards. Threading `IsVoluntary` through every discard
> caller (cleanup, Mind Rot, Liliana's Caress, cast-trigger
> discard-to-draw) is out of scope for this sprint.
>
> Decision: S17 ships Library of Leng wired into the cleanup-step
> discard path only (where voluntariness is implicit). An
> involuntary-discard follow-up issue tracks the gap; proper CR
> 701.8a/c detection lands when a second voluntary-discard
> replacement enters the catalog.

### 10a. Amendment, 2026-09-16: §10 misread the card and the rules

*Amendment, 2026-09-16, branch `docs/discard-rules-library-of-leng`.
Closes [#160](https://github.com/krakenhavoc/cmd_and_ctrl/issues/160)
as not planned. The work moves to
[#650](https://github.com/krakenhavoc/cmd_and_ctrl/issues/650) and
[#651](https://github.com/krakenhavoc/cmd_and_ctrl/issues/651).*

Nothing §10 describes ever shipped. Library of Leng was dropped from
S17 (see Consequences), it is still not in the catalog (batch 28,
[#390](https://github.com/krakenhavoc/cmd_and_ctrl/issues/390), skipped
it), and `IsVoluntary` was never added. The problem is that the
follow-up §10 created (#160) planned the wrong work.

**What the card says.** Oracle text: "You have no maximum hand size.
If an effect causes you to discard a card, discard it, but you may put
it on top of your library instead of into your graveyard."

- **What matters is effect, cost or game rule, not voluntary or
  involuntary.** The rules have no such thing as a voluntary discard,
  and the discard keyword action §10 cites doesn't say otherwise. An
  effect is what a spell or ability does (CR 609.1). A cost is
  something a player pays (CR 118.1). The cleanup hand-size discard is
  a turn-based action (CR 514.1, 703.1). Leng replaces only a discard
  caused by an effect. From the Gatherer rulings (2004-10-04): "You
  can't use the Library of Leng ability to place a discarded card on
  top of your library when you discard a card as a cost, because costs
  aren't effects."
- **It applies to Mind Rot**, which §10 listed as a discard Leng must
  not replace. The same rulings: "The ability applies any time a spell
  or ability has you discard as part of its effect. It does not matter
  if you or your opponent control the spell or ability." Looting
  ("draw, then discard") is also an effect discard.
- **It never applies to the cleanup discard**, which is the one path
  §10 wired it to. That discard is a turn-based action (CR 514.1,
  703.1), not an effect. Leng usually removes your maximum hand size
  (CR 402.2), but a later hand-size effect can set one again: hand-size
  effects apply in timestamp order, so Null Profusion entering after
  Leng makes your maximum hand size two (ruling, 2009-10-01). You can
  still owe a cleanup discard then, and Leng doesn't replace it.
- **The card goes on top of the library, not "top or bottom".** The
  replacement is a "may", which fits the existing
  `optional_replacement` prompt. When one effect discards several
  cards, the player decides for each card and chooses the order of
  the ones that go to the library (ruling). A card put there isn't
  revealed unless the discarding effect says so.
- **Some discards during resolution are costs.** In "[do X]. If you
  do, …" and "… unless [a player does X]", X is a cost paid while the
  spell or ability resolves (CR 118.12, 118.12a). So a discard that
  happens during resolution isn't automatically an effect discard.
  Thirst for Knowledge's "unless you discard an artifact card" branch
  is a cost, and Leng doesn't apply to it.
- The wording #160 quoted as Leng's ("if a spell or ability an
  opponent controls causes you to discard") belongs to Obstinate
  Baloth, Loxodon Smiter, Wilt-Leaf Liege and Dodecapod.

**What the engine has to know about each discard** is its cause
(effect, cost or turn-based action), the source, and who controls the
source. Leng needs the cause. The Obstinate Baloth family also needs
the controller. Madness (CR 702.35a) needs neither, because it
replaces every discard. Rest in Peace needs no cause either, but all
of these need discards to go through the replacement window first.

**None of it can be written today, because no discard reaches the
CR 614 window.** All four places that discard call
`MoveCard(hand, graveyard)` directly and then emit `EventDiscardCard`
without a `Source` (develop at `7c1ae9b`):

| site | what discards there |
|---|---|
| `game/mutations.go` `DiscardSelection` | the cleanup hand-size discard (CR 514.1) |
| `game/effect_api.go` `discardPicksLocked` | every effect discard the player chooses — Mind Rot, looting (#651 moved these off `DiscardSelection`) |
| `game/effect_api.go` `DiscardRandomForEffect` | random discards |
| `game/additional_cost.go` `payAdditionalCostLocked` | discard as an additional cost to cast a spell |
| `game/pending_choice.go` `PendingChoiceDiscardFromHand` | Thoughtseize-style "you choose, they discard" |

`routeCardToZoneLocked` (`game/zone_route.go`) already opens the
window for mill, countered spells, exile and bounce
([#529](https://github.com/krakenhavoc/cmd_and_ctrl/issues/529)), but
it has no discard flag. Two more graveyard arrivals skip the window
too: a library search that puts the card into a graveyard (Entomb,
`effect_api.go`) and surveil (`ResolveSurveil` in `pending_choice.go`).
So Rest in Peace also needs those two.

**There was also a bug underneath, now fixed (#651).** An effect
discard did not block the game: `DiscardChoiceForEffect` added to
`DiscardPending`, the map the cleanup step uses, nothing read that map
outside cleanup, and entering cleanup reset it to the hand-size count
— so players could pass priority while a Mind Rot discard was still
owed, and at cleanup the discard was lost. An effect's discard is now
a `PendingChoice` (ADR 0010 §10 amendment), so it is owed, gated and
never erased, and it has somewhere stable to pause when the
replacement window arrives. `DiscardPending` is cleanup-only.

**Where the work went.**

- [#650](https://github.com/krakenhavoc/cmd_and_ctrl/issues/650):
  ADR for discard as a replaceable event with a cause. It decides the
  cause taxonomy, a single route for all four sites, the CR 118.12
  cost case, picking random discard sets up front, the CR 903.9 prompt
  on a discarded commander, and what bots answer. It also decides
  whether that lands as another amendment here or as a new ADR.
  Depends on #651.
- [#651](https://github.com/krakenhavoc/cmd_and_ctrl/issues/651): bug.
  Effect discards can be passed through and are wiped at cleanup.
- [#657](https://github.com/krakenhavoc/cmd_and_ctrl/issues/657):
  madness, which needs #650's routing but not the cause.
- Library of Leng itself becomes an ordinary catalog card once #650
  lands. It stays on #390's skip list until then.

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
- ~~**Library of Leng voluntary-vs-involuntary discard precision** —
  S17 ships cleanup-path only; proper CR 701.8a/c detection
  tracked as a follow-up issue.~~ Withdrawn 2026-09-16: the premise
  was wrong, see §10a. Tracked now as
  [#650](https://github.com/krakenhavoc/cmd_and_ctrl/issues/650) and
  [#651](https://github.com/krakenhavoc/cmd_and_ctrl/issues/651).
- **Non-self repeat replacements** — if a future card needs
  "replace the event again" behavior beyond CR 614/616, flag for
  rules review.

## Consequences (as shipped 2026-04-23)

- **9 catalog entries ship**: six new cards (Doubling Season,
  Hardened Scales, Branching Evolution, Stasis, Kismet, Fog) plus
  three S14 finishers (Cultivate / Path to Exile / Solemn Simulacrum)
  that now set `SearchLibrary.TappedOnEntry`. Hangarback Walker,
  Champion of Lambholt, Library of Leng, and the Mycosynth Lattice
  clauses deferred — see Out of Scope below. (Library of Leng is
  still not in the catalog as of 2026-09-16; see §10a.)
- **Seventh catalog hook** `CatalogReplacements` joins the
  S14/S15/S16 six.
- **`Spec.Replacements []game.ReplacementEffect`** new field on the
  catalog spec; parallel to `Static`.
- **Two new PendingChoiceKinds**: `replacement_order` (CR 616
  multi-effect ordering) + `optional_replacement` (CR 614.10 yes/no).
  Each has a resume method (`ResolveReplacementOrder`,
  `ResolveOptionalReplacement`) and a dispatcher leg. Wire
  projection adds `ReplacementOptions`.
- **`ReplacementEffect.Optional bool`** (sub-PR 6) flags effects
  that need owner opt-in before firing. Used today by the CR 903.9
  commander-zone built-in; future home for "may exile instead" cards.
- **S13.1 commander-zone replacement deleted**; built-in
  replacement takes over AND widened: fires on EVERY commander
  move (spell-driven, SBA-driven, admin-driven) instead of only
  admin-flagged moves. Closes [#164](https://github.com/krakenhavoc/cmd_and_ctrl/issues/164).
- **Battlefield-entry pipeline routing** is inline at each entry
  site (cast resolve, land play, admin move). The planned
  `enterBattlefieldLocked` shared-helper refactor was scoped out
  in favor of targeted per-site calls — less churn, same behavior.
- **`routeBattlefieldCardToOwnerGraveyardLocked` routes through
  the pipeline** so SBA deaths + wrath destroys fire the commander-
  zone replacement. New `executeBattlefieldLeaveLocked` helper
  runs the physical move post-pipeline.
- **`MarkCombatDamage(source, target, delta)`** new wrapper around
  `MarkDamage`; flags `IsCombatDamage=true` for Fog / combat-only
  prevention.
- **`RepEventStepTransition`** engine-only event kind fired at the
  top of `runStepEntryHooksLocked` so Stasis can cancel StepUntap.
  Step advance still does not emit a wire `Event`.
- **`Game.TurnScopedReplacements`** new per-turn replacement slot
  cleared at `StepCleanup`. Fog-class transient effects live here.
- **`SearchLibrary` primitive gains `TappedOnEntry bool`**;
  Cultivate / Path to Exile / Solemn Simulacrum stop shipping the
  "enters untapped" sandbox note.
- **CR 514.1 cleanup-step discard scoped to active player only**
  (sub-PR 6 bug fix). Pre-S17 the check populated `DiscardPending`
  for every seated player; the correct rule is active-player-only.
- **Wire P/T includes counter delta** — `viewOfCard` sources from
  `c.CurrentPower()` / `c.CurrentToughness()` so the on-card P/T
  pip matches the SBA + combat-damage reads. Pre-existing S13.2
  bug surfaced by S17's first counter-placing scenario.
- **Cast-payload retry stash** in `Game.svelte.sendAction` — the
  insufficient-mana "Cast anyway" and "Auto-tap & cast" retries
  now replay the original cast_spell payload (targets, modes, X,
  distribution) instead of hand-rolling a bare payload. Pre-
  existing S15 bug surfaced by a Path to Exile / Swords manual
  test.
- **`CastSpell` target-required guard** rejects cast_spell with
  empty `params.Targets` when the card's `target_mode` is non-
  empty. Turns silent "no effect" resolutions into visible
  `ErrInvalidParam` toasts.
- **Shift+click per-card `+1/+1` counter** in `PlayerPanel.svelte`
  as a debug affordance until the right-click admin context menu
  mini-sprint ([#170](https://github.com/krakenhavoc/cmd_and_ctrl/issues/170)) ships.
- **`CounterPips.svelte` rendering** disambiguated — abbr + count
  as two spans with a tabular-numeric badge (the previous `·`
  separator read as `+` at 10px font).
- **`EventEffectError` logged via `slog.Warn`** at emission time —
  until the client learns to render effect errors as toasts,
  server terminal output is the diagnostic surface.
- **S18** inherits combat-keyword behavior + the Mycosynth Lattice
  "lands tap for any color" + "no mana ability adds non-colorless"
  clauses (re-homed from S17's provisional sub-PR 6).
- **S24** inherits Mind Control + aura attachment infrastructure
  (bundled out of S17 during planning).
- **S28** inherits cost-replacement effects (Trinisphere, Thalia,
  Spellshift, Kambal's life-gain-on-cast).
- **S30** inherits damage-prevention shields with charges, built on
  top of the S17 `MarkCombatDamage` hook + turn-scoped prevention
  pattern.
- **No protocol version bump.** New PendingChoice kinds + wire
  fields are additive; old clients ignore them.

## Decision log (planning round, 2026-04-22)

1. Aura / Mind Control → S24, not S17.
2. Cost-replacement effects → S28.
3. Damage prevention split — S17 ships the hook + Fog; S30 ships
   the shield mechanic with charges.
4. Library of Leng → cleanup-path only; strict voluntariness
   deferred. *Withdrawn 2026-09-16: see §10a.*
5. Mycosynth Lattice clauses → own sub-PR, scope decision deferred
   to PR-open.
6. Enters-tapped for fetched lands → `SearchLibrary.TappedOnEntry`
   primitive flag, not the generic replacement pipeline.
7. Once-per-event tracking applies to all fired replacements, not
   only `SelfReplacement`.
8. No protocol version bump; `replacement_order` is additive.
