---
title: "One mana and life ward payment"
date: 2026-10-13
issues: [2945]
---
**One mana and life ward payment** (#2945, [ADR 0161](decisions/0161-one-mana-and-life-ward-payment.md)) — `effects.WardManaAndLife` declares both components in one ward trigger. `CounterUnlessPaidPrompt.LifeCost` reuses the existing pending-choice life scalar. The shared payment validates the entire mana plan and mandatory plus optional Phyrexian life before charging either. A failed or declined payment and a departed guarded spell consume neither component. Legal moves price the total life; human prompts and bot/MCP views show it. **Cards:** Gisa, the Hellraiser and Captain Howler, Sea Scourge (Full).
