# ADR 0086 — Storm, and the turn's cast order

**Status:** Accepted · 2026-09-23 · S46 — modal spells, multi-target
clauses and copy effects · tracked on
[#1238](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1238),
sprint trackers
[#888](https://github.com/krakenhavoc/cmd_and_ctrl/issues/888) (modal
spells, multi-target clauses and copy effects) and
[#884](https://github.com/krakenhavoc/cmd_and_ctrl/issues/884) (turn
machinery and per-turn accounting).

**Numbering:** on 2026-09-23, every remote head was checked with the
AGENTS.md §4 loop — `git ls-remote --heads origin` returned 405
branches, and `git log --all --diff-filter=A --name-only -- docs/decisions/`
over the fetched refs lists 87 ADR files, the highest of which is
`0085-life-total-cant-change.md` (on `develop`, merged as #1254).
No branch anywhere holds 0086 or higher. 0086 is the first free
number.

**Builds on:** [ADR 0059](0059-turn-machinery.md) (the per-turn tally
and its reset at the real turn boundary); [ADR 0043](0043-copy-effects.md)
as amended 2026-09-18 (#920, a spell copying itself), 2026-09-22
(#1223, the shared copy frame) and by #1206 / #1235 (`copyFrame`,
`resolveCopyTargetsLocked`, `createCopyLocked`);
[ADR 0065](0065-modal-and-multi-target-clauses.md) §2 (target slots
live on the `TargetRef`) and its copy-retarget note;
[ADR 0018](0018-triggers-on-the-stack.md) (a triggered ability uses
the stack and can be answered); [ADR 0033](0033-ai-bot-seat.md) §1
(the enumerator answers every blocking prompt) and §4 (the public
log).

**Amends:** nothing. It adds one field to `game.TurnTally`, one read,
one event kind, one log kind and one catalog constructor.

---

## The rule, and the number it is filed under

The pinned Comprehensive Rules for this tree are the edition effective
**August 7, 2026** (`MagicCompRules 20260819.txt`, AGENTS.md §6). In
that edition storm is **CR 702.40**, not 702.39:

> **702.40a** Storm is a triggered ability that functions on the
> stack. "Storm" means "When you cast this spell, copy it for each
> other spell that was cast before it this turn. If the spell has any
> targets, you may choose new targets for any of the copies."
>
> **702.40b** If a spell has multiple instances of storm, each
> triggers separately.

CR 702.39 in this edition is **provoke**. Issue #1238 and the roadmap
comments that fed it cite 702.39 from an older edition; every citation
this ADR and the code it describes make is 702.40, checked against the
pinned text as AGENTS.md §6 requires. There is no 702.40c: the
"a countered storm spell still has its trigger" fact people file under
that number is **CR 113.7a** ("once activated or triggered, an ability
exists on the stack independently of its source"), and it is cited
that way below.

Three readings of 702.40a do the work:

1. **"each other spell"** — every player's, not just the storm
   spell's controller's, and not the storm spell itself.
2. **"that was cast"** — a spell that was cast and then countered was
   still cast. The count does not care where the card is now.
3. **"before it"** — the count is about this spell's position in the
   turn's casts, which is a question about an *order*, not a total.

## Context

`game.Game` already keeps two per-turn cast records and neither can
answer (1)–(3) together:

- `SpellsCastThisTurn map[uuid.UUID]CastTally` (`game.go:182`,
  bumped in `CastSpell`, `mutations.go:1306`) is `{Total,
  Noncreature}` **per player**. It answers "is this your first
  noncreature spell this turn" and nothing about order or about the
  table.
- `TurnTally` (`turn_tally.go`, ADR 0059 / #586) counts what has
  happened this turn and slices `EventsThisTurn()` for the filtered
  long tail. Its `EventCast` arm bumps no counter at all today — it is
  there only as a decision notch for the CR 726 loop breaker
  (`turn_tally.go:648`).

The catalog's one existing storm-shaped card, **Thousand-Year
Storm** (`thousand_year_storm.go`), scans `EventsThisTurn()` through
`b12InstantsAndSorceriesCastBeforeThisTurn`
(`batch12_helpers.go:139`), which **looks each cast card up** to test
whether it is an instant or a sorcery. That is correct for
Thousand-Year Storm, whose text has a type filter, and it is honest
about its own weakness in its doc comment ("weaker, never stronger: a
spell that has since left every tracked zone is not counted"). It is
wrong for storm, which has no type filter and must not undercount a
countered spell.

## Decisions

### 1. The turn tally records the ORDER of this turn's casts

```go
type TurnTally struct {
    // …
    // Casts is every spell cast this turn, table-wide, in cast order.
    Casts []uuid.UUID `json:"casts,omitempty"`
}

// SpellsCastBeforeThisTurn reports how many spells were cast this
// turn, by any player, BEFORE the spell `spellID` was cast
// (CR 702.40a). Caller must hold g.mu.
func (g *Game) SpellsCastBeforeThisTurn(spellID uuid.UUID) int
```

Appended by the **existing** `EventCast` arm of `turnTallyListener`,
which is the one place the engine already reacts to a cast and which
runs ahead of the trigger harvester — so the storm spell's own entry
is in the list before its storm trigger is built. Reset by
`resetTurnTallyLocked` with everything else in the tally, at the real
turn boundary (ADR 0059, #1009), and deep-copied by `cloneTurnTally`,
which is what carries it through `Clone` / `RestoreFrom` (undo) and
through the snapshot, where `TurnTally` rides as one JSON value.

**Instance IDs, in order, not a count.** "Before it" is a question
about an object's position, and the answer is an index. A scalar
cannot produce one, and a per-object counter would need a second read
to order the objects anyway.

**Table-wide, not per player.** CR 702.40a says "each other spell".
`SpellsCastThisTurn` stays exactly as it is: the two answer different
questions and merging them would make the per-player tally lie for
"first spell you cast each turn".

**Not an `EventsThisTurn` scan.** The scan is what Thousand-Year
Storm does, and it drops a cast whose card the engine can no longer
find. Storm must not: a Grapeshot countered by an opponent is still
a spell that was cast before the Tendrils. A recorded ID is immune to
everything that happens to the card afterwards, which is the whole
point of recording it at the cast.

**Not merged with Thousand-Year Storm's helper.** That card needs the
type test and therefore needs the lookup; this one must not have it.
They stay two readers.

**The MOST RECENT cast of an ID, when there are two.** A spell
countered or bounced back to its owner's hand (Remand) and cast again
the same turn is two casts of one card, and a card keeps its instance
ID across the round trip — so the list holds that ID twice. The
question is always about the later one, and the earlier cast's own
trigger cannot still be waiting: CR 603.3 puts a cast trigger on the
stack *above* its own spell, so it resolves before anything can bounce
that spell. Taking the first index instead would undercount every
recast, which is the case that actually happens at a table. The first
cast *counts toward* the second, because CR 400.7 made the card a new
object when it left the stack: it really is "another spell that was
cast before it".

**Cost.** One `uuid.UUID` (16 bytes) per spell cast per turn, freed at
the turn boundary. A storm turn is the worst case and is the turn the
field exists for.

### 2. The count is read when the storm trigger RESOLVES

The trigger's `Build` captures the spell's **instance ID** and
nothing else. `SpellsCastBeforeThisTurn` is called from the trigger's
`Effect`, as it resolves.

CR 608.2h settles when a resolving ability reads information from the
game: once, when the effect is applied. That is the read this
implements.

The two reads happen to agree — the list is append-only and the
spell's index in it is frozen by its own cast, so nothing between the
announce and the resolution can move it. That is a property worth
**pinning with a test** rather than a licence to take the cheaper
read, and taking the resolution read buys three things:

- the trigger carries one immutable value, so a second instance of
  storm on the same spell (CR 702.40b) and a doubled trigger
  (`Spec.TriggerDoublers`) each re-derive the same number rather than
  sharing a captured one;
- a spell cast **in response** to the storm trigger is visibly not
  counted, because it is appended after the storm spell's own entry
  and the index does not move — which is the rule, and is now an
  assertion rather than a coincidence of when a closure ran;
- the number the log announces and the number of copies made come
  from one call, so they cannot disagree.

**Explicitly rejected:** #1238's own sketch — capture the
*pre-increment* `SpellsCastThisTurn` count at `Build`. It is per
player (storm counts everyone), it is a scalar (so a second storm
trigger must re-derive it), and "pre-increment" is a fact about the
order of two statements in `CastSpell` rather than about the rules.
The tally bump moved once already (S19 sub-PR 6 put it *before*
`EventCast` on purpose); a storm count that depends on it is a card
that breaks when that line moves again.

### 3. Storm is a declaration, not card code

```go
Triggered: []game.TriggeredAbility{Storm()},
Triggered: []game.TriggeredAbility{Storm(), Storm()},  // CR 702.40b
```

`effects.Storm()` (`cards/effects/storm.go`) is the same shape
`Cascade()` has and for the same reason: `FromStack: true`, watching
`EventCast` for the source's own cast, because the ability fires while
its spell is still on the stack and the battlefield scan cannot see
it. `Spec.Triggered` already accepts an arbitrary
`game.TriggeredAbility`, so **no new `Spec` slot**: a card with
printed storm is its printed effect plus one line.

The trigger's controller is `ev.Actor` — the player who cast the
spell — not `source.Controller`, for `Cascade()`'s reason: a spell
cast off an opponent's library under an impulse grant belongs to the
caster.

### 4. One re-target prompt per copy, not a batched one

CR 702.40a says "you may choose new targets for **any** of the
copies", and CR 707.10c makes that offer per copy. `CopySpell{Count:
n, ChooseNewTargets: true}` is already a loop of independent
`CopySpellForEffect` calls (ADR 0043; `spell_copy.go`), each of which
opens its own `PendingChoicePickTarget` carrying its own `copyFrame`.
Storm uses that unchanged.

A batched "choose new targets for the copies" prompt was considered
and rejected. ADR 0065 §2's slot dimension lives on the `TargetRef`
(`Slot`, `Mode`); nothing anywhere carries a COPY dimension, and a
batched prompt would have to invent one, thread it through
`resolveCopyTargetsLocked`, and answer a question the rules do not
ask — whether the copies' targets are chosen simultaneously or one at
a time. The per-copy shape is also the one that is already
**enumerable**: `legal/choices.go`'s `PendingChoicePickTarget` case
answers one prompt with one ref, so a bot walks N prompts with no
change, and the client's existing picker renders each in turn.
Increasing Vengeance ("copy that spell twice instead") and
Thousand-Year Storm both ship this shape today.

Empty the Warrens prints no re-target clause because it has no
targets. It still declares `Storm()` and gets no prompts, because the
engine skips the offer for a spell with no chosen target
(`itemHasChosenTarget`) — the card says nothing special and needs to
say nothing special.

### 5. The table is told the count

A new engine event kind, `EventStorm`, and a new public log kind,
`storm`, rendering **"Grapeshot — storm count 3"**. Emitted once, by
the one engine entry point, as the trigger resolves.

The log needs its own line here, and the existing silences say why
rather than cover it:

- `EventTrigger` is a deliberate silence because "a trigger reaching
  the stack is told by the `LogResolve` of the ability it becomes"
  (`log_event_kind_gate_test.go`). For an **ability** that
  `LogResolve` carries no `card_id` and no label, so it renders as
  "a card resolved" — true, and it does not say the number.
- A spell **copy** emits no event at all (`createSpellCopyLocked`
  deliberately emits no `EventCast`, CR 707.10). So without this line
  the table watches three Grapeshots appear from nowhere with nothing
  written down about why there are three.

The count is the card. `saga_chapter` and `class_level` are the
precedent: a mechanic whose number is otherwise unobservable gets a
line.

### 6. A countered storm spell keeps its trigger — and, today, makes no copies

CR 113.7a: the storm ability exists on the stack independently of its
source, so countering the spell does not remove the trigger. That is
pinned by a test.

What the trigger then does is the open half, and it is left open on
purpose. `CopySpellForEffect` answers `ErrCardNotFound` for a spell
that is no longer on the stack, and `CopySpell.Apply` treats that as
"the copy effect did nothing" — the convention Reverberate has
shipped since S30. CR 608.2h's last-known-information reading says the
copies should be made anyway.

It is not fixed here because the fix is a change to the **shared**
copy path that Reverberate must not take: Reverberate *targets* the
spell, so CR 608.2b already counters it by game rules when that target
is gone, and widening `stackSpellLocked` to reach a countered spell
would resurrect a target the rules say has left. The right shape is a
last-known-stack record consulted only by non-targeting copy effects,
which is its own piece of work. Filed separately.

The ordinary line of play is unaffected: CR 603.3 puts the storm
trigger on the stack **above** its own spell, so it resolves first and
the spell is still there. Reaching the open case at all takes a
counterspell cast in response to the trigger.

## What this does not build

- **Storm's count as a number other cards can read.** Nothing but
  storm asks "how many spells were cast before this one". Cards that
  ask "how many spells have been cast this turn" (Aetherflux
  Reservoir's shape) would read `len(TurnTally.Casts)`, which is one
  line, but no catalog card needs it today and an unused accessor is
  an unused accessor.
- **A `Keyword` field on `game.TriggeredAbility`.** `Cascade()` and
  `Storm()` are both keyword triggers with no machine-readable name,
  which means `cards/coverage`'s mechanic probe table cannot ask "does
  this card have storm" the way it asks "does this card have
  flashback" (`altCost`). Cascade has shipped without one since S28.
  Filed rather than built, because the field is only worth adding once
  something reads it.
- **Copies of the storm trigger itself.** `Strionic Resonator` on a
  storm trigger is the ability-copy path (#1223) and needs nothing
  from here: the copied ability re-reads the same list and produces
  the same number.

## Amendment 2026-09-23 — #1255 and #1257: Decision 6's open half is closed, and Decision 5's premise changed · Accepted · S45

Issues [#1255](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1255)
and [#1257](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1257),
both filed while closing #1238. No new ADR number: each is the
follow-up this ADR named, and the copy half is written up in
[ADR 0043's amendment of the same date](0043-copy-effects.md#amendment-2026-09-23-1255-a-spell-that-left-the-stack-is-copied-from-last-known-information-cr-6082h),
which owns the copy path.

### Decision 6, revised: a countered storm spell IS copied

Decision 6 left "a countered storm spell keeps its trigger — and makes
no copies" open because the only fix in reach was a change to the
shared copy lookup that Reverberate must not take. The fix is the
shape Decision 6 described: a last-known-stack record that only a
non-targeting copy effect consults.

- `Game.lastKnownStack` records the card and stack item of a spell that
  leaves the stack **without resolving**, at the one choke point every
  such exit passes (`routeCardToZoneLocked` with `DropStackMeta`:
  counterspells, Remand-style returns, the sandbox move). It lives for
  the turn.
- `CopyLastKnownSpellForEffect` reads the stack, then the #920
  resolving slot, then the record. `CopySpellForEffect` is unchanged.
- `stormItem` sets `CopySpell.FromLastKnown`. Storm's trigger names the
  spell ("copy **it**") and does not target it, so CR 608.2h governs:
  the copies are made from the spell as it last stood on the stack,
  with its targets, modes and X, and each still gets its own CR 707.10c
  re-target prompt (Decision 4).

`TestACounteredStormSpellKeepsItsTriggerOnTheStack` was written to
change the day this landed, and it did: with one spell before it, a
Grapeshot countered in response to its trigger now deals 1 (its copy)
rather than 0. `TestBrainFreezeCounteredInResponseStillMills` is the
same rule on the storm card a counterspell is most often aimed at.

The record is `dropped` by the snapshot, for `resolving`'s reason:
every reader of it is a stack item or delayed trigger whose behaviour
is a closure, already counted in `ContinuationCensus`, so a snapshot
that could need the record is not a restore point.

### Decision 5, revised premise: the resolve line now names the ability

Decision 5 argued for the `storm` log kind partly on the ground that an
ability's `resolve` entry "carries no `card_id` and no label, so it
renders as 'a card resolved'". #1257 changed that for every ability:
the entry now carries the ability's source as `card_id` and its stack
label as `label`, and reads "Grapeshot — storm resolved".

Decision 5 still stands, on its first ground alone, which was always
the stronger one: the resolve line names the **ability**, and the
count is the **card**. "Grapeshot — storm resolved" followed by three
Grapeshots still does not say why there are three. So the table now
reads both lines — "Grapeshot — storm count 3", then "Grapeshot —
storm resolved" — and neither is redundant.

The two deliberate log silences that rested on the old line
(`EventTrigger`, `EventActivateAbility` in
`log_event_kind_gate_test.go`) keep their silence; their stated reason
— "told by the LogResolve of the ability it becomes" — is now true.
The redaction and wire-cost argument for the new fields is in
[docs/protocol.md](../protocol.md) under "An ability's `resolve` /
`fizzle`".
