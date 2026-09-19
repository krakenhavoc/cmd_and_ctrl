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
