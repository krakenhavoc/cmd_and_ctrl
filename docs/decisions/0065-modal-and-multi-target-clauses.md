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
the order chosen (CR 608.2c, and CR 700.2d lets one repeat), is asked
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
   indices ascending, which destroys the order CR 608.2c resolves in.

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
- **Resolution order is ~~announce order~~ printed order (CR 608.2c;
  amended 2026-09-30, see the amendment at the end).** Options with an
  `Effect` run once per occurrence. A card that
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
  *Lifted by #2622 (2026-10-09):* the head-only prompt had since been
  checked against the whole target list, so every answer to a copy of
  a two-clause spell (Bite Down) or of targets in two modes (Dromoka's
  Command) was refused and a bot table stopped. The prompt now walks
  every step of the copied announcement, one prompt per step, asking
  for exactly the number of targets the original chose there and
  offering the step's original targets beside its legal new ones (CR
  707.10c: a target left unchanged may stay even if illegal). A step
  with nothing to change to is not asked. See `offerCopyTargetsLocked`.

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

---

## Amendment (2026-09-27, #1590): a conditional mode count — "you may choose both instead"

**Sprint:** S45 — Modal spells, multi-target clauses and copy effects.
Tracker [#888](https://github.com/krakenhavoc/cmd_and_ctrl/issues/888);
the proof card is Jeska's Will on the Edea deck tracker
[#1565](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1565).

§3's `ModeSpec` bounds the count with two integers fixed at
`Register`. The Commander Legends Will cycle prints a bound that
depends on the board: *"Choose one. If you control a commander as you
cast this spell, you may choose both instead."* Flame of Anor prints
the same sentence about a Wizard. With nowhere for the condition to
live, all six catalogued cards on this shape (Jeska's Will, Akroma's
Will, Drown in Dreams, Will of the Mardu, Will of the Abzan, Flame of
Anor) shipped as "choose one" with a caveat — weaker than printed,
the #259 posture.

### Decision: `RaisedMax` + `RaiseMaxIf` on the spec, read by one accessor

```go
type ModeSpec struct {
    // … Prompt, Options, Min, Max, Repeatable …
    RaisedMax  int
    RaiseMaxIf ModeCountCondition // a registered KEY, not a func
}

// effects/modes.go, once
var YouControlACommander = game.ModeCondition("you-control-a-commander", controlsACommander)

// card file
Modes: ChooseOne(
    ModeDoing("Add {R} for each card in target opponent's hand.", …),
    ModeDoing("Exile the top three cards of your library. …", nil, …),
).OrUpToIf(2, YouControlACommander),
```

`(*Game).modeMaxLocked(ms, chooser)` — exported as `ModeMaxForEffect`
— is the only reader: `RaisedMax` while the named condition holds for
the chooser, `Max` otherwise. `Min` never moves, because every printed
card on the shape says the caster *may* choose both. `OrUpToIf`
mutates and returns the spec, like `TargetSpec.WithCount`, so the card
file reads like its oracle text.

**Why a key and not a func.** A `ModeSpec` is reachable from `Game`
through a trigger's paused `mode_pick` frame, so a func-typed field on
it is a new closure route, and ADR 0041 phase 3's ratchet
(`testdata/closure_fields.txt`, `closureClassCeilings`) lets the
blocker classes only shrink. `game.ModeCondition(key, fn)` registers
the predicate once at init in a package-level registry and returns a
`ModeCountCondition` whose key is unexported — the `BodyRef` idiom
ADR 0041's tier 2 uses for delayed triggers — so the spec carries data
and a func literal on it does not compile. An unregistered key holds
for nobody: the printed bound, the weaker reading. The key is never
persisted (the spec is re-derived from the catalog), so it needs no
ledger.

The predicate takes the CHOOSER, not the card, for the same reason
`AlternativeCost.Condition` does (S28's Fierce Guardianship): "you
control" is a question about a player. `effects.YouControlACommander`
registers the free-spell cycle's existing `controlsACommander`, so
"control a commander" means the same thing in both places: a
commander PERMANENT the player controls, anybody's (a stolen one
counts, an opponent's on their own side and one in the command zone
do not). `effects.YouControlAWizard` registers `ControlsA("Wizard")`,
Snuff Out's condition shape, for Flame of Anor.

### Decision: read at the choice, never again

The bound is read at the moment the choice is made — CR 601.2b's
announce in `castSpellLocked`, CR 602.2b's activation in
`activateCatalogAbilityLocked`, CR 603.3c's `mode_pick` prompt for a trigger —
and passed to `validateModes(spec, max, modes)`, which no longer reads
`spec.Max` itself. The answer lands on `StackItem.Modes` and nothing
re-asks the condition, so "as you cast this spell" is a check and not
a duration: a commander that dies in response does not take a bullet
off the stack. Nothing new is stored — the multiset of chosen modes
already IS the record, and a restore or a copy (CR 707.10) carries it
as it always has.

A trigger's `mode_pick` prompt carries the raised bound in
`PendingChoice.ModeMax`, so `validateModePick` and the enumerator's
answer walk (`legal/choices.go`) read the number the prompt was
queued with. No catalogued trigger uses this today (SOLDIER Military
Program is the first one that would), but a ModeSpec has three owners
(§3) and a bound only one of them honoured would be the #544 bug
waiting for its card.

### Decision: the wire's `max` is the caster's bound; the public copy is the printed one

`ModeSpecView.max` is stamped per caster from `ModeMaxForEffect` in
`castStampsFor` and in the activated-ability rows, so
`ModePickerModal.svelte` — which already bounds its selection by
`spec.max` — offers "both" exactly when the gate would accept it,
with no client change. `publicModeSpec` (#1172) puts the printed
`Max` back on the public copy: the raise is the asking seat's answer,
exactly as `legal_targets` is, even though the board it reads is
public. No new wire field.

### Decision: the enumerator widens by the same accessor

`legal.legalModeSets` takes its upper bound from `ModeMaxForEffect`
for the enumerating seat, so a bot without a commander is never
offered Jeska's Will's "both" and a bot with one is.
`TestJeskasWillBothIsOfferedOnlyWithACommander` dispatches every
offer against the engine either way (#544).

### Decision: `Register` refuses a half-declared or non-raising count

`checkRaisedModeMax` panics at boot on `RaisedMax` without
`RaiseMaxIf` (or the reverse), on a raise that does not exceed a
bounded `Max`, and — for a non-repeatable spec — on a raise past the
number of printed bullets. Each would register silently and ship a
mode count the printed card does not have.

### Consequences

- Jeska's Will, Akroma's Will, Drown in Dreams, Will of the Mardu,
  Will of the Abzan and Flame of Anor ship `CompletenessFull`. The
  four whose bodies read `item.Targets[0]` or walked every target
  against every chosen bullet now read each bullet's own target group
  through `effects.OptionTargets(ctx, option)`, in PRINTED order from
  `OnResolve` — which is observable on Will of the Mardu (the Warriors
  are made before the damage counts your creatures) and Drown in
  Dreams (draw, then mill).
- `docs/protocol.md`'s `cast_spell` row says what `modes.max` means.

### Out of scope, stated

- **"Choose both instead" with no "may"** — the teamwork cards (HULK
  SMASH!, Go Nuts!), Inscription of Ruin's "if this spell was kicked,
  choose any number instead", Depth Defiler, Pyrrhic Strike. Their
  condition is the caster's own optional-cost choice made in the SAME
  announcement (ADR 0073), and some force the higher count rather than
  permit it, so the bound depends on `params.OptionalCosts` and `Min`
  moves too. The predicate here takes only the board.
- **Board conditions on a trigger that FORCE the higher count** —
  Prophetic Titan's delirium ("choose both instead", no "may"). The
  trigger owner already reads the raised bound as the trigger goes on
  the stack, but only `Max` is raised; a forced count needs `Min`
  raised with it. Captain Kirk's "choose one or more instead" keeps
  `Min` at one and would fit as is. Neither is catalogued by this
  change.

---

## Amendment (2026-09-28, #1563): divided damage — "divided as you choose" (CR 601.2d / 700.2i)

**Sprint:** S45 — Modal spells, multi-target clauses and copy effects.
Tracker [#888](https://github.com/krakenhavoc/cmd_and_ctrl/issues/888);
the proof card is Shatterskull Smashing on the Edea deck tracker
[#1565](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1565).

"Out of scope, stated" above named this: `StackItem.Distribution` rode
the stack item — deep-cloned, snapshotted, copied, remapped by a CR
115.7 retarget — and nothing wrote it at announce or read it at
resolution. Every catalogued divided card (Fury, Dragonlord Atarka,
Shatterskull Smashing, and Abzan Charm's "distribute" bullet) shipped
the division made FOR the player — an even split in pick order, or all
of it on one target — with a caveat. Weaker than printed, the #259
posture, and it is what this amendment removes.

### Decision: the amount is a clause property, as data

```go
type DivideSpec struct {
    Total       int  // a fixed amount (Fury's 4)
    FromX       bool // the announced X (Fire Covenant)
    DoubleFromX int  // with FromX: twice X from this X up (Shatterskull Smashing's 6)
}

type TargetSpec struct {
    // … every #764 / #1559 field …
    Divide *DivideSpec
}

// card file
Targets: TargetPermanent("up to two target creatures and/or planeswalkers",
    Or(Creature(), Planeswalker())).WithCount(0, 2).Dividing(DivideXDoublingFrom(6)),
```

On the CLAUSE, because CR 601.2d divides among the targets a clause
chose, and a clause is what every reader already walks: the announce
gate, the CR 608.2b re-check, the view projection and the enumerator
all iterate `AnnouncedClauses`, so a divided clause on a MODE (Abzan
Charm's third bullet) needed no second path. `effects.Divide(n)`,
`DivideX()` and `DivideXDoublingFrom(n)` are the constructors;
`game.(*DivideSpec).TotalFor(x)` is the one place the amount is
computed, mirrored once in the client.

**Data, not a func.** A trigger's clause is reachable from `Game`
through the `pick_target` resume frame, and ADR 0041 phase 3's ratchet
lets func routes only shrink. Every printed divided card measured
against the Sep-23 dump is one of the three shapes above — a constant,
X, or Shatterskull's doubled X — so a `func(x int) int` would have
bought nothing but a closure route. And the client has to compute the
same amount from the X it collected, which a func cannot ship.

`Register` refuses a fixed amount below 1, a doubling without
`FromX`, a divided clause with `AllowSame` (the division is keyed by
target id, so two picks of one object could not be told apart), and
`DivideX` on a trigger's clause (a trigger announces no X).

### Decision: the division is announced with the targets, and one gate judges it

`settleDistribution(steps, targets, dist, x)` runs straight after
`validateAnnouncedTargetsLocked` at all three announce points — a cast
(CR 601.2d), an activation (CR 602.2b) and each step of a trigger's
`pick_target` walk (CR 603.3d, `ResolvePickTargetsDivided`). For every
divided step it demands each target at least 1 and the shares summing
to the clause's amount under the announced X, and it refuses a step
with more targets than the amount (some target would get 0). A share
naming anything that is not a target of a divided step is refused, as
is a division sent with no divided clause at all — a confused client,
not something to drop quietly. One convenience: a step with ONE target
and no share named takes the whole amount, because there is nothing
to choose.

The settled map is what `StackItem.Distribution` stores. A trigger
walk accumulates it on the `pickTargetFrame` (`dist`, deep-copied by
`clonePickTargetFrame`) and stamps it on the item with the targets.

The S13.1 free-form path — a card with no structured clause — keeps
the division it was sent, unjudged, exactly as it keeps its targets:
the sandbox records it for the table.

### Decision: resolution honours the announcement; a departed share is lost

`effects.DealDividedDamage(ctx)` (and `PutDividedCounters(ctx, kind)`
for "distribute") walks `ctx.LegalTargets()` — the CR 608.2b re-check
— and deals each survivor exactly its announced share. A target that
left or stopped qualifying takes nothing, and its share is NOT moved
to the others: the division was made at announce and CR 608.2b says
only that an illegal target is unaffected. That is observably
different from the old even split over the SURVIVORS, which dealt a
departed target's share to someone else.

Nothing else was needed for undo, snapshot and copy: the field was
already carried by all three, and CR 707.10's copy (and 707.10c's new
targets, through `remapDistributionLocked`) keeps it. Those paths are
now pinned by tests rather than assumed.

### Decision: the wire says the amount; the client asks for the split

`LegalTargetsView.divide` — `{ total?, from_x?, double_from_x? }` —
rides every target clause's view, including a `pick_target` prompt's.
`cast_spell` and `activate_ability` already had a `distribution` field
(S13.1 data capture); `resolve_choice` gains one for the `pick_target`
answer. The client's target walk, on completing a divided step with
two or more picks, opens a small per-target number picker seeded with
the even split and confirms only when every share is at least 1 and
they sum to the amount; one pick needs no picker.

### Decision: the bot announces the even split

`game.EvenDistribution(steps, targets, x)` — the amount split as
evenly as possible, the remainder one point at a time to the earliest
targets — is what `internal/legal` puts on every divided cast,
activation and `pick_target` answer, and it returns `ok = false` for a
target set larger than the amount, which the enumerator then does not
offer (#544). A `pick_target` prompt's upper bound is also capped at
the amount, so "any number of targets" (Bogardan Hellkite) does not
spend the expansion budget on sets the gate refuses. The even split is
a legal DEFAULT, not a policy: a smarter split (lethal first) is an
aiseat question for later.

### Consequences

- Fury, Dragonlord Atarka, Shatterskull Smashing and Abzan Charm ship
  `CompletenessFull`. `b22DamageDividedEvenly` is deleted.
- New: Inferno Titan (a bounded divided trigger), Bogardan Hellkite
  ("any number of targets"), Fire Covenant (an X paid in life, the
  amount the same X) and Mogg Mob (the activated-ability path).
- `x_matters_guard_test.go` counts a `FromX` declaration as reading X,
  because the engine reads it for the card at announce.

### Out of scope, stated

- **An amount read off the board or a paid cost** — Lathiel's "up to
  that many" (life gained this turn, and "up to"), Ureni's "X is the
  number of …", Orca's "equal to its power", Avacyn's Judgment's
  "if this spell's madness cost was paid, X instead".
  `DivideSpec` has a constant and an announced X; none of these is
  either, and Lathiel's "up to" also lets the shares sum to LESS than
  the amount. Lathiel keeps its caveat.
- **Division across two clauses**, and a divided clause that allows
  the same object twice. No printed card does either; both are refused
  rather than guessed at.
- **Prevention divided among targets** (Remedy) and **"distribute
  counters" with no targets** (Feast of the Victorious Dead) — the
  first is a prevention shield, the second is not a target clause at
  all.

## Amendment (2026-09-28 (b), #1657): a divided amount read off the board or a paid cost, and "up to" division

**Sprint:** S45 — Modal spells, multi-target clauses and copy effects.
Tracker [#888](https://github.com/krakenhavoc/cmd_and_ctrl/issues/888).

The amendment above left two shapes out, by name: an amount that is
neither a constant nor the announced X, and Lathiel's "up to that
many", whose shares may sum to less than the amount.

    Ureni, the Song Unending  X damage divided …, where X is the number of lands you control
    Orca, Siege Demon         When Orca dies, it deals damage equal to its power divided …
    Lathiel, the Bounteous    distribute up to that many +1/+1 counters among any number
      Dawn                    of other target creatures (that many = life gained this turn)
    Avacyn's Judgment         2 damage divided …; if its madness cost was paid, X instead

### Decision: an amount RULE, named by a key, that answers in the existing data shapes

```go
type DivideSpec struct {
    Total, FromX, DoubleFromX …  // unchanged
    AmountKey DivideAmount       // a registered rule; replaces the three above
    UpTo      bool               // the shares may sum to at most the amount
}

var DivideLandsYouControl = game.RegisterDivideAmount("lands-you-control",
    func(g *game.Game, a game.DivideAmountArgs) game.DivideSpec {
        return game.DivideSpec{Total: b02CountLandsControlledBy(g, a.Controller)}
    })

Targets: TargetPermanent(…).WithCount(0, 0).Dividing(DivideBy(DivideLandsYouControl)),   // Ureni
Targets: TargetCreature(…).WithCount(0, 0).Dividing(UpTo(DivideBy(DivideLifeYouGainedThisTurn))), // Lathiel
```

A **key, not a func**, for the reason `ModeCountCondition` (#1590) and
`LifeCostCount` (#1594) are keys: a trigger's clause is reachable from
`Game` through the `pick_target` resume frame, and ADR 0041 phase 3's
closure ratchet admits no new func-typed route. The functions live in a
registry in `game/divide.go`, registered once at init by
`RegisterDivideAmount`; the cards' rules are in
`effects/divide_amounts.go`.

The rule reads `DivideAmountArgs` — the announcing player, the source,
a trigger's source LAST-KNOWN characteristics (CR 603.10; Orca's power
after it died, counters included), and the claimed alternative cost
(the key that lands on `StackItem.AltCost`). It **answers in the data
shapes the engine already has**: `{Total: n}`, or `{FromX: true}` for
Avacyn's Judgment's madness cast, so the announced X is still applied by
`TotalFor` and — the reason this matters — the client can still apply
it to the X it collected. The earlier amendment's argument against a
func ("the client has to compute the same amount from the X it
collected, which a func cannot ship") is met: the server evaluates the
rule and ships its answer, never the rule.

`Register` refuses a clause that names a rule AND a fixed or X amount.

### Decision: the rule is read ONCE, at announce, into the steps

`bindDivideAmountsLocked(steps, args)` replaces each step's
`Clause.Divide` with the rule's answer — a fresh spec, since the steps
hold clause copies (`AnnouncedClauses`), so the catalog's clause is
never touched. It runs at the three announce points:

- a cast, straight after X is bound (`CastSpell`), with the claimed
  alternative cost;
- an activation (`activateCatalogAbilityLocked`);
- a trigger, when its target walk OPENS (`queuePickTargetLocked`), with
  the harvester's LKI.

The last is the load-bearing choice. CR 603.3d puts the targets and the
division on the ability "as it is put on the stack", which in this
engine is the walk; the frame carries the bound steps, so the prompt,
its wire projection (`pick_target.divide`), the enumerator's cap and
the gate all read one number, and a land that arrives while the prompt
is open changes nothing. That is Ureni's ruling ("The value of X won't
change even if the number of lands you control changes after that
point") and Lathiel's ("Gaining more life in response … won't change how
many counters will be distributed"). Nothing re-reads the rule at
resolution: the item carries `Distribution`, as before.

`settleDistribution` and `EvenDistribution` are unchanged in shape —
they read bound steps. The frame is cloned by `clonePickTargetFrame`
(the bound spec is immutable and shared), and a snapshot drops the
frame exactly as it always has (the census records it), so undo and
snapshot need nothing new.

### Decision: "up to" relaxes the sum, not the minimum

`UpTo` changes one comparison in the gate: the shares must sum to **at
most** the amount rather than exactly. Each chosen target still gets at
least 1 — CR 601.2d, and Lathiel's own ruling ("Each target must
receive at least one +1/+1 counter") — so an up-to clause still takes
no more targets than its amount. Zero targets is the clause's Min, as
for any clause ("You may choose no targets if you want"). The bot's
even split of the whole amount is a legal up-to answer, so the
enumerator needs nothing new; it offers the empty answer because Min is
0.

### Decision: the view resolves the rule for the viewer

`viewOfTargetClause` and `abilityLegalTargets` stamp the resolved
amount (`stampDivideAmount`) for the source's controller and — on an
`alternative_costs[]` entry — that offer's key, so Avacyn's Judgment's
hand card says `{total: 2}` and its madness offer says `{from_x:
true}`. `DivideView` gains `up_to`. Writing this found that
`abilityClauseView` never carried `divide` at all, so a single-clause
divided ability (Mogg Mob) reached the client with no amount and the
picker never asked for the split the gate demands for two or more
targets; `abilityLegalTargets` now stamps it.

The client's `divisionProblem` takes `upTo`, the divide modal and the
targeting banner say "up to", and `hasXCost` asks the claimed offer's
own mana cost, because Avacyn's Judgment prints `{1}{R}` and its madness
cost is `{X}{R}`: the X prompt has to open for the madness cast.

### Consequences

- Lathiel, the Bounteous Dawn ships `CompletenessFull`; its round-robin
  body is deleted.
- New, all `full`: Ureni, the Song Unending, Orca, Siege Demon, and
  Avacyn's Judgment (#653's "if its madness cost was paid").
- The enumerator binds the same amount before the even split, in the
  cast (with the offer's key) and activation expansions.

### Out of scope, stated

- **Polukranos, World Eater** — "When Polukranos becomes monstrous, it
  deals X damage divided …", X being the monstrosity activation's X.
  The divided half would be `DivideBy` a rule reading that X off the
  triggering event; the engine has no monstrosity (CR 701.37) at all,
  and no "becomes monstrous" event. The card waits on that keyword.
- **Divided prevention** (Remedy) and **distribute without targets**
  (Feast of the Victorious Dead), as above.

---

## Amendment (2026-09-28 (c), #1655): a mode count read off the optional costs, and a forced higher count

**Sprint:** S45 — Modal spells, multi-target clauses and copy effects.
Tracker [#888](https://github.com/krakenhavoc/cmd_and_ctrl/issues/888).

The 2026-09-27 amendment's "Out of scope" named two shapes #1590 could
not express, and they share a fix:

- *"Choose one. If this spell was kicked, choose any number instead"*
  (the Inscription cycle), *"If it was kicked, choose both instead"*
  (Depth Defiler). The condition is the caster's own optional-cost
  choice, made in the same CR 601.2b announcement as the modes, not
  the board.
- *"If there are four or more card types among cards in your
  graveyard, choose both instead"* (Prophetic Titan). There is no
  "may", so the higher count is **forced** and `Min` has to rise with
  `Max`.

### Decision: `RaisedMin` beside `RaisedMax`, and three declarations

```go
type ModeSpec struct {
    // … Min, Max, Repeatable …
    RaisedMax  int
    RaisedMin  int                // new: 0 = Min does not move
    RaiseMaxIf ModeCountCondition // governs both; the #1590 name is kept
}

ChooseOne(…).OrUpToIf(2, YouControlACommander) // "you may choose both instead" — #1590, unchanged
ChooseOne(…).InsteadIf(2, WasKicked)           // "choose both instead": Min = Max = 2
ChooseOne(…).AnyNumberIf(WasKicked)            // "choose any number / one or more instead": Max = every bullet, Min stays
```

"Choose any number instead" keeps a minimum of one. The spell is still
a "choose one" card whose count widens; announcing no bullet at all is
refused, kicked or not. `checkRaisedModeMax` now also refuses a
`RaisedMin` that does not exceed `Min` or that exceeds `RaisedMax`.

### Decision: one reader over a query, not the chooser

`modeMaxLocked(ms, chooser)` becomes `modeBoundsLocked(ms, q)`
(`ModeBoundsForEffect` outside the package). It returns `(lo, hi)` for
a `ModeCountQuery`:

```go
type ModeCountQuery struct {
    Chooser       uuid.UUID
    OracleID      string // the card whose OptionalCosts index space OptionalCosts names
    OptionalCosts []int  // what was announced WITH the modes
}
```

The predicate registry keys a `func(*Game, ModeCountQuery) bool`.
`game.ModeCondition(key, func(g, chooser))` keeps its #1590 signature
and adapts. The new `game.ModeConditionOnAnnouncement(key, func(g, q))`
registers a predicate that reads the announcement. It is still a
**key, not a closure** (ADR 0041 phase 3: the closure ratchet gains no
route). `q.Kicked()` / `q.OptionalCostTimes(key)` count by the cost's
`Key` rather than its index, so a condition never knows its card's
declaration order. `effects.WasKicked` is the first announcement
condition. `effects.DeliriumForModes` and `effects.Descended8` are
board conditions.

Where each owner gets its query:

- **A spell** (`castSpellLocked`): `CastSpellParams.OptionalCosts`.
  Mode validation moved BELOW `validateOptionalCostChoice` and the gift
  check, so the count is read against an optional-cost choice that has
  already been validated. CR 601.2b announces the modes and the
  optional costs in one step, and ADR 0089's gift already reads that
  step's optional costs to rewrite the target clause (CR 601.2c comes
  after). The mode count is the same read, one clause earlier.
- **A trigger** (`queueModePickLocked`) and **an activated ability**
  (`activateCatalogAbilityLocked`): `modeQueryForSourceLocked(source)`
  reads the source spell's `PaidCost.OptionalCosts` while it is on the
  stack. That is Depth Defiler's "when you cast this spell … if it was
  kicked": its trigger is harvested with the spell still on the stack
  and its payment recorded. Once the source has entered, the query
  reads the permanent's `CastProvenance.OptionalCosts` (ADR 0073 §5).
  No printed card needs the second path yet; it is one line and it
  keeps the three owners from answering "was it kicked" three ways.

### Decision: a forced count on a trigger asks for what is on offer

CR 603.3c: a bullet whose targets cannot be chosen "can't be chosen";
it does not remove the ability. So when a raised minimum exceeds the
number of choosable bullets, the `mode_pick` prompt's `ModeMin` drops
to the choosable count, and never below the printed `Min`, which
`EnoughChoosableModes` has already enforced. A spell has no such
clamp. A kicked forced-both spell whose second bullet has no target is
refused at announce, like any spell that cannot fill its targets, and
the enumerator offers no move for it (`legalModeSets` returns nothing
when `lo` exceeds the choosable options).

### Decision: the wire carries both ranges; the picker switches on the kicker

`ModeSpecView.min` / `.max` are the caster's bounds with **no**
optional cost announced; `min` can now differ from the printed one (a
board-driven forced count). `if_optional_paid: {min, max}` carries the
bounds with every optional cost the card offers announced once. It is
stamped only when those differ, so it is absent for every card but the
kicker-driven ones. The client's cast flow asks the optional costs
BEFORE the modes (`confirmAltCost` → `continueCast`), so
`modesUnderChoices` opens the picker on `if_optional_paid` when the
caster ticked one. A picker whose `min == max > 1` reads "choose N"
rather than "choose up to N". `publicModeSpec` restores the printed
`min` and `max` and drops `if_optional_paid`, as it drops #1590's raised `max`.

"Every optional cost once" is exact for every card on this shape: each
offers one optional cost. A card that offered two, with a count
depending on only one of them, would need a per-cost list. None is
printed.

### Decision: bullets whose order is observable run in printed order

The engine walked `ModeOption.Effect` in announce order (§3 above;
superseded by the 2026-09-30 amendment below).
Three of this change's cards have bullets that see each other: Depth
Defiler's bounce feeds its discard, Wail of the Forgotten's bounce
feeds its discard, and Inscription of Abundance's counters feed its
"greatest power" and its fight. They declare their bullets with `Mode`
and run them through `effects.BulletsInPrintedOrder` from `OnResolve`
(a spell) or the ability's `Effect` (a trigger), as the #1590 Wills do
with `if ctx.HasMode(i)` chains. The other three use `ModeDoing`,
because their bullets cannot observe one another.

### Consequences

- Inscription of Ruin, Inscription of Abundance, Depth Defiler,
  Prophetic Titan, Let's Play a Game and Wail of the Forgotten ship
  `CompletenessFull`.
- `legal.legalModeSets` takes the query, so each optional-cost
  announcement (`optionalCostSets`) gets its own mode range. Unkicked
  Inscription moves offer exactly one bullet and kicked ones up to
  three, and every move dispatches (#544,
  `TestInscriptionOfRuinModeRangesFollowTheKicker`).
- A trigger prompt's `ModeMin` is persisted as it always was
  (`snapshot.go`), so a forced minimum survives a restore and an undo
  (`TestPropheticTitanForcedCountSurvivesRestoreAndUndo`).
- `docs/protocol.md`'s `cast_spell` row documents `if_optional_paid`
  and the moving `min`.

### Out of scope, stated

- **Teamwork** (HULK SMASH!, Go Nuts!, Atlantis Attacks, Murdock's
  Crusade, Widow's Bite: "you may tap any number of creatures you
  control with total power N or more") and **blight** (Pyrrhic
  Strike: "you may put two -1/-1 counters on a creature you
  control"). Both are optional costs `AdditionalCost` has no component
  for (a tap-creatures-by-total-power payment, a counter placement).
  Once either exists, the card's count is `InsteadIf(2, …)` over a
  condition that reads `q.OptionalCostTimes`. Nothing in this
  amendment changes for them.
- **The engine's announce-order walk of `ModeOption.Effect`.** This
  amendment works around it card by card and does not re-decide §3.
- **Inscription of Insight.** Its "scry 2, then draw two cards" bullet
  queues the scry prompt and draws in the continuation, so a kicked
  cast that also takes "X is the number of cards in their hand" would
  count the hand BEFORE the draw. Printed order needs a bullet walk
  that waits for a prompt to be answered, and nothing has one yet.

## Amendment 2026-09-30: chosen modes run in printed order (#1653)

**Decision (owner, 2026-09-30).** A modal spell or ability carries out
its chosen modes in the order they are WRITTEN on the card, not the
order the player announced them in. CR 608.2c: "The controller of the
spell or ability follows its instructions in the order written." This
reverses §3's "resolution order is announce order", which the #1655
cards had to work around.

**What changed.**

- `game.PrintedModeOrder(modes)` gives the occurrence indexes sorted by
  option, with repeats of one option (CR 700.2d) keeping their announce
  order. `runChosenModeEffectsLocked` walks it, for spells, triggers
  and activated abilities alike, so every `ModeOption.Effect`
  (`ModeDoing`) runs in printed order.
- **Storage is unchanged.** `StackItem.Modes` and `TargetRef.Mode` stay
  in announce order, because an occurrence's index is what pairs it
  with its target group. Only the order the bodies run in moves, so
  `ModeTargets(occurrence)` and `item.Targets` still name the right
  targets after the reorder. The announce-time prompts, the wire's
  `modes` and the bot's move list are untouched. `mode_labels` on the
  stack item is now in resolution (printed) order, which is what the
  overlay is for.
- `effects.Context.ModeOccurrences()` is the same order for an
  `OnResolve` that walks occurrences by hand. `if ctx.HasMode(i)`
  chains were already printed order.
- `effects.BulletsInPrintedOrder` is deleted. Depth Defiler, Wail of
  the Forgotten, Inscription of Abundance, Atlantis Attacks, HULK
  SMASH!, Murdock's Crusade, Pyrrhic Strike and Widow's Bite now
  declare `ModeDoing` bullets and no `OnResolve`. (Depth Defiler keeps
  an empty trigger `Effect`, which a row with a `Build` must declare.)

Tests: `modes_printed_order_test.go` announces a spell, a triggered
ability and an activated ability in reverse and a repeatable mode
interleaved, and checks both the run order and each occurrence's
target.

## Amendment 2026-10-09: a later bullet waits for an earlier bullet's prompt (#2789)

**The gap.** The 2026-09-30 amendment made the bullets RUN in printed
order, but a bullet that asks something (a library search, a scry, a
discard) only queues its prompt and returns, and the walk then ran the
next bullet before the answer. Titania's Command put its +1/+1 counters
out while its land search was open, so a creature land it found missed
them; Insatiable Avarice drew three before the tutored card was on top.
The "Out of scope, stated" note above (Inscription of Insight: "Printed
order needs a bullet walk that waits for a prompt to be answered, and
nothing has one yet") named the same gap. CR 608.2c's "follows its
instructions in the order written" was already decided; this amendment
only makes the walk honour it across a prompt.

**What changed.**

- `runChosenModeEffectsLocked` checks, before each bullet, whether the
  resolution is paused on one of its own prompts (#1289's
  `pausedResolutionChoiceLocked`). If it is, the bullets not yet run
  are parked in `Game.pausedModeWalk` (`game/mode_walk.go`) and the
  walk returns. The resolution stays open, so the CR 704.3 boundary is
  held, and a prompt the answer chains on is stamped as part of the
  resolution and keeps the walk parked.
- When the last such prompt is answered or withdrawn,
  `holdForOpenResolutionLocked` resumes the walk instead of closing the
  resolution. A resumed bullet may pause again; the boundary runs only
  after the last bullet.
- Checked before each bullet, not after, so an `OnResolve` or trigger
  body that paused holds the bullets too.
- The parked walk is data: the stack item (through the ordinary stack
  item mirror, which re-derives the bullets from the oracle ID or the
  ability's catalog row), the spell's catalog key and the occurrences
  still to run. Clone and the snapshot both carry it
  (`GameSnapshot.PausedModeWalk`, additive in schema v7), so it adds no
  restore-point blocker. A binary from before this change drops the key
  and would lose the parked bullets of a game restored mid-prompt.
- Applies to `ModeDoing` bodies only. An `if ctx.HasMode(i)` chain in
  one `OnResolve` cannot be suspended, so Titania's Command moved to
  `ModeDoing` bullets. The other modal cards whose prompting bullet is
  followed by another bullet already declare `ModeDoing` and are fixed
  by the engine change: Insatiable Avarice, Kolaghan's Command,
  Mishra's Command, Prismari Command and Kozilek's Command.

Tests: `titanias_command_test.go` (the counters and the Bears wait for
the search; a creature land found gets the counters) and
`TestInsatiableAvariceDrawsTheTutoredCard`.


## Amendment — 2026-10-11: related target clauses (#1943, #2033)

Sprint S58, tracker #2077. A later target can now depend on a required,
single target in an earlier clause of the same mode occurrence. The
catalog declares `TargetSpec.Relation` as data: an earlier clause slot and
either `controllers_graveyard` or `attached_to`. The card-facing builders
are `InGraveyardOfTargetsController` and `AttachedToTarget`. Registration
rejects a forward reference, a non-single anchor, an unknown relationship,
or a dependent clause with more than one pick.

Announcement and resolution check the relationship against the current
objects (CR 601.2c, 608.2b). A missing anchor supplies no relationship.
Required dependent clauses also narrow the earlier clause's legal set:
a first pick with no possible second pick is never offered. The bot
walk and trigger prompts apply the same gate to their chosen prefix.
The wire clause includes `relation: {slot, pairs}`, where `pairs` maps
public anchor IDs to compatible target IDs; the client narrows its next
picker and skips an optional clause with no compatible candidates.

Copy and retarget prompts judge the completed target set (CR 115.7).
They can change both halves of a related pair. A related retarget walk
holds `PendingChoice.RetargetNext` as data and commits it only when all
slots have been answered. Clone deep-copies that prefix; schema v8
carries it across restore. A v7 binary would lose the uncommitted prefix,
so the schema bump makes that reader refuse the newer restore point. An unchanged illegal target
may remain, but changing the anchor rechecks a previously legal dependent relationship;
an unchanged target that was already illegal may still remain (CR 115.7c).

Goblin Welder is the first graveyard caller. Both targets must remain
legal before the exchange begins. Its entry batch settles the incoming
artifact's replacement window against the original board, then a routed
sacrifice settles the outgoing artifact's window. Neither card moves
while either window waits for a choice. The settled incoming entry travels
on the exit route; both land before either move is announced. The new
arrival is excluded from look-back leaves/dies harvesting, while ordinary
observers see the completed board. Existing replacement resume cloning
copies the paired entry data for undo. The settled entry carries only
landing data, with no new closure route through the replacement event. Sacrifice protection or a replacement
of either move does not cancel the other instruction.

Fiery Annihilation uses the attachment relationship. If its creature
target is illegal, it does nothing, including leaving the Equipment alone,
as the Foundations release notes specify. Legal damage is followed by
Equipment exile and the existing exile-instead-of-dying rider.

This was a partial seam: Breaking of the Fellowship waited for a
same-controller relationship that can read the first target's last-known
controller after that target becomes illegal. The 2026-10-15 (b) amendment
below adds it.

## Amendment 2026-10-15 (b): the same-controller relation and an illegal anchor (#1943)

Sprint S58, tracker #2077. This closes the partial seam the 2026-10-11
amendment left open. A third relationship, `same_controller`
(`game.TargetControlledBySameController`, card-facing builder
`ControlledByTargetsController`), is "another target creature that player
controls", where that player controls the earlier target. Registration
accepts it under the same rules as the other two: a required single
anchor and at most one dependent pick.

**"That player" is read from the anchor as CR 608.2h reads any object.**
While the anchor is on the battlefield, its current controller decides,
even if the anchor has become an illegal target (it gained hexproof, or
changed controller to another opponent). Once it has left the
battlefield, its last-known controller decides
(`Game.LastKnownPermanentForEffect`, the turn-long record #1379 added). Only this relationship falls back to last-known
information: the attachment and graveyard relationships still need a
present anchor, because an Equipment cannot be attached to a creature
that is gone and Goblin Welder's exchange needs both halves.

The fallback matters because it decides whether the spell resolves. If
Breaking of the Fellowship's first creature dies in response, the second
creature, still under that player's control, is still a legal target
(CR 608.2b), so the spell resolves and the Ring tempts its caster. No
damage is dealt, because an illegal target performs no action. Had the
relation answered "no information", both targets would be illegal, the
spell would not resolve and the Ring would not tempt.

The wire, the client and the bot need nothing new: `relation.pairs`
already carries the compatible second picks for each anchor, and the
enumerator already narrows a first pick that has no possible second.

Card: Breaking of the Fellowship (Full). Tests:
`TestBreakingOfTheFellowshipRechecksTheRelationOnResolution` (either
creature changing controller, the first dying, the first dying and the
second changing controller, the first gaining hexproof),
`TestBreakingOfTheFellowshipRefusesAnUnrelatedSecondTarget`,
`TestBreakingOfTheFellowshipEnumeratesSameControllerPairs` and
`TestSameControllerRelationViewShipsCompatiblePairs`.

## Amendment 2026-10-15: tiered is Spree with exactly one bullet (#3777)

CR 702.183a: "Tiered represents a static ability found on modal spells.
'Tiered' means 'As an additional cost to cast this spell, choose one of
its modes and pay the additional cost associated with that mode.'"
That is the 2026-09-23 Spree shape with the count fixed at one, so it
is a constructor and not a new field:

```go
Modes: Tiered(
    SpreeModeDoing("Cross-Slash — Destroy target tapped creature.", "{0}", …),
    SpreeModeDoing("Blade Beam — …", "{1}", …),
    SpreeModeDoing("Omnislash — Destroy all tapped creatures.", "{3}{W}", nil, …),
),
```

`effects.Tiered` is `Min: 1, Max: 1` with `effects.TieredPrompt`, which
the roadmap's `tiered` row matches the way the spree row matches
"Spree". Everything else is Spree's and unchanged: the bullet's cost is
summed into the cast price at the optional-cost point (ADR 0073 §3), the
enumerator prices each selection on its own (#544), and the client's
mode picker shows each bullet's cost.

A `{0}` bullet is written `"{0}"`, not `""`: both price the same, but
the printed tier is a cost and every tiered bullet prints one, which
`TestTieredIsExactlyOneMode` pins.

Proof cards (all Full): Cloud's Limit Break, Thunder Magic, Fire Magic,
Tifa's Limit Break, Restoration Magic.

## Amendment 2026-10-16: entwine is an optional cost that raises the count to every mode (#3866)

**Sprint:** S76 — Cards: incubate, face-down and outside the game.

CR 702.42a: "'Entwine [cost]' means 'You may choose all modes of this
spell instead of just the number specified. If you do, you pay an
additional [cost].' Using the entwine ability follows the rules for
choosing modes and paying additional costs in rules 601.2b and
601.2f–h." CR 702.42b: "If the entwine cost was paid, follow the text of
each of the modes in the order written on the card when the spell
resolves."

"Out of scope, stated" above left entwine for its own design, as closer
to an `AlternativeCost` than to Spree's per-bullet price. Neither is the
shape it ended up with: entwine is one flat cost paid once, chosen at CR
601.2b together with the modes, and paid at 601.2f–h like any additional
cost. That is exactly an ADR 0073 optional additional cost, and the mode
choice it buys is exactly the 2026-09-28 (c) amendment's
announcement-driven count with the minimum raised too.

### Decision

- **The cost** is an optional additional cost keyed `game.EntwineKey`
  ("entwine"), paid at most once: `effects.Entwine("{2}")`, or
  `effects.EntwineSacrifice(n, label, preds…)` for Betrayal of Flesh's and
  Solar Tide's "Entwine—Sacrifice N lands" (the counted clause
  `SacrificeNCost` builds). Its mana joins the total at CR 601.2f through
  the one pricer, its sacrifice rides `sacrifice_ids` and is paid with the
  spell on the stack, and it is recorded on `PaidCost.OptionalCosts`, so a
  CR 707.10 copy of an entwined spell is entwined.
- **The count** is `effects.Entwining(ms)`: `ms.InsteadIf(len(ms.Options),
  effects.WasEntwined)`. `WasEntwined` is a `ModeConditionOnAnnouncement`
  reading the entwine cost off the announcement (CR 601.2b), as `WasKicked`
  does. Unentwined, the printed Min..Max stands; entwined, every bullet and
  no fewer. A count between them is refused — Kaya's Guile ("Choose two …
  Entwine {3}") is two or all four, never three. Choosing every bullet
  without paying is refused because it exceeds the printed Max.
- **The order** is CR 702.42b's, which is the printed-order walk every
  modal spell takes since the 2026-09-30 amendment; nothing new.
- **Register** (`checkEntwine`) refuses either half without the other, two
  entwine costs, a repeatable cost, a Repeatable spec, entwine beside
  escalate, and an entwine count on a trigger's or an activated ability's
  modes (CR 702.42a is a static ability of a modal spell).

### What it reuses, unchanged

The cast gate (`modeBoundsLocked` with the announced optional costs), the
hand card's `modes.if_optional_paid`, the client's optional-cost offer
("Entwine {2}" beside kicker) and its `modesUnderChoices` switch, and the
bot enumerator's one expansion per announced optional-cost set — all
already read one `game.ModeBoundsForEffect`, so none can offer a selection
another refuses. No wire field, no snapshot field and no closure route is
added. The one client change is a convenience: when the ticked optional cost
raises the count to every option, the picker opens with them all chosen.

### Alternatives considered

- **A `ModeSpec.Entwine` cost, like `ModeSpec.Escalate`.** Rejected: it
  would duplicate what ADR 0073 already prices, validates, pays, records and
  shows, and a non-mana entwine (sacrifice three lands) would need a second
  sacrifice wire beside the optional cost's — the reason escalate's struct
  stayed mana, discards and taps only.
- **Choosing every mode as the announcement, with the cost implied.** Equal
  rules outcome, but the cost would then be priced from the mode count
  rather than announced, and the client and enumerator would each need a
  rule for "all modes means pay more". Announcing the cost and deriving the
  count is the direction the 2026-09-28 (c) amendment already built.

Proof card: Spectral Shift (Full), with the other entwine cards whose
bullets the engine expresses.

## Amendment 2026-10-16: a Do's later steps wait for an earlier step's prompt (#4012)

**The gap.** The 2026-10-09 amendment parked a modal item's later
bullets behind an earlier bullet's prompt. One level down, the steps
of a single effect still ran back to back. `effects.Do(steps...)` is
how most catalog abilities spell "do this, then that", and a step that
asks something (a scry with no `Then`, a CR 616 ordering prompt on its
damage, counters or life, a prevention follow-up's question from
ADR 0108's 2026-10-16 amendment) only queues the prompt and returns.
The next step then ran before the answer. Vivi Ornitier dealt its
damage while its +1/+1 counter waited on the Doubling Season /
Hardened Scales order. A `Do(Scry{N: 2}, DrawCards{N: 1})` would draw
the old top card. CR 608.2c is the same rule as before. This
amendment applies the mode walk's mechanism to steps.

**What changed.**

- Before each step, the `Do` asks the engine whether the resolution
  is paused on one of its own prompts
  (`Game.ParkEffectStepsForEffect`, `game/step_walk.go`). If it is,
  the position of that step is parked in `Game.pausedStepWalk` and the
  `Do` returns. The resolution stays open (#1289), so the CR 704.3
  boundary is held.
- `holdForOpenResolutionLocked` resumes the parked steps through
  `game.ResumeEffectSteps` (wired by `effects/do_walk.go`). They run
  after the damage instruction's own parked continuation
  (`followUpParks`, the earlier instruction) and before a modal item's
  next bullet and the resolution's tail, which are later ones. A resumed
  step may pause again. `finishResolutionTailLocked` waits for the walk
  as it waits for the bullets.
- Before the next step runs, the prevention follow-ups still owed run
  (`resumeStepWalkLocked`). A CR 615.13 announcement (#2047, Selfless
  Squire) or another follow-up held behind the answered question was
  owed by an earlier step's damage, so it happens before the next step
  (CR 615.5). If one of them asks, the walk stays parked behind that
  question.
- The parked position is data: the stack item and a path of step
  indexes. No closure is stored, so the closure ratchet does not move
  (`Game.pausedStepWalk` and `stepWalk.item` are `keyed`, like the mode
  walk's) and no census counter rises. Clone and the snapshot carry it
  (`GameSnapshot.PausedStepWalk`, additive in schema v9). A binary from
  before this change drops the key and would lose the parked steps of a
  game restored mid-prompt.
- **Only the `Do` that is the item's own Effect parks.** Resuming has
  to find the same steps again from data. A stamped ability's Effect is
  re-derived from its catalog row on restore (ADR 0041 P9), and the
  `Do` answers the existing steps probe (`doStepsOf`, #1968) with the
  same steps. The check is identity of the steps' backing array, so a
  second `Do` with equal primitives is not mistaken for it. A `Do` run
  from a hand-written closure (`Do(...)(g, item)`, an `InterveningIf`
  around a hand-written Effect, a `MayAtResolution` wrapper, a delayed
  trigger's keyed body) cannot be found again, so it runs its steps back
  to back as before. That is the old behaviour, not a new failure.
- **An `If` step is walked one level down.** The path has one index
  per level, so the steps inside an `If`'s `Then` park and resume too.
  `AtYourEndStepIfACreatureDied` wraps its steps that way. The condition
  is checked once, as the `If` step begins, and is not checked again on
  resume.

**Cards whose behaviour changes.** A catalog row whose Effect is a `Do`
with a step that can ask before its last step:

- CR 616 counter order: Vivi Ornitier.
- Damage to a player (CR 616 order, a prevention follow-up): Aurelia,
  the Law Above; Vindictive Vampire; Midnight Reaper.
- A life gain with two replacements to order, then something else:
  Dovin Baan, Orazca Relic, Woodweaver's Puzzleknot (both abilities),
  Guide of Souls, Reservoir Walker, Tatyova, Benthic Druid, Vaultborn
  Tyrant, Lotho, Corrupt Shirriff, Tangled Skyline.
- A draw or mill whose replacement asks, then something else: Bilbo's
  Ring, Chulane, Teller of Tales, Goldbug, Scrappy Scout, Insidious
  Fungus, Rogue Refiner, Spelunking, Inverted Iceberg.
- A colour pick for added mana, then the rest: Domri, Anarch of Bolas.

The seventeen drain rows (`EachOpponentLosesLife` or
`TargetPlayerLosesLife`, then `GainLife`) change only when a life-loss
replacement asks. On every one of these the later step now happens after
the answer. Before, it happened before the answer.

Tests: `effects/do_walk_test.go`. They cover a Khârn the Betrayer
question mid-`Do`, live and across a strict restore; Vivi's damage
waiting on the CR 616 counter order, live and across an undo, with the
census unchanged; a scry with no `Then` mid-`Do`, live and across a
strict restore; a scry inside an `If`'s steps across a restore; and a
first step whose one instance deals damage to Khârn and to a player
with Selfless Squire's shield, where the next step sees the Squire's
announcement already made, live and across a strict restore.

## Amendment 2026-10-19: a spell's OnResolve and a bullet's own instructions wait too (#4019)

**The gap.** The 2026-10-16 amendment parked a `Do`'s later steps only
when the `Do` was the item's own Effect, because only that `Do` could
be found again from data. Everything else still ran its next instruction
before an earlier one's answer: a spell's `OnResolve` (Electrolyze drew
its card while Khârn the Betrayer's question was open), the body of a
`ModeDoing` bullet (Demonic Pact gained its 4 life before the same
question was answered, because the mode walk parks between bullets, not
inside one), a `Do` wrapped by `InterveningIf` (it became one opaque
step), and an `If` inside any of those.

**Measured.** A scan of the catalog's function literals on 2026-10-19
found 380 `OnResolve` bodies, 147 trigger and activated `Effect`
bodies, 31 `ModeDoing` bodies and 70 `Then` bodies with two or more
statements that call an effect primitive. Almost any primitive can ask
(a CR 616 order on damage, counters, tokens, life or a draw; a
prevention follow-up's question; dredge), so each is a candidate. No
engine change can stop a Go function part-way and resume it after a
restore. So the fix is two more sites that can park, and a recipe for
converting rows to them. The rows are converted card by card.

**What changed.**

- **A spell's OnResolve as steps.** `effects.ResolveSteps(steps...)`
  is an `OnResolve` the registry can read: it answers a probe with its
  steps, and the registry gives the card's `CardDef` a `ResumeResolve`
  that runs them from a path. The engine opens the spell's step site
  while it runs the spell's catalog hook (`runSpellResolverLocked`: the
  item's ID and the key it resolved under), and
  `Game.ParkSpellStepsForEffect` parks only inside that site and only
  when the key can find the steps again. The parked walk records the
  key (`stepWalk.key`). A spliced text or a fused half finds no site,
  or a key with no `ResumeResolve`, and runs back to back as before.
- **A bullet as steps.** `effects.ModeSteps(label, targets, steps...)`
  is a `ModeOption` whose body answers the same probe. The mode walk
  opens the bullet's site around that occurrence (`runBulletLocked`).
  When a step parks (`Game.ParkBulletStepsForEffect`), the walk parks
  itself with that bullet first and the position inside it
  (`modeWalk.inner`), and resumes it through `game.ResumeBulletSteps`
  before the next bullet. A step reads its occurrence with
  `ctx.Occurrence()`.
- **A hand-written step.** `effects.Instruction(fn)` makes one function
  a step. It runs whole, and the walk may stop before or after it. A
  converted body is split at its asking calls, one `Instruction` per
  piece. The function belongs to the catalog row, so a resume finds it
  again. It must not pass anything to the next step through a Go
  variable.
- **`InterveningIf` keeps a `Do`'s steps.** A `Do` it wraps becomes the
  `If`'s `Then` instead of one opaque step, as `IfYouHaveTheInitiative`
  already did (it now calls `InterveningIf`). A hand-written Effect is
  still one step.
- **An `If` inside any site** is walked one level down, as before.

**What it costs.** The step sites are plain data set around one call
(`Game.stepSites`, dropped by the snapshot because it is zero between
actions). A parked position is still an item, a key and a path. No
closure is stored, no census counter rises, and `closure_fields.txt`
does not change. Two additive fields in schema v9
(`pausedStepWalk.key`, `pausedModeWalk.inner`). A binary from before
this change drops them. It would resume a spell's parked steps as the
item's Effect (a spell has none, so nothing runs) and a parked bullet
from its start.

**Cards converted.** Electrolyze and Tune the Narrative
(`ResolveSteps`), Demonic Pact's first bullet (`ModeSteps`), and Imoen,
Mystic Trickster (its two printed instructions as two steps under
`InterveningIf`). Every other row behaves as it did before.

Tests: `effects/resolution_order_test.go`. Electrolyze and Demonic Pact
against Khârn's question, live and across a strict restore; Tune the
Narrative and Imoen against a dredge offer, live and on an undo (the
dredge offer is not a restore point), with the census unchanged; an
`If` inside `ResolveSteps`, and a bullet's later step before the next
bullet, across a strict restore.

## Amendment 2026-10-19 (b): a continuation's instructions wait, and a step reads what an earlier one recorded (#4162, #4163)

**The gap.** The two amendments above park a step sequence the engine
can find again from data: an item's Effect, a spell's `OnResolve`, a
bullet. Two shapes were left.

- **A continuation** (#4162). A primitive's `Then`, an `OnYes` or
  `OnAccept`, and the callback a `...ThenForEffect` method runs are Go
  closures riding their prompt. When one runs two instructions and the
  first asks something (a dredge offer on a draw, a CR 616 order on
  counters, tokens or life), the second ran before the answer. Arcane
  Denial drew the caster's card with the victim's dredge offer open.
  Scavenging Ooze gained its life while its counter waited on the
  Doubling Season / Hardened Scales order.
- **A value carried between steps** (#4163). A step must not hand the
  next anything through a Go variable, because a restore taken between
  them loses it. Some later steps need what the game keeps no record of:
  the number of cards discarded (Windfall, Dark Deal, Tolarian Winds),
  the Soldiers just made (Martial Coup), the opponents who really drew
  (Cut a Deal), the number a player chose (Arcane Denial), the
  sacrificed creature's power (the Disciples).

**What changed.**

- **Continuation steps.** `effects.RegisterThenSteps(key, steps...)`
  names a step sequence with an on-disk key. Its `Run` walks the steps
  as a `Do` does and parks between them through
  `Game.ParkThenStepsForEffect`, as the key, the bullet occurrence and a
  path (`stepWalk.then`, `stepWalk.occ`). `game.ResumeThenSteps` finds
  them again by the key. The closure that calls `Run` still rides its own
  prompt, unchanged and counted where it was. What is new is that the
  rest of the continuation, once a step of it has asked, is data. The key
  shares the effect-key namespace and ledger (`steps <key>` in
  `testdata/effect_keys.txt`). A restore point naming a key this binary
  lacks is refused (`checkEffectKeys`), because the walk would run
  nothing after the answer.
- **The parked walks are a chain** (`stepWalk.outer`). A continuation
  runs while another walk may be parked: the item's own `Do`, waiting
  behind the very prompt whose answer ran the continuation. And a step
  may run a continuation that parks before the step's own walk does. The
  continuation is the earlier instruction in both cases (CR 608.2c), so
  it resumes first. Each walk takes a mark (`Game.StepWalkMarkForEffect`)
  as each step begins. A walk that parks goes directly above that mark,
  beneath the walks the step parked and above those parked before it
  (`parkStepWalkLocked`). The walks above the mark are copied, not
  relinked, because a clone may share them. `holdForOpenResolutionLocked`
  resumes the chain top first (`resumeStepWalksLocked`) until a step
  pauses again or none is left. An item's Effect, a spell's steps and a
  continuation all park this way. Before, a second walk could not park
  and ran on.
- **The step record.** `Game.RecordStepForEffect(item, key, rec)` files
  a `StepRecord` (IDs and numbers) under the resolving item and a name the
  card chooses. `StepRecordForEffect` reads it back, and in the effects
  package these are `ctx.RecordStep` and `ctx.StepRecord`. A record lives
  for one resolution. It is dropped as the resolution closes and as a new
  one opens, so no other item reads it. A step that needs an event's
  outcome records the event log's last sequence number and reads the log
  after it: Martial Coup spares the tokens the caster created after its
  record, and Cut a Deal counts the opponents with a draw event after
  its record (a dredge is not a draw).

**What it costs.** No closure is stored, and no census counter rises.
`closure_fields.txt` gains one `keyed` line, `stepWalk.outer`, which is
not a ceiling. Schema v9 gains additive fields
(`pausedStepWalk.then`, `.occ`, `.outer` and `stepRecords`), and the
fixtures are not edited. A binary from before this change drops them. It
would resume one parked walk, the top one, as an item's Effect (a
continuation's path into the wrong steps, or nothing), and its later
steps would read a zero record. That is the rollback a restore point
written by a newer binary always risks, and `checkEffectKeys` cannot
catch it from the older side.

**The scan.** `steps_scan_test.go` no longer reads a switch's clauses as
a sequence (`clauseList`): they are exclusive arms. Abrade, Boros Charm,
Izzet Charm, Archdruid's Charm and Rip Apart leave `stepsSkipped` for
that reason, and the "a Then" and "carried" reasons are gone.

**Cards converted.** Continuations: Arcane Denial, Braids, Arisen
Nightmare, Forgotten Ancient, Twilight Prophet, Scavenging Ooze, Seize the
Spotlight, Disciple of Freyalise, Disciple of Bolas, Replicating Ring,
Bonehoard Dracosaur, Midnight Clock, Court of Ambition, Transpose and
Sylvan Library. The chained ones (Braids, Court of Ambition, Forgotten
Ancient, Seize the Spotlight, Sylvan Library) carry the queue of players,
creatures or fortunes still owed in the record and run their steps again
for the next one. Records: Martial Coup, Windfall, Dark Deal, Tolarian
Winds and Cut a Deal, now `ResolveSteps`. Currency Converter's flagged
pair was two arms of one `if`, now one instruction. Genesis Wave, Coiling
Oracle and Animist's Awakening stay as they are, because their reveal asks
nothing.

Tests: `effects/then_steps_test.go`. A probe whose `Do` asks a number
and whose continuation scries then draws that many: the continuation
parks above the `Do`, the draw reads the number from the record, and the
`Do`'s last step runs after the draw, live and across a strict restore
taken while the scry is open. Seize the Spotlight's fortunes on a dredge
offer, Scavenging Ooze on a CR 616 counter order, Martial Coup on a
CR 616 token order (Doubling Season and Ojer Taq) and Cut a Deal on a
dredge offer, both answers, live and on an undo.
`game/step_walk_chain_test.go`: the chain order, copy on park, and a
record dropped as its resolution closes.

## Amendment 2026-10-23: the step record holds words, and four bodies stop carrying Go state (#4228)

#4155 left four bodies that handed a Go variable from one prompt to the
next: Rhys, the Evermore, Cecily, Haunted Mage, The Ur-Sphinx and Flamewar,
Streetwise Operative.

- **`StepRecord.Strs`.** The record gains a list of words beside its IDs and
  numbers (`json:"strs,omitempty"`), carried by `Clone` and the snapshot.
  Schema v9 gains `stepRecords[].record.strs`, additive, and the fixtures are
  not edited. A binary from before it drops the words. The only reader is
  Rhys's keyed question (below), whose continuation key that binary also does
  not have, so it refuses the file (`checkEffectKeys`) instead of restoring a
  question whose answer would read no kinds. `ctx.RecordStepRecord` files a
  whole record.
- **Rhys** files the creature, the counter kinds it had (sorted) and how far
  the questions have got. Each question is a keyed option pick
  (`option-pick/rhys-the-evermore-remove`, whose Carry is the resolving item,
  because a keyed continuation runs with no item of its own). The answer
  removes the counters and asks the next kind from the record. A table
  waiting on any question is a restore point; before, each question's
  continuation was a closure holding the kinds.
- **The Ur-Sphinx** is a `Do` of one counting step, a mill step for each seat
  (`game.MaxPlayers`; a seat the table lacks does nothing) and an offer step
  for each seat. The first step files the event log's cursor and how many
  cards each player mills. An offer step reads the player's `EventMill`
  events after the cursor and offers the nonland cards still in their
  graveyard, as a keyed choose-cards pick
  (`choose-cards/the-ur-sphinx-cast`) whose answer grants the free cast. The
  next offer is the next step, which the walk runs after the answer.
- **Flamewar** files the log's cursor, exiles, and the next step marks the
  cards that moved from the library to exile after the cursor. It no longer
  holds the exiled IDs in a pointer across the two statements.
- **Cecily** is `Do(DrawCards, the life loss, the free-cast offer)`. The
  registered `DrawThen` `cecily-haunted-mage-after-draw` is gone: a
  `DrawThen` rides a replacement event in memory and is in no snapshot or
  ledger, and nothing else named it.

The closure counts do not move: no field is added to `Game` that reaches a
func, and the closures these bodies used are replaced by keys and steps.

Tests: `effects/resolution_steps_state_test.go`. Rhys across a strict
restore before each of three questions and on an undo. The Ur-Sphinx across
a strict restore before each of four offers, a milled land not offered. Cecily on a dredge offer, live and on an undo.
Flamewar's marking step on the live table, an undo and a restored table.

## Amendment 2026-10-30: "choose up to X", where X is the spell's X (#4617)

**Sprint:** S79 — Card coverage: every Standard set, newest first. Seam
found during #4455 (Marvel Super Heroes).

The Ruinous Wrecking Crew ({X}{B}{R}): "When The Ruinous Wrecking Crew
enters, choose up to X —" with four bullets. Every count before this was a
printed `Min..Max`, raised by a condition at most (the 2026-09-27 and
2026-09-28 (c) amendments). This one is a number, and the number is a
payment. CR 107.3m: "If an object's enters-the-battlefield triggered ability
or replacement effect refers to X, and the spell that became that object as
it resolved had a value of X chosen for any of its costs, the value of X for
that ability is the same as the value of X for that spell, although the value
of X for that permanent is 0."

### Decision: `ModeSpec.UpToCastX`, read by the one bounds reader

- **The flag.** `ModeSpec.UpToCastX`, built with `effects.ChooseUpToX(…)`:
  `Min` 0 and `Max` the printed bullets, so a reader that never asks for the
  bounds still sees a bounded count. `ModeCountQuery` gains `X`, and
  `modeBoundsLocked` answers `0..min(q.X, len(Options))` for the flag. The
  cap is CR 700.2d: no bullet may be chosen twice, so X = 6 is all four.
- **Where X comes from.** `queueModePickLocked` sets `q.X` from the source's
  `CastProvenance.X` (`Card.CastX`, #1312), which the entry finisher stamps
  before `EventETB`, so the dispatch's copy of the entering permanent already
  carries it. It is read once, as the trigger is put on the stack (CR
  603.3c), and the prompt carries the bound (`ModeMax`), so the answer gate,
  the bot's enumerator, a restore point and an undo all see one number. A
  permanent that was not cast (a reanimation, a blink) has no spell X and
  reads 0, which is CR 107.3m's "the spell that became that object".
- **X = 0 asks nothing.** "Choose up to 0" can choose no mode, and CR 700.2b
  / 603.3c say a modal triggered ability with no mode chosen is removed from
  the stack. `queueModePickLocked` returns false, as it does for too few
  choosable bullets, and no prompt is queued. The bound's 0 never reaches a
  reader that would take it for "unbounded".
- **The prompt says the X.** Its reason is the printed prompt with the X:
  "Choose up to X (X is 2)". `ModeMax` is also capped at the bullets on offer
  (a bullet whose target can't be chosen is not offered, CR 700.2a), so the
  picker's count says what is reachable.
- **Register** (`checkUpToCastX`) keeps it to an enters trigger: refused on a
  spell's and an activated ability's modes, on a trigger that watches
  anything but `EventETB` (CR 107.3m names only the enters trigger), on a
  granted trigger (CR 107.3j: a gained ability's X is its own, or 0), and
  beside `Repeatable`, a raised count, `AtRandom` or `NotChosen`.

### Decision: an empty answer removes the ability

`ResolveModePick` passed the chosen modes on to `buildOrPickTriggerLocked`,
which reads a nil list as "not asked yet". An empty answer to a "choose up
to" prompt became nil and the same question was asked again, for every
"choose up to one" trigger already in the catalog (Ertai Resurrected,
Hullbreaker Horror, Ravager Wurm). The empty answer now removes the ability
(CR 603.3c, CR 700.2b): it records no mode, is never built, and never reaches
the stack.

### Wire and client

No field is added. `mode_min` is omitted at 0 (as it always was), so the
client's picker now reads an absent `mode_min` as 0 instead of 1: before,
nobody could answer a "choose up to" prompt with no bullet. With a minimum of
0 the dock's button reads "Choose none" until a bullet is picked, the hint
says up to how many, and a single-bullet picker takes a bullet back on a
second click. `docs/protocol.md` documents the absent `mode_min` and the X in
`reason`.

### Bots

`legal` offers every selection of the prompt's bullets from none up to
`ModeMax`, the same `ModePickSelections` walk as any `mode_pick`, so a bot is
never offered more than X and is never asked at X = 0. The heuristic still
answers a mode_pick with the enumerator's first selection, which for an
"up to" prompt is none. Valuing bullets is a bot change of its own.

### Out of scope, stated

- **Other "choose up to X" counts.** Riku of Many Paths ("where X is the
  number of times you chose a mode for that spell") reads the triggering
  spell, and Bumi, King of Three Trials ("the number of Lesson cards in your
  graveyard") reads the board. Neither is in the catalog. Either would be a
  keyed count beside this flag, read in the same `modeBoundsLocked`.
- **A copy of the spell.** A token copy of a permanent spell enters with no
  `CastProvenance.X` (`resolvePermanentSpellCopyLocked` carries the optional
  costs and the gift, not X), so its "enters with X counters" and its "choose
  up to X" both read 0. CR 707.10 copies the value of X. That gap is every X
  creature's, not this count's.

Tests: `game/up_to_cast_x_test.go` (the bound at X = 0, 1, 2, 4 and 6; no
prompt and no item at X = 0; the reason; the cap at the bullets on offer; an
empty answer to "up to X" and "up to one" removes the ability),
`cards/effects/the_ruinous_wrecking_crew_test.go` (X = 0, 2 and 6 through the
engine, choosing none, a Crew that was not cast, a restore and an undo, and
the `Register` refusals) and `legal/up_to_cast_x_test.go` (the enumerator
offers up to X and every move dispatches; nothing owed at X = 0).

Proof card: The Ruinous Wrecking Crew (Full).
