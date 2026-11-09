# ADR 0221 — Block taxes: "creatures can't block unless their controller pays {N}" (CR 509.1d)

**Status:** Accepted · 2026-11-09 · S79 — Card coverage: every Standard set, newest first
**Issues:** [#5375](https://github.com/krakenhavoc/cmd_and_ctrl/issues/5375) (the seam), found by [#5368](https://github.com/krakenhavoc/cmd_and_ctrl/issues/5368) (Archangel of Tithes)
**Numbering:** checked on 2026-11-09 after `git fetch --all --prune`, with the loop in
[README.md](README.md#picking-a-number) over every remote branch (81 heads, listed in the PR).
The highest number anywhere was **0220**; this is **0221**.
**Related:** [ADR 0080](0080-attack-taxes.md) (the attack tax this mirrors),
[ADR 0045](0045-combat-restrictions.md) Decisions 15 and 38 (block requirements and the
per-defender completion point), [ADR 0033](0033-ai-bot-seat.md) (the enumerator's promise).

## Context

Archangel of Tithes prints both halves of a combat tax:

> As long as this creature is untapped, creatures can't attack you or planeswalkers you
> control unless their controller pays {1} for each of those creatures.
> As long as this creature is attacking, creatures can't block unless their controller
> pays {1} for each of those creatures.

The attack half is ADR 0080's `AttackTax`. The block half had nothing: CR 509.1d — "If any
of the chosen creatures require paying costs to block, the defending player determines the
total cost to block. … Once the total cost is determined, it becomes 'locked in.'" — and
CR 509.1f — "Once the player has enough mana in their mana pool, they pay all costs in any
order. Partial payments are not allowed." Shipping the card without it is stronger than
printed for every defender.

The declaration paths differ. An attack is staged and paid by one verb call. Blocks are
STAGED by `declare_blocker` / `declare_blockers` and stay revisable (`remove_from_combat`)
until the defender's declaration completes (`block_completion.go`, ADR 0045 Decision 38):
on `finish_blocks`, on a pending defender's pass, or as the step ends.

## Decisions

### 1. `BlockTax` is ADR 0080's shape with no structural "you"

`game/block_tax.go`: a struct of hooks (`Label`, `AppliesTo`, `ManaCost`, `ActiveWhen`) on
a permanent, collected through `CatalogBlockTaxes` / `BlockTaxesForCard` (the one accessor,
so an ability-removed permanent stops taxing and the engine, the enumerator and the view
price alike). `BlockTaxQuery` carries the source, the blocking creature and its controller.

There is no structural "you", because the printed clause has none: "creatures can't block"
taxes every blocking creature, whoever controls the source. The condition ("as long as this
creature is attacking") is an ordinary predicate on the live source, `game.WhileSourceAttacking`.

`ManaCost` is a per-blocker cost string, concatenated per blocking creature, for ADR 0080
§1's reasons. A creature blocking several attackers (#1706) is ONE creature and pays once.

### 2. Paid where the declaration completes, never where a block is staged

Staging charges nothing: the blocks are not determined yet. The price is determined and paid
at the completion point (CR 509.1d–f), after the CR 509.1c requirement checkpoint and before
anything is announced:

- **`finish_blocks`** (`FinishBlocksWith`) pays with the action's posture — `auto_tap`,
  `locked_sources`, `phyrexian_life`, ADR 0080 §4's trio, strict as the zero value. Unpaid,
  it is refused with `*BlockTaxUnpaidError` (`ErrBlockTaxUnpaid`, wire code
  `block_tax_unpaid`); nothing is spent and the declaration stays pending with its blocks
  staged. Which blocks to drop is the defender's choice (ADR 0080 §3).
- **A pending defender's pass** pays with the zero posture (the pool, strictly) and is
  refused the same way.
- **The step ending** with the defender still pending cannot be refused — a step cannot end
  on an unfinished turn-based action. The pool is tried strictly; if it is short the
  defender's staged blocks are taken back and they have declared none, the one declaration
  that owes nothing and the only one the engine can make without choosing which blocks they
  keep.

The payer is `payAbilityManaCostLocked` with the zero spend context, as for the attack tax,
so mana triggers fire for its taps.

### 3. Requirements are not owed through a tax (CR 509.1c)

"If a creature can't block unless a player pays a cost, that player is not required to pay
that cost, even if blocking with that creature would increase the number of requirements
being obeyed." The requirement search leaves out a creature a block tax would charge for
unless it already blocks. Without this a "must block if able" creature under an attacking
Archangel would make the checkpoint refuse `finish_blocks` until it blocked, and the tax
refuse it once it did.

### 4. The wire

- `finish_blocks` takes optional params `{ auto_tap?, locked_sources?, phyrexian_life? }`.
  A payload without them is the strict posture.
- `TurnView.block_taxes`: one row per pending defender whose blocks would be taxed —
  `{ seat, cost?, each?, labels }`: the staged blocks' price (what `finish_blocks` charges),
  the per-blocker price when uniform, and the clauses.
- Error code `block_tax_unpaid`, with the price in `reason` and the shortfall in `missing`.
- A block move's `MoveCost.mana` is what it adds to the price; `finish_blocks`' is the whole
  price, with `auto_tap` set when the pool alone is short.

### 5. The enumerator and the bot

`legal/combat.go` offers a block only while the seat could pay for every block it has staged
plus this one (pool, then the tapper, then life for Phyrexian symbols — the attack tax's
test). So a bot never stages blocks its `finish_blocks` cannot pay for, and the finish stays
AlwaysLegal: nobody holds priority while the declaration is open, so nothing changes what
the seat can pay between the last block and the finish. The heuristic subtracts the block's
`MoveCost.mana` from its value at `AttackTaxPenalty`, the attack tax's weight.

The client shows the price on block staging (the dock's question and the list sheet) from
`block_taxes` and sends `finish_blocks` with `auto_tap: true` when blocks are taxed.

### 6. Out of scope

- A non-mana block tax: none printed.
- Re-pricing after the declaration: once the total is locked in, later changes are ignored
  (CR 509.1d), which needs no code — the tax is only asked at completion.
- A block tax with a duration (a War Tax mirror): none printed; `ModAttackTax`'s
  `ScopedEffect` path is the model when one is.

## Consequences

- Archangel of Tithes ships `full`.
- No snapshot change: nothing new is stored. The price is computed when asked; the payment
  is a mana spend.
- Every table without a block tax is unchanged: the price is `""` and the payer is never
  called.
