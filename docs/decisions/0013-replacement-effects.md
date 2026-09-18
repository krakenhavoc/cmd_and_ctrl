# ADR 0013 — Replacement effects engine (S17)

**Status:** Implemented · 2026-04-22 (planned), 2026-04-23 (shipped) · Sprint S17
**Amended:** 2026-09-16 · Branch `docs/discard-rules-library-of-leng` — §10 withdrawn, see [§10a](#10a-amendment-2026-09-16-10-misread-the-card-and-the-rules)
**Amended:** 2026-09-17 · Branch `fix/792-identical-replacements-no-prompt` — §5's prompt has two exceptions now, see [§5a](#5a-amendment-2026-09-17-when-the-616-prompt-has-only-one-answer)
**Amended:** 2026-09-17 · Branch `fix/799-853-discard-helper` — §5f's open note is closed: a discard is an exit too, see [§5g](#5g-amendment-2026-09-17-a-discard-is-an-exit-too)
**Amended:** 2026-09-17 · Branch `fix/815-847-replacement-outcomes` — §5a's "never collapsed" limit now holds on the multi-effect order paths too, see [§5h](#5h-amendment-2026-09-17-a-may-inside-a-chosen-order-is-still-a-question)
**Amended:** 2026-09-17 · Branch `fix/815-847-replacement-outcomes` — §5d's landed-outcome rule now decides the destruction COUNT as well, see [§5i](#5i-amendment-2026-09-17-destroyed-this-way-counts-what-was-destroyed)

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

### 5a. Amendment, 2026-09-17: when the 616 prompt has only one answer

*Amendment, 2026-09-17, branch `fix/792-identical-replacements-no-prompt`.
Closes [#792](https://github.com/krakenhavoc/cmd_and_ctrl/issues/792) and
records [#710](https://github.com/krakenhavoc/cmd_and_ctrl/issues/710),
which shipped the first half of this and did not amend §5.*

§5 says "when ≥2 replacements apply in the same iteration, the
affected player picks the order", and the engine queued the prompt on
that count alone. CR 616.1 gives the affected player a choice; it does
not require asking a question whose answers are indistinguishable. Two
exceptions now apply the gathered order inline instead, both in
`applyReplacementsLocked`:

1. **Every applicable effect is a pure cancel** (#710) — whichever is
   put first cancels the event and ends the apply-loop.
   `ReplacementEffect.PureCancel` is the card's declaration that its
   `Replace` does nothing but `ev.Cancel()`. Two "skip your draw step"
   enchantments under one controller.
2. **Every applicable effect is the same declared effect** (#792) —
   two Doubling Seasons, two Rhox Faithmenders, two Hardened Scales.
   Same modification, applied N times, in any order.

"The same declared effect" is `replacementIdentity` in
`replacements.go`: the source's `CatalogAbilityKey`, the slot index
into that entry's `Replacements` slice, and the controller of the
object contributing it. It is the engine's existing catalog-
replacement key with the OBJECT dropped — `ReplacementEffectID` says
"the Season in battlefield slot 3", the identity says "Doubling
Season's counter-doubling, controlled by Ian". The controller belongs
in it because replacements read their own controller (Notion Thief
writes `src.Controller` straight into the event), so two copies under
different controllers are two different modifications. Built-in,
turn-scoped and test replacements are registered per instance rather
than declared on a catalog entry and have no identity, so they never
collapse.

Three deliberate limits:

- **A mixed window still prompts with everything listed.** Two Seasons
  and a Hardened Scales do not collapse to two entries: CR 616.1 lets
  the affected player interleave, and Season → Scales → Season is 6
  counters, which neither Season → Season → Scales (5) nor
  Scales → Season → Season (8) can reach. Collapsing would remove a
  legal outcome, which is worse than one extra prompt.
- **An effect that asks its controller a question is never
  collapsed** — a CR 614.10 "may", a shockland's pay-life, a copy
  selector. Skipping the ordering prompt would skip its question too.
  Shared by both exceptions as `asksItsOwnQuestion`.
- **An effect whose `Replace` writes its own SOURCE into the event is
  not covered.** "That damage is dealt to this creature instead" from
  two copies of one card would share an identity and would not be
  interchangeable. Nothing in the catalog does that today — the two
  source-reading replacements that exist are additive, and Arwen is
  legendary — and AGENTS.md §7 tells card authors to raise it rather
  than ship it quietly. If one ever lands, the fix is a declared flag
  in the `PureCancel` mould, not a per-card special case.

### 5b. Amendment, 2026-09-17: what a paused life change owes its caller, and why a cost never pauses

*Amendment, 2026-09-17, branch `fix/793-life-change-continuation`.
Closes [#793](https://github.com/krakenhavoc/cmd_and_ctrl/issues/793),
which [#482](https://github.com/krakenhavoc/cmd_and_ctrl/issues/482)
(PR #790) created by routing every life change through this pipeline.*

§5 says the affected player picks the order and the event lands from the
resume. Everything in §5 is still true; what it did not say is what
happens to the code that ASKED for the event while it is waiting.

Before #482, a life change could not pause, so a caller could change a
life total and read it back on the next line to find out how much
actually moved. "Each opponent loses X life. You gain life equal to the
life lost this way" — Exsanguinate, Debt to the Deathless, Gray Merchant
of Asphodel, Kokusho — is that shape, and after #482 it reads a total
that has not moved yet and gains nothing. Two decisions:

**1. A life change carries its continuation, the way a damage event
carries its tail.** `ReplacementEvent.lifeTail` holds `then(g, applied)`
and `runLifeTailLocked` is the one place it runs, reached from every
TERMINAL outcome of a life event: landed (`applyResolvedLifeChangeLocked`,
the same function the CR 616 resume calls), replaced away under
CR 614.10, or its player gone. `applied` is the post-replacement delta,
signed the way the event is, and `0` for the two non-outcomes — the
caller is told either way, because a batch waiting on several players
must not stall on the one that moved nothing.

The public surface is `ChangePlayerLifeThenForEffect` for one player and
`LoseLifeEachThenForEffect` for "each opponent", the second built on the
first so there is one implementation of waiting rather than one per
card. The batch drains players IN SEQUENCE, each from the previous
one's continuation: that is what keeps the running total a value carried
forward rather than a shared accumulator, which is what makes an undo
across the prompt replay identically. The observable cost is that a
paused first opponent delays the rest of the drain until the prompt is
answered; the losses are simultaneous in the rules and sequential in
this engine either way, and doing them all on the far side of the prompt
is the closer of the two.

`Then` takes the live `*Game`, the contract every other continuation
frame in the engine follows. Nothing new is snapshotted: the tail rides
the `ReplacementEvent` on the existing `replacementResume` frame, which
`ContinuationCensus.ChoiceResumeFrames` already counts and `Snapshot`
already refuses to call a restore point.

One thing DID have to change for undo. `Clone` shallow-copied the
`PendingChoice`, so the undo snapshot shared the in-flight
`ReplacementEvent` with the live game — and answering a prompt mutates
that event in place. Undoing the answer and answering again therefore
doubled an already-doubled delta and found the continuation consumed.
`cloneReplacementResume` now gives the snapshot its own copy of the
event (the gathered `applicable` list stays shared; a resume only reads
it). That was a latent bug for counters and damage too, not just life.

*Corrected by §5e (#808):* the event was not the only thing the answer
consumed. An effect that applied on its own BEFORE the prompt was queued
is marked only in `Game.replacementsAppliedThisEvent`, which `Clone` did
not copy and the answer deletes when the event settles, so undoing into
the open prompt and answering again fired that effect a second time.
The map is now part of the undo snapshot, and "undo across the prompt
replays the same way" holds for that case too.

**2. Paying life is a life LOSS, but a cost never pauses.**
CR 119.4: "If a player pays life, the payment is subtracted from their
life total; in other words, the player loses that much life." So the
window runs on a payment and a life-loss replacement sees it; the issue's proposal to route costs
around the pipeline as "not a life-change effect" would have been a
rules change. What a payment is *not* is life GAIN, so Rhox Faithmender
and Alhammarret's Archive never touch it — which is the half the issue
was actually worried about, and it needs no special case at all.

What the payment cannot do is stop to ask a question. CR 601.2h pays a
spell's costs as one indivisible step of casting it and CR 601.2 rewinds
the announcement if they cannot all be paid; CR 602.2b says the same for
an activated ability, and CR 605.3b resolves a mana ability immediately
and without the stack. A CR 616 ordering prompt in the middle of any of
those leaves a half-paid cost that no rewind can take back. So
`PayLifeForEffect` sets `ReplacementEvent.mustSettleNow` and the
apply-loop settles without prompting:

- an ordering window applies in the order it was gathered — the escape
  §5a already uses for a pure-cancel set, for identical effects, and for
  an eliminated chooser — one effect at a time, re-gathering after each
  (CR 616.1f; §5e);
- anything that would ask its own question (`asksItsOwnQuestion`: a
  CR 614.10 "may", a shockland's pay-life, a copy selector) is skipped
  un-applied, the weaker-never-stronger posture
  `optionalReplacementResumableLocked` takes for an entry with nothing
  to resume it.

The affected player gives up their CR 616 ordering choice on a cost
payment. That is a real, declared simplification: arbitrary but
deterministic, reachable only with two DIFFERENT life-loss replacements
on one table, and the alternative is a spell stuck on the stack.

**A payment the window cancels is refused, not free** (§5e, #808).
CR 119.8: if a player can't lose life, "a cost that involves having that
player pay life can't be paid", and CR 614.17b: "If an event can't
happen, a player can't choose to pay a cost that includes that event."
The engine models "your life total can't change" as a null replacement,
so `PayLifeForEffect` returns `ErrInvalidParam` when the window cancels
the payment. A replacement that only changes the amount is still a
payment. This amendment originally said nothing about the cancelled
case and the code paid the cost for nothing.

`mustSettleNow` is deliberately a property of the EVENT rather than of
life: any future entry point with no resume and no rewind can set it,
and the two branches that honour it are three lines each.

### 5c. Amendment, 2026-09-17: damage owes its caller the same continuation

*Amendment, 2026-09-17, branch `fix/708-807-damage-path`.
Closes [#807](https://github.com/krakenhavoc/cmd_and_ctrl/issues/807).*

§5b is about life. Everything in it is true of DAMAGE, word for word,
because damage has run this window since S22 (players) and S30
(permanents) and therefore pauses the same way. Creeping Bloodsucker —
"this creature deals 1 damage to each opponent. You gain life equal to
the damage dealt this way" — read each opponent's life total back on the
line after damaging them, which is #793's shape with
`DealDamageToPlayerForEffect` in place of `ChangePlayerLifeForEffect`,
and gained nothing for an opponent whose damage was still waiting on a
CR 616 prompt.

**The decision is to make it one idiom rather than two.**
`damageTail` already existed (#694) and carried what the ENGINE owes a
settled damage event; it now also carries `then(g, dealt)`, the CALLER's
half, exactly as `lifeTail` does. `runDamageTailLocked` is the one place
it runs, reached from every TERMINAL outcome — landed, fully prevented,
replaced away, target gone — with `dealt` the post-replacement amount
and `0` for the three non-outcomes.

The public surface mirrors the life side name for name:
`DealDamageToPlayerThenForEffect`, `DealDamageToCreatureThenForEffect`,
and `DealDamageEachThenForEffect` for "each opponent", the batch built
on the single forms and sequenced through their continuations so the
running total is a value carried forward rather than a shared
accumulator. Card authors learn `...ThenForEffect` once.

Two things fell out of doing it this way:

- **Six copies of the pipeline dance became one.**
  `damageThroughReplacementsLocked` is `changeLifeThroughReplacementsLocked`
  for damage, and all six entry points (two effect deals, three combat
  paths, the manual sandbox mark) now call it. Adding a terminal outcome
  to a damage event is one edit, not six — which is the property #694
  was reaching for and did not quite get.
- **`Clone` deep-copies the `damageTail`.** The life tail is cleared by
  nilling the POINTER on the event, which the snapshot already has its
  own copy of; the damage tail is cleared by nilling `then` INSIDE it,
  because the rest of the tail is still being read when the continuation
  runs. Sharing the struct would let the live game's run consume the
  snapshot's continuation, so an undone answer would land the damage and
  skip the rest of the card.

Nothing new is snapshotted: the tail rides the `replacementResume`
frame, already counted by `ContinuationCensus.ChoiceResumeFrames`.

The catalog lint from #793 grew a second pattern rather than a second
file — `life_continuation_guard_test.go` now scans for a `.Life` or
`.DamageMarked` read after a damage call as well as after a life change,
and recognises the `DealDamage{…}.Apply(ctx)` primitive spelling. One
card needed converting; the sweep found no others.

### 5d. Amendment, 2026-09-17: a destruction clears damage only when it lands

*Amendment, 2026-09-17, branch `fix/708-807-damage-path`.
Closes [#708](https://github.com/krakenhavoc/cmd_and_ctrl/issues/708),
noticed while fixing [#605](https://github.com/krakenhavoc/cmd_and_ctrl/issues/605)
(PR #701).*

§6 lists `routeBattlefieldCardToOwnerGraveyardLocked` as a pipeline
integration point: a battlefield exit runs the CR 614 window so the
CR 903.9 commander built-in — and any future "if it would be destroyed,
instead …" — can replace it. What §6 did not notice is that the same
function cleared `Card.DamageMarked` on the line that found the owner,
*before* opening that window.

That is the ordering inverted. The damage was gone before the
replacements could see it, and it was gone whether or not the permanent
actually left. A destruction a replacement rewrote left the permanent
standing with its marks erased, which is wrong twice: CR 514.2 removes
marked damage at the cleanup step, not when something tried and failed
to destroy the permanent; and the CR 704.5g lethal-damage check would
not see it again on the next pass.

**Decision: the clear is a TERMINAL-OUTCOME mutation, like landing
damage and landing a life change.** It moved into
`executeBattlefieldLeaveLocked`, the one function that actually performs
the move and the one both the unpaused path and the CR 903.9 resume go
through. `clearBattlefieldDamageLocked` (permanent_damage.go, next to
the marking it undoes) is the named verb, and its doc comment is the
short list of who may call it.

*Extended, 2026-09-17 (#816):* "the one function that actually performs
the move" was only true of the DESTROY route. Every other exit —
exile, bounce, tuck, mill, the sandbox move — reaches the battlefield
through `MoveCard`'s CR 400.7 cleanup instead, and cleared nothing, so
an exiled or bounced permanent carried its damage into the new zone,
showed it there, and brought it back onto the battlefield when the card
was replayed. (Not every return: the exile → battlefield helper scrubs
the card as it mints the new instance ID, so a blink was fine and a
recast was not — per-path coverage, which is the thing being ended.)
The clear now lives
in that ONE exit cleanup (`clearBattlefieldDamage`, card-level, called
from `MoveCard` and from the CR 514.2 sweep), which is still the landed
outcome and is now the landed outcome of every route; it takes the
CR 702.2c deathtouch flag with it. Nothing above changes: a replaced
destruction still never reaches a move, so it still keeps its damage.

Consequences, checked caller by caller:

| Caller | Before | After |
| --- | --- | --- |
| SBA lethal-damage / deathtouch destruction | cleared, then maybe moved | cleared iff it moves |
| `DestroyPermanentForEffect` (Doom Blade) | same | same |
| `DestroyPermanentsForEffect` (Wrath) | same, per card | same, per card |
| Sacrifice (`sacrifice.go`) | cleared on the shared ramp | unchanged — it always lands |
| Legend rule, aura/equipment SBA | unchanged | unchanged |
| A REPLACED exit | damage lost | damage kept, and readable by the replacement |

LKI is deliberately untouched: the clear sits just before
`snapshotLKILocked`, where the old one effectively did, so a
dies-trigger sees the same permanent it has always seen. Whether a dying
creature's last-known information should show the damage that killed it
is a separate question with nobody asking it.

**Regeneration does not ride on this.** CR 701.15a makes removing all
damage part of the regeneration shield's own replacement, not part of
being destroyed — which is exactly why the destroy path must not do it.
Regeneration is still unmodelled (keywords.go's canonical set does not
contain it); when it lands, it clears the damage in its own `Replace`.

**The SBA does not loop.** A permanent that survives a replaced
destruction with lethal damage still on it would be doomed again on the
next pass, so the recurrable cases were enumerated: indestructible is
filtered out of `doomed` before anything is destroyed (CR 702.12b), a
paused exit is skipped by #605's `zoneChangePausedLocked`, and an "exile
it instead" removes the permanent so there is nothing to re-doom. What
remains is a replacement that cancels a destruction outright without
regenerating, which no rules text describes and no catalog card
registers; `runStateChecksLocked`'s 32-pass bound is the backstop, and
there is a test that says so.

### 5e. Amendment, 2026-09-17: a player leaving mid-event, and CR 616.1f

*Amendment, 2026-09-17, branch `fix/808-life-continuation-leave`.
Closes [#808](https://github.com/krakenhavoc/cmd_and_ctrl/issues/808),
the post-merge review of #806 (§5b).*

§5b says the continuation runs on every TERMINAL outcome of a life
event, and §5c says the same for damage. Leaving the game was a terminal
outcome neither had:

**1. A dropped prompt finishes its event.** `cleanupStackForEliminatedLocked`
drops every prompt the departed player owed (the S31 fuzzer fix), and
that used to discard the `replacementResume` frame with it. The drain
batches run each leg from the previous leg's continuation, so dropping
the first opponent's frame dropped every later opponent's loss and the
caster's gain. `finishDroppedReplacementLocked` now settles a dropped
life or damage event:

- when the departed player is the one the event happens to (the CR 616
  chooser always is), nothing lands — CR 800.4a takes them out of the
  game — and the continuation runs with zero;
- when they only owned a CR 614.10 "may" on somebody else's event, the
  pipeline resumes and the existing gone-chooser escapes decide for them
  (the "may" is declined), and the event lands with its continuation.

Other event kinds are unchanged apart from clearing their once-per-event
entry, which nothing would clear otherwise.

**2. An eliminated player is gone** (CR 800.4a). An eliminated seat stays
in `g.Seats`, so the `playerByIDLocked == nil` checks never fired for a
player who had conceded. The drain and damage batches skip eliminated
players, the single `...ThenForEffect` entry points treat one as a no-op
with a zero continuation, and the landing functions
(`applyResolvedLifeChangeLocked`, `applyResolvedDamageToPlayerLocked`)
return `ErrPlayerEliminated` after running the continuation with zero,
which the CR 616 resume logs and moves past.

**3. CR 616.1f re-checks after each applied effect.** "Once the chosen
effect has been applied, this process is repeated (taking into account
only replacement or prevention effects that would now be applicable)."
Every path that applied several effects without a prompt — pure
cancels, identical effects, an eliminated chooser, a cost — applied the
whole gathered list back to back, so an effect the first one had switched
off still fired. Those paths now apply the first gathered effect and go
round the apply-loop again (`applyFirstGatheredLocked`). The CR 616
ordering answer keeps "one prompt, one ordering decision" but checks each
chosen effect still applies (`stillAppliesLocked`) before firing it; an
effect that no longer does is left unmarked for the re-entered loop.

**4. Undo into an open prompt replays exactly.** See the correction
under §5b: `Game.replacementsAppliedThisEvent` is deep-copied by `Clone`
and restored by `RestoreFrom`. It stays out of the persisted snapshot —
between actions it is non-empty only while a replacement prompt is open,
and `Snapshot` already refuses to call that a restore point.

**5. The iteration cap on a resume.** Every unpaused entry point lets an
event through as-is when the apply-loop hits its cap; the two resumes
(`ResolveReplacementOrder`, `ResolveOptionalReplacement`) returned the
error instead, dropping the event and its continuation on a prompt that
was already dequeued. They now tolerate it the same way.

### 5f. Amendment, 2026-09-17: the CR 903.9 resume finishes a move from any zone

*Amendment, 2026-09-17, branch `fix/707-816-battlefield-exit`.
Closes [#707](https://github.com/krakenhavoc/cmd_and_ctrl/issues/707),
noticed while fixing [#605](https://github.com/krakenhavoc/cmd_and_ctrl/issues/605)
(PR #701).*

§8 makes the commander-zone rewrite a CR 614.10 "may", and #529 (see
the note in `zone_route.go`) moved the window that offers it down into
the shared exit primitive so CR 903.9's "from ANYWHERE" holds for every
route: countered, fizzled, exiled, bounced, tucked, milled. Asking is
not the same as finishing, and one caller was still asking on its own.

`MoveCardByIDAsCommander` — the sandbox `move_card` verb, §6's second
integration point — kept its own pipeline call, on the stated grounds
that it accepts an arbitrary source AND destination, the battlefield
included, which an exit primitive has no business expressing. What it
did not keep is a continuation: `applyResolvedReplacementEventLocked`
could finish a BATTLEFIELD exit (`executeBattlefieldLeaveLocked`) or a
move carrying a `zoneRoute`, and this one was neither. So a commander
moved by hand out of a graveyard, a hand, a library or the stack paused
on the prompt and then did nothing at all — the prompt closed, the card
stayed where it was, and BOTH answers lost the move.

**Decision: the sandbox move is two verbs wearing one name, and only
one of them is an exit.** A destination of battlefield or stack is an
ENTRY — it owes enters-tapped, enters-with-counters and the ETB fire,
and none of CR 903.9's four destinations are among them — and stays
inline where it was. Every other destination IS an exit and now goes
through `routeCardToZoneLocked` like every other exit in the engine.
The resume comes with it: the `zoneRoute` frame records what the move
asked for and `ev.OldZone` records where the card was, so
`executeZoneRouteLocked` finishes it from whatever zone the window
opened over. No second resume, and the `RepEventMove` branch of the
answer path is now one general case plus the two older hand-rolled
ones (the battlefield entry and the destroy / sacrifice / SBA leave).

Three things fell out of it:

- **The per-zone details survive the pause, because the route already
  carries them.** "Library (bottom)" was a post-move reorder (ADR 0028
  §7) that ran while the paused card was still in its old zone, found
  nothing, and let the commander land on TOP when its owner declined;
  it is now `zoneRoute.ToBottom`, honoured against the settled
  destination. A card moved off the stack by hand now retires its
  `StackMeta` entry (`DropStackMeta`) instead of leaving a ghost item
  on the client's stack.
- **A stale prompt is still pruned** (#701). `pausedZoneChangeStaleLocked`
  keys on `ev.OldZone` and was already zone-general, and the exit
  primitive calls the prune on every landing, so a card that leaves by
  another route while the prompt is open takes the prompt with it —
  and an answer that races the prune is dropped with a breadcrumb
  rather than moving the card a second time out of a zone nobody asked
  about.
- **The `src` ZoneRef stays load-bearing.** The exit primitive finds
  the card by scan, so the sandbox verb checks the card is in the
  named source zone before routing: a stale client request naming a
  zone the card has already left is still `ErrCardNotFound`, not a
  move out of wherever it ended up.

What is deliberately NOT fixed here: a DISCARD does not open the window
at all (`effect_api.go` `discardPicksLocked`, `mutations.go`'s
`DiscardCards`, `pending_choice.go`'s discard leg all call `MoveCard`
raw), so a discarded commander is never offered the command zone. That
is a missing window rather than a missing resume, and every one of
those sites would have to learn to tolerate a pause — the cost path
among them, which CR 601.2h says must not pause at all (§5b).
*Closed by [§5g](#5g-amendment-2026-09-17-a-discard-is-an-exit-too).*

### 5g. Amendment, 2026-09-17: a discard is an exit too

*Amendment, 2026-09-17, branch `fix/799-853-discard-helper`.
Closes [#853](https://github.com/krakenhavoc/cmd_and_ctrl/issues/853)
and [#799](https://github.com/krakenhavoc/cmd_and_ctrl/issues/799).
This is the paragraph §5f left open.*

§5f named the last exit that did not go through the primitive and left
it there. It is fixed now, and the reason it took a second PR is the
reason §5f gave: there was no ONE place to put the window. Five sites
discarded — the CR 514.1 cleanup discard, the effect-discard
continuation (#797), the revealed-hand leg of `ResolvePendingChoice`,
the random discard, and the discard component of an additional cost —
and each moved the card with a raw `MoveCard` in its own three-line
loop. #799 folded them onto `discardCardsLocked(player, cards, opts)`
in `discard.go` first; #853 is then four lines of `zoneRoute` in one
function.

**Decision: a discard is an ordinary exit, and `zoneRoute.Discard` is
the flag that keeps its event shape.** A routed move emits one event,
never two — `EventCounterSpell` for a counter, `EventMill` for a mill,
otherwise `EventZoneMove` — because Syr Konrad and Bloodchief Ascension
watch the whole family and a second event double-counts. A discard
takes the same slot with `EventDiscardCard`, so nothing downstream sees
a discard differently than it did before the window existed.

Where it does NOT follow the mill is the redirect. `Mill` is honoured
only when the card really lands in a graveyard, because CR 701.17a
defines the keyword action by its destination; `Discard` is honoured
wherever the card lands, because CR 701.8a defines a discard by its
SOURCE — "move it from its owner's hand". A commander whose owner takes
the CR 903.9 offer was still discarded, so Megrim and Containment
Construct still see it, and the event's `NewZone` says `command` for a
listener that cares.

Three things fell out of it:

- **A discard can pause, so a multi-card discard is sequenced through
  the resume.** The card whose owner is being asked has not moved, and
  the answer arrives an action later, so the cards after it in the
  batch cannot be discarded on the next line. `zoneRoute.then` carries
  the rest of the batch — the remaining slice, one card shorter each
  time — and the last card's continuation runs the discard's own
  `Then` (#797). A value carried forward rather than a shared cursor,
  for `LoseLifeEachThenForEffect`'s reason (§5b): an undo across the
  open prompt has nothing to put back. Mill made the opposite call in
  #529 and was right to — a paused mill must not re-read the top of
  the library, so it proceeds around the paused card — but a discard
  reads a list that was fixed before the first card moved, so
  sequencing costs it nothing and buys an honest "then".
- **The cost site settles instead of asking.** `zoneRoute.MustSettleNow`
  sets the event's `mustSettleNow`, so the CR 903.9 "may" is skipped
  un-applied and a commander pitched to an additional cost goes to the
  graveyard. Weaker than printed, never stronger; the same posture and
  the same rule (CR 601.2h) as `payLifeAsCostLocked` (§5b), which is
  the other half of the same cost line.
- **The undo snapshot needs its own copy of the route.**
  `cloneReplacementResume` shared the `zoneRoute` on the stated grounds
  that it is written once and only read afterwards. That stopped being
  true the moment it carried a continuation: `runRouteTailLocked` clears
  `then` THROUGH the pointer, exactly as `runDamageTailLocked` does
  (§5c), so a shared struct would let the live game's answer consume
  the snapshot's continuation and a replayed answer would move the card
  and skip the rest of the discard.

What is still NOT fixed: the event a discard pushes through the window
is a plain `RepEventMove` with no CAUSE on it. The destination is
replaceable — which is all CR 903.9 needs — but "if you would discard a
card" is not, so Library of Leng and madness still wait on #650 and
§10a still stands. No catalog replacement fires on a discard today
except the built-in: every `RepEventMove` watcher in the catalog gates
on `OldZone == ZoneBattlefield` or `NewZone == ZoneBattlefield`.

### 5h. Amendment, 2026-09-17: a "may" inside a chosen order is still a question

*Amendment, 2026-09-17, branch `fix/815-847-replacement-outcomes`.
Closes [#847](https://github.com/krakenhavoc/cmd_and_ctrl/issues/847),
noticed by the #802/#801 agent in PR #845 and not fixed there.*

§5a's third limit says an effect that asks its controller a question is
never collapsed, "because skipping the ordering prompt would skip its
question too". That was true of the two paths §5a was about. It was not
true of the two paths that apply several effects once the prompt has
been answered or ruled out, and both of them answered a CR 614.10 "may"
on its controller's behalf, in the direction that favours it.

**1. The chosen-order loop asks.** `ResolveReplacementOrder` fires the
whole submitted order in one pass (§5, "one prompt = one ordering
decision"), and it already refused to fire a copy selector or a
shockland's pay-life blind — each queues its own prompt and bails, and
the effects later in the order are left unapplied for the apply-loop
re-entry to pick up. `Optional` had no such branch, so a "may" that
shared a window with any other effect simply happened. Now it takes the
identical branch, through `offerOptionalReplacementLocked` — the same
helper the single-effect path in `applyReplacementsLocked` now uses, so
there is one place that knows when the question is asked and when it is
declined inline (chooser gone, or an entry with nothing to resume it —
§5b's "weaker, never stronger", #359).

The resume is the existing one: `ResolveOptionalReplacement` fires or
skips, then re-enters the apply-loop, which re-gathers. No second
resume and no new frame — the effects after the "may" in the chosen
order are unapplied, so the gather finds them exactly as it finds the
ones after a copy selector. The cost of that, shared with the two older
branches, is that an order of THREE or more whose middle effect pauses
offers the remaining two as a fresh ordering prompt rather than
remembering the tail of the first answer. CR 616.1 re-chooses after
every applied effect anyway, so re-asking is the rules-faithful
direction; it is recorded here because "one prompt, one ordering
decision" is the thing it bends.

**2. The eliminated-chooser fallback drops the question.**
§5b gave `mustSettleNow` a fallback that applies the gathered order
through `skipQuestionsLocked` — anything that would ask its own
question is skipped un-applied rather than fired. The gone-chooser
fallback beside it (§5, the S31 fuzzer finding) passed the gathered
list straight to `applyFirstGatheredLocked`, so the same "may" that a
cost would have skipped was fired when the affected player had left the
table. It routes through `skipQuestionsLocked` too now. The two
fallbacks are the same sentence: an order nobody chose cannot carry an
effect whose answer nobody gave.

A "may" whose own controller is still seated is declined here even
though they could have answered it, because the event they would be
answering about belongs to a player who is gone (CR 800.4a) and the
engine has no way to sequence one question inside an order it is
applying un-prompted. Weaker than printed, never stronger — the posture
every other un-prompted path in this ADR takes.

Nothing new is snapshotted: the "may" pauses on the same
`PendingChoiceOptionalReplacement` and the same `replacementResume`
frame it has used since S17, which `ContinuationCensus.ChoiceResumeFrames`
already counts and `cloneReplacementResume` already copies. The undo
contract from §5b/§5e carries over unchanged — rewinding into the
prompt and answering again lands the same event, because the ordering
answer's applied-effect marks live in `Game.replacementsAppliedThisEvent`,
which `Clone` deep-copies.

### 5i. Amendment, 2026-09-17: "destroyed this way" counts what was destroyed

*Amendment, 2026-09-17, branch `fix/815-847-replacement-outcomes`.
Closes [#815](https://github.com/krakenhavoc/cmd_and_ctrl/issues/815),
noticed by the #708/#807 agent in PR #813 and not fixed there.*

§5d moved the marked-damage clear onto the landed outcome of a
destruction. This is the same move for the destruction's NUMBER.

`destroyPermanentsLocked` counted a leg as destroyed whenever
`routeBattlefieldCardToOwnerGraveyardLocked` returned no error, and
that call returns no error for three things that are not a
destruction: one the CR 614 window cancelled outright (indestructible
granted mid-window, "it isn't destroyed instead"), one a replacement
sent somewhere other than a graveyard, and one that has merely PAUSED
on the CR 903.9 prompt and has not happened yet. Every card that reads
"for each creature destroyed this way" — Fumigate, Deadly Tempest,
Bane of Progress, Blood Money, The Battle of Bywater — was paid for
permanents that were still standing.

**1. The rule is CR 701.7a, read off the landed outcome.** "To destroy
a permanent, move it from the battlefield to its owner's graveyard."
So `destroyedThisWayLocked` asks where the permanent actually ended up:

| Where it landed | Destroyed? |
| --- | --- |
| a graveyard | yes |
| the command zone | yes — CR 903.9, see below |
| exile, hand, library (a replacement rewrote the destination) | no |
| still on the battlefield (the window cancelled it) | no |

The exile row is the judgement, and it is the one the issue left open.
A permanent that an "exile it instead" replacement removed really did
leave the battlefield, and it is tempting to call that destroyed
because the destruction is what moved it. CR 701.7a names the
graveyard, so the engine does not: it was never put into one, so it
was not destroyed. The consequence is deliberate and small —
Fumigate gains nothing for it, Deadly Tempest charges nobody for it,
and a "whenever a creature dies" trigger never saw it either, because
it did not die.

The command-zone row is a **declared carry-over**, not something
CR 701.7a compels: by the letter of the same sentence a commander put
into the command zone was not put into a graveyard either. The engine
has counted it since S23 on the stated grounds that "CR 903.9 replaces
the zone change, not the destruction" (simultaneous.go's header), the
catalog's notes say so on the cards, and changing it would quietly
change every wrath payoff's number for commanders. It stays, in ONE
place — `destroyedThisWayLocked` — which is where to change it if it
ever flips.

The outcome is read off the LIVE BOARD in the continuation rather than
off the settled event, because that is the reading that is still true
after an undo rewinds into an open CR 903.9 prompt and the answer is
replayed. A captured event pointer is the snapshot's problem (§5b);
the board is not.

**2. It is a continuation, because any leg can pause.** Same idiom,
same names as §5b and §5c: `DestroyPermanentsThenForEffect(ids,
then(destroyed []uuid.UUID))` beside the fire-and-forget
`DestroyPermanentsForEffect(ids) int`. The legs are destroyed IN
SEQUENCE, each from the previous one's continuation, with the landed
list carried forward BY VALUE — the property that makes an undo across
the prompt replay identically, for the reason
`LoseLifeEachThenForEffect` gives.

The continuation rides `zoneRoute.then`, the frame #856 added for the
discard batch, so there is one continuation idiom on the exit side and
not two. The destroy / sacrifice / SBA route still performs its own
move (§5f says why folding it into `routeCardToZoneLocked` is a change
of its own), so the route it now carries is flagged
`ViaBattlefieldLeave` and exists only to hold `then`:
`routeBattlefieldExitThenLocked` puts it on the event and
`finishBattlefieldLeaveLocked` runs it from the one place both the
unpaused path and the CR 903.9 resume land in.

**3. The pause does not break the simultaneity batch.** Sequencing a
batch through a prompt would normally close the batch on the way out
and re-open it — or not — on the far side, which is exactly what
simultaneous.go exists to prevent: the copies are taken BEFORE the
first move, and a batch re-taken afterwards has lost the creatures
that already left. So the copies are taken once and carried forward in
the continuation with the landed list (`simultaneousExitSnapshotLocked`
/ `publishSimultaneousExitLocked`), and every leg re-publishes the same
ones. A Blood Artist still sees the whole board die with it, on either
side of a commander's prompt.

**4. What can now pause that could not before.** A `DestroyAllMatching`
with a `Then` clause — and only one with a `Then`; the fire-and-forget
sweeps are untouched. Five cards: Fumigate, Deadly Tempest, Bane of
Progress, Blood Money, The Battle of Bywater. When a commander is
caught in one of those wipes, the rest of the sweep AND the "for each"
clause now happen when its owner answers CR 903.9, an action later,
instead of happening immediately with the commander counted on the
strength of the prompt having been queued. The observable cost is the
one the drain (§5b), the damage batch (§5c) and the discard batch
(§5g) already pay, and it buys the only thing that can report a true
number.

`Then` also receives the swept CARDS narrowed to the ones that were
destroyed, so the clause that reads the cards (Deadly Tempest's
per-controller tally, The Battle of Bywater's "creatures you control")
and the clause that reads the count can no longer disagree. That was
already the stated intent of the indestructible filter in mass.go; it
is now true of the permanents the window takes away mid-sweep too.

**5. The fire-and-forget count and the SBA flag.**
`DestroyPermanentsForEffect` keeps its `int` and it now means the same
thing — how many of the legs that SETTLED landed as destructions —
which is every leg unless one paused. A card that reads the number
must use the `Then` form; that is what its doc comment says and what
AGENTS.md §7 says. The state-based-action sweep never wanted a
destruction count: its question is "did this pass do anything", and it
now answers that from the `doomed` set it collected rather than from a
number that deliberately excludes a cancelled destruction the sweep
still has to look at again.

**Not fixed here**, and listed so the next reader does not assume it
was: `ExileCardsForEffect` and `BounceCardsToHandForEffect`
(simultaneous.go) count a paused or cancelled leg exactly the way the
destroy sweep used to, and a paused exit whose prompt is DROPPED —
its chooser left the game (`finishDroppedReplacementLocked`,
mutations.go) or it went stale (`dropStaleReplacementResumeLocked`,
pending_choice.go) — never runs its route's continuation, so a
sequenced batch behind it stalls. That hole is #853's as much as this
one's; both are filed separately. (Both are now closed — §5j and §5k.)

### 5j. Amendment, 2026-09-17: a prompt taken away is still an outcome

*Amendment, 2026-09-17, branch `fix/865-866-route-tail`.
Closes [#865](https://github.com/krakenhavoc/cmd_and_ctrl/issues/865),
filed by the #847/#815 agent in PR #863 and listed as not-fixed at the
end of §5i.*

§5g gave an exit a continuation (`zoneRoute.then`) and §5i put a
sequenced batch on it. Both assumed a paused prompt is eventually
ANSWERED. Two things take a prompt away instead, and both simply threw
the frame out:

- **dropped** — its chooser left the game, so nobody can answer it
  (`cleanupStackForEliminatedLocked` → `finishDroppedReplacementLocked`,
  mutations.go, CR 800.4a);
- **pruned** — the card moved by some other route while the question
  was open, so the move the prompt asks about can never happen
  (`pruneStaleZoneChangeChoicesLocked`, #605/#701, and its answer-path
  twin `dropStaleReplacementResumeLocked`).

The continuation went out with the frame. A two-card discard stopped
after the first card, a `DestroyAllMatching` with a `Then` stopped
after the commander, and the caller's own clause — "then draw two",
"for each creature destroyed this way" — never ran at all. The batch
did not fail; it waited forever.

**1. Abandoning is a terminal outcome, and has one function.** #808
made "the player left" a terminal outcome of the life and damage tails
for exactly this reason, and this is the same move for the exit:
`abandonZoneRouteLocked(frame)` (zone_route.go) releases the CR 614.5
once-per-event entry and runs `runRouteTailLocked`. All three sites
call it; none of them has a copy of the reasoning.

**2. Nothing moves, and the leg reports itself as not landed.** A
paused route has moved nothing — the card is still in its old zone and
no event has been emitted (`routeCardToZoneLocked`) — so an abandoned
route is shaped exactly like a CR 614.10 cancellation: no
`EventDiscardCard`, no `EventZoneMove`, no `EventLTB`. It follows that
the leg is not counted, and it follows without a flag: every reader of
a route's outcome reads the LIVE BOARD (`destroyedThisWayLocked`,
`landedInZoneLocked` — §5i, §5k), and the board still has the card
where it was. A commander whose owner conceded mid-prompt is not
"destroyed this way", and a discard that never left the hand did not
happen.

**3. Every other exit of the frame, too.** Once "the tail runs at every
terminal outcome" is the rule, the exceptions are bugs waiting to be
found rather than decisions. `routeCardToZoneLocked` and
`executeZoneRouteLocked` now run the tail from a deferred terminal
outcome covering every return — a card that is nowhere, a destination
that cannot be resolved, a failed `MoveCard`, the apply-loop erroring —
rather than from the two branches that remembered to. The pause is the
one exit that is deliberately not terminal, and it is the one the
defer skips. `finishBattlefieldLeaveLocked` already made this call for
the other exit ("the tail runs even when the move failed"); this is
that rule, applied to the shared primitive.

**4. The prune sweeps before it runs anything.** A tail may queue the
next leg's prompt, and when that leg lands it re-enters
`pruneStaleZoneChangeChoicesLocked` through `executeZoneRouteLocked`.
So the prune collects its frames and drops every one of them from the
queue FIRST, then runs the tails — the shape
`cleanupStackForEliminatedLocked` already used for the dropped frames
of #808, and for the same reason: walking a slice by index while a
callee mutates it is the bug this avoids.

**5. What this does not change.** Nothing about who is asked, when, or
what an answered prompt does. The only observable difference is that a
batch behind an abandoned prompt now finishes, with the abandoned leg
counted as nothing.

### 5k. Amendment, 2026-09-17: one batch body, and "this way" means arrived

*Amendment, 2026-09-17, branch `fix/865-866-route-tail`.
Closes [#866](https://github.com/krakenhavoc/cmd_and_ctrl/issues/866),
filed by the #847/#815 agent in PR #863 and listed as not-fixed at the
end of §5i.*

§5i fixed the destroy sweep's count and left the exile and bounce
sweeps counting the way it used to: a leg the CR 614 window had
cancelled, and a leg that had merely PAUSED on the CR 903.9 prompt,
were both counted as having moved. Settle the Wreckage
(`b29ExileAttackersThenTheyFetchBasics`) read one of those numbers
synchronously, so it handed its victim a basic land for every attacking
creature whose owner had been *asked* about the command zone.

**1. The rule for these two is CR 400.7, not CR 701.7a.** A card that
changes zones becomes a new object in the zone it arrives in, so "for
each card exiled this way" means the cards that reached EXILE.
`landedInZoneLocked` (simultaneous.go) asks exactly that, against the
destination the route requested:

| Where it landed | Exiled this way? |
| --- | --- |
| exile | yes |
| the command zone (CR 903.9 took the offer) | **no** |
| anywhere else (a replacement rewrote the destination) | no |
| still where it was (cancelled, or the prompt abandoned — §5j) | no |

The command-zone row is where this parts company with destroy, and the
difference is not an inconsistency. §5i counts a commander as destroyed
because CR 903.9 replaces the zone change and not the destruction — the
permanent was still destroyed. Nothing correspondingly replaces the
fact that an exile put the card in exile: it did not, so it was not
exiled this way. The two rules live one function apart
(`destroyedThisWayLocked`, `landedInZoneLocked`) and
`routeLegLandedLocked` picks between them.

**2. One body, not three.** Writing §5i's sequencing twice more would
have been three copies of one loop. It is now
`routeAllThenLocked(route, ids, then(landed))` plus its fire-and-forget
twin `routeAllLocked(route, ids) int`, parameterised by a `zoneRoute`
TEMPLATE — the description of what each leg is — and three small
switches on it: which mover the leg takes (`routeLegLocked`), what
counts as nothing to do (`routeLegNothingToDoLocked`), and what counts
as landed (`routeLegLandedLocked`). The destroy sequencing from §5i was
rebased onto it and `destroyEachStepLocked` is gone; the destroy
fire-and-forget loop the SBA sweep shares went the same way. The
templates are `destroyRoute`, `exileRoute`, `bounceRoute`, and the
destroy one carries no destination at all — `ViaBattlefieldLeave` says
the move belongs to `executeBattlefieldLeaveLocked`, which names its
own (§5i.2, §5f).

A leg with nothing to do is now SKIPPED rather than routed — a card
that is not where the route expects it, or is already at the
destination. That was true of the destroy sweep's sequenced form
already; extending it is what stops a batch from opening a CR 614
window (and a commander's prompt) for a move that cannot happen.

**3. The pause keeps the batch open on this side too.** `zoneRoute`
already carried `simultaneousExit` for §5i's destroy legs;
`executeZoneRouteLocked` now publishes it around the move and the
continuation the way `finishBattlefieldLeaveLocked` does, so an exile
or bounce batch that stops for a CR 903.9 answer is still one event to
the watchers in it.

**4. What can now pause that could not before.** An `ExileAllMatching`
or `BounceAllMatching` with a `Then`, and `ReturnAllToHand` with one —
and only those; every fire-and-forget sweep in the catalog (Farewell,
Merciless Eviction, Evacuation, Cyclonic Rift, River's Rebuke, Whelming
Wave, Wash Out, Aetherize, Aether Gale, Selective Obliteration,
Desynchronization, Wave Goodbye, Perplexing Test) is untouched. One
card reads such a count today: **Settle the Wreckage**. When an
attacking commander is caught in it, the rest of the sweep and the
"search for that many basic lands" now happen when its owner answers
CR 903.9, an action later, instead of happening immediately with the
commander counted on the strength of the prompt having been queued.

**5. The fire-and-forget counts.** `ExileCardsForEffect` and
`BounceCardsToHandForEffect` keep their `int` and their signatures, and
the number now means what `DestroyPermanentsForEffect`'s has meant
since §5i: how many of the legs that SETTLED landed where the route
asked. A cancelled leg is not in it; a paused leg cannot be, which is
what the `Then` forms are for.

*#870: the SINGLE-CARD read-back joins this body rather than keeping
its own — `ExileCardThenForEffect` is a wrapper over
`ExileCardsThenForEffect` with a batch of one, so one card gets the
same CR 400.7 answer and the same replay-under-undo as a sweep.*

### 5l. Amendment, 2026-09-18: a mill counts what landed

*Amendment, 2026-09-18, branch `fix/893-894-mill-and-living-death`.
Closes [#893](https://github.com/krakenhavoc/cmd_and_ctrl/issues/893),
filed by the #870/#877 agent in PR #878.*

§5i fixed the destroy sweep's number and §5k fixed exile's and
bounce's. The mill was the fourth verb with a "this way" clause on it
and the one nobody had looked at: `MillToZoneForEffect` returned every
leg that had not PAUSED, so the list it handed back held a card the
CR 614 window had cancelled, a card a replacement had sent somewhere
else, and — because a pause is the one thing it did check — nothing at
all for the leg that mattered most. Oona, Queen of the Fae reads that
list as "for each card of the chosen color exiled this way" and made a
Faerie for a commander whose owner was still being ASKED about the
command zone.

**1. The rule is CR 400.7, read against the destination the mill asked
for.** Not a new rule and not a new function: `landedInZoneLocked` is
the one §5k wrote, and the mill asks it the same question exile and
bounce do.

| Where it landed | Milled this way? |
| --- | --- |
| the destination the mill named (a graveyard, or exile for "exile the top N") | yes |
| the command zone (CR 903.9 took the offer) | no |
| anywhere else — "if a card would be put into a graveyard from anywhere, exile it instead" | no |
| still on the library (cancelled, or the prompt abandoned — §5j) | no |

The event shape already agreed with this and had since #529:
`zoneRoute.Mill` is honoured only when the card really reaches a
graveyard, because CR 701.17a defines the keyword action by its
destination. What did not agree was the number the caller was handed.
Now one answer serves both.

**2. One body, and no second mill.** The mill joins
`routeAllThenLocked` with a fourth template, `millRoute(player, dest)`
— a function rather than a package var because two things about a mill
belong to the caller: the destination (a graveyard for CR 701.17a's
mill, exile for "exile the top N cards of your library", which is not a
mill) and the Actor, which is the player whose library is being read
and not the card's owner.

`MillToZoneThenForEffect(player, n, dest, until, then)` is the
continuation form, `MillToZoneForEffect` keeps its signature and its
slice, and both plan the mill through `millPlanLocked` so they cannot
disagree about what a mill of n is. The fire-and-forget form needed the
IDs rather than a count, so `routeAllLocked`'s loop became
`routeAllLandedLocked` and the count is its length — one loop, not two.

**3. `until` is answered before the first move, and that is exact.**
Helm of Obedience's "mills until a creature card is put into their
graveyard" is a predicate on the card that came OFF THE LIBRARY, not on
where that card ended up, and the batch was already chosen up front
(#529). So the run is truncated in the plan, which is what lets the
plan be a plain list of IDs — the only thing the shared body takes —
and it is identical to the old post-move loop card for card.

Its one deviation is now declared on the card: a commander that takes
the command zone was never put into a graveyard, so by CR 701.17a's
letter the Helm should keep milling, and it stops instead. That is the
caveat Helm of Obedience carries, rewritten to say so.

**4. The two forms differ in one thing, and #529 chose it.** A paused
mill must not re-read the top of the library, and the plan-up-front
answers that for both forms. What the FIRE-AND-FORGET form keeps is
#529's second half: it proceeds AROUND the paused card, milling the
rest on this line, because nothing is waiting on its answer. The `Then`
form cannot — a list that is still being decided is not a list — so it
sequences, and the rest of the mill happens when the prompt is
answered. That is the same trade §5i's destroy batch and §5g's discard
batch make, and it is the only version that can report a true list.

**5. What can now pause that could not before.** Five catalog readers,
and only readers; every fire-and-forget mill in the catalog is
untouched.

- **Oona, Queen of the Fae** — the issue. The Faeries wait for the
  answer and a commander that takes the command zone pays for none.
- **Helm of Obedience** — "a creature card put into their graveyard" is
  now read after the answer, so a commander that goes to the graveyard
  IS reanimated. It used to read the list with the prompt open, find
  nothing, and drop its own second half.
- **Sphinx's Tutelage** — the repeat is recursion through the
  continuation rather than a loop, so "if two nonland cards that share
  a color were milled this way" is asked about cards that arrived.
- **Dread Summons** — the seats are milled in sequence, each from the
  previous one's continuation, with the creature tally carried forward
  BY VALUE for the reason `LoseLifeEachThenForEffect` gives (§5b).
- **Oblivion Sower** — it walks the exile zone rather than the returned
  slice, which is the same read one line too early.

**6. Nothing new is snapshotted.** The mill rides `zoneRoute.then` and
the `replacementResume` frame every other exit has used since §5g, and
`cloneReplacementResume` already gives an undo snapshot its own copy of
the route. The undo contract is the one §5k's exile batch signs, and
`milled_this_way_test.go` pins it: rewind into the open prompt, answer
again, and the same cards are milled and the same list reported,
because the landed list is carried forward by value.

### 5m. Amendment, 2026-09-18: a card that exiles and then uses the card waits

*Amendment, 2026-09-18, branch `fix/893-894-mill-and-living-death`.
Closes [#894](https://github.com/krakenhavoc/cmd_and_ctrl/issues/894),
filed by the #870/#877 agent in PR #878.*

§5k gave the batch a continuation and #870 gave the single card one.
This is the two catalog callers that still wrote the next line instead
of handing one over, and the first of them lost a card doing it.

**1. Living Death, and the only bug in this pair that changes an
outcome.** "Each player exiles all creature cards from their graveyard,
then sacrifices all creatures they control, then puts all cards they
exiled this way onto the battlefield." The three passes were three
loops, and the first loop's legs can PAUSE: a commander card in any
graveyard asks its owner about the command zone. The swap did not wait.
It sacrificed, it reanimated what was in exile at that moment, and it
finished — and the commander then landed in exile with the step that
would have returned it already over. Stranded, with nothing in the game
that could ever move it again.

Now the exile is ONE batch (`ExileCardsThenForEffect`) and the other
two passes are its continuation. Three things fall out of it:

- **The reanimated set is the LANDED list**, which is precisely what
  "all cards they exiled this way" means (CR 400.7, §5k). A commander
  that takes the command zone is not in it and does not come back —
  the right answer, where the old loop had no answer at all.
- **Both remaining passes share ONE continuation**, because a
  battlefield ENTRY cannot pause:
  `ReturnFromExileToBattlefieldForEffect` consults the pipeline before
  the card leaves exile and DROPS the return if a CR 616 prompt is
  queued rather than waiting on it. If that ever changes, the
  reanimation half needs a continuation of its own.
- **Both sides are gathered before anything moves**, the graveyard side
  in particular. The creatures sacrificed in pass 2 land in graveyards,
  and a set re-gathered on the far side of the exile would reanimate
  them too — the trick the card is built on, and it is now protected by
  a value carried into the continuation rather than by the passes
  happening to run back to back.

The sacrifice pass is still per-card and fire-and-forget. A commander
sacrificed there is asked its own CR 903.9 question and lands a beat
later, which strands nothing — the card is where it was until the
answer — and nothing in Living Death reads the sacrifice. There is no
batch-with-continuation form of sacrifice to use; when one exists this
is a caller for it.

**2. Path to Exile, and the reason it is here anyway.** The outcome was
always right: "Exile target creature. Its controller may search their
library for a basic land card." The search ran on the next line, so a
commander's owner was offered the search while their own command-zone
question was still open — two prompts at once, in the wrong order, at a
table where the second one is a reasonable thing to answer first. It
moves into `ExileTarget.Then` and the `exiled` answer is deliberately
IGNORED: the search is a separate sentence, not an "if you do", so it
is offered whichever way the question is answered. Only the ORDER
changes, which is the whole of the fix and the whole of its risk.

**3. The `Flicker` primitive is Living Death with one card.** "Exile
it, then return it" was the same two lines, and a flickered commander
was asked about the command zone, had its return run against an empty
exile, and stayed in exile for good. It is `ExileTarget.Then` now,
gated on `exiled` this time — a commander that takes the command zone
stays there, which is what CR 903.9 says happens and not a card lost.
One catalog card uses it today (Y'shtola Rhul); every future one gets
the fix for free, which is the argument for putting it in the primitive
rather than in the card.

**4. What can now pause where it could not before.** Living Death's
sacrifice and reanimation halves, Path to Exile's search, and any
`Flicker`. In each case the delay is one action and it happens only
when a commander is actually caught in the effect.

**5. What was deliberately NOT converted**, so the next reader does not
assume it was missed:

- **The mass exiles with no `Then`** — Farewell's up-to-four halves,
  Merciless Eviction, Selective Obliteration. §5k already declared the
  fire-and-forget sweeps untouched; no half reads another's outcome and
  nothing is stranded, so what a paused leg costs them is interleaving,
  not correctness.
- **Exile-then-a-silent-unconditional-clause** — Swords to Plowshares,
  Solitude, Resculpt, Anguished Unmaking, Ashes to Ashes, Cemetery
  Reaper, Heritage Reclamation, Patron of the Vein, Deathrite Shaman's
  two graveyard modes. The clause is not gated on the exile and asks
  nobody anything, so the order is unobservable.
- **Exile-then-a-clause-about-the-CARD** — Cling to Dust, Scavenging
  Ooze, Deluge of the Dead ("if it was a creature card"). The condition
  reads the card's type, which is known before the move, and CR 903.9
  is a "may" the clause does not mention. Gating these on the landed
  outcome is a rules question rather than a plumbing one, and it is not
  this issue's.
- **Deathrite Shaman's mana mode.** Its target is a land card in a
  graveyard, which the CR 903.9 built-in can only fire for if a land is
  flagged as a commander, and the follow-up is MANA — delaying that
  behind a prompt in the middle of paying for something is a change
  with a bigger blast radius than the ordering it would fix.

**6. Nothing new is snapshotted.** Both cards ride
`ExileCardsThenForEffect` / `ExileCardThenForEffect`, which ride
`zoneRoute.then` and the `replacementResume` frame from §5g. The undo
contract is §5k's and `paused_exile_continuations_test.go` pins it on
the card: rewind into the open prompt, answer the other way, and the
board follows that answer.

### 5n. Amendment, 2026-09-18: a tuck waits for the CR 903.9 answer

*Amendment, 2026-09-18, branch `fix/783-478-tuck-and-fetch-pause`.
Closes [#783](https://github.com/krakenhavoc/cmd_and_ctrl/issues/783),
filed while reviewing PR #773.*

A library is a CR 903.9 destination. §5f made a tuck pausable and §5m
made the exile's callers wait; the tuck's callers were still writing the
next instruction on the next line, because `TuckToLibraryForEffect`
returns `nil` whether the card moved or a prompt was queued. Three
catalog cards read straight past the question.

**1. Chaos Warp, and CR 608.2c.** "The owner of target permanent
shuffles it into their library, then reveals the top card of their
library. If it's a permanent card, they put it onto the battlefield."
The instructions run in order, and the shuffle-in cannot be finished
while it is still a question. The old code shuffled and revealed with
the commander still on the battlefield; an owner who then declined had
their commander land on **top** of the already-shuffled library rather
than shuffled into it. That was the card's declared caveat and the only
reason it was not `full`. The shuffle and the reveal are now the tuck's
continuation, and the answer is deliberately **ignored**: the shuffle is
one sentence about the library, not an "if you do", so a commander that
goes to the command zone instead still leaves its owner shuffling and
revealing. Chaos Warp is `full`.

**2. The God-Eternals, and a positioned landing.** "Put it into its
owner's library third from the top" was a tuck to the TOP followed by a
remove-and-reinsert on the next line. With the prompt open the card was
still in the graveyard or in exile, so the reinsert's `Remove` returned
`ErrCardNotFound` and the trigger logged an `EventEffectError`; when the
owner then declined, the God-Eternal landed on top. God-Eternals are
commonly commanders, so this was a normal path rather than a corner.
`TuckToLibraryAtDepthForEffect` already existed (§5f's `zoneRoute.Depth`
rides the route and is applied against the SETTLED destination);
`b22TuckThirdFromTop` simply predated it. The reposition is now a
**positioned landing** — the card is placed once, where the card says —
and no continuation is needed, because nothing follows the tuck.

**3. Aetherspouts counts what landed.** Each owner's "top or bottom for
each" is asked as a scry over the cards that reached their library, and
the count included an attacking commander whose tuck had merely PAUSED.
The owner scried a library card they had no right to look at, and the
commander landed on top afterwards, never ordered. The scry is now the
batch's continuation over the LANDED list — CR 400.7's reading, §5k's.

**4. Sylvan Library's chain.** The put-back leg tucked and then asked
the next card's question on the next line, so a drawn commander put back
was asked about the command zone while the second "pay 4 life or put it
back" was already on the table. The next link now hangs off the tuck's
continuation, for the same reason the first link hangs off the answer
before it.

**5. One route template, two halves.** Rather than a fifth copy of the
sequencing, the tuck is a `zoneRoute` template beside `destroyRoute`,
`exileRoute`, `bounceRoute` and `millRoute` (§5k, §5l), with `TuckOptions`
naming the position (top, bottom, Nth from the top). Both halves are
built from it: `TuckCardsToLibraryThenForEffect` /
`TuckToLibraryThenForEffect` sequence through `routeAllThenLocked`, and
the fire-and-forget `TuckToLibraryForEffect` /
`TuckToLibraryAtDepthForEffect` keep their signatures and their single
`routeCardToZoneLocked` call. The single-card `Then` form is a wrapper
over the batch, exactly as `ExileCardThenForEffect` is (#870).

**6. What was deliberately NOT converted.** Sensei's Divining Top
("draw a card, then put this artifact on top of its owner's library")
and Mistveil Plains — in both the tuck is the LAST instruction, so
fire-and-forget is the right form and a pause costs nothing. Teferi,
Hero of Dominaria was already on the at-depth entry point.

**7. Nothing new is snapshotted.** The tuck rides `zoneRoute.then` and
the `replacementResume` frame from §5g, with §5k's undo contract;
`paused_tuck_continuation_test.go` (engine) and
`paused_tuck_continuations_test.go` (catalog) pin it.

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
  Depends on #651. *Partly delivered by #799/#853, see
  [§5g](#5g-amendment-2026-09-17-a-discard-is-an-exit-too): the single
  route, the up-front random set, the CR 903.9 prompt and the cost
  case are done. What #650 still owes is the CAUSE on the event.*
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
