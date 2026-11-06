# ADR 0021 — Additional costs to cast

**Status:** accepted (S21 sub-PR 5)
**Extends:** [ADR 0020](0020-activated-abilities.md) (the cost model
for activated abilities), [ADR 0019](0019-structured-targeting.md)
(the announce-time validation pattern).

## Context

Three cards in the Pirates decklist — Thrill of Possibility, Big
Score, Unexpected Windfall — read:

> As an additional cost to cast this spell, discard a card.
> Draw two cards.

The tempting shortcut is to ignore the first line and write
`OnResolve` as "draw two, then discard one". It produces the same
hand size, and it is wrong in three ways that this deck notices:

1. **Ordering.** The cost is paid while the spell is on the stack,
   so Mary Read and Anne Bonny's Treasure and Marauding Mako's
   counter arrive *before* the spell resolves. Under the shortcut
   they arrive after, and the Treasure isn't available to hold up a
   response.
2. **Counterability.** A countered Thrill still costs you the card.
   Under the shortcut, countering it refunds the discard.
3. **Choice.** You choose what to discard knowing only your current
   hand — not the two cards you're about to draw. The shortcut lets
   you loot, which is strictly better and a different card.

The engine already had two kinds of cost: a spell's mana cost (S15)
and an activated ability's `AbilityCost` (S21 sub-PR 2). Neither
covers this, because both are attached to something other than a
cast.

## Decisions

### 1. A struct on the Spec, not a parsed cost string

`Spec.AdditionalCost *game.AdditionalCost`, wired through
`CatalogAdditionalCost` exactly as `Spec.Modes` and `Spec.Targets`
are. Today it carries one component:

```go
type AdditionalCost struct {
    DiscardCards int
    Label        string
}
```

The same reasoning as `AbilityCost`: the shapes are few, and a
second cost mini-language would have to be maintained against a
handful of cards. `DiscardCost(n)` is the constructor; it fills the
label ("Discard a card") so the client's prompt reads like the card.

Deliberately **not** added yet: sacrifice-as-an-additional-cost,
exile-from-graveyard, "reveal a card". No card in the current
decklists needs them, and a constructor with no caller is a guess
about an API rather than an API. `AdditionalCost` is a struct
precisely so they can arrive as fields without churning the shape.

### 2. Validate at announce, pay once the spell is on the stack

`CastSpellParams.DiscardIDs` carries the caster's picks.
`CastSpell` validates them with the rest of the announce-time
choices — before any cost is paid, and before the spell moves — so
a rejected cast is total: no card leaves the hand, and the spell
stays where it was. This is the same validate-all-then-pay
discipline `ActivateCatalogAbility` uses, and for the same reason:
a half-paid cost that then fails would strand the board in a state
no rule describes.

Payment then happens *after* `MoveCard(src, g.Stack, …)` and the
`StackMeta` entry, before `EventCast`. That ordering is the whole
point of the ADR: CR 601.2a puts the spell on the stack first, and
601.2h pays the costs last, so the discard trigger goes on the
stack above the spell and resolves first.

Two consequences fall out for free:

- The spell being cast is never a legal discard — it isn't in hand
  any more. The validator rejects it explicitly rather than relying
  on that, since validation runs before the move.
- A card with no additional cost that arrives *with* `discard_ids`
  is rejected rather than silently ignored. A client that sends
  them is confused about something, and a silent success hides it.

### 3. The cost picker opens first in the client, and that's a UI choice

`DiscardCostModal` opens before the X prompt, the mode picker and
targeting. The rules order is the other way round — targets at
CR 601.2c, costs at 601.2h — but the whole cast rides one
`cast_spell` message, so collection order is not observable in the
game state. It's ordered this way because the discard is the choice
most likely to make a player back out, and because it matches the
sacrifice-cost picker that S21 sub-PR 2 already put in front of
targeting.

The picked IDs thread through `TargetingState.discardIDs` the way
`xValue` and `modes` already do, so a future card that has both an
additional cost and a target needs no new plumbing.

### 4. A cost is not a target

The picker is a plain list of the caster's hand, not the
board-click targeting flow. Nothing can make a card in your own
hand an illegal choice, a cost can't be responded to, and hexproof
is irrelevant. Same posture as `SacrificeCostModal`.

## Consequences

- Thrill of Possibility, Big Score and Unexpected Windfall ship,
  and the decklist's discard payoffs see them.
- `Blood`'s "discard a card" activation cost — a declared gap in
  S21 sub-PR 4 — now has a model to copy. It isn't done here:
  `AbilityCost` would need its own `DiscardCards` field and the
  ability path its own picker, which is a separate change.
- Kicker, escalate, and other optional additional costs are still
  out. They need a *choice* of whether to pay, which changes the
  spell's effect, not just its price — that's a mode-plus-cost
  shape and wants its own design.

## Alternatives considered

**Fold the discard into `OnResolve`.** Rejected for the three
reasons in Context. The archetype is built on the ordering.

**A `PendingChoice` for the discard, like the cleanup-step
discard.** That queues a prompt and resolves it asynchronously,
which would leave the spell announced but not paid for while the
prompt sits open — a state the rules don't have, and one that
another player could act in. Announce-time collection keeps the
cast atomic.

## Amendment — sacrifice as an additional cost (S21 sub-PR 6)

`AdditionalCost` gains a second component, `Sacrifice *TargetSpec`,
for Village Rites, Altar's Reap and Deadly Dispute. The decision
above holds unchanged; this records what the second component
settled.

**It is the same decision, one zone over.** The three reasons the
discard is a cost rather than an `OnResolve` effect apply verbatim to
the sacrifice: the creature dies with the spell on the stack, so
Blood Artist and Zulaport Cutthroat drain before the cards are drawn;
a countered Village Rites still costs the creature; and with nothing
to sacrifice the spell cannot be cast at all. The aristocrats
archetype is built on the first of those the way the Pirates list is
built on the discard ordering, so it is asserted directly — the test
checks that Blood Artist's target prompt exists *while Village Rites
is still on the stack*, which an `OnResolve` implementation cannot
produce.

**Validation is shared, not duplicated.** The clause reuses
`validateSacrificeCostLocked`, the activated-ability validator, so
"you may only sacrifice what you control" (CR 701.21a) and the
predicate check live in one place across all three cost sites
(spell, activated ability, mana ability). `SacrificeCost` likewise
builds its spec with the same `sacrificeSpec` helper the abilities
use.

**A sacrifice needs no self-exclusion.** The discard component has to
exclude the spell being cast; the sacrifice component gets that for
free, because the spell is on the stack and the clause only matches
permanents on the battlefield.

**Still out, and still for the reason above:** kicker and other
*optional* additional costs. A choice of whether to pay changes the
spell's effect rather than its price.

### Known debt

On the client, `xValue`, `discardIDs` and `sacrificeIDs` are now
three parallel announce-time payments threaded side by side through
`begin` / `beginForMode` / `continueCast`. A fourth should bundle
them into one `CastChoices` object rather than adding a parameter;
noted in `targeting.ts` at the declaration.

## Addendum (2026-09-17): sacrificing N permanents as an additional cost (#747)

**Status:** Accepted · 2026-09-17 · tracked on
[#747](https://github.com/krakenhavoc/cmd_and_ctrl/issues/747). This
status covers this section only. The design, and the owner's and the
lead's decisions on its open questions, are in
[ADR 0020's #747 addendum](0020-activated-abilities.md#addendum-2026-09-17-sacrifice-costs-of-n-permanents-747)
(§12–§17), because the validator, the enumerator helper and the client
picker are shared by all three sacrifice cost sites. This section records
what that design means for a spell.

- **The clause carries the count.** "As an additional cost to cast this
  spell, sacrifice two creatures" is
  `AdditionalCost{Sacrifice: sacrificeSpec(label, preds...).WithCount(2, 2)}`,
  built by `SacrificeNCost(2, "creatures", Creature())`.
  `SacrificeCost(label, preds...)` still builds the count-1 clause, and
  Village Rites, Altar's Reap and Deadly Dispute do not change.
- **§2 above still holds.** `validateAdditionalCostLocked` validates
  `sacrifice_ids` at announce, through the shared
  `validateSacrificeCostLocked`, which now requires exactly N distinct
  permanents. `payAdditionalCostLocked` pays them only after the spell is
  on the stack. A wrong count refuses the cast, and the card stays in its
  zone. The spell still needs no check against being named, because it is
  on the stack and the clause matches only permanents on the battlefield.
- **The N permanents are sacrificed as one simultaneous exit** (ADR 0020
  §14), inside `payAdditionalCostLocked`. A Blood Artist sacrificed with
  another creature to pay for a spell drains for both, whatever order the
  IDs arrive in. The drains still go on the stack above the spell, as the
  amendment above requires.
- **The enumerator's cast site** (`legal/cast.go:202`) takes N from the
  clause, offers no cast when the caster controls fewer than N matching
  permanents, and offers one payment per cast for N ≥ 2 (ADR 0020 §15).
- **The client** keeps threading `sacrificeIDs` through `targeting.ts`,
  which is already a `string[]`. `Board.svelte`'s cast-time
  `SacrificeCostModal` gets the same multi-select as the ability one, with
  the count from `additional_cost.sacrifice_options.max`, and the same
  "Choose for me" button (owner decision, ADR 0020 §16). The button fills
  the first N entries of `additional_cost.sacrifice_options.cards`, which
  the server sends in ADR 0020 §15's order, and never confirms. No fourth
  parallel payment is added, so the known debt above does not grow.
- **Still out:** variable counts ("sacrifice any number of creatures",
  "sacrifice X"), and kicker-style *optional* sacrifices. Both change
  what the caster announces, not only how many permanents they pick.

## Amendment (2026-09-18): optional additional costs are in, and live in ADR 0073 (#664)

**Status:** Superseded in part · 2026-09-18 · this section covers the
deferral only.

Three paragraphs of this ADR say kicker and the rest of the *optional*
additional costs are out of scope — the Consequences bullet above
("Kicker, escalate, and other optional additional costs are still out…"),
the S21 sub-PR 6 amendment's "Still out, and still for the reason above",
and the #747 addendum's "Still out" bullet. All three are now answered by
[ADR 0073 — Optional additional costs and the announce-time cast
gate](0073-optional-additional-costs-and-the-cast-gate.md).

What changed, and what did not:

- **The reasoning in those paragraphs stands.** A choice of whether to pay
  does change what the spell does rather than only what it costs, which is
  why ADR 0073 records the choice on the stack item
  (`PaidCost.OptionalCosts`) and carries it onto an entering permanent
  (`Card.PaidOptionalCosts`) instead of treating it as pricing.
- **It is not a new kind of cost.** ADR 0073 §1 puts an `Optional` flag on
  *this* `AdditionalCost` struct, so Constant Mists' "Buyback—Sacrifice a
  land" is the `Sacrifice` component this ADR added in S21 sub-PR 6 with one
  bool set. The validator and the payer this ADR describes are still the only
  ones; they take an ordered payment plan (mandatory first, then the chosen
  optional costs) rather than a single cost.
- **`ManaCost` is new on the component**, because every additional cost this
  ADR shipped is non-mana and kicker is mana. It is added to the total at
  CR 601.2f, where this ADR always said additional costs go.
- **The known-debt note above is discharged.** The client's announce-time
  payments were bundled into one `CastChoices` object in #874; ADR 0073's
  optional-cost choice rides that bundle rather than adding a fifth
  parameter.
- **Still out, and now for a narrower reason:** escalate and entwine. Their
  cost is per extra *mode*, which is a mode-and-cost product that ADR 0073's
  index-list announcement cannot express (ADR 0073, Consequences).

## Amendment (2026-09-30): either/or additional costs (ADR 0100 sub-PR 3, #1732)

"As an additional cost to cast this spell, sacrifice an artifact or discard a
card" (Demand Answers) is a mandatory cost that offers a choice, which the
struct of §1 could not say: it demanded every component it named. [ADR
0100](0100-delve-either-or-and-variable-sacrifice-costs.md) §2 adds it without
a second kind of cost:

- **`AdditionalCost.Either`.** A cost with branches has no components of its
  own; each branch is an ordinary `AdditionalCost` with a `Key`. It stays in
  the mandatory slot, because the cost is mandatory and CR 601.2b asks only
  *which* branch.
- **Two components join the struct:** a fixed `PayLife` (CR 119.4, Bitter
  Triumph's "pay 3 life"), and a mandatory `ManaCost` inside a branch
  (Lightning Axe's "pay {5}"), priced at CR 601.2f next to the optional costs'
  mana by `game.AdditionalCostMana`. Blight, until now optional only, may be a
  branch.
- **§2 stands.** The branch is announced (`CastSpellParams.CostBranch`,
  required on a branched card and refused on any other) and the chosen branch
  becomes the plan's mandatory entry, so the validator and the payer this ADR
  describes are still the only ones, and the flat `discard_ids` /
  `sacrifice_ids` lists are walked against the branch that was announced.
- **The discard is now recorded.** `PaidCost.Discarded` keeps the cards the
  additional cost discarded, for Grab the Prize's "if the discarded card wasn't
  a land card" (ADR 0100 owner decision 6), and `PaidCost.CostBranch` keeps the
  branch.

## Amendment (2026-09-30): a variable sacrifice count (ADR 0100 sub-PR 4, #1732)

§1's sacrifice component paid a printed count. [ADR
0100](0100-delve-either-or-and-variable-sacrifice-costs.md) §3 lets a cast's
mandatory additional cost print two variable counts, with no new component:

- **"Sacrifice any number of …"** (Vicious Betrayal) and its "you may"
  spellings (Torgaar, Famine Incarnate; Plumb the Forbidden's "you may
  sacrifice one or more") are one clause whose count runs from zero:
  `effects.SacrificeAnyNumberCost`, a `Sacrifice` spec with `Min` 0 and `Max`
  0 (`game.SacrificeAnyNumber`). Sacrificing none is not paying, so "you may"
  needs no optional slot.
- **"Sacrifice X …"** (Devastating Summons, Eliminate the Competition) is
  `effects.SacrificeXCost`, the `CountFromX` clause; X is the announced
  `CastSpellParams.XValue` (CR 107.3a), the same X the spell's text reads
  (CR 107.3i).
- **§2 stands.** There is still one validator and one payer. A variable
  clause takes whatever is left of the flat `sacrifice_ids` list, which is
  unambiguous because `effects.Register` holds a cast's plan to at most one
  variable clause and, with one, no other sacrifice (`checkVariableSacrificePlan`).
  The count lands in `PaidCost.Sacrificed`, and a per-sacrifice discount reads
  it at CR 601.2f through `CostQuery.Sacrificing` (`effects.CostsLessPerSacrificed`).

## Amendment (2026-10-15): a value fixed as the spell is cast (#3781)

"Draw X cards, where X is the greatest power among creatures you controlled as
you cast this spell" (Lifestream's Blessing, #3656) reads the board at a moment
no hook could reach: the cast path built the stack item and kept the costs it
paid (the `Paid` record, ADR 0100) but nothing that was merely *true* of the
game as the cast finished. The spell fell back to the board at resolution, so
a creature that died or grew in response changed the answer (CR 601.2i fixes it
at the cast).

- **Declaration.** `Spec.CastTimeValues` / `CardDef.CastTimeValues` is a keyed
  list of `game.CastTimeValue{Key, Value func(*Game, *StackItem) int}`. Keyed
  so one card can record two facts; `Register` refuses a blank key, a nil
  `Value` or a duplicate (`checkCastTimeValues`).
- **The stamp.** `CastSpell` calls `stampCastTimeValuesLocked` as the last act
  before `EventCast`: every cost is paid (so a creature sacrificed to pay is
  not counted, which is what "as you cast" says), the spell is on the stack,
  nobody has had priority. `Value` runs under the game lock and reads through
  the same accessors any catalog hook does.
- **The memory.** `StackItem.CastValues map[string]int`, nil for every spell
  that declares nothing. Carried by `Clone` (deep copy), by the snapshot
  (`castValues`, additive within v9: absent reads as nil, and a binary that
  drops the key falls back as below), and by a spell copy: CR 707.10 copies the
  values the original fixed, and a copy is never cast, so there is no second
  moment to recompute at.
- **The reader.** `effects.Context.CastValue(key) (int, bool)`. `ok` is false
  only for a game restored from before the value existed; the card then falls
  back to the board at resolution, the pre-amendment behaviour.
- **Not the same as `AdditionalCost` or X.** X is announced and paid for
  (CR 107.3); a cast-time value is *computed by the engine* from the game, not
  chosen by the caster. A card whose text reads the board at resolution
  ("greatest power among creatures you control", Traverse the Outlands) does
  not belong here.

Card: Lifestream's Blessing is Full. Other "as you cast this spell" statements
(a count of lands you controlled as you cast it, and so on) now declare one
`CastTimeValue` and read it back; none was found carrying a caveat for the gap.

## Amendment (2026-11-06): a cost sized by the chosen targets (#5240)

Urgent Necropsy's "As an additional cost to cast this spell, collect evidence
X, where X is the total mana value of the permanents this spell targets" is a
cost whose amount is not known until the targets are. The rules allow it: the
targets are announced at CR 601.2c and the total cost is determined at CR
601.2f and paid at 601.2h. The ruling (2024-02-02) says X is locked in once the
targets are chosen and before costs are paid. The engine already validates the
whole cast in one `cast_spell`, targets before costs, so the decision is only
where the floor is read and in which order the client and the bot ask.

- **No new cast step.** The amount is re-derived from the announced targets at
  announce, not stored. `game.ExileCost.ManaValueFromTargets` (built by
  `game.CollectEvidenceFromTargetsCost`, `effects.CollectEvidenceFromTargetsCost`)
  is collect evidence's any-number graveyard clause with no printed floor;
  `CastSpell` fixes the floor with `Game.CollectEvidenceTargetsFloorForEffect`
  right after the CR 601.2c target check and checks `exile_ids` against it with
  the existing validator. X is the sum of the mana values of the distinct
  permanents among the targets (a permanent chosen for two "target" words is one
  permanent the spell targets, CR 115.3), read through `ManaValueForEffect`. A
  parked announcement already carries `Targets` and `ExileIDs`, so a cast parked
  mid-announcement re-derives the same X when it is made again; nothing is added
  to the snapshot.
- **Every such cast collects evidence**, collect evidence 0 included, so
  `EventCollectEvidence` is emitted for it as for an announced optional collect
  evidence.
- **Castable without a target search.** `effects.Register` allows the clause
  only alone on a spell's mandatory slot, on a spell whose every target clause is
  "up to": X = 0 is then always a legal announcement, so the cast gate needs no
  search over target sets. It refuses it on an optional cost, an either/or
  branch, an ability, beside another component, on a spell with no target
  clause and on one with a clause that must be filled.
- **§3 stands, with one exception.** The cost pickers still open before
  targeting, because the order is not observable in the game state. A
  target-sized cost cannot be asked before its targets, so for this cost alone
  the client skips the graveyard picker in the pre-targeting chain and opens it
  when the targets are confirmed, sized from the view's
  `exile_cost_target_mana_values` (each permanent's mana value as the engine
  reads it, so the client's X is the engine's). X = 0 asks nothing. The one
  `cast_spell` carries both.
- **The bot chooses targets first, then the payment.** The enumerator takes the
  graveyard in fuel order as the pool, drops a target set whose X the graveyard
  cannot reach (CR 118.3), and names the cheapest-first cards that reach it for
  each set it offers.

Card: Urgent Necropsy, Full.
