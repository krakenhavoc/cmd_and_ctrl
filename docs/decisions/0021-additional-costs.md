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
