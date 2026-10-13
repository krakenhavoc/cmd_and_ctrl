# ADR 0161 — One mana and life ward payment

Date: 2026-10-13
Status: Accepted
Sprint: S58 — Deck requests, October batch (follow-up)
Issues: #2945, #2910

## Context

Gisa, the Hellraiser has one ward cost: {2} and paying 2 life. Two ward
triggers or sequential independent payment offers can charge only part of
that cost. The existing pay-unless path already plans the complete mana
payment before tapping or spending, and can add an optional Phyrexian-symbol
life payment. It is the existing shared payment boundary, not a new prompt.

## Decision

1. `WardManaAndLife` declares the supported compound cost. Ward permits this
   pair while preserving all existing single-component shapes and rejecting
   unsupported mixtures. The card has one ward trigger and one payment offer.
2. `CounterUnlessPaidPrompt.LifeCost` forwards the fixed amount into the
   existing `PendingChoice.LifeCost`. The same scalar already travels through
   clone, snapshot and protocol projection; document its pay-unless meaning.
   The mandatory amount is server data, not a client-selected amount or a
   count of Phyrexian symbols. No new answer field is needed.
3. A successful payment requires the complete remaining mana cost and the
   sum of fixed life plus any claimed Phyrexian-symbol life to be payable.
   Validate life and the complete mana plan before mutating anything. Pay
   the combined life once through the existing life-as-cost path, then commit
   the planned mana taps/spend. Life cost settlement cannot pause. Decline,
   an unaffordable ordinary acceptance or a departed guarded object consumes
   no life, mana or taps; malformed optional claims keep the prompt open.
4. The legal enumerator uses the same complete life amount and prices it on
   the offered move. Humans and MCP seats receive the mandatory life amount
   beside the mana cost; the human prompt renders both and retains the
   existing optional Phyrexian-symbol picker separately. The computed wire flag
   `life_cost_unpayable` disables acceptance when a life lock or insufficient
   life prevents the fixed payment; it is not persisted state.
5. Gisa uses the corrected crime annotation from ADR 0157, the existing
   once-per-turn tally, tribal static layers, and one creation instruction
   for two tapped blue and black Zombie Rogue tokens. Skeleton/Zombie type
   changes include Gisa herself when appropriate; her text says no "other".

## Compatibility

This reuses the existing serialized pending-choice life scalar and its wire
field. The prompt structure is runtime data and pay-unless continuations
already prevent saving an unresolved prompt as a restorable point. Snapshot
schema v8 remains unchanged; no frozen fixtures are rewritten.

## Validation

Cover successful payment, decline, absent mana, insufficient and exactly
enough life, life locks, a canceled life-cost replacement, departed guarded
objects, combined fixed and Phyrexian life, clone/undo and prompt projection.
Confirm legal move affordability/pricing and the human/MCP representation.
Exercise Gisa's ward, crime limit, two tapped two-color tokens, and both
tribal buffs including a type-changed Gisa. Existing ward shapes retain their
tests. Close #2945 only after the implementation PR merges.

## Number reservation

Initially reserved 0158 after checking all 13 remote refs on 2026-10-13.
A signed-in-pages ADR merged with that number before this PR. Renumbered this
ADR and its inbound links to 0161 after fetching and checking all 12 remaining
remote refs; 0158–0160 were occupied, so 0161 was the first unused number.
