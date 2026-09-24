---
title: "X on an activated ability"
date: 2026-09-17
issues: [550]
pr: 768
legacy_order: 108
---
**X on an activated ability** (PR #550) - `AbilityCost` reads `{X}` out of its mana string rather than a flag a card could forget: `DemandsX`, `XSlots` (Treasure Vault's `{X}{X}` is two) and `FloorX`, with a `MinX` field for a printed floor ("X can't be 0"). X is announced and validated at activation (CR 602.2b), locked onto `StackItem.XValue`, and read back with `ctx.X()`; the enumerator offers one X per ability, the largest affordable ([ADR 0033](decisions/0033-ai-bot-seat.md)). Shipped on Treasure Vault, Helm of Obedience and Soothsaying. The other cards from the old row were not re-triaged: Blast Zone (#308) and Magus of the Candelabra (#398) no longer wait on the cost, and Gogo, Master of Mimicry (#396) still waits on "Ability-copy".
