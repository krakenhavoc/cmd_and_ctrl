# ADR 0065 — Modal and multi-target clauses

**Status:** Accepted · 2026-09-18 · S45 — Modal spells, multi-target
clauses and copy effects · tracked on
[#764](https://github.com/krakenhavoc/cmd_and_ctrl/issues/764), sprint
tracker [#888](https://github.com/krakenhavoc/cmd_and_ctrl/issues/888).

**Numbering:** on 2026-09-18, all 263 remote heads (`git ls-remote
--heads origin`, then `git log --all --name-only -- docs/decisions/`
over the fetched refs) were checked with the AGENTS.md §4 loop. The
highest number in use anywhere is 0064: `0060-leaving-the-game`,
`0061-token-creation-and-discard-are-replaceable-events`,
`0062-abilities-and-special-actions-from-the-hand`,
`0063-durations-and-control` and `0064-emblems` are all held by
parallel branches. 0065 is the first free number.

**Lifts the deferral in:** [ADR 0019 §8 "Out of scope"](0019-structured-targeting.md)
— "per-mode target slots … needs `StackItem.Targets` grouped by mode
and a two-step picker". That is exactly what this ADR builds.

**Amends:** [ADR 0019](0019-structured-targeting.md) §1 (a `TargetSpec`
is now a clause *list*), §7 (the one-targeted-option limit is gone) and
§8 (slots are per clause, not per pick);
[ADR 0018](0018-triggers-on-the-stack.md) §6 (a new blocking prompt
kind); [ADR 0020](0020-activated-abilities.md) (an activated ability
can be modal); [ADR 0026](0026-delayed-triggers.md) (a triggered
ability can be modal).

**Builds on:** [ADR 0033](0033-ai-bot-seat.md) §1 (the enumerator's
`MaxExpansionPerSource` budget), [ADR 0041](0041-game-persistence.md)
and [ADR 0044](0044-surviving-a-deploy.md) (snapshot / restore, the
continuation census), [ADR 0043](0043-copy-effects.md) (copy
retargeting).

**Explicitly NOT reused:** the `option_pick` pending-choice kind
(#568). It was not on `develop` at this branch's base (`11f4c3d5`) —
`grep -r option_pick` over the whole tree returned nothing — and it
landed while this branch was being written, so the decision below was
made without it and the rebase kept it. This ADR adds a kind of its
own, `mode_pick`: the name `pending_choice.go` has held in reserve for
exactly this since S20
([§4](#4-modal-triggers-the-mode-is-chosen-as-the-trigger-goes-on-the-stack-cr-6033c)).

The two are not the same question and folding them together would
lose both halves. `option_pick` is answered with ONE index, is asked
at RESOLUTION (CR 608.2), is frequently addressed to somebody other
than the controller, and its options carry cards — its own commit
message says a modal spell's "choose one" is deliberately not that
kind. `mode_pick` is answered with a bounded MULTISET of indices in
the order chosen (CR 700.2c, and CR 700.2d lets one repeat), is asked
as the ability is put on the stack (CR 603.3c), always goes to the
ability's controller, and each option carries a target clause rather
than a card list. If a later change gives `option_pick` bounds and an
ordered multi-answer, `mode_pick` is the obvious first thing to fold
into it.

---

## Context

Four gaps, one shape. Measured on `develop` at `11f4c3d5`:

1. **A target clause is one predicate.** `game.TargetSpec`
   (`targets.go:28`) has one `CardOK` / `PlayerOK`, one `Zones`, one
   `Min`/`Max`. A card with two differently-constrained slots — Bite
   Down's "target creature you control" and "target creature or
   planeswalker you don't control" — has to widen the spec to the
   union of both and check the per-slot half at *resolution*, where
   the only available answer is "the spell does nothing". Three
   catalog cards ship that caveat today (Bite Down, Soul's Fire,
   Resourceful Defense).

2. **Modes cannot repeat, and at most one may target.**
   `validateModes` (`modes.go:66`) rejects a repeated index, so CR
   700.2d ("you may choose the same mode more than once", Mystic
   Confluence) is unrepresentable. `castTargetSpec` (`modes.go:89`)
   returns `ErrInvalidParam` when two chosen options both target, and
   `effects.Register` (`registry.go:45`) panics at boot on a card that
   declares one, so Kolaghan's Command cannot even be written down.

3. **Only spells are modal.** `TriggeredAbility` (`triggers.go:73`)
   has `Targets` and `OptionalPrompt` and a comment at `:119` saying
   modal triggers "need a separate ModePrompt slot".
   `effects.ActivatedAbility` (`spec.go:520`) has no mode slot either.
   Gala Greeters takes the first unused mode in printed order;
   Aetheric Amplifier offers only its first bullet.

4. **The wire and the picker are single-clause.** `cast_spell`
   carries one flat `targets` array with no slot dimension;
   `TargetingState` (`client/src/lib/targeting.ts:131`) has one
   `legal` set, one `min`/`max`, one `picked`; `ModePickerModal`
   (`:63`) hard-refuses a second targeted mode and sorts the chosen
   indices ascending, which destroys the order CR 700.2c resolves in.

The issue's audit counts 54 cards where this is the only core blocker
and 58 where it is any core blocker, with a verified estimate of
**33–38 real** after misattributions.

---

## Decisions

### 1. A `TargetSpec` *is* a clause, and a clause may be followed by more

```go
// TargetClause is ONE clause of a target statement: one predicate,
// chosen Min..Max times. It is an alias for TargetSpec.
type TargetClause = TargetSpec

type TargetSpec struct {
    Mode       string          // client hint, derived
    Label      string          // "target creature you control"
    Players    bool
    Zones      []ZoneKind
    CardOK     func(g *Game, caster uuid.UUID, c Card, zone ZoneKind) bool
    PlayerOK   func(g *Game, caster uuid.UUID, p *Player) bool
    Min, Max   int
    AllowSame  bool
    CountFromX bool

    // Distinct: this clause's picks must differ from every EARLIER
    // clause's picks — "a SECOND target permanent you control"
    // (CR 601.2c: one object can't fill two "target" words unless
    // the card says otherwise). Added by #764.
    Distinct bool

    // Rest holds clauses 2..n. A single-clause statement — which is
    // every clause in the catalog before #764, every cost-payment
    // predicate, and every mode option that targets once — leaves it
    // empty and is byte-for-byte what it was. Added by #764.
    Rest []TargetClause
}

func (s *TargetSpec) Clauses() []TargetClause   // head, then Rest
```

**Why head-and-tail rather than `struct{ Clauses []TargetClause }`.**
`TargetSpec` has two jobs in this tree and only one of them is
targeting. It is also the reusable *predicate over permanents* for
every cost payment: `AbilityCost.SacrificeOther`,
`ManaAbilityShape.SacrificeOther`, `AdditionalCost.Sacrifice`,
`AlternativeCost.ExileFromGraveyard`, `TapPermanentsCost.Spec`,
`CounterRemovalCost.From`. Those are single-clause by nature — a
sacrifice cost is one predicate, not a list — and wrapping them in a
one-element list would be ceremony at ~40 declaration sites and every
reader. Making the spec *be* its first clause keeps all of that, and
every one of the ~500 catalog card files, compiling unchanged, while
there is exactly **one** predicate struct in the tree and exactly
**one** walk over it (`Clauses()`).

**The shim, precisely.** It is not a second model and there is no
second code path:

- `TargetClause` is a Go *type alias*, not a distinct type. There is
  one struct.
- Every consumer — the legal-target enumeration, the announce gate,
  the CR 608.2b re-check, the protocol projection, the bot enumerator
  — iterates `spec.Clauses()`. A one-clause spec yields a one-element
  slice, so the "old" behaviour is the general code with `n == 1`.
- The constructors in `cards/effects/targets.go` (`TargetAny`,
  `TargetCreature`, …) still return a one-clause `*TargetSpec`.
  `Clauses(first, then…)` is the new constructor that hangs more
  clauses off the first, and `.WithCount` still edits the head.
- `Register` refuses a nested `Rest` (a clause inside `Rest` with its
  own `Rest`) at boot. The list is flat by construction.

This is the migration path: a card moves from one wide clause to two
narrow ones by changing its `Targets:` line and deleting the caveat.
Nothing else in the tree changes with it.

### 2. Slots live on the `TargetRef`, not in a second container

`StackItem.Targets` stays one flat `[]TargetRef` in announce order.
Each ref gains two small integers:

```go
type TargetRef struct {
    Kind TargetKind
    ID   uuid.UUID

    // Slot is the index of the CLAUSE this ref answers, within the
    // clause list that governed the announcement.
    Slot int
    // Mode is the index into StackItem.Modes — the mode OCCURRENCE,
    // not the option — whose clause list Slot indexes. 0 for a
    // non-modal item.
    Mode int
}
```

A `[]TargetGroup` was the alternative and was rejected: `ctx.Target()`,
`ctx.Targets()`, `ctx.LegalTargets()`, `item.Targets[0]`, the wire's
`targets` array, `cloneStackItem`, the snapshot encoder, the copy
retargeter and about ninety card files all read the flat list, and
grouping it would touch every one of them to express something two
`int`s already say. With the ints, a positional reader
(`item.Targets[0]` is the biter, `[1]` the victim) is *still correct*,
because announce order is clause order — which is what makes the
conversion of Bite Down a one-line change to its `Targets:` slot.

Zero values mean "clause 0 of the only clause list", so every ref that
predates this ADR — in a snapshot, on the wire, in a test fixture —
reads correctly without a migration.

The CR 608.2b re-check resolves the clause per ref
(`clauseForRefLocked`): the announced mode occurrence's clause list if
the item is modal, the item's own clause list otherwise, then
`Clauses()[ref.Slot]`. **This is the whole point of the change**: a
two-clause spell's second slot is now re-checked against its *own*
predicate at resolution, not against a union.

### 3. One `ModeSpec`, three owners

```go
type ModeOption struct {
    Label   string
    Targets *TargetSpec                       // this bullet's clause list
    Effect  func(g *Game, item *StackItem) error  // optional bullet body
}

type ModeSpec struct {
    Prompt     string
    Options    []ModeOption
    Min, Max   int
    Repeatable bool   // CR 700.2d
}
```

The same struct is read by `Spec.Modes` (spells, unchanged),
`TriggeredAbility.Modes` (new) and `effects.ActivatedAbility.Modes`
(new). There is no per-owner variant and no per-card special case.

- **"Choose one"** is `Min 1 / Max 1`; **"choose two"** `2 / 2`;
  **"choose one or both"** `1 / 2`; **"choose one or more"**
  (Sublime Epiphany) `1 / len(Options)`; **"choose up to one"**
  `0 / 1`.
- **Repeated modes (CR 700.2d).** `Repeatable` drops the
  distinctness check in `validateModes`. `StackItem.Modes` is then a
  **multiset in announce order** — Mystic Confluence choosing its
  draw mode three times is `[0, 0, 0]` — and each occurrence gets its
  own target group (`ref.Mode == 0, 1, 2`). `ctx.HasMode(i)` keeps
  its meaning ("some occurrence chose option i"); `ctx.ModeCount(i)`
  is the count, and `ctx.Modes()` is the ordered multiset.
- **Resolution order is announce order (CR 700.2c).** Options with an
  `Effect` run in `item.Modes` order, once per occurrence. A card that
  prefers the old shape keeps writing `if ctx.HasMode(0) { … }` in
  `OnResolve`; both read the same data, and `Effect` exists so a
  *trigger* or *activated ability* — which has no `OnResolve` to hang
  an if-chain on without hand-writing a switch — declares its bullets
  the same way a spell does.
- **Per-mode targets** are picked in mode order *after* the modes are
  chosen, one clause at a time: the announcement is a flat walk over
  the steps `(occurrence 0, clause 0), (occurrence 0, clause 1),
  (occurrence 1, clause 0), …`. That flat walk is the "two-step
  picker" ADR 0019 named, and it is the same walk a non-modal
  multi-clause card takes with one occurrence.
- `effects.Register`'s panic on "targeted modes with `Max > 1`"
  (`registry.go:45`) is **deleted**. In its place `Register` refuses a
  `Repeatable` spec whose `Max` is 1 (nothing to repeat) and a nested
  `Rest`.

### 4. Modal triggers: the mode is chosen as the trigger goes on the stack (CR 603.3c)

One new pending-choice kind, `PendingChoiceModePick`
(`"mode_pick"`, the name reserved at `pending_choice.go:38`), queued by the harvester at the point CR 603.3c
puts the ability on the stack — **after** the CR 603.5 "you may"
prompt and **before** the CR 603.3d target pick. The order in
`dispatchTriggerInstanceLocked` becomes:

1. targeted with no legal target for *any* choosable mode → the
   trigger is removed with no prompt at all (CR 603.3d);
2. `OptionalPrompt` → the yes/no prompt; "yes" re-enters at 3;
3. `Modes` → `mode_pick` to the controller (the source's
   controller, exactly as the target pick is);
4. `Targets` (card-level, or the chosen occurrences') → the
   `pick_target` walk;
5. `Build` → queue.

Two consequences, both deliberate:

- **An untargeted modal trigger takes the same path.** Black Mark and
  Gala Greeters get a real `mode_pick` prompt at step 3 and
  nothing at step 4. A chain of `confirm` prompts could approximate
  the choice but not the *timing*: it would ask at resolution, a
  priority round too late, and a player could not respond to the
  announced mode. The kind exists so the timing is right.
- **A mode with no legal target is not offered.** `mode_pick`
  ships only the options that are choosable right now: for a targeted
  option, one whose legal set is non-empty. If that leaves fewer than
  `Min`, the trigger is removed (CR 603.3d) without a prompt — the
  same rule the single-clause path already applies, now evaluated per
  option.

**Blocking and enumeration.** `mode_pick` blocks the table
(`choiceGateDecisions`) — an unanswered mode choice is an ability that
is not yet on the stack, so nothing else may happen — and gets a case
in `legal.choiceMoves` offering every legal multiset, bounded by
`MaxExpansionPerSource`. Both are what
`TestEveryChoiceKindIsClassifiedAndEnumerated` demands.

Reflexive triggers (#795 / #636) are unaffected: their payload is a
record of what happened, not a target, and they carry no modes.

### 5. Modal activated abilities: modes and targets at activation (CR 602.2b)

`ActivateAbilityParams` gains `Modes []int`, validated by the same
`validateModes` and the same clause walk the cast path uses, in the
CR 602.2b order — **modes, then targets, then costs**. The announcement
is indivisible, so there is no prompt here at all: the client sends
the whole activation, exactly as it sends a whole cast. The
`mode_pick` prompt is for the one owner that cannot be asked in
advance (a trigger, which the engine puts on the stack itself).

### 6. Bots: bounded enumeration, with a stated preference

`legal`'s cast and activation enumerators expand
`modes × (targets per clause, per occurrence)` under one
`MaxExpansionPerSource` budget (ADR 0033 §1, default 12). Policy,
documented in `docs/bot.md` and implemented in `legalModeSets`:

- **Prefer modes that have legal targets.** An option whose clause
  has an empty legal set is dropped from the candidate list before
  any combination is built, so the budget is never spent on
  selections the engine would refuse.
- **For a repeatable spec, offer the "all one mode" selections
  first**, then the mixed ones: when only one option is legal, "the
  same mode `Max` times" is the only selection, and it should not be
  crowded out by an enumeration that walks mixed multisets first.
- The budget is spent **modes-outermost**: every mode selection gets
  at least one target set before any selection gets a second, so a
  bot is never offered only the first mode of a charm.

### 7. Wire and client: one flat list of clause steps

- `LegalTargetsView` is unchanged and is now explicitly **per clause**.
  `CardView`, `ModeOptionView` and `ActivatedAbilityView` gain
  `clauses: LegalTargetsView[]`, present only when the statement has
  more than one clause; `legal_targets` stays and is clause 0, so
  every existing client path and test keeps working.
- `ModeSpecView` gains `repeatable`.
- `cast_spell` / `activate_ability` `targets[]` entries gain
  `slot` and `mode`, both optional and both defaulting to 0.
- `StackItemView` gains `mode_labels: string[]` — the oracle text of
  the chosen bullets, in announce order — so the stack overlay stops
  printing `modes: 0, 2` at players who cannot see the caster's hand.
- The client's picker becomes one walk. `TargetingState` carries
  `steps: TargetStep[]` (each with its clause's label, mode hint,
  legal set, min/max, `distinct`, and the `mode`/`slot` it answers),
  a `step` cursor, and `done: TargetRef[][]`. The mode picker feeds
  the step list; the board-click flow advances the cursor instead of
  firing, and fires when the last step completes. A one-clause
  non-modal cast is a one-step walk and behaves exactly as it did.
- The mode picker stops sorting the chosen indices and stops refusing
  a second targeted option; when `repeatable` it offers a count per
  option instead of a toggle.

### 8. Undo, clone and snapshot

- `TargetRef.Slot` / `.Mode` are plain ints on a carried field:
  carried, with no new census entry.
- `ModeSpec.Repeatable` and `ModeOption.Effect` are catalog data, not
  game state; they are never serialised.
- `StackItem.modeSpec` (the `*ModeSpec` an ability item was announced
  under) is classified `rebuilt` for the same reason `targetSpec` is:
  a spell's is re-derived from the catalog by oracle ID, and an
  ability's is counted in `ContinuationCensus` alongside its
  `targetSpec` when it is not.
- The `mode_pick` prompt's resume frame is a closure, so it is
  counted in `ContinuationCensus` exactly as the `pick_target` and
  `trigger_prompt` frames are, and a game holding one is not a
  restore point.

---

## Out of scope, stated

- **Divided damage and pawprint modes** (CR 700.2i). `Distribution`
  rides the stack item and nothing reads it; unchanged here.
- **Modes chosen by another player** (CR 700.2e, Seize the
  Spotlight). `mode_pick` has a `Chooser` field and the prompt
  would go to a different seat, but no card in the catalog needs it
  and nothing is built for it.
- **"Choose one that hasn't been chosen this turn"** (CR 700.2 with a
  per-source tally: Monument to Endurance, Silent Hallcreeper). The
  hook is `ModeSpec.Available` and it is **not** added; Gala Greeters
  converts to a real prompt with all three bullets offered every
  time, which is *stronger* than printed and therefore stays a
  declared caveat until the tally lands.
- **Set-level `Validate` over the clause list** (Secret Tunnel's
  "share a creature type"), the targeting twin of #624. The clause
  list is the right place for it and the field is not added.
- **Per-clause copy retargeting.** `CopySpellForEffect`'s re-target
  prompt walks the copied item's clause list head only; a two-clause
  spell copied by Reverberate re-targets its first clause and keeps
  the rest. Noted as a caveat on the copy path, not fixed here.

---

## Consequences

Good:

- The per-slot constraint is enforced **at announce**, which is where
  CR 601.2c puts it, so "a pair that doesn't fit makes the spell do
  nothing" stops being a thing a player can do.
- Three caveats clear on cards that were already shipping, two on
  cards that were shipping a wrong mode, and 33–38 cards become
  writable.
- One struct, one walk, one picker, three owners.

Costs:

- `TargetRef` grows by two ints and appears in a lot of places. The
  zero values are the old behaviour, which is what makes that
  affordable.
- `mode_pick` is a 26th pending-choice kind, and every kind is a
  thing the bot, the gate and the client have to know about.
- The clause list is flat by construction and enforced at boot rather
  than by the type system, because the alias is what buys the
  compatibility.

---

## Amendment (2026-09-23, #1330): Spree — a mode with its own cost (CR 702.172a)

**Sprint:** S45 — Modal spells, multi-target clauses and copy effects.
Tracker [#888](https://github.com/krakenhavoc/cmd_and_ctrl/issues/888).

§3 shipped one `ModeSpec`, three owners, and every option in it was free
to choose beyond the spell's own printed cost. CR 702.172a's Spree keyword
("Choose one or more modes. As an additional cost to cast this spell, pay
the costs associated with those modes chosen this way.") is the first
printed shape where choosing a bullet is not free, and it is why Three
Steps Ahead — this issue's proof card — could not be registered at all:
`game.ModeOption` had nowhere to put "+ {1}{U}".

### Decision: `ModeOption.Cost string`, mana only

```go
type ModeOption struct {
    Label   string
    Targets *TargetSpec
    Effect  func(g *Game, item *StackItem, occurrence int) error

    // Cost is CR 702.172a's Spree: the additional mana cost paid IF
    // AND ONLY IF this bullet is chosen, in brace notation, on top of
    // the spell's own cost and every OTHER chosen bullet's. Empty for
    // an ordinary modal bullet — every modal card before S45.
    Cost string
}
```

A bare `string`, not a reused `*AdditionalCost` (the shape ADR 0073 §1
built for kicker and buyback), and that is the one decision in this
amendment worth arguing. `AdditionalCost` carries `DiscardCards`,
`Sacrifice`, `PayLifeX`, `Optional`, `Key` and `Repeat` alongside
`ManaCost`, and every printed Spree card checked against the Scryfall
dump at the time of writing — Three Steps Ahead, Explosive Derailment,
Insatiable Avarice, Caught in the Crossfire, Phantom Interference,
Requisition Raid, Lively Dirge, Rustler Rampage, Final Showdown, One
Last Job, Getaway Glamer, Jailbreak Scheme, Metamorphic Blast, and
Unfortunate Accident — prices every bullet in mana alone. Reusing
`AdditionalCost` would have meant a boot-time panic disabling five of
its six fields on every card that will ever declare one, which is
worse than the honest field: the day a card needs "discard a card" per
bullet (Duskmourn's own "Bake into a Pie" is exactly that shape), the
change is `Cost string` → `Cost *AdditionalCost` in one place, not a
retrofit of a struct that had been carrying dead weight the whole
time.

**Two constructors, both in `effects/modes.go`** (the card-file layer,
not `game`), matching `Mode` / `ModeDoing`'s existing split:

```go
Modes: Spree(
    SpreeModeDoing("Counter target spell.", "{1}{U}",
        TargetSpell("target spell"), CounterTheModesTarget),
    SpreeModeDoing("Create a token that's a copy of target artifact or creature you control.", "{3}",
        TargetPermanent("target artifact or creature you control", And(Or(Artifact(), Creature()), YouControl())),
        TokenCopyTheModesTarget),
    SpreeModeDoing("Draw two cards, then discard a card.", "{2}", nil,
        func(item *game.StackItem, ctx *Context, occ int) error {
            if err := (DrawCards{Player: item.Controller, N: 2}).Apply(ctx); err != nil {
                return err
            }
            ctx.Game.QueueDiscardChoiceForEffect(game.DiscardPrompt{
                Player: item.Controller, Source: item.SourceCardID, N: 1,
            })
            return nil
        }),
),
```

`Spree(options...)` is `Min: 1, Max: len(options)` and NOT
`Repeatable` — CR 702.172a's "one or more" is not "the same one
twice", and no printed Spree card says otherwise. `SpreeMode` /
`SpreeModeDoing` are `Mode` / `ModeDoing` with `Cost` set; they exist
so a card file reads like its oracle text ("+ {1}{U} — Counter target
spell.") instead of a struct literal with a bare field assignment.

`CounterTheModesTarget` and `TokenCopyTheModesTarget` are two NEW
shared bullet bodies in `modes.go`, promoted from private closures
that were about to be duplicated: Sublime Epiphany already had a
"counter target spell" bullet and a "token copy of target creature you
control" bullet, byte-for-byte the same body Three Steps Ahead needed
one clause narrower. `internal/cards/coverage`'s exact-clone detector
caught the duplication in review — the two cards now both call the
shared functions, which is the `DestroyTheModesTarget` /
`BounceTheModesTarget` pattern this ADR already established, applied
to the two bodies that happened not to exist yet.

### Decision: the price joins the total where ADR 0073 §3 already puts an extra

`game.AddModeCostMana(cost ParsedCost, ms *ModeSpec, modes []int) (ParsedCost, error)`
sums every chosen occurrence's `Cost` into the running total, called
from `printedCostLocked` in the exact spot `AddOptionalCostMana` is —
after the alternative-cost swap and the commander tax, before the cost
modifiers (CR 601.2f: an additional cost joins the total before
Thalia or Trinisphere read it). One function, so the cast path, the
bot enumerator and the auto-tap preview cannot disagree about what a
Spree selection costs (#544) — the identical invariant ADR 0073 §3
states for kicker, reused rather than re-derived.

No new `PaidCost` field. Kicker needed `PaidCost.OptionalCosts` because
"was this kicked" is asked by a card's own resolution and by an
entering permanent's ETB trigger, neither of which can be inferred any
other way. A Spree bullet's price is not asked anywhere after the
fact — `ctx.HasMode(i)` already answers "was this bullet chosen", and
CR 702.172a is silent about anything reading the money afterward — so
the modes multiset already carried on `StackItem.Modes` is the whole
record.

### Decision: the enumerator prices PER MODE SELECTION, not once for the card

This is the one place the change reaches past the card layer. Every
modal card before this amendment had ONE price regardless of which
modes were chosen, so `legal/cast.go`'s `castMovesPayingOptional`
priced the cast once, before `modeSets` was even computed, and searched
X against that one number. A Spree selection's price depends on WHICH
modes are in it, so that ordering is now backwards: `modeSpec` /
`modeSets` are computed first, and `game.AddModeCostMana` is called
**inside** the `for _, modes := range modeSets` loop to produce
`modeCost`, which THEN feeds the `!perTarget` X search and the
per-target repricing loop that already existed for §14's target-priced
cards. A selection this seat cannot afford is `continue`d rather than
crashing the whole card's enumeration — Explosive Derailment's two
bullets are each individually affordable off three Mountains and never
offered together, which
[`spree_test.go`](../../server/internal/legal/spree_test.go) pins by
funding a seat to exactly one bullet's price, then to both, and
`dispatchAll`-ing every offered move against the real engine either
way (#544).

### Decision: no card-shaped mode cost, no repeatable mode cost, no owner but Spec.Modes

Three refusals at boot, matching the "declare it when a card needs it"
discipline this file already keeps:

- **A trigger's or an activated ability's mode may not carry a Cost.**
  CR 702.172a is a static ability printed on SPELLS. No trigger or
  activated ability in Magic charges more for choosing one of its
  modes, and `checkModeCost` panics if one tries — the same "the field
  exists, using it wrong is a boot error" posture ADR 0073 §1 takes
  with `AdditionalCost.Optional` on the mandatory slot.
- **`Optional`, `Key`, `Repeat`, `DiscardCards` and `Sacrifice` have no
  equivalent here**, because the field is a bare mana string. A future
  non-mana Spree card is the trigger for widening it, not a reason to
  guess its shape now.
- **The `ModeSpec` itself is not marked `Repeatable`.** `Register`
  already panics on `Repeatable` with `Max == 1`; nothing new was
  needed to keep Spree off that combination, because `Spree()` never
  sets the flag.

### Consequences

- Three Steps Ahead, Explosive Derailment, Insatiable Avarice and
  Caught in the Crossfire ship at `CompletenessFull`. #1306's tracker
  is updated.
- The wire gains `ModeOptionView.Cost` (`cost`, omitempty, brace
  notation) — printed text, scoped `surfacePublicPile` beside `label`
  and `target_mode` exactly as ADR 0073's `AlternativeCostView.mana_cost`
  is, never `surfacePrivate` like `legal_targets`. `docs/protocol.md`'s
  `cast_spell` row says so.
- `ModePickerModal.svelte` shows each bullet's own cost beside its
  label and a running concatenation of the chosen bullets' costs
  beside Confirm — a preview, not a computation; the server still
  prices the real sum (#544).
- Sublime Epiphany's two bullets that used to be private closures are
  now `CounterTheModesTarget` / `TokenCopyTheModesTarget` in
  `modes.go`, available to every future card that prints either
  clause.

### Out of scope, stated

- **A card-shaped Spree cost** ("discard a card" per bullet — "Bake
  into a Pie"'s real shape). `ModeOption.Cost` is a mana-only string on
  purpose; widening it to `*AdditionalCost` is a follow-up, not a
  speculative field sitting unused today.
- **Escalate and entwine**, ADR 0073's Consequences named them as the
  same family and left them for whoever built this. Entwine ("choose
  both, pay the entwine cost") is closer to an `AlternativeCost` than
  to a per-mode price — one flat surcharge for taking every mode,
  not a sum of independently-priced bullets — and does not reuse this
  shape without its own design. Escalate ("choose one, then pay
  {cost} for each other mode you choose") is closer to what Spree
  built, but "the SAME price for every mode after the first" is a
  detail Spree's per-bullet `Cost` cannot express without a bullet
  declaring a different cost depending on how many others are already
  chosen — left for that card.
